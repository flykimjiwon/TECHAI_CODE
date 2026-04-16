package tools

import (
	"fmt"
	"strings"
)

// GenerateUnifiedDiff produces a simple unified diff between old and new content.
// Returns a human-readable diff string with +/- markers.
func GenerateUnifiedDiff(path, oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Find changed region
	// Leading common lines
	commonStart := 0
	for commonStart < len(oldLines) && commonStart < len(newLines) && oldLines[commonStart] == newLines[commonStart] {
		commonStart++
	}

	// Trailing common lines
	commonEnd := 0
	for commonEnd < len(oldLines)-commonStart && commonEnd < len(newLines)-commonStart &&
		oldLines[len(oldLines)-1-commonEnd] == newLines[len(newLines)-1-commonEnd] {
		commonEnd++
	}

	// Extract changed sections
	removedStart := commonStart
	removedEnd := len(oldLines) - commonEnd
	addedStart := commonStart
	addedEnd := len(newLines) - commonEnd

	if removedStart >= removedEnd && addedStart >= addedEnd {
		return "" // no diff
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", path, path))

	// Context: show 3 lines before
	ctxStart := commonStart - 3
	if ctxStart < 0 {
		ctxStart = 0
	}
	ctxEnd := len(oldLines) - commonEnd + 3
	if ctxEnd > len(oldLines) {
		ctxEnd = len(oldLines)
	}
	newCtxEnd := len(newLines) - commonEnd + 3
	if newCtxEnd > len(newLines) {
		newCtxEnd = len(newLines)
	}

	b.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
		ctxStart+1, ctxEnd-ctxStart,
		ctxStart+1, newCtxEnd-ctxStart-(removedEnd-removedStart)+(addedEnd-addedStart)))

	// Leading context
	for i := ctxStart; i < commonStart; i++ {
		b.WriteString(fmt.Sprintf(" %s\n", oldLines[i]))
	}

	// Removed lines
	for i := removedStart; i < removedEnd; i++ {
		b.WriteString(fmt.Sprintf("-%s\n", oldLines[i]))
	}

	// Added lines
	for i := addedStart; i < addedEnd; i++ {
		b.WriteString(fmt.Sprintf("+%s\n", newLines[i]))
	}

	// Trailing context
	trailStart := len(oldLines) - commonEnd
	trailEnd := trailStart + 3
	if trailEnd > len(oldLines) {
		trailEnd = len(oldLines)
	}
	for i := trailStart; i < trailEnd; i++ {
		b.WriteString(fmt.Sprintf(" %s\n", oldLines[i]))
	}

	return b.String()
}
