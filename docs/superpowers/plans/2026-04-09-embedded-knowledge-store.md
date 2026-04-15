# Embedded Knowledge Store Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Go `embed`로 개발 레퍼런스 문서 + 스킬을 바이너리에 내장하여, 폐쇄망 LLM이 질문 키워드에 맞는 문서를 컨텍스트로 받아 정확한 코드를 생성하도록 한다.

**Architecture:** `knowledge/` 디렉토리를 `//go:embed`로 내장. `internal/knowledge/` 패키지가 키워드 추출 → 인덱스 검색 → 토큰 버짓 내 문서 선택 → 시스템 프롬프트 주입을 담당. `app.go`의 `sendMessage()`에서 주입기를 호출하여 매 요청마다 관련 지식을 동적 추가.

**Tech Stack:** Go 1.26, `embed.FS`, `encoding/json`, `runtime.GOOS`, `strings`, `regexp`

---

## File Structure

```
택가이코드/
├── knowledge/                          # //go:embed 대상 (신규 디렉토리)
│   ├── docs/                           # 레퍼런스 문서
│   │   ├── bxm/                        # Tier 0
│   │   │   └── (13 files)
│   │   ├── go/                         # Tier 1
│   │   │   └── (4 files)
│   │   ├── javascript/                 # Tier 1
│   │   │   └── (3 files)
│   │   ├── typescript/                 # Tier 1
│   │   │   └── (3 files)
│   │   ├── react/                      # Tier 1
│   │   │   └── (4 files)
│   │   ├── css/                        # Tier 1
│   │   │   └── (4 files)
│   │   ├── charts/                     # Tier 1-2
│   │   │   └── (5 files)
│   │   ├── vue/                        # Tier 2
│   │   │   └── (3 files)
│   │   ├── java/                       # Tier 2
│   │   │   └── (3 files)
│   │   ├── python/                     # Tier 3
│   │   │   └── (3 files)
│   │   ├── terminal/                   # OS별
│   │   │   └── (4 files)
│   │   └── tools/                      # 공통
│   │       └── (3 files)
│   ├── skills/                         # 작업 패턴
│   │   └── (7 files)
│   └── index.json                      # 키워드 → 문서 매핑
├── internal/knowledge/                 # Go 패키지 (신규)
│   ├── store.go                        # embed.FS + Document 구조체 + 로드
│   ├── store_test.go                   # store 테스트
│   ├── extractor.go                    # 키워드 추출기
│   ├── extractor_test.go              # 추출기 테스트
│   ├── injector.go                     # 프롬프트 주입 (모드+키워드 하이브리드)
│   └── injector_test.go               # 주입기 테스트
├── cmd/build-index/                    # 인덱스 빌드 도구 (신규)
│   └── main.go                         # knowledge/ 스캔 → index.json 생성
├── cmd/scrape-bxm/                     # BXM 문서 스크래핑 도구 (신규)
│   └── main.go                         # SWLab HTML → Markdown 변환
├── internal/app/app.go                 # 수정: sendMessage에 주입기 통합
└── Makefile                            # 수정: build-index 타겟 추가
```

---

## Task 1: knowledge 패키지 기본 구조 (store.go)

**Files:**
- Create: `internal/knowledge/store.go`
- Create: `internal/knowledge/store_test.go`
- Create: `knowledge/docs/.gitkeep` (embed가 빈 디렉토리 무시하므로 플레이스홀더)

- [ ] **Step 1: 테스트용 샘플 문서 생성**

```bash
mkdir -p knowledge/docs/go knowledge/skills
```

`knowledge/docs/go/stdlib.md`:
```markdown
# Go Standard Library Quick Reference

## fmt
- `fmt.Sprintf(format, args...)` — 포맷된 문자열 반환
- `fmt.Fprintf(w, format, args...)` — io.Writer에 출력
- `fmt.Errorf(format, args...)` — 에러 생성 (%w로 래핑)

## strings
- `strings.Contains(s, substr)` — 포함 여부
- `strings.Split(s, sep)` — 분리
- `strings.TrimSpace(s)` — 양쪽 공백 제거
- `strings.ReplaceAll(s, old, new)` — 전체 치환

## os
- `os.ReadFile(name)` — 파일 전체 읽기
- `os.WriteFile(name, data, perm)` — 파일 쓰기
- `os.Getenv(key)` — 환경변수
- `os.Getwd()` — 현재 디렉토리
```

`knowledge/skills/debugging.md`:
```markdown
# Debugging Skill

## 절차
1. 에러 메시지 정확히 읽기
2. 재현 조건 확인 (입력값, 환경)
3. 최소 재현 케이스 만들기
4. 이분법(bisect)으로 원인 좁히기
5. 가설 → 검증 → 수정 → 테스트

## 도구
- `grep_search`: 에러 메시지로 관련 코드 검색
- `file_read`: 스택 트레이스의 파일/라인 확인
- `shell_exec`: `git log --oneline -20`으로 최근 변경 확인
```

- [ ] **Step 2: store.go 실패 테스트 작성**

`internal/knowledge/store_test.go`:
```go
package knowledge

import (
	"testing"
)

func TestNewStore(t *testing.T) {
	s, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	if len(s.docs) == 0 {
		t.Fatal("NewStore() should load at least one document")
	}
}

func TestStoreSearch(t *testing.T) {
	s, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	results := s.Search([]string{"go", "fmt"}, 4096)
	if len(results) == 0 {
		t.Fatal("Search for 'go', 'fmt' should return docs/go/stdlib.md")
	}
	found := false
	for _, doc := range results {
		if doc.Path == "docs/go/stdlib.md" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Search should include docs/go/stdlib.md")
	}
}

func TestStoreForOS(t *testing.T) {
	s, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	osDocs := s.ForOS()
	// Should return 0 or more docs filtered by runtime.GOOS
	for _, doc := range osDocs {
		if doc.OS != "" && doc.OS != s.osTag {
			t.Errorf("ForOS() returned doc with OS=%q, want %q", doc.OS, s.osTag)
		}
	}
}
```

- [ ] **Step 3: 테스트 실행하여 실패 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go test ./internal/knowledge/ -v`
Expected: FAIL — package does not exist

- [ ] **Step 4: store.go 구현**

`internal/knowledge/store.go`:
```go
package knowledge

import (
	"embed"
	"encoding/json"
	"io/fs"
	"runtime"
	"strings"
)

