package tools

import (
	"strings"
	"testing"
)

func TestORPatternSplit(t *testing.T) {
	SetUserContext("RWA_IBS_DMB_CMM_MAS 테이블의 DMB_K 컬럼 사용하는 프로그램")
	ResetFailedPatterns()

	result := Execute("grep_search", `{"pattern":"RWA_IBS_DMB_CMM_MAS.*DMB_K|DMB_K.*RWA_IBS_DMB_CMM_MAS","include":"*.go"}`)
	if len(result) == 0 {
		t.Fatal("Empty result")
	}
	// Should NOT contain "DMB_K|DMB_K" — the old bug
	if strings.Contains(result, "DMB_K|DMB_K") {
		t.Fatal("OR pattern not properly split — found DMB_K|DMB_K as term")
	}
	t.Logf("OR pattern result length: %d", len(result))
}

func TestDuplicateBlock(t *testing.T) {
	ResetFailedPatterns()
	Execute("grep_search", `{"pattern":"ZZZZZ_NEVER_EXISTS_12345"}`)
	Execute("grep_search", `{"pattern":"ZZZZZ_NEVER_EXISTS_12345"}`)
	result := Execute("grep_search", `{"pattern":"ZZZZZ_NEVER_EXISTS_12345"}`)
	if !strings.Contains(result, "Already searched") {
		t.Fatalf("Expected blocked message, got: %s", result)
	}
	t.Logf("Blocked: %s", result)
}

func TestAutoKeywordExtraction(t *testing.T) {
	terms := extractSearchTerms("RWA_IBS_DMB_CMM_MAS 테이블의 DMB_K 컬럼")
	if len(terms) != 2 {
		t.Fatalf("Expected 2 terms, got %d: %v", len(terms), terms)
	}
	if terms[0] != "RWA_IBS_DMB_CMM_MAS" || terms[1] != "DMB_K" {
		t.Fatalf("Wrong terms: %v", terms)
	}
	t.Logf("Extracted: %v", terms)
}

func TestAutoKeywordKoreanAttached(t *testing.T) {
	terms := extractSearchTerms("RWA_IBS_DMB_CMM_MAS테이블의DMB_K컬럼사용")
	if len(terms) != 2 {
		t.Fatalf("Expected 2 terms from attached Korean, got %d: %v", len(terms), terms)
	}
	t.Logf("Extracted from attached: %v", terms)
}
