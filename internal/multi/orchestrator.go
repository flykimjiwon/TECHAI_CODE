package multi

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/llm"
	"github.com/kimjiwon/tgc/internal/tools"
)

const maxToolIter = 20

// Orchestrator manages parallel agent execution.
type Orchestrator struct {
	client       *llm.Client
	model1       string // Super model (primary, write access)
	model2       string // Dev model (secondary, read-only)
	synthModel   string // Model used for synthesis (defaults to model2)
	workDir      string // working directory for scan partitioning
	progress     chan AgentProgress
}

// NewOrchestrator creates an orchestrator with two model identifiers.
func NewOrchestrator(client *llm.Client, superModel, devModel, workDir string) *Orchestrator {
	return &Orchestrator{
		client:     client,
		model1:     superModel,
		model2:     devModel,
		synthModel: devModel, // synthesis uses the faster/cheaper model
		workDir:    workDir,
		progress:   make(chan AgentProgress, 64),
	}
}

// Progress returns the channel that receives real-time agent updates.
func (o *Orchestrator) Progress() <-chan AgentProgress {
	return o.progress
}

// Run executes the multi-agent pipeline for the given strategy.
func (o *Orchestrator) Run(ctx context.Context, strategy Strategy, history []openai.ChatCompletionMessage, mode int) MergedResult {
	start := time.Now()
	config.DebugLog("[MULTI] Run strategy=%s model1=%s model2=%s historyLen=%d",
		strategy, o.model1, o.model2, len(history))

	var result MergedResult
	result.Strategy = strategy

	switch strategy {
	case StrategyReview:
		result = o.runReview(ctx, history, mode)
	case StrategyConsensus:
		result = o.runConsensus(ctx, history, mode)
	case StrategyScan:
		result = o.runScan(ctx, history, mode)
	}

	result.Elapsed = time.Since(start)
	result.TotalTokens = result.Agent1.Tokens + result.Agent2.Tokens

	// LLM synthesis: combine both outputs into one cohesive response
	o.sendProgress(Agent1Super, "synthesizing", "응답 종합중", result.TotalTokens, time.Since(start))
	result.FinalOutput = MergeWithSynthesis(ctx, o.client, o.synthModel, strategy, result.Agent1, result.Agent2)

	config.DebugLog("[MULTI] complete strategy=%s elapsed=%v totalTokens=%d",
		strategy, result.Elapsed, result.TotalTokens)

	close(o.progress)
	return result
}

// runReview: Agent1 generates with full tools → Agent2 reviews the result read-only.
func (o *Orchestrator) runReview(ctx context.Context, history []openai.ChatCompletionMessage, mode int) MergedResult {
	var result MergedResult

	// Phase 1: Agent1 (Super) with full tool access
	o.sendProgress(Agent1Super, "streaming", "", 0, 0)
	result.Agent1 = o.runAgent(ctx, Agent1Super, o.model1, history, tools.AllTools(), mode)
	o.sendProgress(Agent1Super, "done", "", result.Agent1.Tokens, result.Agent1.Elapsed)

	if result.Agent1.Err != nil {
		// Skip Agent2 if Agent1 failed
		o.sendProgress(Agent2Dev, "error", "Agent1 실패로 건너뜀", 0, 0)
		return result
	}

	// Phase 2: Agent2 (Dev) reviews Agent1's output — read-only tools only
	o.sendProgress(Agent2Dev, "streaming", "", 0, 0)

	reviewHistory := make([]openai.ChatCompletionMessage, len(history))
	copy(reviewHistory, history)
	reviewPrompt := fmt.Sprintf(
		`아래는 다른 AI가 생성한 응답입니다. 당신은 검토자(Reviewer)입니다.

## 검토 규칙
1. 원본 응답의 내용을 절대 반복하지 마세요.
2. 오직 다음 항목만 작성하세요:
   - 오류/버그가 있으면 구체적으로 지적 (어디서, 왜, 수정 방안)
   - 보안 취약점이 있으면 지적
   - 성능 개선이 가능하면 제안
   - 빠진 중요 내용이 있으면 보충
   - 더 나은 접근법이 있으면 제안
3. 문제가 없으면 "검토 완료: 특이사항 없음"이라고만 쓰세요.
4. 간결하게 작성하세요. 칭찬이나 요약은 불필요합니다.

## 원본 응답:
%s`, result.Agent1.Content,
	)
	reviewHistory = append(reviewHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: reviewPrompt,
	})

	result.Agent2 = o.runAgent(ctx, Agent2Dev, o.model2, reviewHistory, tools.ReadOnlyTools(), mode)
	o.sendProgress(Agent2Dev, "done", "", result.Agent2.Tokens, result.Agent2.Elapsed)

	return result
}