//go:embed all:knowledge
var knowledgeFS embed.FS

// Document represents a single knowledge document.
type Document struct {
	Path     string   `json:"path"`
	Title    string   `json:"title"`
	Content  string   `json:"-"`
	Tier     int      `json:"tier"`
	OS       string   `json:"os"`
	Keywords []string `json:"keywords"`
	Tokens   int      `json:"tokens"`
}

// IndexFile is the structure of index.json.
type IndexFile struct {
	Documents []Document `json:"documents"`
}

// Store holds all embedded knowledge documents with a keyword index.
type Store struct {
	docs    []Document
	index   map[string][]int // keyword → doc indices
	osTag   string           // runtime.GOOS
}

// NewStore loads documents from embedded FS and builds the keyword index.
func NewStore() (*Store, error) {
	s := &Store{
		index: make(map[string][]int),
		osTag: runtime.GOOS,
	}

	// Try loading index.json for keyword metadata
	indexData, err := fs.ReadFile(knowledgeFS, "knowledge/index.json")
	indexed := make(map[string]Document)
	if err == nil {
		var idx IndexFile
		if jsonErr := json.Unmarshal(indexData, &idx); jsonErr == nil {
			for _, doc := range idx.Documents {
				indexed[doc.Path] = doc
			}
		}
	}

	// Walk all .md files in knowledge/
	err = fs.WalkDir(knowledgeFS, "knowledge", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}

		content, readErr := fs.ReadFile(knowledgeFS, path)
		if readErr != nil {
			return nil // skip unreadable files
		}

		// Strip "knowledge/" prefix for consistent paths
		relPath := strings.TrimPrefix(path, "knowledge/")

		doc := Document{
			Path:    relPath,
			Content: string(content),
			Tokens:  estimateTokens(string(content)),
		}

		// Extract title from first # heading
		for _, line := range strings.SplitN(string(content), "\n", 10) {
			if strings.HasPrefix(line, "# ") {
				doc.Title = strings.TrimPrefix(line, "# ")
				break
			}
		}

		// Merge metadata from index.json if available
		if meta, ok := indexed[relPath]; ok {
			doc.Tier = meta.Tier
			doc.OS = meta.OS
			doc.Keywords = meta.Keywords
		} else {
			// Infer from path
			doc.Tier, doc.OS, doc.Keywords = inferMetadata(relPath)
		}

		s.docs = append(s.docs, doc)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Build keyword index
	for i, doc := range s.docs {
		for _, kw := range doc.Keywords {
			lower := strings.ToLower(kw)
			s.index[lower] = append(s.index[lower], i)
		}
	}

	return s, nil
}

// Search returns documents matching any keyword, sorted by tier (lower first),
// trimmed to fit within the token budget.
func (s *Store) Search(keywords []string, budget int) []Document {
	scored := make(map[int]int) // doc index → match count
	for _, kw := range keywords {
		lower := strings.ToLower(kw)
		if indices, ok := s.index[lower]; ok {
			for _, idx := range indices {
				scored[idx]++
			}
		}
	}

	if len(scored) == 0 {
		return nil
	}

	// Sort by: tier ASC, then match count DESC
	type entry struct {
		idx   int
		score int
	}
	var entries []entry
	for idx, score := range scored {
		entries = append(entries, entry{idx, score})
	}
	sortEntries(entries, s.docs)

	// Collect within budget
	var result []Document
	used := 0
	for _, e := range entries {
		doc := s.docs[e.idx]
		if used+doc.Tokens > budget {
			continue
		}
		used += doc.Tokens
		result = append(result, doc)
	}
	return result
}

// ForOS returns terminal docs matching the current OS.
func (s *Store) ForOS() []Document {
	var result []Document
	for _, doc := range s.docs {
		if doc.OS == "" || doc.OS == s.osTag {
			if strings.HasPrefix(doc.Path, "docs/terminal/") {
				result = append(result, doc)
			}
		}
	}
	return result
}

// sortEntries sorts by tier ASC, then score DESC (simple insertion sort, small N).
func sortEntries(entries []entry, docs []Document) {
	for i := 1; i < len(entries); i++ {
		key := entries[i]
		j := i - 1
		for j >= 0 && (docs[entries[j].idx].Tier > docs[key.idx].Tier ||
			(docs[entries[j].idx].Tier == docs[key.idx].Tier && entries[j].score < key.score)) {
			entries[j+1] = entries[j]
			j--
		}
		entries[j+1] = key
	}
}

// estimateTokens gives a rough token count (~4 chars per token for mixed content).
func estimateTokens(s string) int {
	return len(s) / 4
}

// inferMetadata derives tier, os, and keywords from the file path.
func inferMetadata(path string) (tier int, osTag string, keywords []string) {
	parts := strings.Split(path, "/")

	// Determine category from path
	if len(parts) >= 2 {
		category := parts[1] // "bxm", "go", "css", etc.
		keywords = append(keywords, category)

		switch category {
		case "bxm":
			tier = 0
		case "go", "javascript", "typescript", "react", "css":
			tier = 1
		case "charts":
			tier = 1
		case "vue", "java":
			tier = 2
		case "python":
			tier = 3
		case "terminal":
			// Infer OS from filename
			if len(parts) >= 3 {
				fname := strings.TrimSuffix(parts[2], ".md")
				switch fname {
				case "windows":
					osTag = "windows"
				case "linux":
					osTag = "linux"
				case "macos":
					osTag = "darwin"
				}
				keywords = append(keywords, fname)
			}
		}
	}

	// Add filename as keyword
	if len(parts) > 0 {
		fname := strings.TrimSuffix(parts[len(parts)-1], ".md")
		keywords = append(keywords, fname)
	}

	// Skills tier
	if strings.HasPrefix(path, "skills/") {
		tier = 1
	}

	return
}
```

주의: `//go:embed` 지시문은 `knowledge/` 디렉토리가 **패키지 디렉토리 기준 상대경로**에 있어야 합니다. `internal/knowledge/store.go`에서 `knowledge/`를 embed하려면 심볼릭 링크가 필요하거나, embed를 루트 패키지에 두고 전달해야 합니다. 이를 해결하기 위해:

**embed를 cmd/tgc/main.go 또는 별도 루트 레벨 패키지에 배치합니다:**

`knowledge.go` (프로젝트 루트 `github.com/kimjiwon/tgc` 패키지):
```go
package tgc

import "embed"

//go:embed all:knowledge
var KnowledgeFS embed.FS
```

