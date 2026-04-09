package knowledge

import (
	"path/filepath"
	"strings"
)

// Injector builds knowledge context strings for LLM system prompts.
type Injector struct {
	store       *Store
	tokenBudget int
}

// NewInjector creates an injector with the given token budget.
func NewInjector(store *Store, tokenBudget int) *Injector {
	return &Injector{
		store:       store,
		tokenBudget: tokenBudget,
	}
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

	header := "\n\n## Knowledge Context\n(아래는 질문과 관련된 레퍼런스 문서입니다. 코드 생성 시 참고하세요.)\n"
	headerTokens := estimateTokens(header)

	remaining := inj.tokenBudget - headerTokens
	if remaining <= 0 {
		return ""
	}

	b.WriteString(header)

	for _, doc := range docs {
		title := extractTitle(doc)
		section := "\n### " + title + "\n\n" + doc.Content + "\n"
		sectionTokens := estimateTokens(section)

		if sectionTokens > remaining {
			continue
		}

		b.WriteString(section)
		remaining -= sectionTokens
	}

	// If only the header was written with no doc sections, return empty.
	if b.Len() == len(header) {
		return ""
	}

	return b.String()
}

// extractTitle returns the first # heading from the document content.
// If no heading is found, it returns the filename from the document path.
func extractTitle(doc Doc) string {
	lines := strings.SplitN(doc.Content, "\n", 20)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
	}
	// Fallback: use filename without extension
	base := filepath.Base(doc.Path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
