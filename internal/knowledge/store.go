// Package knowledge provides an embedded knowledge store that loads .md
// documents from an fs.FS (backed by embed.FS at the root package level),
// builds a keyword index, and supports search by keyword within a token budget.
package knowledge

import (
	"encoding/json"
	"io/fs"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Tier represents document priority. Lower is higher priority.
const (
	Tier0 = 0 // BXM (product-specific, always injected)
	Tier1 = 1 // Daily use: Go, JS, TS, React, CSS, Charts, Skills
	Tier2 = 2 // Frequent: Vue, Java
	Tier3 = 3 // Reference: Python
)

// Doc represents a single knowledge document.
type Doc struct {
	Path     string   // e.g. "knowledge/docs/go/stdlib.md"
	Content  string   // raw markdown content
	Tier     int      // priority tier (0-3)
	OS       string   // empty = all OS, or "windows"/"linux"/"darwin"
	Keywords []string // search keywords
}

// IndexEntry represents a document entry in index.json.
type IndexEntry struct {
	Path     string   `json:"path"`
	Tier     int      `json:"tier"`
	OS       string   `json:"os,omitempty"`
	Keywords []string `json:"keywords"`
}

// Store holds all loaded knowledge documents and a keyword index.
type Store struct {
	docs     []Doc
	kwIndex  map[string][]*Doc // keyword -> matching docs
}

// NewStore creates a new Store by walking the given fs.FS under
// the "knowledge/" prefix, loading all .md files, and building
// a keyword index from index.json (if present) or inferred from path.
// If allowedPacks is non-empty, only documents under those subdirectories
// are loaded (e.g. ["react", "database"] loads knowledge/docs/react/* and
// knowledge/docs/database/*). An empty slice loads everything.
func NewStore(fsys fs.FS, allowedPacks ...string) (*Store, error) {
	s := &Store{
		kwIndex: make(map[string][]*Doc),
	}

	// Build pack filter set
	packFilter := make(map[string]bool)
	for _, p := range allowedPacks {
		packFilter[strings.ToLower(strings.TrimSpace(p))] = true
	}

	// Try to load index.json for explicit metadata
	indexMap := make(map[string]IndexEntry)
	if data, err := fs.ReadFile(fsys, "knowledge/index.json"); err == nil {
		var entries []IndexEntry
		if err := json.Unmarshal(data, &entries); err == nil {
			for _, e := range entries {
				indexMap[e.Path] = e
			}
		}
	}

	// Walk all .md files under knowledge/
	err := fs.WalkDir(fsys, "knowledge", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Pack filter: check if this doc belongs to an allowed pack.
		// Path format: "knowledge/docs/<pack>/file.md" or "knowledge/skills/<pack>/file.md"
		if len(packFilter) > 0 {
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				pack := strings.ToLower(parts[2]) // e.g. "react", "database"
				if !packFilter[pack] {
					return nil // skip this doc
				}
			}
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		doc := Doc{
			Path:    path,
			Content: string(data),
		}

		// Use index.json metadata if available, otherwise infer
		if entry, ok := indexMap[path]; ok {
			doc.Tier = entry.Tier
			doc.OS = entry.OS
			doc.Keywords = entry.Keywords
		} else {
			doc.Tier, doc.OS, doc.Keywords = inferMetadata(path)
		}

		s.docs = append(s.docs, doc)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Build keyword index (references into s.docs slice)
	for i := range s.docs {
		for _, kw := range s.docs[i].Keywords {
			lkw := strings.ToLower(kw)
			s.kwIndex[lkw] = append(s.kwIndex[lkw], &s.docs[i])
		}
	}

	return s, nil
}

// DocCount returns the number of loaded documents.
func (s *Store) DocCount() int {
	return len(s.docs)
}

// Search returns documents matching any of the given keywords, sorted by
// tier ASC then match count DESC, trimmed to fit within the token budget.
func (s *Store) Search(keywords []string, budget int) []Doc {
	if len(keywords) == 0 || budget <= 0 {
		return nil
	}

	// Count matches per doc
	type scored struct {
		doc   *Doc
		count int
	}
	seen := make(map[*Doc]*scored)

	for _, kw := range keywords {
		lkw := strings.ToLower(kw)
		for _, doc := range s.kwIndex[lkw] {
			if sc, ok := seen[doc]; ok {
				sc.count++
			} else {
				seen[doc] = &scored{doc: doc, count: 1}
			}
		}
	}

	if len(seen) == 0 {
		return nil
	}

	// Collect and sort: tier ASC, match count DESC
	results := make([]*scored, 0, len(seen))
	for _, sc := range seen {
		results = append(results, sc)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].doc.Tier != results[j].doc.Tier {
			return results[i].doc.Tier < results[j].doc.Tier
		}
		return results[i].count > results[j].count
	})

	// Collect docs within token budget
	var out []Doc
	remaining := budget
	for _, sc := range results {
		tokens := estimateTokens(sc.doc.Content)
		if tokens > remaining {
			continue
		}
		out = append(out, *sc.doc)
		remaining -= tokens
	}

	return out
}