그리고 `internal/knowledge/store.go`에서:
```go
package knowledge

import (
	"io/fs"
	// ...
)

// NewStore는 외부에서 embed.FS를 받아서 초기화
func NewStore(fsys fs.FS) (*Store, error) {
	// fsys는 knowledge/ 하위를 가리킴
	// ...
}
```

`internal/app/app.go`에서:
```go
import tgc "github.com/kimjiwon/tgc"

knowledgeStore, _ := knowledge.NewStore(tgc.KnowledgeFS)
```

- [ ] **Step 5: 테스트 실행하여 통과 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go test ./internal/knowledge/ -v`
Expected: PASS (3 tests)

- [ ] **Step 6: 빌드 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go build -o /dev/null ./cmd/tgc/`
Expected: 정상 빌드

- [ ] **Step 7: 커밋**

```bash
git add knowledge.go internal/knowledge/store.go internal/knowledge/store_test.go knowledge/
git commit -m "feat(knowledge): add embedded knowledge store with keyword index

Loads .md files from embedded knowledge/ directory, builds keyword
index from index.json or path inference, searches by keyword within
token budget. Tier-based priority (0=BXM, 1=daily, 2=frequent, 3=ref).

Constraint: embed.FS at root package level (Go embed path limitation)
Confidence: high
Scope-risk: narrow

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: 키워드 추출기 (extractor.go)

**Files:**
- Create: `internal/knowledge/extractor.go`
- Create: `internal/knowledge/extractor_test.go`

- [ ] **Step 1: 실패 테스트 작성**

`internal/knowledge/extractor_test.go`:
```go
package knowledge

import (
	"testing"
)

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		query    string
		expected []string
	}{
		{
			query:    "tailwind로 반응형 카드 만들어줘",
			expected: []string{"tailwind", "반응형"},
		},
		{
			query:    "BXM Bean에서 다건 조회 패턴 알려줘",
			expected: []string{"bxm", "bean", "다건", "조회"},
		},
		{
			query:    "recharts로 bar chart 그려줘",
			expected: []string{"recharts", "chart"},
		},
		{
			query:    "Go에서 goroutine 사용법",
			expected: []string{"go", "goroutine"},
		},
		{
			query:    "IP 주소 확인하는 법",
			expected: []string{"ip"},
		},
		{
			query:    "React useState 사용법",
			expected: []string{"react", "usestate"},
		},
		{
			query:    "Spring Boot REST API 만들기",
			expected: []string{"spring", "rest", "api"},
		},
		{
			query:    "Vue 3 composition API 예제",
			expected: []string{"vue", "composition"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := ExtractKeywords(tt.query)
			for _, want := range tt.expected {
				found := false
				for _, g := range got {
					if g == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ExtractKeywords(%q) missing %q, got %v", tt.query, want, got)
				}
			}
		})
	}
}
```

- [ ] **Step 2: 테스트 실행하여 실패 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go test ./internal/knowledge/ -run TestExtractKeywords -v`
Expected: FAIL — ExtractKeywords undefined

- [ ] **Step 3: extractor.go 구현**

`internal/knowledge/extractor.go`:
```go
package knowledge

import (
	"strings"
	"unicode"
)

// techDictionary maps tech terms (lowercase) → canonical keyword.
// Includes Korean aliases.
var techDictionary = map[string]string{
	// BXM (Tier 0)
	"bxm": "bxm", "뱅크웨어": "bxm", "bankware": "bxm",
	"bean": "bean", "빈": "bean", "@bxmbean": "bean",
	"dbio": "dbio",
	"service": "service", "서비스": "service",
	"centercut": "centercut", "센터컷": "centercut",
	"io": "io", "dto": "dto",

	// Go (Tier 1)
	"go": "go", "golang": "go",
	"goroutine": "goroutine", "고루틴": "goroutine",
	"channel": "channel", "채널": "channel",
	"concurrency": "concurrency", "동시성": "concurrency",

	// JavaScript/TypeScript (Tier 1)
	"javascript": "javascript", "js": "javascript",
	"typescript": "typescript", "ts": "typescript",
	"node": "node", "nodejs": "node", "node.js": "node",
	"npm": "node", "npx": "node",

	// React (Tier 1)
	"react": "react", "리액트": "react",
	"usestate": "react", "useeffect": "react", "usememo": "react",
	"useref": "react", "usecallback": "react", "usetransition": "react",
	"nextjs": "nextjs", "next.js": "nextjs", "next": "nextjs",
	"app router": "nextjs", "rsc": "nextjs",

	// CSS (Tier 1)
	"tailwind": "tailwind", "tw": "tailwind", "테일윈드": "tailwind",
	"shadcn": "shadcn", "shadcn/ui": "shadcn",
	"bootstrap": "bootstrap", "부트스트랩": "bootstrap",
	"반응형": "responsive", "responsive": "responsive",
	"다크모드": "darkmode", "dark mode": "darkmode",
	"css": "css",

	// Charts (Tier 1-2)
	"recharts": "recharts",
	"chart.js": "chartjs", "chartjs": "chartjs",
	"chart": "chart", "차트": "chart", "그래프": "chart",
	"d3": "d3", "d3.js": "d3",
	"echarts": "echarts", "apache echarts": "echarts",
	"nivo": "nivo",
	"tremor": "tremor",

	// Vue (Tier 2)
	"vue": "vue", "뷰": "vue", "vue3": "vue", "vue.js": "vue",
	"composition": "composition",
	"nuxt": "nuxt", "nuxt.js": "nuxt",
	"pinia": "pinia",

	// Java (Tier 2)
	"java": "java", "자바": "java",
	"spring": "spring", "스프링": "spring",
	"spring boot": "spring", "springboot": "spring",
	"maven": "build", "gradle": "build",

	// Python (Tier 3)
	"python": "python", "파이썬": "python",
	"fastapi": "fastapi",
	"django": "django",

	// Terminal/OS
	"ip": "ip", "아이피": "ip",
	"terminal": "terminal", "터미널": "terminal",
	"powershell": "windows", "cmd": "windows",
	"bash": "linux", "shell": "terminal",
	"git": "git", "깃": "git",

	// Tools
	"vite": "vite", "docker": "docker", "sql": "sql",
	"rest": "rest", "api": "api",

	// Skills
	"tdd": "tdd", "테스트": "tdd",
	"debug": "debugging", "디버깅": "debugging", "디버그": "debugging",
	"review": "code-review", "리뷰": "code-review", "코드리뷰": "code-review",
	"refactor": "refactoring", "리팩토링": "refactoring",
	"security": "security", "보안": "security",

	// Patterns
	"다건": "다건", "paging": "paging", "페이징": "paging",
	"조회": "조회", "select": "select",
	"트랜잭션": "transaction", "transaction": "transaction",
	"배치": "batch", "batch": "batch",
	"예외": "exception", "exception": "exception",
}

// ExtractKeywords extracts technical keywords from a user query.
// Returns deduplicated, lowercased canonical keywords.
func ExtractKeywords(query string) []string {
	lower := strings.ToLower(query)
	seen := make(map[string]bool)
	var result []string

	add := func(kw string) {
		if !seen[kw] {
			seen[kw] = true
			result = append(result, kw)
		}
	}

	// Phase 1: Multi-word matches (longest first)
	multiWord := []string{
		"spring boot", "app router", "vue.js", "next.js",
		"node.js", "chart.js", "d3.js", "dark mode",
		"shadcn/ui", "apache echarts", "@bxmbean",
	}
	for _, mw := range multiWord {
		if strings.Contains(lower, mw) {
			if canonical, ok := techDictionary[mw]; ok {
				add(canonical)
			}
		}
	}

	// Phase 2: Tokenize and match single words
	tokens := tokenize(lower)
	for _, tok := range tokens {
		if canonical, ok := techDictionary[tok]; ok {
			add(canonical)
		}
	}

	return result
}

// tokenize splits a string into words, handling Korean + English mixed text.
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder

	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '/' || r == '@' || r == '-' {
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()

	return tokens
}
```

