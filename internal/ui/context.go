package ui

// ContextWarnLevel describes how "full" the model context window is.
// The UI maps each level to a color so users get an at-a-glance signal.
type ContextWarnLevel int

const (
	ContextNormal   ContextWarnLevel = iota // < 70%
	ContextWarn                             // 70% – 79%
	ContextCritical                         // >= 80%
)

// Context thresholds. These are tuned to match Claude Code's UX: yellow
// warning at 70% so the user has time to react, red at 80% so they know it
// is time to compact, and auto-trigger at 90% to prevent crashes.
const (
	contextWarnThreshold     = 70
	contextCriticalThreshold = 80
	contextAutoThreshold     = 90
)

// ContextPercent returns how full the context window is, as an integer
// 0-100. It returns 0 when window is non-positive (no capability info
// yet) and caps at 100 so the HUD never shows "120%".
func ContextPercent(tokens, window int) int {
	if window <= 0 || tokens <= 0 {
		return 0
	}
	pct := (tokens * 100) / window
	if pct > 100 {
		return 100
	}
	return pct
}

// ContextLevel classifies a percentage into Normal / Warn / Critical so
// the status bar can pick a color without knowing the thresholds.
func ContextLevel(pct int) ContextWarnLevel {
	switch {
	case pct >= contextCriticalThreshold:
		return ContextCritical
	case pct >= contextWarnThreshold:
		return ContextWarn
	default:
		return ContextNormal
	}
}

// ShouldAutoCompact reports whether the caller should proactively run the
// stage-3 LLM summary compaction. Returns true at and above 90%.
func ShouldAutoCompact(pct int) bool {
	return pct >= contextAutoThreshold
}
