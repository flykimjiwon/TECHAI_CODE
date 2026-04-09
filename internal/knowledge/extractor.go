package knowledge

import (
	"strings"
	"unicode"
)

// techDictionary maps known technical terms (lowercased) to their canonical keyword.
var techDictionary = map[string]string{
	// BXM (Tier 0)
	"bxm": "bxm", "뱅크웨어": "bxm", "bankware": "bxm",
	"bean": "bean", "빈": "bean", "@bxmbean": "bean",
	"dbio": "dbio",
	"service": "service", "서비스": "service",
	"centercut": "centercut", "센터컷": "centercut",
	"io": "io", "dto": "dto",

	// Go
	"go": "go", "golang": "go",
	"goroutine": "goroutine", "고루틴": "goroutine",
	"channel": "channel", "채널": "channel",
	"concurrency": "concurrency", "동시성": "concurrency",

	// JavaScript/TypeScript
	"javascript": "javascript", "js": "javascript",
	"typescript": "typescript", "ts": "typescript",
	"node": "node", "nodejs": "node", "node.js": "node",
	"npm": "node", "npx": "node",

	// React
	"react": "react", "리액트": "react",
	"usestate": "react", "useeffect": "react", "usememo": "react",
	"useref": "react", "usecallback": "react", "usetransition": "react",
	"nextjs": "nextjs", "next.js": "nextjs", "next": "nextjs",
	"app router": "nextjs", "rsc": "nextjs",

	// CSS
	"tailwind": "tailwind", "tw": "tailwind", "테일윈드": "tailwind",
	"shadcn": "shadcn", "shadcn/ui": "shadcn",
	"bootstrap": "bootstrap", "부트스트랩": "bootstrap",
	"반응형": "responsive", "responsive": "responsive",
	"다크모드": "darkmode", "dark mode": "darkmode",
	"css": "css",

	// Charts
	"recharts": "recharts",
	"chart.js": "chartjs", "chartjs": "chartjs",
	"chart": "chart", "차트": "chart", "그래프": "chart",
	"d3": "d3", "d3.js": "d3",
	"echarts": "echarts", "apache echarts": "echarts",
	"nivo": "nivo", "tremor": "tremor",

	// Vue
	"vue": "vue", "뷰": "vue", "vue3": "vue", "vue.js": "vue",
	"composition": "composition",
	"nuxt": "nuxt", "nuxt.js": "nuxt", "pinia": "pinia",

	// Java
	"java": "java", "자바": "java",
	"spring": "spring", "스프링": "spring",
	"spring boot": "spring", "springboot": "spring",
	"maven": "build", "gradle": "build",

	// Python
	"python": "python", "파이썬": "python",
	"fastapi": "fastapi", "django": "django",

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

	// BXM patterns
	"다건": "다건", "paging": "paging", "페이징": "paging",
	"조회": "조회", "select": "select",
	"트랜잭션": "transaction", "transaction": "transaction",
	"배치": "batch", "batch": "batch",
	"예외": "exception", "exception": "exception",
	"insert": "insert", "update": "update", "delete": "delete",
	"등록": "insert", "수정": "update", "삭제": "delete",
	"studio": "studio", "스튜디오": "studio",
	"lock": "lock", "락": "lock",
}

// multiWordTerms holds compound terms sorted by descending length so that
// the longest match is attempted first. Built once at init time.
var multiWordTerms []string

func init() {
	// Collect dictionary keys that contain a space (multi-word terms).
	for k := range techDictionary {
		if strings.Contains(k, " ") {
			multiWordTerms = append(multiWordTerms, k)
		}
	}
	// Sort descending by length so longest match wins.
	sortDescByLen(multiWordTerms)
}

// sortDescByLen sorts strings by descending length, then lexicographically
// for determinism among equal-length strings.
func sortDescByLen(ss []string) {
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0; j-- {
			if len(ss[j]) > len(ss[j-1]) || (len(ss[j]) == len(ss[j-1]) && ss[j] < ss[j-1]) {
				ss[j], ss[j-1] = ss[j-1], ss[j]
			} else {
				break
			}
		}
	}
}

// ExtractKeywords takes a user query string and returns canonical technical
// keywords for searching the knowledge store.
//
// It performs two passes:
//  1. Multi-word matching (longest match wins) on the lowercased query.
//  2. Single-word tokenization on the remaining text, looking up each token
//     in the tech dictionary.
//
// Returns deduplicated, lowercased canonical keywords.
func ExtractKeywords(query string) []string {
	if query == "" {
		return nil
	}

	lower := strings.ToLower(query)
	seen := make(map[string]bool)
	var result []string

	addKeyword := func(canonical string) {
		if !seen[canonical] {
			seen[canonical] = true
			result = append(result, canonical)
		}
	}

	// Pass 1: multi-word matching (longest match wins).
	// Replace matched regions with spaces so they are not re-matched in pass 2.
	remaining := lower
	for _, term := range multiWordTerms {
		idx := strings.Index(remaining, term)
		if idx == -1 {
			continue
		}
		canonical := techDictionary[term]
		addKeyword(canonical)
		// Blank out the matched region with spaces to prevent re-matching.
		remaining = remaining[:idx] + strings.Repeat(" ", len(term)) + remaining[idx+len(term):]
	}

	// Pass 2: single-word tokenization on the remaining text.
	tokens := tokenize(remaining)
	for _, tok := range tokens {
		if canonical, ok := techDictionary[tok]; ok {
			addKeyword(canonical)
		}
	}

	return result
}

// tokenize splits s into tokens. A character belongs to the current token if
// it is a letter, digit, '.', '/', '@', or '-'. Everything else flushes the
// current token and starts a new one.
//
// Additionally, script boundaries between CJK (Hangul/Han/Katakana/Hiragana)
// and Latin/digit characters cause a token flush, so "tailwind로" becomes
// ["tailwind", "로"] rather than a single token.
func tokenize(s string) []string {
	var tokens []string
	var buf strings.Builder
	prevCJK := false
	first := true

	for _, r := range s {
		if !isTokenChar(r) {
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
			first = true
			continue
		}

		curCJK := isCJK(r)

		// Flush on script boundary (CJK <-> non-CJK) within a token.
		if !first && curCJK != prevCJK && buf.Len() > 0 {
			tokens = append(tokens, buf.String())
			buf.Reset()
		}

		buf.WriteRune(r)
		prevCJK = curCJK
		first = false
	}
	if buf.Len() > 0 {
		tokens = append(tokens, buf.String())
	}
	return tokens
}

// isTokenChar reports whether r is part of a token.
func isTokenChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '/' || r == '@' || r == '-'
}

// isCJK reports whether r is a CJK character (Hangul, Han, Katakana, Hiragana).
// Used to detect script boundaries so "tailwind로" splits into ["tailwind", "로"].
func isCJK(r rune) bool {
	return unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hiragana, r)
}