- [ ] **Step 4: 테스트 실행하여 통과 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go test ./internal/knowledge/ -run TestExtractKeywords -v`
Expected: PASS (8 sub-tests)

- [ ] **Step 5: 커밋**

```bash
git add internal/knowledge/extractor.go internal/knowledge/extractor_test.go
git commit -m "feat(knowledge): add keyword extractor with tech dictionary

Dictionary-based extraction supports Korean aliases (뱅크웨어→bxm,
테일윈드→tailwind, 리액트→react). Multi-word matching for compound
terms (spring boot, app router). Returns canonical lowercase keywords.

Confidence: high
Scope-risk: narrow

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: 프롬프트 주입기 (injector.go)

**Files:**
- Create: `internal/knowledge/injector.go`
- Create: `internal/knowledge/injector_test.go`

- [ ] **Step 1: 실패 테스트 작성**

`internal/knowledge/injector_test.go`:
```go
package knowledge

import (
	"strings"
	"testing"
)

func TestInjectorInject(t *testing.T) {
	s, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	inj := NewInjector(s, 8192)

	// BXM query should return BXM docs
	result := inj.Inject(0, "BXM Bean 작성법 알려줘")
	if !strings.Contains(result, "## Knowledge Context") {
		t.Error("Inject should include Knowledge Context header")
	}
}

func TestInjectorTokenBudget(t *testing.T) {
	s, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	// Very small budget
	inj := NewInjector(s, 100)

	result := inj.Inject(0, "Go fmt 사용법")
	tokens := estimateTokens(result)
	if tokens > 150 { // some slack for headers
		t.Errorf("Inject exceeded budget: got ~%d tokens, want ≤150", tokens)
	}
}

func TestInjectorEmptyQuery(t *testing.T) {
	s, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	inj := NewInjector(s, 8192)

	// No keywords → should return empty or minimal
	result := inj.Inject(0, "안녕하세요")
	if len(result) > 100 {
		t.Errorf("Inject for generic greeting should be minimal, got %d chars", len(result))
	}
}
```

- [ ] **Step 2: 테스트 실행하여 실패 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go test ./internal/knowledge/ -run TestInjector -v`
Expected: FAIL — NewInjector undefined

- [ ] **Step 3: injector.go 구현**

`internal/knowledge/injector.go`:
```go
package knowledge

import (
	"fmt"
	"strings"
)

// Injector builds knowledge context strings for LLM system prompts.
type Injector struct {
	store       *Store
	tokenBudget int
}

// NewInjector creates an injector with the given token budget.
func NewInjector(store *Store, tokenBudget int) *Injector {
	return &Injector{store: store, tokenBudget: tokenBudget}
}

// Inject returns a knowledge context string for the given mode and user query.
// Returns empty string if no relevant documents found.
func (inj *Injector) Inject(mode int, userQuery string) string {
	keywords := ExtractKeywords(userQuery)
	if len(keywords) == 0 {
		return ""
	}

	docs := inj.store.Search(keywords, inj.tokenBudget)
	if len(docs) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n\n## Knowledge Context\n")
	b.WriteString("(아래는 질문과 관련된 레퍼런스 문서입니다. 코드 생성 시 참고하세요.)\n\n")

	usedTokens := estimateTokens(b.String())

	for _, doc := range docs {
		if usedTokens+doc.Tokens > inj.tokenBudget {
			break
		}

		section := fmt.Sprintf("### %s\n\n%s\n\n", doc.Title, doc.Content)
		b.WriteString(section)
		usedTokens += doc.Tokens
	}

	return b.String()
}
```

- [ ] **Step 4: 테스트 실행하여 통과 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go test ./internal/knowledge/ -run TestInjector -v`
Expected: PASS (3 tests)

- [ ] **Step 5: 커밋**

```bash
git add internal/knowledge/injector.go internal/knowledge/injector_test.go
git commit -m "feat(knowledge): add prompt injector with token budget management

Extracts keywords from user query, searches store, builds context
string within token budget. Tier 0 (BXM) docs get priority.
Returns empty string for unrelated queries (no wasted tokens).

Confidence: high
Scope-risk: narrow

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: app.go 통합

**Files:**
- Modify: `knowledge.go` (프로젝트 루트 — embed.FS 선언)
- Modify: `internal/app/app.go:40-50` (Model 구조체에 injector 추가)
- Modify: `internal/app/app.go:105-135` (NewModel에서 Store/Injector 초기화)
- Modify: `internal/app/app.go:652-668` (sendMessage에서 지식 주입)

- [ ] **Step 1: 루트 패키지에 embed.FS 선언**

`knowledge.go` (프로젝트 루트):
```go
package tgc

