package multi

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/llm"
)

// reviewKeywords triggers Review strategy when found in user input.
var reviewKeywords = []string{
	// 한국어
	"리뷰", "검토", "리팩토링", "리팩터", "분석", "코드 리뷰",
	"평가", "점검", "진단", "살펴", "확인해", "봐줘", "체크",
	"문제 없", "괜찮", "개선점", "수정해", "고쳐", "최적화",
	"개선", "파악", "설명해", "알려줘", "구조", "아키텍처",
	"테스트", "버그", "에러", "오류", "원인", "왜 안",
	"성능", "느려", "빠르게", "줄여", "정리",
	// 영어
	"refactor", "review", "audit", "analyze", "analysis",
	"optimize", "improve", "fix", "check", "explain",
	"architecture", "structure", "test", "bug", "error",
	"performance", "clean up", "simplify",
}

// consensusKeywords triggers Consensus strategy.
var consensusKeywords = []string{
	// 한국어
	"비교", "보안", "디버그", "의견", "판단", "결정",
	"어떤 게 나을", "뭐가 나을", "장단점", "차이",
	"어떻게 생각", "어느 쪽", "둘 중", "선택",
	"트레이드오프", "고민",
	// 영어
	"compare", "security", "debug", "pros cons",
	"difference", "trade-off", "which is better",
}

// scanKeywords triggers Scan strategy.
var scanKeywords = []string{
	// 한국어
	"전체", "모든 파일", "프로젝트 전체", "전수",
	"모든 코드", "전부 찾", "다 찾아", "전체 검색",
	"어디에 있", "어디서 쓰", "사용처", "의존성",
	"임포트", "참조",
	// 영어
	"all files", "codebase", "find all", "search all",
	"where is", "usage", "dependencies", "references",
}

// simpleKeywords: 이 키워드가 있으면 멀티 불필요 (단순 대화/질문)
var simpleKeywords = []string{
	"안녕", "뭐야", "뭘 알", "어떤 기능", "도움말",
	"hi", "hello", "thanks", "고마워", "감사",
	"help", "ㅎㅇ", "ㄱㅅ", "ㅂㅇ",
	"읽어줘", "열어줘", "보여줘",
}

// AutoDetect returns true when multi-agent mode should activate.
// Hybrid approach: keyword matching first, then LLM fallback.
func AutoDetect(input string, history []openai.ChatCompletionMessage, tokenCount int, contextWindow int) bool {
	// 0. Skip simple greetings/questions
	lower := strings.ToLower(input)
	for _, kw := range simpleKeywords {
		if strings.Contains(lower, kw) {
			return false
		}
	}

	// 1. Context usage >= 60%
	if contextWindow > 0 && tokenCount > 0 {
		pct := (tokenCount * 100) / contextWindow
		if pct >= 60 {
			return true
		}
	}

	// 2. Keyword match (instant)
	for _, kw := range reviewKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	for _, kw := range consensusKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	for _, kw := range scanKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	// 3. Input length > 500 characters
	if utf8.RuneCountInString(input) > 500 {
		return true
	}

	// 4. Conversation > 30 turns
	userTurns := 0
	for _, msg := range history {
		if msg.Role == openai.ChatMessageRoleUser {
			userTurns++
		}
	}
	if userTurns > 30 {
		return true
	}

	return false
}

// AutoDetectWithLLM is the hybrid fallback: when keyword matching fails,
// ask the LLM (dev model, fast/cheap) to classify the intent.
// Returns true if the LLM thinks multi-agent would help.
func AutoDetectWithLLM(client *llm.Client, devModel string, input string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	classifyPrompt := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `You are a classifier. Given a user input, decide if it needs multi-agent processing (two AI models working together).
Reply ONLY "YES" or "NO".

YES cases: code review, refactoring, security audit, comparing approaches, analyzing codebase, debugging complex issues, large file scanning
NO cases: simple questions, greetings, single file reads, short answers, general chat`,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		},
	}

	resp, err := client.Chat(ctx, devModel, classifyPrompt)
	if err != nil {
		config.DebugLog("[MULTI-LLM] classify error: %v", err)
		return false // fail-safe: don't activate on error
	}

	result := strings.TrimSpace(strings.ToUpper(resp))
	isMulti := strings.HasPrefix(result, "YES")
	config.DebugLog("[MULTI-LLM] classify input=%q result=%q multi=%v", truncateInput(input, 80), result, isMulti)
	return isMulti
}

// AutoDetectStrategy picks the best strategy based on input keywords.
// Falls back to StrategyReview if no specific pattern matches.
func AutoDetectStrategy(input string) Strategy {
	lower := strings.ToLower(input)

	// Check scan keywords first (most specific)
	for _, kw := range scanKeywords {
		if strings.Contains(lower, kw) {
			return StrategyScan
		}
	}

	// Then consensus
	for _, kw := range consensusKeywords {
		if strings.Contains(lower, kw) {
			return StrategyConsensus
		}
	}

	// Default to review
	return StrategyReview
}

// IsLargeProject checks if the working directory has many files,
// which benefits from multi-agent parallel scanning.
// Returns true if there are 500+ files (quick count, stops early).
func IsLargeProject(dir string) bool {
	count := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() && (strings.HasPrefix(name, ".") || name == "node_modules" || name == "dist" || name == "__pycache__" || name == "vendor" || name == ".git") {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			count++
		}
		if count >= 500 {
			return filepath.SkipDir // stop early
		}
		return nil
	})
	return count >= 500
}

// SplitSubdirs returns top-level subdirectories split into two groups
// for parallel scanning by two agents.
func SplitSubdirs(dir string) (group1, group2 []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}

	var dirs []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() {
			continue
		}
		// Skip noise
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "dist" || name == "__pycache__" || name == "vendor" {
			continue
		}
		dirs = append(dirs, name)
	}

	mid := len(dirs) / 2
	if mid == 0 {
		mid = 1
	}
	return dirs[:mid], dirs[mid:]
}

func truncateInput(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
