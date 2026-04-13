package multi

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/llm"
)

// synthesisTimeout is the max time allowed for the LLM synthesis call.
const synthesisTimeout = 30 * time.Second

// MergeWithSynthesis uses an LLM to combine both agent outputs into one
// cohesive response. Falls back to simple merge on error.
// The parent ctx is used so that user cancellation (Ctrl+C) stops the synthesis call.
func MergeWithSynthesis(ctx context.Context, client *llm.Client, model string, strategy Strategy, a1, a2 AgentResult) string {
	// If either agent errored, fall back to simple merge
	if a1.Err != nil || a2.Err != nil {
		return simpleMerge(strategy, a1, a2)
	}

	// If Agent2 has no content, just return Agent1
	a2Trimmed := strings.TrimSpace(a2.Content)
	if a2Trimmed == "" {
		return a1.Content
	}

	// Skip synthesis when Agent2 found no issues (wasteful LLM call)
	if strings.Contains(a2Trimmed, "특이사항 없음") {
		return a1.Content
	}

	ctx, cancel := context.WithTimeout(ctx, synthesisTimeout)
	defer cancel()

	var synthesisPrompt string

	switch strategy {
	case StrategyReview:
		synthesisPrompt = fmt.Sprintf(`당신은 두 AI 에이전트의 응답을 하나로 종합하는 편집자입니다.

Agent1(Super)이 원본 응답을 작성했고, Agent2(Dev)가 검토 의견을 제출했습니다.

## Agent1(Super) 응답:
%s

## Agent2(Dev) 검토:
%s

## 지시사항:
- Agent1의 원본 응답을 기반으로, Agent2의 검토에서 유효한 지적사항을 반영하여 하나의 완성된 응답을 작성하세요.
- "Agent1이...", "Agent2가..." 같은 메타 언급을 하지 마세요. 마치 처음부터 하나의 AI가 작성한 것처럼 자연스럽게 통합하세요.
- Agent2가 지적한 문제점이 있으면 수정된 내용을 반영하세요.
- Agent2가 추가한 유용한 정보는 포함하세요.
- 불필요한 반복은 제거하세요.
- 코드 블록이 있으면 Agent2의 개선 제안을 반영한 최종 버전만 포함하세요.
- 한국어로 작성하세요.`, a1.Content, a2.Content)

	case StrategyConsensus:
		synthesisPrompt = fmt.Sprintf(`당신은 두 AI 에이전트의 독립적인 응답을 비교 종합하는 분석가입니다.

두 에이전트가 같은 질문에 독립적으로 답변했습니다.

## Agent1(Super) 응답:
%s

## Agent2(Dev) 응답:
%s

## 지시사항:
- 두 응답의 공통점과 차이점을 분석하세요.
- 각각의 강점을 살려 하나의 최적 응답으로 종합하세요.
- 의견이 다른 부분은 양쪽 관점을 간결히 비교하세요.
- "Agent1이...", "Agent2가..." 대신 "한 관점에서는...", "다른 관점에서는..." 같은 표현을 사용하세요.
- 한국어로 작성하세요.`, a1.Content, a2.Content)

	case StrategyScan:
		synthesisPrompt = fmt.Sprintf(`당신은 두 AI 에이전트의 병렬 탐색 결과를 통합하는 편집자입니다.

두 에이전트가 프로젝트의 서로 다른 영역을 탐색한 결과입니다.

## Agent1(Super) 탐색 결과:
%s

## Agent2(Dev) 탐색 결과:
%s

## 지시사항:
- 두 탐색 결과를 하나의 종합 보고서로 통합하세요.
- 중복 내용은 제거하고, 각 에이전트가 발견한 고유 정보를 모두 포함하세요.
- 논리적 순서로 재구성하세요.
- "Agent1이...", "Agent2가..." 같은 메타 언급을 하지 마세요.
- 한국어로 작성하세요.`, a1.Content, a2.Content)

	default:
		config.DebugLog("[MULTI-SYNTHESIS] unknown strategy %s, falling back to simple merge", strategy)
		return simpleMerge(strategy, a1, a2)
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: synthesisPrompt,
		},
	}

	config.DebugLog("[MULTI-SYNTHESIS] starting model=%s strategy=%s", model, strategy)

	resp, err := client.Chat(ctx, model, messages)
	if err != nil {
		config.DebugLog("[MULTI-SYNTHESIS] error: %v, falling back to simple merge", err)
		return simpleMerge(strategy, a1, a2)
	}

	config.DebugLog("[MULTI-SYNTHESIS] done len=%d", len(resp))
	return resp
}

// simpleMerge is the fallback when LLM synthesis fails.
func simpleMerge(strategy Strategy, a1, a2 AgentResult) string {
	switch strategy {
	case StrategyReview:
		return simpleMergeReview(a1, a2)
	case StrategyConsensus:
		return simpleMergeConsensus(a1, a2)
	case StrategyScan:
		return simpleMergeScan(a1, a2)
	default:
		return simpleMergeReview(a1, a2)
	}
}

func simpleMergeReview(a1, a2 AgentResult) string {
	var b strings.Builder

	if a1.Err != nil {
		b.WriteString(fmt.Sprintf("## Agent1(Super) 오류\n%v\n\n", a1.Err))
	} else {
		b.WriteString(a1.Content)
	}

	if a2.Err != nil {
		b.WriteString(fmt.Sprintf("\n\n---\n## 검토 오류\n%v\n", a2.Err))
	} else if a2.Content != "" {
		b.WriteString("\n\n---\n## 검토 의견\n")
		b.WriteString(a2.Content)
	}

	return b.String()
}

func simpleMergeConsensus(a1, a2 AgentResult) string {
	var b strings.Builder

	b.WriteString("## 관점 1\n")
	if a1.Err != nil {
		b.WriteString(fmt.Sprintf("오류: %v\n", a1.Err))
	} else {
		b.WriteString(a1.Content)
	}

	b.WriteString("\n\n---\n## 관점 2\n")
	if a2.Err != nil {
		b.WriteString(fmt.Sprintf("오류: %v\n", a2.Err))
	} else {
		b.WriteString(a2.Content)
	}

	return b.String()
}

func simpleMergeScan(a1, a2 AgentResult) string {
	var b strings.Builder
	b.WriteString("## 병렬 스캔 결과\n\n")

	if a1.Err != nil {
		b.WriteString(fmt.Sprintf("### 영역 1\n오류: %v\n\n", a1.Err))
	} else if a1.Content != "" {
		b.WriteString("### 영역 1\n")
		b.WriteString(a1.Content)
		b.WriteString("\n\n")
	}

	if a2.Err != nil {
		b.WriteString(fmt.Sprintf("### 영역 2\n오류: %v\n\n", a2.Err))
	} else if a2.Content != "" {
		b.WriteString("### 영역 2\n")
		b.WriteString(a2.Content)
		b.WriteString("\n\n")
	}

	return b.String()
}