import "embed"

//go:embed all:knowledge
var KnowledgeFS embed.FS
```

- [ ] **Step 2: Model 구조체에 injector 필드 추가**

`internal/app/app.go` — Model 구조체에 추가:
```go
import "github.com/kimjiwon/tgc/internal/knowledge"

// Model 구조체 내부에 추가:
knowledgeInj *knowledge.Injector
```

- [ ] **Step 3: NewModel()에서 Store/Injector 초기화**

`internal/app/app.go` — `NewModel()` 내부, `projectCtx += llm.GatherSystemContext()` 이후:
```go
	// Initialize knowledge store
	var knowledgeInj *knowledge.Injector
	if knowledgeStore, err := knowledge.NewStore(tgc.KnowledgeFS); err == nil {
		knowledgeInj = knowledge.NewInjector(knowledgeStore, 8192)
		config.DebugLog("[KNOWLEDGE] loaded %d documents", len(knowledgeStore.DocCount()))
	} else {
		config.DebugLog("[KNOWLEDGE] failed to load: %v", err)
	}
```

Model 초기화에 추가:
```go
	m := Model{
		// ... 기존 필드 ...
		knowledgeInj: knowledgeInj,
	}
```

- [ ] **Step 4: sendMessage()에서 지식 주입**

`internal/app/app.go` — `sendMessage()` 함수 내, history에 user 메시지 추가 후:
```go
func (m *Model) sendMessage(input string) tea.Cmd {
	m.msgs = append(m.msgs, ui.Message{
		Role: ui.RoleUser, Content: input, Timestamp: time.Now(),
	})

	// Inject knowledge context into system prompt
	if m.knowledgeInj != nil {
		knowledgeCtx := m.knowledgeInj.Inject(m.activeTab, input)
		if knowledgeCtx != "" {
			// Update system message with knowledge context
			mode := llm.Mode(m.activeTab)
			sysPrompt := llm.SystemPrompt(mode) + m.projectCtx + knowledgeCtx
			m.history[0] = openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleSystem, Content: sysPrompt,
			}
			config.DebugLog("[KNOWLEDGE] injected %d chars for query: %s",
				len(knowledgeCtx), truncate(input, 50))
		}
	}

	m.history = append(m.history, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser, Content: input,
	})
	// ... 나머지 기존 코드 ...
}
```

헬퍼 함수 (app.go 하단):
```go
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
```

- [ ] **Step 5: 빌드 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go build -o /dev/null ./cmd/tgc/`
Expected: 정상 빌드

- [ ] **Step 6: 수동 테스트**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && make build && ./techai`
테스트:
1. "Go fmt 사용법 알려줘" → debug.log에 `[KNOWLEDGE] injected` 로그 확인
2. "안녕하세요" → `[KNOWLEDGE]` 로그 없음 확인 (키워드 없으므로 미주입)

- [ ] **Step 7: 커밋**

```bash
git add knowledge.go internal/app/app.go
git commit -m "feat(knowledge): integrate injector into app message flow

Knowledge context is dynamically injected into system prompt on each
user message. Keywords extracted → docs searched → appended to prompt.
Debug log shows injection status. No injection for generic queries.

Constraint: System prompt updated per-message (not cached) for dynamic context
Directive: knowledgeInj nil-check required — store init can fail gracefully
Confidence: high
Scope-risk: moderate

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: 인덱스 빌드 도구

**Files:**
- Create: `cmd/build-index/main.go`
- Modify: `Makefile` (build-index 타겟 추가)

- [ ] **Step 1: build-index 도구 작성**

`cmd/build-index/main.go`:
```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	Path     string   `json:"path"`
	Title    string   `json:"title"`
	Tier     int      `json:"tier"`
	OS       string   `json:"os,omitempty"`
	Keywords []string `json:"keywords"`
	Tokens   int      `json:"tokens"`
}

type IndexFile struct {
	Documents []Document `json:"documents"`
}

func main() {
	knowledgeDir := "knowledge"
	if len(os.Args) > 1 {
		knowledgeDir = os.Args[1]
	}

	var idx IndexFile

	err := filepath.Walk(knowledgeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "WARN: skip %s: %v\n", path, readErr)
			return nil
		}

		relPath := strings.TrimPrefix(path, knowledgeDir+"/")
		doc := Document{
			Path:   relPath,
			Tokens: len(content) / 4,
		}

		// Extract title
		for _, line := range strings.SplitN(string(content), "\n", 10) {
			if strings.HasPrefix(line, "# ") {
				doc.Title = strings.TrimPrefix(line, "# ")
				break
			}
		}

		// Infer tier and OS from path
		parts := strings.Split(relPath, "/")
		if len(parts) >= 2 {
			category := parts[1]
			doc.Keywords = append(doc.Keywords, category)

			switch category {
			case "bxm":
				doc.Tier = 0
			case "go", "javascript", "typescript", "react", "css":
				doc.Tier = 1
			case "charts":
				doc.Tier = 1
			case "vue", "java":
				doc.Tier = 2
			case "python":
				doc.Tier = 3
			case "terminal":
				if len(parts) >= 3 {
					fname := strings.TrimSuffix(parts[2], ".md")
					switch fname {
					case "windows":
						doc.OS = "windows"
					case "linux":
						doc.OS = "linux"
					case "macos":
						doc.OS = "darwin"
					}
					doc.Keywords = append(doc.Keywords, fname)
				}
			}
		}

		// Add filename as keyword
		if len(parts) > 0 {
			fname := strings.TrimSuffix(parts[len(parts)-1], ".md")
			doc.Keywords = append(doc.Keywords, fname)
		}

		if strings.HasPrefix(relPath, "skills/") {
			doc.Tier = 1
		}

		idx.Documents = append(idx.Documents, doc)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join(knowledgeDir, "index.json")
	data, _ := json.MarshalIndent(idx, "", "  ")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s: %d documents indexed\n", outPath, len(idx.Documents))
}
```

- [ ] **Step 2: Makefile에 타겟 추가**

`Makefile`에 추가:
```makefile
# Knowledge index
.PHONY: build-index
build-index:
	@echo "Building knowledge index..."
	go run ./cmd/build-index/
	@echo "Done."

# build depends on index
build: build-index
	go build $(LDFLAGS) -o $(BINARY) ./cmd/tgc/
```

