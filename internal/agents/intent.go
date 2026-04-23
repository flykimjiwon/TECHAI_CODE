package agents

import (
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/llm"
)

// Intent represents the classified user intent for a message in Super mode.
type Intent int

const (
	IntentUnclear Intent = iota
	IntentChat
	IntentPlan
	IntentAuto
)

// String returns a short human label for the intent.
func (i Intent) String() string {
	switch i {
	case IntentChat:
		return "chat"
	case IntentPlan:
		return "plan"
	case IntentAuto:
		return "auto"
	default:
		return "unclear"
	}
}

// Keyword dictionaries (Korean + English). Kept intentionally small and
// opinionated — intent detection is a hint, not a hard classifier.
var (
	chatKeywords = []string{
		"뭐야", "뭐지", "알려줘", "설명", "어떻게", "왜", "언제", "어디",
		"무엇", "누구", "차이", "의미",
		"what", "why", "how", "when", "where", "explain", "tell me",
		"describe", "difference", "meaning",
	}

	planKeywords = []string{
		"계획", "플랜", "단계", "순서", "먼저", "설계", "아키텍처", "구조",
		"plan", "outline", "architecture", "design", "roadmap", "blueprint",
		"step by step", "step-by-step",
		"만들자", "만들어", "만들어줘", "만들어볼", "구현", "구축", "설정",
		"세팅", "셋업", "초기화", "시작", "생성",
		"프로젝트", "project", "앱", "app ", "서비스", "시스템",
		"create", "build ", "setup", "set up", "initialize", "scaffold",
		"implement", "develop",
	}

	autoKeywords = []string{
		"전부", "모두", "다 ", "전체", "자동", "싹 다", "끝까지", "완전히",
		"마이그레이션", "리팩터", "리팩토링",
		"all ", "every", "whole", "entire", "autonomous", "end-to-end",
		"fix all", "refactor", "migrate", "cleanup", "rewrite",
		"모든", "싹", "전면적",
		"고쳐줘", "수정해", "정리",
		"everything", "all files", "across", "throughout",
	}
)

// DetectIntentLocal runs a fast, deterministic keyword-based classification
// and returns IntentUnclear when no strong signal is found.
func DetectIntentLocal(message string) Intent {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return IntentUnclear
	}
	lower := strings.ToLower(trimmed)

	// Short messages ending with '?' are almost always chat.
	if len([]rune(trimmed)) < 30 && strings.Contains(trimmed, "?") {
		return IntentChat
	}

	chatScore := countMatches(lower, chatKeywords)
	planScore := countMatches(lower, planKeywords)
	autoScore := countMatches(lower, autoKeywords)

	// Very long messages (>200 chars) without explicit chat markers lean auto.
	if len([]rune(trimmed)) > 200 && chatScore == 0 && planScore == 0 && autoScore == 0 {
		return IntentAuto
	}

	maxScore := chatScore
	winner := IntentChat
	if planScore > maxScore {
		maxScore = planScore
		winner = IntentPlan
	}
	if autoScore > maxScore {
		maxScore = autoScore
		winner = IntentAuto
	}
	if maxScore == 0 {
		return IntentUnclear
	}
	return winner
}

func countMatches(text string, keywords []string) int {
	count := 0
	for _, k := range keywords {
		if strings.Contains(text, k) {
			count++
		}
	}
	return count
}

// DetectIntentLLM performs a one-shot LLM classification for ambiguous cases.
// Uses the internal LLM endpoint (works in closed-network environments).
// Any error returns IntentUnclear so callers can gracefully fall back.
func DetectIntentLLM(ctx context.Context, client *llm.Client, model, message string) Intent {
	if client == nil {
		return IntentUnclear
	}
	prompt := `You are an intent classifier for a coding agent. Classify the user message into ONE category:

- chat: simple question, conversation, or request for explanation. Short answer expected.
- auto: multi-step task with clear scope that benefits from autonomous execution (bug fixes, refactors, well-defined implementations).
- plan: large, ambiguous task that benefits from explicit planning before execution (architecture, multi-file features, complex migrations).

Respond with ONLY one word: chat, auto, or plan.

Message: "` + message + `"`

	resp, err := client.Chat(ctx, model, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	})
	if err != nil {
		return IntentUnclear
	}

	resp = strings.ToLower(strings.TrimSpace(resp))
	switch {
	case strings.Contains(resp, "chat"):
		return IntentChat
	case strings.Contains(resp, "plan"):
		return IntentPlan
	case strings.Contains(resp, "auto"):
		return IntentAuto
	}
	return IntentUnclear
}

// Suggest returns a short UI hint describing the recommended mode.
func (i Intent) Suggest() string {
	switch i {
	case IntentPlan:
		return "Plan 모드를 추천합니다 — Tab으로 전환"
	case IntentAuto:
		return "Deep Agent 모드를 추천합니다 — Tab으로 전환"
	}
	return ""
}

// SuggestedTab returns the tab index that should be activated when the user
// accepts the hint. Returns -1 if no switch is recommended.
func (i Intent) SuggestedTab() int {
	switch i {
	case IntentPlan:
		return 2 // ModePlan
	case IntentAuto:
		return 1 // ModeDev (Deep Agent)
	}
	return -1
}
