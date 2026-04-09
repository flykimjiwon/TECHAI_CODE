package knowledge

import (
	"strings"
	"testing"
	"testing/fstest"
)

func injectorTestFS() fstest.MapFS {
	return fstest.MapFS{
		"knowledge/docs/go/stdlib.md": &fstest.MapFile{
			Data: []byte("# Go Standard Library Quick Reference\n\n## fmt\n- `fmt.Sprintf` formats strings\n- `fmt.Fprintf` writes to a writer\n"),
		},
		"knowledge/docs/go/concurrency.md": &fstest.MapFile{
			Data: []byte("# Go Concurrency Patterns\n\n## goroutine\nUse `go func()` to launch a goroutine.\n"),
		},
		"knowledge/skills/debugging.md": &fstest.MapFile{
			Data: []byte("# Debugging Skill\n\n## 절차\n1. 에러 메시지 읽기\n2. 스택 트레이스 확인\n"),
		},
		"knowledge/docs/bxm/overview.md": &fstest.MapFile{
			Data: []byte("# BXM Overview\n\nBXM은 뱅크웨어 글로벌의 핵심 프레임워크입니다.\n## Bean\n## DBIO\n## Service\n"),
		},
		"knowledge/docs/bxm/bean.md": &fstest.MapFile{
			Data: []byte("# BXM Bean Guide\n\n@BxmBean 어노테이션을 사용하여 빈을 등록합니다.\n"),
		},
		"knowledge/docs/python/basics.md": &fstest.MapFile{
			Data: []byte("# Python Basics\n\n## print\n- print(value)\n"),
		},
		"knowledge/docs/css/no-heading.md": &fstest.MapFile{
			Data: []byte("This document has no heading.\nJust plain content about CSS.\n"),
		},
	}
}

func TestInjectorInject(t *testing.T) {
	store, err := NewStore(injectorTestFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	inj := NewInjector(store, 8192)

	// Query about Go fmt usage should return knowledge context
	result := inj.Inject(0, "Go fmt 사용법")
	if result == "" {
		t.Fatal("expected non-empty result for 'Go fmt 사용법', got empty")
	}
	if !strings.Contains(result, "## Knowledge Context") {
		t.Error("result should contain '## Knowledge Context' header")
	}
	if !strings.Contains(result, "레퍼런스 문서") {
		t.Error("result should contain Korean description")
	}
}

func TestInjectorInjectBXM(t *testing.T) {
	store, err := NewStore(injectorTestFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	inj := NewInjector(store, 8192)

	// Query with BXM keywords should include BXM docs
	result := inj.Inject(0, "BXM bean 등록 방법")
	if result == "" {
		t.Fatal("expected non-empty result for BXM query, got empty")
	}
	if !strings.Contains(result, "BXM") {
		t.Error("result should contain BXM content")
	}
}

func TestInjectorTokenBudget(t *testing.T) {
	store, err := NewStore(injectorTestFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	// Very small budget: only 100 tokens (~400 chars)
	inj := NewInjector(store, 100)

	result := inj.Inject(0, "Go fmt concurrency goroutine")
	// Result should be within budget. The header alone uses some tokens,
	// so at most one small doc should fit.
	if result != "" {
		totalTokens := estimateTokens(result)
		// Allow some slack for the header overhead, but should be reasonable
		if totalTokens > 120 {
			t.Errorf("result exceeded budget: %d tokens for budget 100", totalTokens)
		}
	}
}

func TestInjectorEmptyQuery(t *testing.T) {
	store, err := NewStore(injectorTestFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	inj := NewInjector(store, 8192)

	// Empty string query -> return ""
	result := inj.Inject(0, "")
	if result != "" {
		t.Errorf("expected empty result for empty query, got %q", result)
	}

	// Query with no tech keywords -> return ""
	result = inj.Inject(0, "안녕하세요")
	if result != "" {
		t.Errorf("expected empty result for non-tech query '안녕하세요', got %q", result)
	}

	// Another non-tech query
	result = inj.Inject(0, "오늘 날씨 어때요")
	if result != "" {
		t.Errorf("expected empty result for non-tech query, got %q", result)
	}
}

func TestInjectorTitleExtraction(t *testing.T) {
	store, err := NewStore(injectorTestFS())
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	inj := NewInjector(store, 8192)

	// Doc with "# Go Standard Library Quick Reference" heading
	result := inj.Inject(0, "Go stdlib fmt")
	if result == "" {
		t.Fatal("expected non-empty result, got empty")
	}
	if !strings.Contains(result, "### Go Standard Library Quick Reference") {
		t.Errorf("expected section header from doc heading, got:\n%s", result)
	}

	// Doc without heading should use filename
	result = inj.Inject(0, "css 반응형 스타일")
	if result == "" {
		t.Fatal("expected non-empty result for CSS query, got empty")
	}
	if strings.Contains(result, "no-heading") {
		// The doc "no-heading.md" has no # heading, so extractTitle should use "no-heading" as filename
		if !strings.Contains(result, "### no-heading") {
			t.Errorf("expected '### no-heading' as fallback title from filename, got:\n%s", result)
		}
	}
}

func TestExtractTitleWithHeading(t *testing.T) {
	doc := Doc{
		Path:    "knowledge/docs/go/stdlib.md",
		Content: "# My Great Title\n\nSome content here.\n",
	}
	title := extractTitle(doc)
	if title != "My Great Title" {
		t.Errorf("extractTitle = %q, want %q", title, "My Great Title")
	}
}

func TestExtractTitleWithoutHeading(t *testing.T) {
	doc := Doc{
		Path:    "knowledge/docs/css/layout-tips.md",
		Content: "No heading here.\nJust content.\n",
	}
	title := extractTitle(doc)
	if title != "layout-tips" {
		t.Errorf("extractTitle = %q, want %q", title, "layout-tips")
	}
}

func TestExtractTitleEmptyContent(t *testing.T) {
	doc := Doc{
		Path:    "knowledge/docs/misc/empty.md",
		Content: "",
	}
	title := extractTitle(doc)
	if title != "empty" {
		t.Errorf("extractTitle = %q, want %q", title, "empty")
	}
}

func TestInjectorNilStore(t *testing.T) {
	// Verify NewInjector works and Inject handles empty store gracefully
	store, err := NewStore(fstest.MapFS{
		"knowledge/placeholder.md": &fstest.MapFile{
			Data: []byte("placeholder"),
		},
	})
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	inj := NewInjector(store, 8192)
	result := inj.Inject(0, "Go fmt")
	// Store has no Go docs, so search should find nothing
	if result != "" {
		t.Errorf("expected empty result from store with no matching docs, got %q", result)
	}
}
