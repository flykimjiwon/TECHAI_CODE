// Author: Kim Jiwon (github.com/flykimjiwon) — forked from hanimo-code
package agents

import "strings"

var (
	// MaxAutoIterations is the hard upper bound on auto-mode tool-loop
	// iterations. Deep Agent mode uses 100 (matching hanimo); /auto uses 20.
	MaxAutoIterations     = 20
	MaxDeepIterations     = 100
)

const (
	AutoCompleteMarker = "[AUTO_COMPLETE]"
	AutoPauseMarker    = "[AUTO_PAUSE]"
	// Deep Agent uses TASK_COMPLETE (matching hanimo's convention)
	TaskCompleteMarker = "[TASK_COMPLETE]"
)

// AutoPromptSuffix is appended to the system prompt when /auto is active.
const AutoPromptSuffix = `

You are in AUTONOMOUS MODE. Complete the task independently:
- Use tools to read, write, and test code
- Run diagnostics to verify your work
- When the task is fully complete, output [AUTO_COMPLETE]
- If you're blocked and need human input, output [AUTO_PAUSE]
- Do NOT ask questions — make decisions and proceed`

// CheckAutoMarkers checks response content for auto-mode control markers.
// Recognizes both /auto markers and Deep Agent's [TASK_COMPLETE].
func CheckAutoMarkers(content string) (complete bool, pause bool) {
	complete = strings.Contains(content, AutoCompleteMarker) ||
		strings.Contains(content, TaskCompleteMarker)
	pause = strings.Contains(content, AutoPauseMarker)
	return
}