- [ ] **Step 3: 인덱스 빌드 테스트**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go run ./cmd/build-index/`
Expected: `Generated knowledge/index.json: N documents indexed`

Run: `cat knowledge/index.json`
Expected: JSON with docs/go/stdlib.md and skills/debugging.md entries

- [ ] **Step 4: 커밋**

```bash
git add cmd/build-index/main.go Makefile knowledge/index.json
git commit -m "feat(knowledge): add index builder tool + Makefile integration

go run ./cmd/build-index/ scans knowledge/ and generates index.json
with path, title, tier, OS tag, keywords, and token estimates.
Makefile build target now depends on build-index.

Confidence: high
Scope-risk: narrow

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: BXM 문서 스크래핑 도구 + Tier 0 문서

**Files:**
- Create: `cmd/scrape-bxm/main.go`
- Create: `knowledge/docs/bxm/overview.md` (+ 12 more BXM docs)

- [ ] **Step 1: scrape-bxm 도구 작성**

`cmd/scrape-bxm/main.go`:
```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var pages = []struct {
	url    string
	output string
}{
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/concepts/overview.html", "overview.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch02/ch02_1.html", "dev-process.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch02/ch02_3.html", "dbio.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch02/ch02_4.html", "bean.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch02/ch02_5.html", "service.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_1.html", "select-multi.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_2.html", "select-paging.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_6.html", "transaction.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_7.html", "exception.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-online/ch03/ch03_8.html", "logging.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-developer-guide-batch/ch01/ch01_4.html", "batch.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-configuration-reference-server/ch04/ch04.html", "config.md"},
	{"https://swlab.bwg.co.kr/web/docs/bxm/swlab-docs-bxm/current/bxm-user-guide-studio/ch01/ch01.html", "studio.md"},
}

func main() {
	outDir := "knowledge/docs/bxm"
	os.MkdirAll(outDir, 0755)

	for _, page := range pages {
		fmt.Printf("Fetching %s → %s\n", page.url, page.output)

		resp, err := http.Get(page.url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", page.url, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		md := htmlToMarkdown(string(body))
		outPath := filepath.Join(outDir, page.output)
		if err := os.WriteFile(outPath, []byte(md), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR writing %s: %v\n", outPath, err)
		}
		fmt.Printf("  → %s (%d bytes)\n", outPath, len(md))
	}

	fmt.Println("\nDone. Run 'go run ./cmd/build-index/' to update index.json")
}

// htmlToMarkdown is a simple HTML→Markdown converter for SWLab docs.
func htmlToMarkdown(html string) string {
	// Remove script/style tags
	reScript := regexp.MustCompile(`(?s)<(script|style|nav|header|footer)[^>]*>.*?</\1>`)
	html = reScript.ReplaceAllString(html, "")

	// Extract main content (article or main tag)
	reMain := regexp.MustCompile(`(?s)<(article|main)[^>]*>(.*?)</\1>`)
	if m := reMain.FindStringSubmatch(html); len(m) > 2 {
		html = m[2]
	}

	// Convert headings
	for i := 6; i >= 1; i-- {
		re := regexp.MustCompile(fmt.Sprintf(`(?s)<h%d[^>]*>(.*?)</h%d>`, i, i))
		prefix := strings.Repeat("#", i) + " "
		html = re.ReplaceAllString(html, "\n"+prefix+"$1\n")
	}

	// Convert code blocks
	rePre := regexp.MustCompile(`(?s)<pre[^>]*><code[^>]*class="[^"]*language-(\w+)"[^>]*>(.*?)</code></pre>`)
	html = rePre.ReplaceAllString(html, "\n```$1\n$2\n```\n")
	rePre2 := regexp.MustCompile(`(?s)<pre[^>]*>(.*?)</pre>`)
	html = rePre2.ReplaceAllString(html, "\n```\n$1\n```\n")

	// Convert inline code
	reCode := regexp.MustCompile(`<code[^>]*>(.*?)</code>`)
	html = reCode.ReplaceAllString(html, "`$1`")

	// Convert lists
	reLi := regexp.MustCompile(`<li[^>]*>(.*?)</li>`)
	html = reLi.ReplaceAllString(html, "- $1")

	// Convert paragraphs
	reP := regexp.MustCompile(`<p[^>]*>(.*?)</p>`)
	html = reP.ReplaceAllString(html, "\n$1\n")

	// Convert tables (basic)
	reTd := regexp.MustCompile(`<t[dh][^>]*>(.*?)</t[dh]>`)
	html = reTd.ReplaceAllString(html, "| $1 ")
	reTr := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
	html = reTr.ReplaceAllString(html, "$1|\n")

	// Strip remaining HTML tags
	reTag := regexp.MustCompile(`<[^>]+>`)
	html = reTag.ReplaceAllString(html, "")

	// Decode HTML entities
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&#39;", "'")
	html = strings.ReplaceAll(html, "&nbsp;", " ")

	// Clean up whitespace
	reBlank := regexp.MustCompile(`\n{3,}`)
	html = reBlank.ReplaceAllString(html, "\n\n")

	return strings.TrimSpace(html)
}
```

- [ ] **Step 2: 스크래핑 실행**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go run ./cmd/scrape-bxm/`
Expected: 13 BXM .md files in `knowledge/docs/bxm/`

- [ ] **Step 3: 결과 검증 + 수동 보정**

Run: `ls -la knowledge/docs/bxm/`
Expected: 13 .md files, 각각 0이 아닌 크기

Run: `head -20 knowledge/docs/bxm/bean.md`
Expected: `# Bean 작성` 등 마크다운 형태

필요 시 수동으로 마크다운 품질 보정 (깨진 테이블, 코드 블록 등).

- [ ] **Step 4: 인덱스 재빌드**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go run ./cmd/build-index/`
Expected: `Generated knowledge/index.json: ~15 documents indexed`

- [ ] **Step 5: BXM 관련 키워드를 extractor.go에 보강**

`internal/knowledge/extractor.go`의 `techDictionary`에 BXM 문서에서 발견된 추가 용어 등록:
```go
// BXM 추가 용어 (스크래핑 후 보강)
"paging":     "paging",
"lock":       "lock",
"페이징":     "paging",
"락":         "lock",
"insert":     "insert",
"update":     "update",
"delete":     "delete",
"등록":       "insert",
"수정":       "update",
"삭제":       "delete",
"studio":     "studio",
"스튜디오":   "studio",
```

- [ ] **Step 6: 커밋**

```bash
git add cmd/scrape-bxm/ knowledge/docs/bxm/ knowledge/index.json internal/knowledge/extractor.go
git commit -m "feat(knowledge): scrape BXM docs from SWLab + update index

