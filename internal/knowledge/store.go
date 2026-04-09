// Package knowledge provides an embedded knowledge store that loads .md
// documents from an fs.FS (backed by embed.FS at the root package level),
// builds a keyword index, and supports search by keyword within a token budget.
package knowledge

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
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
func NewStore(fsys fs.FS) (*Store, error) {
	s := &Store{
		kwIndex: make(map[string][]*Doc),
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