// BM25Search performs a content-based search across all documents using
// simplified BM25 scoring. Used as Level 2 fallback when keyword index
// search returns insufficient results.
//
// Parameters:
//   - query: raw user query string (will be tokenized and lowercased)
//   - budget: maximum total token budget for returned documents
//   - exclude: paths to exclude (already found by Level 1)
//
// Returns documents sorted by BM25 score descending, within budget.
func (s *Store) BM25Search(query string, budget int, exclude map[string]bool) []Doc {
	if query == "" || budget <= 0 {
		return nil
	}

	// Tokenize query into search terms (3+ chars to avoid noise)
	tokens := bm25Tokenize(strings.ToLower(query))
	if len(tokens) == 0 {
		return nil
	}

	// Pre-compute IDF: log(N / df) for each term
	N := float64(len(s.docs))
	if N == 0 {
		return nil
	}

	df := make(map[string]int) // document frequency per term
	for i := range s.docs {
		content := strings.ToLower(s.docs[i].Content)
		for _, tok := range tokens {
			if strings.Contains(content, tok) {
				df[tok]++
			}
		}
	}

	// BM25 parameters
	const k1 = 1.2
	const b = 0.75

	// Compute average document length
	var totalLen float64
	for i := range s.docs {
		totalLen += float64(len(s.docs[i].Content))
	}
	avgDL := totalLen / N

	type scored struct {
		doc   *Doc
		score float64
	}
	var results []scored

	for i := range s.docs {
		if exclude != nil && exclude[s.docs[i].Path] {
			continue
		}

		content := strings.ToLower(s.docs[i].Content)
		dl := float64(len(content))
		score := 0.0

		for _, tok := range tokens {
			termDF := df[tok]
			if termDF == 0 {
				continue
			}

			// Term frequency in this document
			tf := float64(strings.Count(content, tok))
			if tf == 0 {
				continue
			}

			// IDF: log((N - df + 0.5) / (df + 0.5))
			idf := math.Log((N - float64(termDF) + 0.5) / (float64(termDF) + 0.5))
			if idf < 0 {
				idf = 0.01 // floor for very common terms
			}

			// BM25 score component
			tfNorm := (tf * (k1 + 1)) / (tf + k1*(1-b+b*(dl/avgDL)))
			score += idf * tfNorm
		}

		if score > 0 {
			results = append(results, scored{&s.docs[i], score})
		}
	}

	if len(results) == 0 {
		return nil
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Collect within budget
	var out []Doc
	remaining := budget
	for _, sc := range results {
		tokens := estimateTokens(sc.doc.Content)
		if tokens > remaining {
			continue
		}
		out = append(out, *sc.doc)
		remaining -= tokens
	}

	return out
}

// bm25Tokenize splits a query into meaningful tokens for BM25 search.
// Filters out tokens shorter than 2 characters to reduce noise.
func bm25Tokenize(s string) []string {
	raw := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '-' && r != '_'
	})
	var out []string
	seen := make(map[string]bool)
	for _, tok := range raw {
		tok = strings.ToLower(tok)
		if len(tok) >= 2 && !seen[tok] {
			seen[tok] = true
			out = append(out, tok)
		}
	}
	return out
}

