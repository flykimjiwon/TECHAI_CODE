package knowledge

import (
	"io/fs"
	"runtime"
	"testing"
	"testing/fstest"
)

func testFS() fs.FS {
	return fstest.MapFS{
		"knowledge/docs/go/stdlib.md": &fstest.MapFile{
			Data: []byte("# Go Standard Library Quick Reference\n\n## fmt\n- `fmt.Sprintf`\n"),
		},
		"knowledge/skills/debugging.md": &fstest.MapFile{
			Data: []byte("# Debugging Skill\n\n## 절차\n1. 에러 메시지 읽기\n"),
		},
		"knowledge/docs/terminal/windows.md": &fstest.MapFile{
			Data: []byte("# Windows Terminal Commands\n\n## PowerShell\n- Get-ChildItem\n"),
		},
		"knowledge/docs/terminal/linux.md": &fstest.MapFile{
			Data: []byte("# Linux Terminal Commands\n\n## bash\n- ls, cat, grep\n"),
		},
		"knowledge/docs/terminal/macos.md": &fstest.MapFile{
			Data: []byte("# macOS Terminal Commands\n\n## zsh\n- ls, cat, grep\n"),
		},
		"knowledge/docs/bxm/overview.md": &fstest.MapFile{
			Data: []byte("# BXM Overview\n\nProduct-specific documentation.\n"),
		},
		"knowledge/docs/python/basics.md": &fstest.MapFile{
			Data: []byte("# Python Basics\n\n## print\n- print(value)\n"),
		},
		"knowledge/docs/vue/components.md": &fstest.MapFile{
			Data: []byte("# Vue Components\n\n## Template\n- <template>\n"),
		},
	}
}

