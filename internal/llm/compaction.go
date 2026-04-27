package llm

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
)

// Compaction thresholds.
const (
	// compactMinMessages is the minimum conversation length before stage 1
	// (snip) runs. Below this, Compact is a no-op.
	compactMinMessages = 40

	// compactKeepTail is the number of trailing messages that stage 1 never
	// touches — these are the "recent working memory".
	compactKeepTail = 10

	// compactToolSnipThreshold is the minimum tool-message size (in chars)
	// that qualifies for stage 1 snipping. Short tool outputs stay intact.
	compactToolSnipThreshold = 200

	// compactMaxMessageChars is the upper bound on any single message after
	// stage 2. Messages longer than this are head+tail truncated.
	compactMaxMessageChars = 4000

	// compactKeepChars is how many characters are kept from the head and
	// tail of a truncated message.
	compactKeepChars = 2000

	// compactStage3KeepTail is how many recent messages stage 3 preserves
	// verbatim when LLM summarization is triggered.
	compactStage3KeepTail = 10
)

// estimateTokens returns a rough token count using rune-based heuristic.
// Mixed Korean/English text averages ~2 runes per token. Pure English is ~4 chars/token
// but rune count handles CJK correctly (1 CJK rune ≈ 1.5 tokens).
// Real token counts come from the API usage field when available.
func estimateTokens(msgs []openai.ChatCompletionMessage) int {
	total := 0
	for _, m := range msgs {
		total += (utf8.RuneCountInString(m.Content) + 1) / 2
	}
	return total
}

// Compact applies stage 1 (snip) and stage 2 (micro truncate) compaction
// in-place and returns the (possibly modified) slice. It is a no-op when the
// conversation has fewer than compactMinMessages messages.
//
// Stage 1 — Snip: for every tool-role message older than the last
// compactKeepTail messages whose content exceeds compactToolSnipThreshold
// chars, replace the content with a "[snipped: N lines]" marker.
//
// Stage 2 — Micro: for every message (regardless of role or position) whose
// content exceeds compactMaxMessageChars, keep the first and last
// compactKeepChars characters joined by a "[truncated]" marker.
//
// Message order, roles, and the system prompt at index 0 are always
// preserved.
func Compact(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	if len(messages) < compactMinMessages {
		return messages
	}

	config.DebugLog("[COMPACT] stage 1 snip | messages=%d", len(messages))

	// Stage 1 — Snip: replace old, large tool messages with a line-count
	// marker. Skip the last compactKeepTail messages to keep recent tool
	// output available to the model.
	boundary := len(messages) - compactKeepTail
	snipped := 0
	for i := 0; i < boundary; i++ {
		if messages[i].Role != openai.ChatMessageRoleTool {
			continue
		}
		if len(messages[i].Content) <= compactToolSnipThreshold {
			continue
		}
		lines := strings.Count(messages[i].Content, "\n") + 1
		messages[i].Content = fmt.Sprintf("[snipped: %d lines]", lines)
		snipped++
	}
	if snipped > 0 {
		config.DebugLog("[COMPACT] snipped %d tool messages", snipped)
	}

	// Stage 2 — Micro: head+tail truncate any oversized message. This runs
	// across ALL messages (including the last compactKeepTail) because a
	// single runaway message can still blow the context window even after
	// stage 1.
	truncated := 0
	for i := range messages {
		if len(messages[i].Content) <= compactMaxMessageChars {
			continue
		}
		content := messages[i].Content
		head := content[:compactKeepChars]
		tail := content[len(content)-compactKeepChars:]
		messages[i].Content = head + "\n...[truncated]...\n" + tail
		truncated++
	}
	if truncated > 0 {
		config.DebugLog("[COMPACT] truncated %d messages (>%d chars)", truncated, compactMaxMessageChars)
	}

	return messages
}

// CompactWithLLM applies stages 1 and 2 via Compact, then — if estimated
// tokens still exceed maxTokens — runs stage 3: summarize the middle of the
// conversation via a separate LLM call, keeping the system prompt at index 0
// and the last compactStage3KeepTail messages verbatim.
//
// On any stage-3 failure (LLM error, too few messages), the function falls
// back to the stage 1+2 result rather than aborting. Callers are expected to
// pass a non-nil client and a model ID suitable for summarization.
func CompactWithLLM(ctx context.Context, client *Client, model string, messages []openai.ChatCompletionMessage, maxTokens int) []openai.ChatCompletionMessage {
	messages = Compact(messages)

	tokens := estimateTokens(messages)
	if tokens <= maxTokens {
		return messages
	}

	config.DebugLog("[COMPACT] stage 3 LLM summary | tokens=%d > max=%d | messages=%d",
		tokens, maxTokens, len(messages))

	// Need at least system + 1 middle + tail to have something to summarize.
	if len(messages) <= compactStage3KeepTail+1 {
		return messages
	}

	sysMsg := messages[0]
	lastN := compactStage3KeepTail
	if len(messages)-1 < lastN {
		lastN = len(messages) - 1
	}
	tail := make([]openai.ChatCompletionMessage, lastN)
	copy(tail, messages[len(messages)-lastN:])

	middle := messages[1 : len(messages)-lastN]
	if len(middle) == 0 {
		return messages
	}

	var sb strings.Builder
	for _, m := range middle {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, m.Content))
	}

	summaryReq := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Summarize the conversation. Preserve: task goal, completed work, current state, decisions. Respond in Korean.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: sb.String(),
		},
	}

	summary, err := client.Chat(ctx, model, summaryReq)
	if err != nil {
		config.DebugLog("[COMPACT] LLM summary failed: %v", err)
		return messages
	}

	config.DebugLog("[COMPACT] LLM summary done | summaryLen=%d | kept=%d tail messages",
		len(summary), lastN)

	result := make([]openai.ChatCompletionMessage, 0, 2+lastN)
	result = append(result, sysMsg)
	result = append(result, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "[Previous conversation summary]\n" + summary,
	})
	result = append(result, tail...)

	return result
}