// DocList returns a formatted list of all documents with index, title, and keywords.
// Used by Level 3 (LLM fallback) to present document choices to the LLM.
func (s *Store) DocList() []DocInfo {
	infos := make([]DocInfo, len(s.docs))
	for i, doc := range s.docs {
		title := ""
		lines := strings.SplitN(doc.Content, "\n", 10)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "# ") {
				title = strings.TrimPrefix(trimmed, "# ")
				break
			}
		}
		if title == "" {
			base := filepath.Base(doc.Path)
			title = strings.TrimSuffix(base, filepath.Ext(base))
		}
		infos[i] = DocInfo{
			Index:    i,
			Path:     doc.Path,
			Title:    title,
			Keywords: doc.Keywords,
		}
	}
	return infos
}

// DocByIndex returns a document by its index in the store.
func (s *Store) DocByIndex(idx int) *Doc {
	if idx < 0 || idx >= len(s.docs) {
		return nil
	}
	return &s.docs[idx]
}

// DocInfo holds lightweight document metadata for LLM selection.
type DocInfo struct {
	Index    int
	Path     string
	Title    string
	Keywords []string
}

// ForOS returns all documents that are either OS-agnostic (OS=="") or
// match the given operating system string (e.g. "darwin", "linux", "windows").
func (s *Store) ForOS(goos string) []Doc {
	var out []Doc
	for _, doc := range s.docs {
		if doc.OS == "" || doc.OS == goos {
			out = append(out, doc)
		}
	}
	return out
}

// estimateTokens approximates token count using ~4 chars per token.
func estimateTokens(content string) int {
	n := len(content)
	if n == 0 {
		return 0
	}
	return (n + 3) / 4 // ceiling division
}

// inferMetadata determines tier, OS, and keywords from the file path.
// Path format: "knowledge/{category}/{subcategory}/filename.md"
func inferMetadata(path string) (tier int, osName string, keywords []string) {
	// Normalize path separators
	p := filepath.ToSlash(path)

	// Strip "knowledge/" prefix
	p = strings.TrimPrefix(p, "knowledge/")

	parts := strings.Split(p, "/")
	if len(parts) == 0 {
		return Tier3, "", nil
	}

	category := parts[0] // "docs" or "skills"
	var subcategory string
	var filename string

	if len(parts) >= 3 {
		subcategory = parts[1]
		filename = parts[len(parts)-1]
	} else if len(parts) == 2 {
		subcategory = ""
		filename = parts[1]
	} else {
		filename = parts[0]
	}

	// Remove .md extension for keyword extraction
	name := strings.TrimSuffix(filename, ".md")

	// Determine tier based on category/subcategory
	switch category {
	case "docs":
		switch subcategory {
		case "bxm":
			tier = Tier0
		case "go", "javascript", "typescript", "react", "css", "charts":
			tier = Tier1
		case "vue", "java":
			tier = Tier2
		case "python":
			tier = Tier3
		case "terminal":
			// OS-specific documents
			tier = Tier1
			switch name {
			case "windows":
				osName = "windows"
			case "linux":
				osName = "linux"
			case "macos":
				osName = "darwin"
			}
		default:
			tier = Tier2
		}
	case "skills":
		tier = Tier1
	default:
		tier = Tier3
	}

	// Build keywords from subcategory and filename
	if subcategory != "" {
		keywords = append(keywords, strings.ToLower(subcategory))
	}
	if name != "" {
		keywords = append(keywords, strings.ToLower(name))
	}

	// Add extra keywords from filename parts (e.g. "error-handling" -> "error", "handling")
	for _, part := range strings.Split(name, "-") {
		lp := strings.ToLower(part)
		if lp != "" && !containsStr(keywords, lp) {
			keywords = append(keywords, lp)
		}
	}

	return tier, osName, keywords
}

// containsStr checks if a string slice contains a value.
func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
