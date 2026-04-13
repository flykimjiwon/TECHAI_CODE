package knowledge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanAndSearchUserDocs(t *testing.T) {
	// Create temp knowledge dir
	tmp := t.TempDir()
	kdir := filepath.Join(tmp, ".tgc", "knowledge")
	if err := os.MkdirAll(kdir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write test docs
	md := `# 배포 가이드

## 빌드 방법
make build로 로컬 바이너리 생성.

## 온프레미스 설정
엔드포인트: https://example.com/v1
모델: gpt-oss-120b
`
	if err := os.WriteFile(filepath.Join(kdir, "deploy.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	txt := `코딩 규칙
Go 1.22 필수
CGO 사용 금지
Bubble Tea v2만 사용
`
	if err := os.WriteFile(filepath.Join(kdir, "rules.txt"), []byte(txt), 0o644); err != nil {
		t.Fatal(err)
	}

	// Override knowledgeDirs for testing
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(origWd) }()

	// Scan
	idx := ScanUserDocs()
	if idx.Count() != 2 {
		t.Fatalf("expected 2 docs, got %d", idx.Count())
	}

	// TableOfContents
	toc := idx.TableOfContents()
	if toc == "" {
		t.Fatal("TableOfContents returned empty")
	}
	t.Logf("TOC:\n%s", toc)

	// Search: keyword match
	results := idx.Search("빌드", 3)
	if len(results) == 0 {
		t.Fatal("search '빌드' returned 0 results, expected at least 1")
	}
	t.Logf("Search '빌드': %d results, first=%s", len(results), results[0].Title)

	// Search: AND match (both terms)
	results2 := idx.Search("온프레미스 모델", 3)
	if len(results2) == 0 {
		t.Fatal("search '온프레미스 모델' returned 0 results")
	}
	t.Logf("Search '온프레미스 모델': %d results", len(results2))

	// Search: no match
	results3 := idx.Search("xyznotexist", 3)
	if len(results3) != 0 {
		t.Fatalf("expected 0 results for nonsense query, got %d", len(results3))
	}

	// ReadFull
	content, ok := idx.ReadFull("deploy.md")
	if !ok {
		t.Fatal("ReadFull('deploy.md') returned false")
	}
	if len(content) == 0 {
		t.Fatal("ReadFull returned empty content")
	}
	t.Logf("ReadFull deploy.md: %d chars", len(content))

	// FormatSearchResults
	formatted := FormatSearchResults(results, "빌드")
	if formatted == "" {
		t.Fatal("FormatSearchResults returned empty")
	}
	t.Logf("Formatted:\n%s", formatted)
}
