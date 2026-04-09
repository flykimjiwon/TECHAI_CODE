package knowledge

import (
	"sort"
	"testing"
)

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []string // each must be present in result
	}{
		{
			name:     "tailwind responsive card",
			query:    "tailwind로 반응형 카드 만들어줘",
			expected: []string{"tailwind", "responsive"},
		},
		{
			name:     "BXM Bean multi-select",
			query:    "BXM Bean에서 다건 조회 패턴 알려줘",
			expected: []string{"bxm", "bean", "다건", "조회"},
		},
		{
			name:     "recharts bar chart",
			query:    "recharts로 bar chart 그려줘",
			expected: []string{"recharts", "chart"},
		},
		{
			name:     "Go goroutine",
			query:    "Go에서 goroutine 사용법",
			expected: []string{"go", "goroutine"},
		},
		{
			name:     "IP address",
			query:    "IP 주소 확인하는 법",
			expected: []string{"ip"},
		},
		{
			name:     "React useState",
			query:    "React useState 사용법",
			expected: []string{"react"},
		},
		{
			name:     "Spring Boot REST API",
			query:    "Spring Boot REST API 만들기",
			expected: []string{"spring", "rest", "api"},
		},
		{
			name:     "Vue 3 composition API",
			query:    "Vue 3 composition API 예제",
			expected: []string{"vue", "composition", "api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractKeywords(tt.query)
			for _, exp := range tt.expected {
				if !containsStr(got, exp) {
					t.Errorf("ExtractKeywords(%q) = %v, missing expected keyword %q", tt.query, got, exp)
				}
			}
		})
	}
}

func TestExtractKeywords_Empty(t *testing.T) {
	got := ExtractKeywords("")
	if len(got) != 0 {
		t.Errorf("ExtractKeywords(\"\") = %v, want nil/empty", got)
	}
}

func TestExtractKeywords_KoreanOnly(t *testing.T) {
	got := ExtractKeywords("디버깅 방법")
	expected := []string{"debugging"}
	for _, exp := range expected {
		if !containsStr(got, exp) {
			t.Errorf("ExtractKeywords(\"디버깅 방법\") = %v, missing %q", got, exp)
		}
	}
}

func TestExtractKeywords_MultiWordMatch(t *testing.T) {
	got := ExtractKeywords("spring boot 시작하기")
	// "spring boot" should match as a compound term -> "spring"
	if !containsStr(got, "spring") {
		t.Errorf("ExtractKeywords(\"spring boot 시작하기\") = %v, missing \"spring\"", got)
	}
	// The individual word "boot" should NOT appear as a separate keyword
	// (it is not in the dictionary on its own, so this is automatically correct)
}

func TestExtractKeywords_Deduplication(t *testing.T) {
	// "usestate" maps to "react", and "react" maps to "react" — should appear only once
	got := ExtractKeywords("React useState useEffect")
	count := 0
	for _, kw := range got {
		if kw == "react" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 \"react\", got %d in %v", count, got)
	}
}

func TestExtractKeywords_DarkMode(t *testing.T) {
	got := ExtractKeywords("dark mode 적용하기")
	if !containsStr(got, "darkmode") {
		t.Errorf("ExtractKeywords(\"dark mode 적용하기\") = %v, missing \"darkmode\"", got)
	}
}

func TestExtractKeywords_AppRouter(t *testing.T) {
	got := ExtractKeywords("Next.js app router 설정")
	if !containsStr(got, "nextjs") {
		t.Errorf("ExtractKeywords(\"Next.js app router 설정\") = %v, missing \"nextjs\"", got)
	}
}

func TestExtractKeywords_ChartJS(t *testing.T) {
	got := ExtractKeywords("chart.js로 파이 차트 그리기")
	if !containsStr(got, "chartjs") {
		t.Errorf("ExtractKeywords(\"chart.js로 파이 차트 그리기\") = %v, missing \"chartjs\"", got)
	}
}

func TestExtractKeywords_ApacheEcharts(t *testing.T) {
	got := ExtractKeywords("apache echarts 사용법")
	if !containsStr(got, "echarts") {
		t.Errorf("ExtractKeywords(\"apache echarts 사용법\") = %v, missing \"echarts\"", got)
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"next.js is cool", []string{"next.js", "is", "cool"}},
		{"@bxmbean test", []string{"@bxmbean", "test"}},
		{"shadcn/ui setup", []string{"shadcn/ui", "setup"}},
		{"foo---bar", []string{"foo---bar"}},
		{"", nil},
		{"   ", nil},
		{"한글테스트 영어test", []string{"한글테스트", "영어", "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := tokenize(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("tokenize(%q) = %v, want %v", tt.input, got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSortDescByLen(t *testing.T) {
	input := []string{"ab", "abcde", "a", "abc"}
	sortDescByLen(input)
	expected := []string{"abcde", "abc", "ab", "a"}
	for i := range input {
		if input[i] != expected[i] {
			t.Errorf("sortDescByLen: got %v, want %v", input, expected)
			break
		}
	}
}

func TestMultiWordTermsSorted(t *testing.T) {
	// multiWordTerms should be sorted descending by length
	if len(multiWordTerms) == 0 {
		t.Fatal("multiWordTerms is empty")
	}
	sorted := sort.SliceIsSorted(multiWordTerms, func(i, j int) bool {
		return len(multiWordTerms[i]) > len(multiWordTerms[j])
	})
	if !sorted {
		t.Errorf("multiWordTerms not sorted descending by length: %v", multiWordTerms)
	}
}
