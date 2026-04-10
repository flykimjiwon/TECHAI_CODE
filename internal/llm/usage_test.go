package llm

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestStreamChunk_HasUsageField(t *testing.T) {
	// RED: StreamChunk must carry an *openai.Usage so the caller can
	// update the real token count when the provider returns one.
	chunk := StreamChunk{
		Done: true,
		Usage: &openai.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}
	if chunk.Usage == nil {
		t.Fatal("StreamChunk.Usage was nil")
	}
	if chunk.Usage.TotalTokens != 150 {
		t.Fatalf("TotalTokens = %d, want 150", chunk.Usage.TotalTokens)
	}
}

func TestTokenCountFromUsage_NilReturnsZero(t *testing.T) {
	// RED: helper must be nil-safe so fallback path (provider that doesn't
	// return usage) stays at zero rather than panicking.
	if got := tokenCountFromUsage(nil); got != 0 {
		t.Fatalf("tokenCountFromUsage(nil) = %d, want 0", got)
	}
}

func TestTokenCountFromUsage_ReturnsTotal(t *testing.T) {
	u := &openai.Usage{
		PromptTokens:     1234,
		CompletionTokens: 567,
		TotalTokens:      1801,
	}
	if got := tokenCountFromUsage(u); got != 1801 {
		t.Fatalf("tokenCountFromUsage = %d, want 1801", got)
	}
}

func TestTokenCountFromUsage_FallbackToPromptPlusCompletion(t *testing.T) {
	// Some providers populate prompt/completion but leave TotalTokens at 0.
	// The helper should fall back to summing them so we still show a real
	// value instead of zero.
	u := &openai.Usage{
		PromptTokens:     800,
		CompletionTokens: 200,
		TotalTokens:      0,
	}
	if got := tokenCountFromUsage(u); got != 1000 {
		t.Fatalf("tokenCountFromUsage fallback = %d, want 1000", got)
	}
}
