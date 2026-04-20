package tools

import (
	"strings"
	"testing"
)

func TestGrepSearch_Parallel(t *testing.T) {
	result, err := GrepSearch("func.*Search", "../..", "", false, 0)
	if err != nil {
		t.Fatalf("GrepSearch failed: %v", err)
	}
	if result == "No matches found." {
		t.Fatal("Expected matches for 'func.*Search'")
	}
	if !strings.Contains(result, "matches") {
		t.Fatal("Expected scan summary in results")
	}
	t.Logf("Result length: %d bytes", len(result))
}

func TestGrepSearch_NoMatch(t *testing.T) {
	// Search in a directory with no Go source matching this regex
	result, err := GrepSearch("^ZZZZZZZZZZZZZZ$", "../..", "*.xyz", false, 0)
	if err != nil {
		t.Fatalf("GrepSearch failed: %v", err)
	}
	if !strings.HasPrefix(result, "No matches") {
		t.Fatalf("Expected no matches but got: %s", result[:min(len(result), 100)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestFuzzyScore(t *testing.T) {
	tests := []struct {
		needle   string
		haystack string
		wantPos  bool
	}{
		{"apgo", "app.go", true},
		{"regst", "internal/tools/registry.go", true},
		{"clnt", "internal/llm/client.go", true},
		{"zzz", "app.go", false},
	}

	for _, tt := range tests {
		score := fuzzyScore(tt.needle, tt.haystack)
		if tt.wantPos && score <= 0 {
			t.Errorf("fuzzyScore(%q, %q) = %d, want > 0", tt.needle, tt.haystack, score)
		}
		if !tt.wantPos && score > 0 {
			t.Errorf("fuzzyScore(%q, %q) = %d, want 0", tt.needle, tt.haystack, score)
		}
	}
}

func TestTruncateLine(t *testing.T) {
	short := "hello world"
	if truncateLine(short) != short {
		t.Error("Short line should not be truncated")
	}

	long := strings.Repeat("a", 3000)
	result := truncateLine(long)
	if len([]rune(result)) > maxLineChars+3 { // +3 for "..."
		t.Errorf("Truncated line too long: %d runes", len([]rune(result)))
	}
}