// runConsensus: Both agents receive the same prompt in parallel.
func (o *Orchestrator) runConsensus(ctx context.Context, history []openai.ChatCompletionMessage, mode int) MergedResult {
	var result MergedResult
	var wg sync.WaitGroup
	var a1, a2 AgentResult

	history1 := make([]openai.ChatCompletionMessage, len(history))
	copy(history1, history)
	history2 := make([]openai.ChatCompletionMessage, len(history))
	copy(history2, history)

	wg.Add(2)

	// Agent1 with full tools
	go func() {
		defer wg.Done()
		o.sendProgress(Agent1Super, "streaming", "", 0, 0)
		a1 = o.runAgent(ctx, Agent1Super, o.model1, history1, tools.AllTools(), mode)
		o.sendProgress(Agent1Super, "done", "", a1.Tokens, a1.Elapsed)
	}()

	// Agent2 with read-only tools
	go func() {
		defer wg.Done()
		o.sendProgress(Agent2Dev, "streaming", "", 0, 0)
		a2 = o.runAgent(ctx, Agent2Dev, o.model2, history2, tools.ReadOnlyTools(), mode)
		o.sendProgress(Agent2Dev, "done", "", a2.Tokens, a2.Elapsed)
	}()

	wg.Wait()
	result.Agent1 = a1
	result.Agent2 = a2
	return result
}

// runScan: Both agents search in parallel. Both use read-only tools.
func (o *Orchestrator) runScan(ctx context.Context, history []openai.ChatCompletionMessage, mode int) MergedResult {
	var result MergedResult
	var wg sync.WaitGroup
	var a1, a2 AgentResult

	history1 := make([]openai.ChatCompletionMessage, len(history))
	copy(history1, history)
	history2 := make([]openai.ChatCompletionMessage, len(history))
	copy(history2, history)

	// Partition by subdirectories for smarter parallel scanning
	if len(history1) > 0 {
		lastIdx := len(history1) - 1
		if history1[lastIdx].Role == openai.ChatMessageRoleUser {
			group1, group2 := SplitSubdirs(o.workDir)
			if len(group1) > 0 && len(group2) > 0 {
				hint1 := fmt.Sprintf("\n\n[SCAN NOTE: 당신은 Agent1입니다. 다음 디렉토리를 중점 탐색하세요: %s]",
					strings.Join(group1, ", "))
				hint2 := fmt.Sprintf("\n\n[SCAN NOTE: 당신은 Agent2입니다. 다음 디렉토리를 중점 탐색하세요: %s]",
					strings.Join(group2, ", "))
				history1[lastIdx].Content += hint1
				history2[lastIdx].Content += hint2
			} else {
				// Fallback: alphabetical split
				history1[lastIdx].Content += "\n\n[SCAN NOTE: 당신은 Agent1입니다. 프로젝트의 전반부(A-M 파일명)를 중점 탐색하세요.]"
				history2[lastIdx].Content += "\n\n[SCAN NOTE: 당신은 Agent2입니다. 프로젝트의 후반부(N-Z 파일명)를 중점 탐색하세요.]"
			}
		}
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		o.sendProgress(Agent1Super, "streaming", "", 0, 0)
		a1 = o.runAgent(ctx, Agent1Super, o.model1, history1, tools.ReadOnlyTools(), mode)
		o.sendProgress(Agent1Super, "done", "", a1.Tokens, a1.Elapsed)
	}()

	go func() {
		defer wg.Done()
		o.sendProgress(Agent2Dev, "streaming", "", 0, 0)
		a2 = o.runAgent(ctx, Agent2Dev, o.model2, history2, tools.ReadOnlyTools(), mode)
		o.sendProgress(Agent2Dev, "done", "", a2.Tokens, a2.Elapsed)
	}()

	wg.Wait()
	result.Agent1 = a1
	result.Agent2 = a2
	return result
}

