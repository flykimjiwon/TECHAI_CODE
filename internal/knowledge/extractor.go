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
	// React legacy & patterns
	"class component": "react-class", "클래스컴포넌트": "react-class",
	"lifecycle": "react-class", "componentdidmount": "react-class",
	"componentwillunmount": "react-class", "setstate": "react-class",
	"hoc": "react-class", "purecomponent": "react-class",
	"compound": "react-patterns", "render props": "react-patterns",
	"컴포넌트패턴": "react-patterns", "custom hook": "react-patterns",
	"forwardref": "react-patterns", "portal": "react-patterns",
	"react.memo": "react-performance", "최적화": "react-performance",
	"리렌더링": "react-performance", "코드스플리팅": "react-performance",
	"code splitting": "react-performance", "virtualize": "react-performance",
	"lazy": "react-performance", "suspense": "react-performance",
	"pages router": "nextjs-pages", "getserversideprops": "nextjs-pages",
	"getstaticprops": "nextjs-pages", "getinitialprops": "nextjs-pages",
	"_app.tsx": "nextjs-pages", "_document.tsx": "nextjs-pages",

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

	// Java / Spring 생태계
	"java": "java", "자바": "java",
	"spring": "spring-core", "스프링": "spring-core",
	"spring core": "spring-core", "스프링코어": "spring-core",
	"di": "spring-core", "의존성주입": "spring-core", "dependency injection": "spring-core",
	"@bean": "spring-core", "@component": "spring-core", "@configuration": "spring-core",
	"spring boot": "spring-boot-ops", "springboot": "spring-boot-ops", "스프링부트": "spring-boot-ops",
	"actuator": "spring-boot-ops", "액추에이터": "spring-boot-ops",
	"spring mvc": "spring-mvc", "컨트롤러": "spring-mvc", "controller": "spring-mvc",
	"interceptor": "spring-mvc", "인터셉터": "spring-mvc",
	"spring security": "spring-security", "스프링시큐리티": "spring-security",
	"인증": "spring-security", "인가": "spring-security", "로그인": "spring-security",
	"jwt": "spring-security", "oauth": "spring-security", "oauth2": "spring-security",
	"csrf": "spring-security", "권한": "spring-security",
	"jpa": "spring-data-jpa", "querydsl": "spring-data-jpa",
	"spring data": "spring-data-jpa", "스프링데이터": "spring-data-jpa",
	"n+1": "spring-data-jpa", "entity": "spring-data-jpa",
	"@transactional": "spring-data-jpa",
	"spring test": "spring-test", "mockbean": "spring-test", "mockmvc": "spring-test",
	"@springboottest": "spring-test", "testcontainers": "spring-test",
	"maven": "build", "gradle": "build",
	// Tmax WAS
	"jeus": "jeus", "제우스": "jeus",
	"tmax": "jeus", "티맥스": "jeus",
	"webtob": "jeus", "웹투비": "jeus",

	// Python
	"python": "python", "파이썬": "python",
	"fastapi": "fastapi", "django": "django",

	// Terminal/OS
	"ip": "ip", "아이피": "ip",
	"terminal": "terminal", "터미널": "terminal",
	"powershell": "windows", "cmd": "windows",
	"git": "git", "깃": "git",
	// Shell/Bash scripting
	"bash": "bash-scripting", "쉘스크립트": "bash-scripting", "셸스크립트": "bash-scripting",
	"shell script": "bash-scripting", "쉘": "bash-scripting",
	"set -e": "bash-scripting", "trap": "bash-scripting", "getopts": "bash-scripting",
	"sed": "shell-tools", "awk": "shell-tools", "jq": "shell-tools",
	"xargs": "shell-tools", "grep": "shell-tools",
	"cron": "shell-ops", "crontab": "shell-ops", "크론": "shell-ops",
	"systemd": "shell-ops", "systemctl": "shell-ops",
	"ssh": "shell-ops", "rsync": "shell-ops",
	"배포스크립트": "shell-ops", "서버관리": "shell-ops",

	// Tools
	"vite": "vite", "docker": "docker",
	"rest": "rest", "api": "api",
	// Database / SQL
	"sql": "sql-core", "쿼리": "sql-core", "query": "sql-core",
	"join": "sql-core", "서브쿼리": "sql-core", "subquery": "sql-core",
	"윈도우함수": "sql-core", "window function": "sql-core",
	"cte": "sql-core",
	"ddl": "sql-ddl-design", "create table": "sql-ddl-design",
	"인덱스": "sql-ddl-design", "index": "sql-ddl-design",
	"정규화": "sql-ddl-design", "normalization": "sql-ddl-design",
	"erd": "sql-ddl-design", "스키마": "sql-ddl-design",
	"postgresql": "postgresql", "postgres": "postgresql", "pg": "postgresql",
	"jsonb": "postgresql",
	"mysql": "mysql", "mariadb": "mysql",
	"innodb": "mysql",
	"oracle": "oracle", "오라클": "oracle",
	"plsql": "oracle", "pl/sql": "oracle",
	"rownum": "oracle", "connect by": "oracle",
	"tibero": "tibero", "티베로": "tibero",
	"tbsql": "tibero",

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

