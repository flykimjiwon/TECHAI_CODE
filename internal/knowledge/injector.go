package knowledge

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// LLMSelectorFunc is a callback that asks the LLM to select relevant documents.
// It receives a prompt string and returns the LLM's response text.
// Used by Level 3 fallback search.
type LLMSelectorFunc func(prompt string) (string, error)

// Injector builds knowledge context strings for LLM system prompts.
// It orchestrates a 3-level search pipeline:
//
//	Level 1: Keyword extraction (techDictionary + synonym expansion)
//	Level 2: BM25 content-based fallback search
//	Level 3: LLM-based document selection (optional, only when L1+L2 fail)
type Injector struct {
	store       *Store
	tokenBudget int
	llmSelector LLMSelectorFunc // optional: set via SetLLMSelector
}

// NewInjector creates an injector with the given token budget.
func NewInjector(store *Store, tokenBudget int) *Injector {
	return &Injector{
		store:       store,
		tokenBudget: tokenBudget,
	}
}

// SetLLMSelector registers the Level 3 LLM fallback callback.
// If not set, Level 3 is skipped.
func (inj *Injector) SetLLMSelector(fn LLMSelectorFunc) {
	inj.llmSelector = fn
}

// Inject returns a knowledge context string for the given mode and user query.
// Uses a 3-level search pipeline for maximum recall:
//
//	Level 1: techDictionary keywords + synonym expansion → kwIndex search
//	Level 2: BM25 full-text fallback (if L1 returns 0 results)
//	Level 3: LLM picks from doc list (if L1+L2 both return 0 results)
//
// Returns empty string if no relevant documents found at any level.
func (inj *Injector) Inject(mode int, userQuery string) string {
	// ── Level 1: Keyword-based search (techDictionary + synonyms) ──
	keywords := ExtractKeywords(userQuery)
	var docs []Doc
	if len(keywords) > 0 {
		docs = inj.store.Search(keywords, inj.tokenBudget)
	}

	// ── Level 2: BM25 content-based fallback ──
	// Triggered when Level 1 returns insufficient results (0-1 docs)
	if len(docs) <= 1 {
		// Build exclusion set from Level 1 results
		exclude := make(map[string]bool, len(docs))
		for _, d := range docs {
			exclude[d.Path] = true
		}

		// Calculate remaining budget after Level 1 docs
		usedTokens := 0
		for _, d := range docs {
			usedTokens += estimateTokens(d.Content)
		}
		remainingBudget := inj.tokenBudget - usedTokens

		bm25Docs := inj.store.BM25Search(userQuery, remainingBudget, exclude)
		docs = append(docs, bm25Docs...)
	}

	// ── Level 3: LLM-based document selection ──
	// Triggered only when L1+L2 both return 0 results AND llmSelector is set
	if len(docs) == 0 && inj.llmSelector != nil {
		docs = inj.llmFallback(userQuery)
	}

	if len(docs) == 0 {
		return ""
	}

	return inj.buildContext(docs)
}

// llmFallback asks the LLM to select relevant documents from the full list.
// This is Level 3: the most accurate but most expensive search method.
func (inj *Injector) llmFallback(userQuery string) []Doc {
	docInfos := inj.store.DocList()
	if len(docInfos) == 0 {
		return nil
	}

	// Build compact doc list for LLM (minimize tokens)
	var list strings.Builder
	for _, info := range docInfos {
		fmt.Fprintf(&list, "%d|%s|%s\n",
			info.Index, info.Title, strings.Join(info.Keywords, ","))
	}

	prompt := fmt.Sprintf(
		"사용자 질문: \"%s\"\n\n"+
			"아래 문서 목록에서 이 질문에 답하는 데 필요한 문서 번호를 최대 3개 골라주세요.\n"+
			"숫자만 쉼표로 답하세요 (예: 2,15,30). 관련 문서가 없으면 \"none\".\n\n%s",
		userQuery, list.String())

	resp, err := inj.llmSelector(prompt)
	if err != nil || resp == "" {
		return nil
	}

	// Parse response: extract numbers
	resp = strings.TrimSpace(resp)
	if strings.ToLower(resp) == "none" {
		return nil
	}

	var docs []Doc
	remaining := inj.tokenBudget
	parts := strings.Split(resp, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		idx, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		doc := inj.store.DocByIndex(idx)
		if doc == nil {
			continue
		}
		tokens := estimateTokens(doc.Content)
		if tokens > remaining {
			continue
		}
		docs = append(docs, *doc)
		remaining -= tokens
		if len(docs) >= 3 {
			break
		}
	}

	return docs
}

// buildContext formats the found documents into a knowledge context string
// suitable for injection into the LLM system prompt.
func (inj *Injector) buildContext(docs []Doc) string {
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
