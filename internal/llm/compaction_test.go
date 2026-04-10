package llm

import (
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

// makeMsg creates a message with role and content for test fixtures.
func makeMsg(role, content string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{Role: role, Content: content}
}

// makeConversation builds a synthetic conversation of size n where
// index 0 is the system prompt, then alternating user/assistant/tool.
func makeConversation(n int, toolContentLen int) []openai.ChatCompletionMessage {
	msgs := make([]openai.ChatCompletionMessage, 0, n)
	msgs = append(msgs, makeMsg(openai.ChatMessageRoleSystem, "system prompt"))
	for i := 1; i < n; i++ {
		switch i % 3 {
		case 0:
			msgs = append(msgs, makeMsg(openai.ChatMessageRoleUser, "user msg"))
		case 1:
			msgs = append(msgs, makeMsg(openai.ChatMessageRoleAssistant, "assistant reply"))
		case 2:
			msgs = append(msgs, makeMsg(openai.ChatMessageRoleTool, strings.Repeat("T", toolContentLen)))
		}
	}
	return msgs
}

func TestEstimateTokens_CharsOverFour(t *testing.T) {
	msgs := []openai.ChatCompletionMessage{
		makeMsg(openai.ChatMessageRoleUser, strings.Repeat("a", 100)), // 25 tok
		makeMsg(openai.ChatMessageRoleAssistant, strings.Repeat("b", 200)), // 50 tok
	}
	got := estimateTokens(msgs)
	want := 75
	if got != want {
		t.Fatalf("estimateTokens = %d, want %d", got, want)
	}
}

func TestCompact_BelowThreshold_NoChange(t *testing.T) {
	// Less than 40 messages → Compact should be a no-op.
	msgs := makeConversation(39, 500)
	before := len(msgs)
	beforeLastToolContent := ""
	for _, m := range msgs {
		if m.Role == openai.ChatMessageRoleTool {
			beforeLastToolContent = m.Content
		}
	}

	out := Compact(msgs)

	if len(out) != before {
		t.Fatalf("len changed: got %d, want %d", len(out), before)
	}
	for _, m := range out {
		if m.Role == openai.ChatMessageRoleTool && m.Content != beforeLastToolContent {
			t.Fatalf("tool message was modified below threshold: %q", m.Content)
		}
	}
}

func TestCompact_SnipsOldLargeToolMessages(t *testing.T) {
	// 50 messages, tool content = 500 chars (> 200 threshold).
	msgs := makeConversation(50, 500)

	out := Compact(msgs)

	// Old tool messages (index < len-10) should be snipped.
	boundary := len(out) - 10
	for i := 0; i < boundary; i++ {
		if out[i].Role != openai.ChatMessageRoleTool {
			continue
		}
		if !strings.HasPrefix(out[i].Content, "[snipped:") {
			t.Errorf("old tool msg[%d] not snipped: %q", i, out[i].Content)
		}
	}
}

func TestCompact_PreservesLastTenMessages(t *testing.T) {
	msgs := makeConversation(50, 500)
	// Capture last 10 tool contents before compaction.
	original := make([]string, len(msgs))
	for i, m := range msgs {
		original[i] = m.Content
	}

	out := Compact(msgs)

	// Last 10 messages should be untouched.
	start := len(out) - 10
	for i := start; i < len(out); i++ {
		if out[i].Content != original[i] {
			t.Errorf("msg[%d] (in last 10) was modified: got %q, want %q", i, out[i].Content, original[i])
		}
	}
}

func TestCompact_DoesNotSnipShortToolMessages(t *testing.T) {
	// Tool content below 200 chars → should NOT be snipped even if old.
	msgs := makeConversation(50, 50)

	out := Compact(msgs)

	boundary := len(out) - 10
	for i := 0; i < boundary; i++ {
		if out[i].Role != openai.ChatMessageRoleTool {
			continue
		}
		if strings.HasPrefix(out[i].Content, "[snipped:") {
			t.Errorf("short tool msg[%d] was snipped (should not be): %q", i, out[i].Content)
		}
	}
}

func TestCompact_TruncatesVeryLongMessages(t *testing.T) {
	// Build 45 msgs where one user message in the "last 10" zone is 10000 chars.
	msgs := makeConversation(45, 100)
	// Index 40 is inside last-10 (boundary = 35). Make it a huge user message.
	huge := strings.Repeat("X", 10000)
	msgs[40] = makeMsg(openai.ChatMessageRoleUser, huge)

	out := Compact(msgs)

	if len(out[40].Content) >= 10000 {
		t.Errorf("huge msg was not truncated: len=%d", len(out[40].Content))
	}
	if !strings.Contains(out[40].Content, "[truncated]") {
		t.Errorf("truncated marker missing: %q", out[40].Content[:min(100, len(out[40].Content))])
	}
}

func TestCompact_PreservesMessageOrder(t *testing.T) {
	msgs := makeConversation(50, 500)
	// Tag each message with its original index so we can detect reordering.
	for i := range msgs {
		if i == 0 {
			continue
		}
		msgs[i].Name = "idx"
	}
	roles := make([]string, len(msgs))
	for i, m := range msgs {
		roles[i] = m.Role
	}

	out := Compact(msgs)

	if len(out) != len(msgs) {
		t.Fatalf("length changed: got %d, want %d", len(out), len(msgs))
	}
	for i := range out {
		if out[i].Role != roles[i] {
			t.Errorf("role reordering at [%d]: got %q, want %q", i, out[i].Role, roles[i])
		}
	}
}

func TestCompact_PreservesSystemPrompt(t *testing.T) {
	msgs := makeConversation(50, 500)
	sysContent := "system prompt"
	msgs[0] = makeMsg(openai.ChatMessageRoleSystem, sysContent)

	out := Compact(msgs)

	if out[0].Role != openai.ChatMessageRoleSystem {
		t.Errorf("system prompt role changed: %q", out[0].Role)
	}
	if out[0].Content != sysContent {
		t.Errorf("system prompt content changed: got %q, want %q", out[0].Content, sysContent)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