13 BXM reference docs scraped and converted to markdown:
overview, dev-process, dbio, bean, service, select-multi,
select-paging, transaction, exception, logging, batch, config, studio.
Keyword dictionary updated with BXM-specific terms.

Constraint: SWLab site must be accessible during scrape (one-time)
Directive: Re-run scrape-bxm if SWLab docs update
Confidence: medium (HTML→MD conversion may need manual fixes)
Scope-risk: narrow

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: Tier 1 핵심 레퍼런스 문서 작성

**Files:**
- Create: `knowledge/docs/css/tailwind-v4.md`
- Create: `knowledge/docs/css/shadcn-ui.md`
- Create: `knowledge/docs/react/hooks.md`
- Create: `knowledge/docs/react/nextjs.md`
- Create: `knowledge/docs/charts/recharts.md`
- Create: `knowledge/docs/charts/chartjs.md`
- Create: `knowledge/docs/typescript/types.md`
- Create: `knowledge/docs/javascript/es2024.md`
- Create: `knowledge/docs/terminal/windows.md`
- Create: `knowledge/docs/terminal/linux.md`
- Create: `knowledge/docs/terminal/macos.md`
- Create: `knowledge/docs/terminal/git.md`

이 태스크는 문서 **내용 작성**에 집중합니다. 각 문서는 LLM이 코드를 생성할 때 참고할 수 있는 API 레퍼런스 + 패턴 + 예제 형태입니다.

- [ ] **Step 1: CSS 문서 작성**

`knowledge/docs/css/tailwind-v4.md` (~2000자):
```markdown
# Tailwind CSS v4 Quick Reference

## Setup (v4 CSS-first)
```css
/* app.css — no tailwind.config.js needed in v4 */
@import "tailwindcss";
@theme {
  --color-primary: #3b82f6;
  --font-sans: "Inter", sans-serif;
}
```

## Layout
- `flex` `flex-col` `flex-row` `items-center` `justify-between` `gap-4`
- `grid` `grid-cols-3` `grid-rows-2` `col-span-2`
- `container` `mx-auto` `px-4`

## Spacing
- `p-{0-96}` `px-` `py-` `pt-` `pb-` `pl-` `pr-`
- `m-{0-96}` `mx-auto` `my-` `mt-` `mb-`
- `space-x-4` `space-y-2` `gap-4`

## Typography
- `text-sm` `text-base` `text-lg` `text-xl` `text-2xl` ... `text-9xl`
- `font-bold` `font-semibold` `font-medium` `font-normal`
- `text-gray-900` `text-white` `text-primary`
- `leading-tight` `tracking-wide` `truncate` `line-clamp-3`

## Colors
- `bg-{color}-{shade}` `text-{color}-{shade}` `border-{color}-{shade}`
- Shades: `50 100 200 300 400 500 600 700 800 900 950`
- Colors: `slate gray zinc neutral stone red orange amber yellow lime green emerald teal cyan sky blue indigo violet purple fuchsia pink rose`

## Borders & Shadows
- `rounded` `rounded-md` `rounded-lg` `rounded-full` `rounded-xl` `rounded-2xl`
- `border` `border-2` `border-t` `border-b`
- `shadow-sm` `shadow` `shadow-md` `shadow-lg` `shadow-xl`

## Responsive
- `sm:` (640px) `md:` (768px) `lg:` (1024px) `xl:` (1280px) `2xl:` (1536px)
- Example: `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3`

## Dark Mode
- `dark:bg-gray-900` `dark:text-white`

## Transitions
- `transition` `duration-300` `ease-in-out`
- `hover:bg-blue-600` `focus:ring-2` `active:scale-95`

## Common Patterns
```html
<!-- Card -->
<div class="rounded-xl border bg-white p-6 shadow-sm dark:bg-gray-800">
  <h3 class="text-lg font-semibold">Title</h3>
  <p class="mt-2 text-gray-600 dark:text-gray-300">Description</p>
</div>

<!-- Button -->
<button class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 transition">
  Click me
</button>

<!-- Responsive Grid -->
<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
  <!-- cards -->
</div>
```
```

`knowledge/docs/css/shadcn-ui.md` (~1500자): shadcn/ui 설치, 주요 컴포넌트 (Button, Card, Dialog, Table, Form, Input, Select, Sheet, Tabs), 사용 패턴.

- [ ] **Step 2: React 문서 작성**

`knowledge/docs/react/hooks.md` (~1500자): useState, useEffect, useMemo, useCallback, useRef, useTransition, use() (React 19), 각각 예제 코드.

`knowledge/docs/react/nextjs.md` (~1500자): App Router, layout.tsx, page.tsx, loading.tsx, error.tsx, route.ts (API), Server Components vs Client Components, 미들웨어.

- [ ] **Step 3: Charts 문서 작성**

`knowledge/docs/charts/recharts.md` (~1200자): BarChart, LineChart, PieChart, AreaChart, ComposedChart 예제, ResponsiveContainer 패턴, 커스텀 tooltip.

`knowledge/docs/charts/chartjs.md` (~1200자): Chart.js 기본 (bar, line, pie, doughnut, radar), React wrapper (react-chartjs-2), 옵션 설정.

- [ ] **Step 4: TypeScript/JavaScript 문서 작성**

`knowledge/docs/typescript/types.md` (~1200자): 기본 타입, 유니온, 인터섹션, 제네릭, Utility Types (Partial, Pick, Omit, Record, Required, Readonly, ReturnType, Parameters).

`knowledge/docs/javascript/es2024.md` (~1000자): ES2024+ 주요 문법 (structuredClone, Array.groupBy, Promise.withResolvers, Temporal API, 데코레이터).

- [ ] **Step 5: Terminal 문서 작성 (OS별)**

`knowledge/docs/terminal/windows.md` (~1000자): PowerShell 명령어 (Get-ChildItem, Get-Content, Set-Content, Get-NetIPAddress, Test-NetConnection, Get-Process, Invoke-WebRequest), cmd 대응표.

`knowledge/docs/terminal/linux.md` (~1000자): bash 명령어 (ls, cat, grep, find, ip addr, ss, ps, curl, systemctl, journalctl).

