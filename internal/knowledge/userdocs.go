package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// UserDoc represents one indexed markdown or text file from the user's
// knowledge folder.
type UserDoc struct {
	Path      string // relative to the knowledge root
	Title     string // first heading or filename
	Headers   []string
	FirstPara string // first ~200 chars of body
	Size      int64
	content   string // full content, lazily loaded
}

// UserDocIndex is the in-memory searchable index of .tgc/knowledge/.
type UserDocIndex struct {
	mu   sync.RWMutex
	docs []UserDoc
	root string
}

// GlobalIndex is the process-wide user-docs index. Set by the app on
// startup via ScanUserDocs() and read by tools.Execute for the
// knowledge_search tool. Nil if no knowledge folder was found.
var GlobalIndex *UserDocIndex

// knowledgeDirs returns the paths to scan for user knowledge, in
// priority order: project-local first, then global.
func knowledgeDirs() []string {
	dirs := []string{".tgc/knowledge"}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".tgc", "knowledge"))
	}
	return dirs
}

// ScanUserDocs walks the knowledge directories and indexes every
// .md / .txt file found. Call on startup and on /knowledge reload.
func ScanUserDocs() *UserDocIndex {
	idx := &UserDocIndex{}
	for _, dir := range knowledgeDirs() {
		abs, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			continue
		}
		idx.root = abs
		_ = filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".md" && ext != ".txt" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(abs, path)
			doc := parseUserDoc(rel, string(data))
			if fi, e := d.Info(); e == nil {
				doc.Size = fi.Size()
			} else {
				doc.Size = int64(len(data))
			}
			idx.docs = append(idx.docs, doc)
			return nil
		})
		break // use first found knowledge dir
	}
	return idx
}

// parseUserDoc extracts title, headers, and first paragraph from raw
// markdown/text content for the index.
func parseUserDoc(path, content string) UserDoc {
	doc := UserDoc{
		Path:    path,
		content: content,
	}
	lines := strings.Split(content, "\n")
	var headers []string
	firstParaDone := false
	var firstPara strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			h := strings.TrimLeft(trimmed, "# ")
			headers = append(headers, h)
			if doc.Title == "" {
				doc.Title = h
			}
		} else if !firstParaDone && trimmed != "" {
			if firstPara.Len() > 0 {
				firstPara.WriteString(" ")
			}
			firstPara.WriteString(trimmed)
			if firstPara.Len() > 200 {
				firstParaDone = true
			}
		} else if !firstParaDone && trimmed == "" && firstPara.Len() > 0 {
			firstParaDone = true
		}
	}
	if doc.Title == "" {
		doc.Title = filepath.Base(path)
	}
	doc.Headers = headers
	fp := firstPara.String()
	if len([]rune(fp)) > 200 {
		fp = string([]rune(fp)[:200]) + "..."
	}
	doc.FirstPara = fp
	return doc
}

// Count returns the number of indexed documents.
func (idx *UserDocIndex) Count() int {
	if idx == nil {
		return 0
	}
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.docs)
}

// Root returns the knowledge directory path.
func (idx *UserDocIndex) Root() string {
	if idx == nil {
		return ""
	}
	return idx.root
}

// TableOfContents returns a compact summary of all indexed docs,
// suitable for injection into the system prompt.
func (idx *UserDocIndex) TableOfContents() string {
	if idx == nil || len(idx.docs) == 0 {
		return ""
	}
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n\n## User Knowledge (%d docs)\n", len(idx.docs)))
	b.WriteString("Use `knowledge_search` tool to read these when relevant.\n\n")
	for _, d := range idx.docs {
		summary := d.Title
		if summary == filepath.Base(d.Path) && d.FirstPara != "" {
			r := []rune(d.FirstPara)
			if len(r) > 60 {
				summary = string(r[:60]) + "..."
			} else {
				summary = d.FirstPara
			}
		}
		b.WriteString(fmt.Sprintf("- %s -- %s\n", d.Path, summary))
	}
	return b.String()
}

// Search performs a simple keyword search across all indexed docs.
func (idx *UserDocIndex) Search(query string, maxResults int) []UserDoc {
	if idx == nil || len(idx.docs) == 0 {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 3
	}
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	terms := splitSearchTerms(query)
	if len(terms) == 0 {
		return nil
	}
	var results []UserDoc
	for _, doc := range idx.docs {
		haystack := strings.ToLower(doc.Path + " " + doc.Title + " " +
			strings.Join(doc.Headers, " ") + " " + doc.content)
		allMatch := true
		for _, t := range terms {
			if !strings.Contains(haystack, t) {
				allMatch = false
				break
			}
		}
		if allMatch {
			results = append(results, doc)
			if len(results) >= maxResults {
				break
			}
		}
	}
	return results
}

// ReadFull returns the full content of a doc by path.
func (idx *UserDocIndex) ReadFull(path string) (string, bool) {
	if idx == nil {
		return "", false
	}
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	for _, d := range idx.docs {
		if d.Path == path {
			return d.content, true
		}
	}
	return "", false
}

// FormatSearchResults renders search results in a tool-friendly format.
func FormatSearchResults(docs []UserDoc, query string) string {
	if len(docs) == 0 {
		return "검색 결과가 없습니다: " + query
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("knowledge_search: %d건 발견\n\n", len(docs)))
	for i, d := range docs {
		b.WriteString(fmt.Sprintf("--- [%d] %s ---\n", i+1, d.Path))
		b.WriteString(fmt.Sprintf("제목: %s\n", d.Title))
		if len(d.Headers) > 1 {
			b.WriteString("목차: " + strings.Join(d.Headers, " / ") + "\n")
		}
		body := d.content
		r := []rune(body)
		if len(r) > 500 {
			body = string(r[:500]) + "\n... (전문은 file_read 사용)"
		}
		b.WriteString("\n" + body + "\n\n")
	}
	return b.String()
}

// splitSearchTerms tokenizes a query into lowercase search terms.
func splitSearchTerms(query string) []string {
	raw := strings.Fields(strings.ToLower(query))
	var out []string
	for _, t := range raw {
		if len([]rune(t)) >= 2 {
			out = append(out, t)
		}
	}
	return out
}