// runAgent executes a single agent's streaming loop with tool calls.
func (o *Orchestrator) runAgent(
	ctx context.Context,
	agent AgentID,
	model string,
	history []openai.ChatCompletionMessage,
	toolDefs []openai.Tool,
	mode int,
) AgentResult {
	start := time.Now()
	var totalContent strings.Builder
	var tokens int
	toolIter := 0

	config.DebugLog("[MULTI-%s] start model=%s historyLen=%d tools=%d",
		agent, model, len(history), len(toolDefs))

	for toolIter < maxToolIter {
		ch := o.client.StreamChat(ctx, model, history, toolDefs)

		var buf strings.Builder
		var lastToolCalls []llm.ToolCallInfo
		var lastUsage *openai.Usage

		for chunk := range ch {
			if chunk.Err != nil {
				config.DebugLog("[MULTI-%s] stream error: %v", agent, chunk.Err)
				return AgentResult{
					Agent:   agent,
					Content: totalContent.String() + buf.String(),
					Tokens:  tokens,
					Elapsed: time.Since(start),
					Err:     chunk.Err,
				}
			}

			if chunk.Content != "" {
				buf.WriteString(chunk.Content)
				tokens++
				o.sendProgress(agent, "streaming", "", tokens, time.Since(start))
			}

			if chunk.Done {
				lastToolCalls = chunk.ToolCalls
				lastUsage = chunk.Usage
				if lastUsage != nil {
					total := lastUsage.TotalTokens
					if total == 0 {
						total = lastUsage.PromptTokens + lastUsage.CompletionTokens
					}
					if total > 0 {
						tokens = total
					}
				}
			}
		}

		content := buf.String()

		// No tool calls — final response
		if len(lastToolCalls) == 0 {
			totalContent.WriteString(content)
			config.DebugLog("[MULTI-%s] done toolIter=%d tokens=%d elapsed=%v",
				agent, toolIter, tokens, time.Since(start))
			return AgentResult{
				Agent:     agent,
				Content:   totalContent.String(),
				ToolCalls: toolIter,
				Tokens:    tokens,
				Elapsed:   time.Since(start),
				History:   history,
			}
		}

		// Process tool calls
		config.DebugLog("[MULTI-%s] tool_call iter=%d calls=%d", agent, toolIter, len(lastToolCalls))

		// Append assistant message with tool calls to history
		assistantMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		}
		var oaiToolCalls []openai.ToolCall
		for _, tc := range lastToolCalls {
			oaiToolCalls = append(oaiToolCalls, openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}
		assistantMsg.ToolCalls = oaiToolCalls
		history = append(history, assistantMsg)

		if content != "" {
			totalContent.WriteString(content)
		}

		// Execute tools
		for _, tc := range lastToolCalls {
			o.sendProgress(agent, "tool_call", tc.Name, tokens, time.Since(start))
			config.DebugLog("[MULTI-%s] executing tool %s", agent, tc.Name)
			output := tools.Execute(tc.Name, tc.Arguments)

			toolMsg := openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    output,
				ToolCallID: tc.ID,
			}
			history = append(history, toolMsg)
		}

		toolIter++
	}

	config.DebugLog("[MULTI-%s] tool loop limit reached (%d)", agent, maxToolIter)
	return AgentResult{
		Agent:     agent,
		Content:   totalContent.String(),
		ToolCalls: toolIter,
		Tokens:    tokens,
		Elapsed:   time.Since(start),
		Err:       fmt.Errorf("tool loop limit (%d iterations)", maxToolIter),
		History:   history,
	}
}

func (o *Orchestrator) sendProgress(agent AgentID, status, detail string, tokens int, elapsed time.Duration) {
	select {
	case o.progress <- AgentProgress{
		Agent:   agent,
		Status:  status,
		Detail:  detail,
		Tokens:  tokens,
		Elapsed: elapsed,
	}:
	default:
		// Non-blocking: drop if channel is full
	}
}