func TestNewStore(t *testing.T) {
	s, err := NewStore(testFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	if s.DocCount() == 0 {
		t.Fatal("expected at least one document, got 0")
	}
	// We have 8 test files
	if s.DocCount() != 8 {
		t.Errorf("expected 8 documents, got %d", s.DocCount())
	}
}

func TestStoreSearch(t *testing.T) {
	s, err := NewStore(testFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Search for Go stdlib
	results := s.Search([]string{"go", "stdlib"}, 10000)
	if len(results) == 0 {
		t.Fatal("expected results for ['go', 'stdlib'], got none")
	}

	found := false
	for _, doc := range results {
		if doc.Path == "knowledge/docs/go/stdlib.md" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected knowledge/docs/go/stdlib.md in search results")
	}
}

func TestStoreSearchBudget(t *testing.T) {
	s, err := NewStore(testFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Search with a very small budget (1 token = ~4 chars)
	// Should return nothing or very few results
	results := s.Search([]string{"go"}, 1)
	totalTokens := 0
	for _, doc := range results {
		totalTokens += estimateTokens(doc.Content)
	}
	if totalTokens > 1 {
		t.Errorf("expected results within budget of 1 token, got %d tokens", totalTokens)
	}
}

func TestStoreSearchEmpty(t *testing.T) {
	s, err := NewStore(testFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Empty keywords
	results := s.Search(nil, 10000)
	if results != nil {
		t.Errorf("expected nil for empty keywords, got %d results", len(results))
	}

	// Zero budget
	results = s.Search([]string{"go"}, 0)
	if results != nil {
		t.Errorf("expected nil for zero budget, got %d results", len(results))
	}

	// Non-existent keyword
	results = s.Search([]string{"nonexistentkeyword"}, 10000)
	if results != nil {
		t.Errorf("expected nil for non-matching keyword, got %d results", len(results))
	}
}

func TestStoreSearchTierOrder(t *testing.T) {
	s, err := NewStore(testFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Search for "overview" (bxm, tier 0) and "basics" (python, tier 3)
	results := s.Search([]string{"overview", "basics"}, 10000)
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}

	// First result should be tier 0 (bxm)
	if results[0].Tier != Tier0 {
		t.Errorf("expected first result tier 0, got tier %d (path: %s)", results[0].Tier, results[0].Path)
	}
}

func TestStoreForOS(t *testing.T) {
	s, err := NewStore(testFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	goos := runtime.GOOS
	results := s.ForOS(goos)

	for _, doc := range results {
		if doc.OS != "" && doc.OS != goos {
			t.Errorf("ForOS(%q) returned doc with OS=%q: %s", goos, doc.OS, doc.Path)
		}
	}

	// All OS-agnostic docs should be included
	agnosticCount := 0
	for _, doc := range s.docs {
		if doc.OS == "" {
			agnosticCount++
		}
	}

	osSpecificCount := 0
	for _, doc := range s.docs {
		if doc.OS == goos {
			osSpecificCount++
		}
	}

	expectedCount := agnosticCount + osSpecificCount
	if len(results) != expectedCount {
		t.Errorf("ForOS(%q): expected %d docs (agnostic=%d + os=%d), got %d",
			goos, expectedCount, agnosticCount, osSpecificCount, len(results))
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"a", 1},        // 1 char -> ceil(1/4) = 1
		{"abcd", 1},     // 4 chars -> 4/4 = 1
		{"abcde", 2},    // 5 chars -> ceil(5/4) = 2
		{"hello world", 3}, // 11 chars -> ceil(11/4) = 3
	}

	for _, tt := range tests {
		got := estimateTokens(tt.input)
		if got != tt.expected {
			t.Errorf("estimateTokens(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestInferMetadata(t *testing.T) {
	tests := []struct {
		path         string
		wantTier     int
		wantOS       string
		wantKeywords []string
	}{
		{
			path:         "knowledge/docs/bxm/overview.md",
			wantTier:     Tier0,
			wantOS:       "",
			wantKeywords: []string{"bxm", "overview"},
		},
		{
			path:         "knowledge/docs/go/stdlib.md",
			wantTier:     Tier1,
			wantOS:       "",
			wantKeywords: []string{"go", "stdlib"},
		},
		{
			path:         "knowledge/docs/javascript/async.md",
			wantTier:     Tier1,
			wantOS:       "",
			wantKeywords: []string{"javascript", "async"},
		},
		{
			path:         "knowledge/docs/vue/components.md",
			wantTier:     Tier2,
			wantOS:       "",
			wantKeywords: []string{"vue", "components"},
		},
		{
			path:         "knowledge/docs/python/basics.md",
			wantTier:     Tier3,
			wantOS:       "",
			wantKeywords: []string{"python", "basics"},
		},
		{
			path:         "knowledge/skills/debugging.md",
			wantTier:     Tier1,
			wantOS:       "",
			wantKeywords: []string{"debugging"},
		},
		{
			path:     "knowledge/docs/terminal/windows.md",
			wantTier: Tier1,
			wantOS:   "windows",
			wantKeywords: []string{"terminal", "windows"},
		},
		{
			path:     "knowledge/docs/terminal/linux.md",
			wantTier: Tier1,
			wantOS:   "linux",
			wantKeywords: []string{"terminal", "linux"},
		},
		{
			path:     "knowledge/docs/terminal/macos.md",
			wantTier: Tier1,
			wantOS:   "darwin",
			wantKeywords: []string{"terminal", "macos"},
		},
		{
			path:         "knowledge/docs/go/error-handling.md",
			wantTier:     Tier1,
			wantOS:       "",
			wantKeywords: []string{"go", "error-handling", "error", "handling"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			tier, osName, keywords := inferMetadata(tt.path)
			if tier != tt.wantTier {
				t.Errorf("tier = %d, want %d", tier, tt.wantTier)
			}
			if osName != tt.wantOS {
				t.Errorf("os = %q, want %q", osName, tt.wantOS)
			}
			if len(keywords) != len(tt.wantKeywords) {
				t.Errorf("keywords = %v, want %v", keywords, tt.wantKeywords)
			} else {
				for i, kw := range keywords {
					if kw != tt.wantKeywords[i] {
						t.Errorf("keywords[%d] = %q, want %q", i, kw, tt.wantKeywords[i])
					}
				}
			}
		})
	}
}
