package ui

import "testing"

func TestContextPercent_ZeroWindow_ReturnsZero(t *testing.T) {
	if got := ContextPercent(100, 0); got != 0 {
		t.Errorf("ContextPercent(100, 0) = %d, want 0", got)
	}
}

func TestContextPercent_NegativeWindow_ReturnsZero(t *testing.T) {
	if got := ContextPercent(100, -1); got != 0 {
		t.Errorf("ContextPercent(100, -1) = %d, want 0", got)
	}
}

func TestContextPercent_ZeroTokens_ReturnsZero(t *testing.T) {
	if got := ContextPercent(0, 128000); got != 0 {
		t.Errorf("ContextPercent(0, 128000) = %d, want 0", got)
	}
}

func TestContextPercent_HalfFull(t *testing.T) {
	if got := ContextPercent(64000, 128000); got != 50 {
		t.Errorf("ContextPercent(64000, 128000) = %d, want 50", got)
	}
}

func TestContextPercent_OverCaps100(t *testing.T) {
	// Over-window tokens should cap at 100% (never display >100%).
	if got := ContextPercent(200000, 128000); got != 100 {
		t.Errorf("ContextPercent(200000, 128000) = %d, want 100 (capped)", got)
	}
}

func TestContextPercent_QwenLargeWindow(t *testing.T) {
	// Qwen3-Coder has a 262144 window; 2621 tokens ≈ 1%.
	if got := ContextPercent(2621, 262144); got != 0 && got != 1 {
		t.Errorf("ContextPercent(2621, 262144) = %d, want 0 or 1", got)
	}
}

func TestContextLevel_Normal_Below70(t *testing.T) {
	for _, pct := range []int{0, 1, 50, 69} {
		if got := ContextLevel(pct); got != ContextNormal {
			t.Errorf("ContextLevel(%d) = %v, want ContextNormal", pct, got)
		}
	}
}

func TestContextLevel_Warn_70to79(t *testing.T) {
	for _, pct := range []int{70, 75, 79} {
		if got := ContextLevel(pct); got != ContextWarn {
			t.Errorf("ContextLevel(%d) = %v, want ContextWarn", pct, got)
		}
	}
}

func TestContextLevel_Critical_80Plus(t *testing.T) {
	for _, pct := range []int{80, 85, 90, 99, 100} {
		if got := ContextLevel(pct); got != ContextCritical {
			t.Errorf("ContextLevel(%d) = %v, want ContextCritical", pct, got)
		}
	}
}

func TestShouldAutoCompact_At90(t *testing.T) {
	if !ShouldAutoCompact(90) {
		t.Error("ShouldAutoCompact(90) = false, want true")
	}
}

func TestShouldAutoCompact_Below90(t *testing.T) {
	for _, pct := range []int{0, 50, 80, 89} {
		if ShouldAutoCompact(pct) {
			t.Errorf("ShouldAutoCompact(%d) = true, want false", pct)
		}
	}
}

func TestShouldAutoCompact_Above90(t *testing.T) {
	for _, pct := range []int{91, 95, 100} {
		if !ShouldAutoCompact(pct) {
			t.Errorf("ShouldAutoCompact(%d) = false, want true", pct)
		}
	}
}