// synonymGroups maps a canonical keyword to additional natural-language
// synonyms that users might type. These expand the search beyond the
// exact techDictionary mappings.
var synonymGroups = map[string][]string{
	"spring-security":    {"인증", "인가", "로그인", "토큰", "권한", "세션", "필터체인", "authentication", "authorization", "login"},
	"spring-data-jpa":    {"엔티티", "레포지토리", "영속성", "지연로딩", "즉시로딩", "fetch", "persistence"},
	"spring-core":        {"의존성", "주입", "빈", "컨텍스트", "프로파일", "설정", "autowired"},
	"spring-mvc":         {"요청", "응답", "매핑", "핸들러", "뷰", "폼", "바인딩"},
	"spring-boot-ops":    {"배포", "운영", "모니터링", "헬스체크", "프로파일", "도커"},
	"spring-test":        {"테스트", "목", "단위테스트", "통합테스트", "슬라이스"},
	"sql-core":           {"쿼리", "조회", "조인", "집계", "그룹", "정렬", "필터", "where", "having"},
	"sql-ddl-design":     {"테이블", "컬럼", "스키마", "설계", "제약조건", "외래키", "기본키"},
	"postgresql":         {"포스트그레스", "피지", "pg"},
	"mysql":              {"마이에스큐엘", "마리아디비"},
	"oracle":             {"오라클", "plsql"},
	"tibero":             {"티베로", "티맥스db", "tmax db"},
	"jeus":               {"제우스", "was", "웹서버", "티맥스was"},
	"bash-scripting":     {"스크립트", "자동화", "배시", "쉘코딩"},
	"shell-tools":        {"텍스트처리", "파싱", "로그분석", "파이프"},
	"shell-ops":          {"서버관리", "배포", "운영", "데몬", "서비스등록"},
	"react-class":        {"클래스", "레거시", "componentdidmount", "생명주기"},
	"react-patterns":     {"패턴", "컴포넌트설계", "아키텍처", "구조"},
	"react-performance":  {"성능", "최적화", "느림", "렌더링", "번들", "로딩"},
	"nextjs-pages":       {"페이지라우터", "getserversideprops", "ssr", "ssg", "isr"},
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

	// Pass 3 (Level 1 enhancement): synonym expansion.
	// If raw tokens didn't match techDictionary, check synonymGroups.
	// e.g. user types "인증" → not in techDictionary for older builds,
	// but synonymGroups["spring-security"] contains "인증" → match.
	if len(result) == 0 {
		allTokens := tokenize(lower)
		for canonical, synonyms := range synonymGroups {
			for _, syn := range synonyms {
				for _, tok := range allTokens {
					if tok == syn || strings.Contains(tok, syn) {
						addKeyword(canonical)
						break
					}
				}
			}
		}
	} else {
		// Even with results, expand via synonyms for better recall.
		// Only add if a raw token matches a synonym for a DIFFERENT canonical.
		allTokens := tokenize(lower)
		for canonical, synonyms := range synonymGroups {
			if seen[canonical] {
				continue // already matched
			}
			for _, syn := range synonyms {
				for _, tok := range allTokens {
					if tok == syn {
						addKeyword(canonical)
						break
					}
				}
			}
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