`knowledge/docs/terminal/macos.md` (~800자): macOS 특화 (brew, open, pbcopy/pbpaste, defaults, networksetup, dscacheutil, mdfind).

`knowledge/docs/terminal/git.md` (~1000자): Git 핵심 (init, clone, branch, checkout, merge, rebase, stash, log, diff, reset, cherry-pick).

- [ ] **Step 6: 인덱스 재빌드 + 빌드 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go run ./cmd/build-index/`
Run: `go build -o /dev/null ./cmd/tgc/`
Expected: 정상

- [ ] **Step 7: 커밋**

```bash
git add knowledge/docs/ knowledge/index.json
git commit -m "feat(knowledge): add Tier 1 reference docs (CSS, React, Charts, TS, Terminal)

12 reference documents covering daily-use technologies:
- CSS: Tailwind v4, shadcn/ui
- React: hooks, Next.js App Router
- Charts: Recharts, Chart.js
- TypeScript: type system, utility types
- JavaScript: ES2024+ features
- Terminal: Windows, Linux, macOS, Git

Each doc is LLM-optimized: API reference + code patterns + examples.

Confidence: high
Scope-risk: narrow

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: Tier 2-3 + Skills 문서 작성

**Files:**
- Create: `knowledge/docs/vue/composition.md`
- Create: `knowledge/docs/java/spring.md`
- Create: `knowledge/docs/charts/d3.md`
- Create: `knowledge/docs/charts/echarts.md`
- Create: `knowledge/docs/python/fastapi.md`
- Create: `knowledge/skills/tdd.md` (이미 있는 debugging.md 외 추가)
- Create: `knowledge/skills/code-review.md`
- Create: `knowledge/skills/git-workflow.md`
- Create: `knowledge/skills/refactoring.md`
- Create: `knowledge/skills/performance.md`
- Create: `knowledge/skills/security.md`

- [ ] **Step 1: Tier 2 문서 작성**

`knowledge/docs/vue/composition.md` (~1200자): Vue 3 Composition API (ref, reactive, computed, watch, onMounted, defineProps, defineEmits).

`knowledge/docs/java/spring.md` (~1200자): Spring Boot (RestController, Service, Repository, JPA Entity, application.yml, 예외 처리).

`knowledge/docs/charts/d3.md` (~1000자): D3.js 핵심 (select, data, enter/exit, scales, axes, transitions).

`knowledge/docs/charts/echarts.md` (~1000자): Apache ECharts (option 구조, series types, responsive, 대규모 데이터).

- [ ] **Step 2: Tier 3 문서 작성**

`knowledge/docs/python/fastapi.md` (~1000자): FastAPI (라우터, Pydantic 모델, 의존성 주입, 비동기, OpenAPI).

- [ ] **Step 3: Skills 문서 작성**

`knowledge/skills/tdd.md` (~800자): Red-Green-Refactor, 테스트 피라미드, 테스트 네이밍.

`knowledge/skills/code-review.md` (~800자): 체크리스트 (로직, 보안, 성능, 가독성, 테스트 커버리지).

`knowledge/skills/git-workflow.md` (~800자): 브랜치 전략, 커밋 컨벤션 (feat/fix/docs/refactor), PR 템플릿.

`knowledge/skills/refactoring.md` (~800자): Extract Method, Replace Conditional with Polymorphism, 코드 스멜 목록.

`knowledge/skills/performance.md` (~800자): 웹 성능 (Core Web Vitals, lazy loading, 번들 사이즈), 서버 성능 (N+1, 인덱스, 캐싱).

`knowledge/skills/security.md` (~800자): OWASP Top 10 요약, SQL Injection, XSS, CSRF, 인증/인가 패턴.

- [ ] **Step 4: 인덱스 재빌드 + 전체 빌드 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go run ./cmd/build-index/ && go build -o /dev/null ./cmd/tgc/`
Expected: 정상, index.json에 ~40+ 문서

- [ ] **Step 5: 커밋**

```bash
git add knowledge/ knowledge/index.json
git commit -m "feat(knowledge): add Tier 2-3 docs + skill templates

Tier 2: Vue Composition API, Spring Boot, D3.js, ECharts
Tier 3: FastAPI
Skills: TDD, code review, git workflow, refactoring, performance, security

Total knowledge base: ~40+ documents, ~9MB embedded.

Confidence: high
Scope-risk: narrow

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 9: 통합 테스트 + 바이너리 검증

**Files:**
- No new files — validation only

- [ ] **Step 1: 전체 테스트 실행**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && go test ./internal/knowledge/ -v`
Expected: ALL PASS

- [ ] **Step 2: 프로덕션 빌드**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && make build`
Expected: 정상 빌드, `./techai` 바이너리 생성

- [ ] **Step 3: 바이너리 사이즈 확인**

Run: `ls -lh techai`
Expected: ~20-25MB (기존 ~15MB + knowledge ~9MB)

- [ ] **Step 4: 수동 E2E 테스트**

Run: `./techai`

테스트 시나리오:
1. "BXM에서 다건 조회 DBIO 작성법 알려줘" → debug.log에 BXM 문서 주입 확인
2. "tailwind로 반응형 카드 컴포넌트 만들어줘" → Tailwind 문서 주입 확인
3. "recharts로 라인 차트 그리는 코드" → Recharts 문서 주입 확인
4. "IP 주소 확인하는 방법" → 현재 OS에 맞는 터미널 문서 주입 확인
5. "안녕하세요" → 문서 주입 없음 확인

각 테스트 후 `~/.tgc/debug.log`에서 `[KNOWLEDGE]` 로그 확인.

- [ ] **Step 5: 크로스 플랫폼 빌드 확인**

Run: `cd /Users/kimjiwon/Desktop/kimjiwon/택가이코드 && make build-all`
Expected: 5개 바이너리 생성 (darwin-arm64, darwin-amd64, linux-amd64, windows-amd64, windows-arm64)

- [ ] **Step 6: 태그 생성**

Run:
```bash
git tag -a v0.5.0 -m "feat: Embedded Knowledge Store — 폐쇄망 LLM 지식 내장

BXM(Tier 0), Tailwind/React/Charts(Tier 1), Vue/Java(Tier 2),
Python(Tier 3), Skills, OS별 터미널 문서 ~40+ 내장.
키워드 매칭 + 모드 하이브리드 컨텍스트 주입."
```
