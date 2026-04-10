// Package agents contains autonomous agent loop helpers.
//
// The /auto mode lets the model act on a task with tools for up to
// MaxAutoIterations rounds without blocking for user input. The model
// signals completion or the need for human help by emitting one of the
// markers in its response; the host app watches for them in the stream
// buffer via CheckAutoMarkers.
package agents

import "strings"

const (
	// MaxAutoIterations caps the number of tool-call rounds /auto can
	// run before it is forcibly stopped, so a confused model cannot
	// loop forever.
	MaxAutoIterations = 20

	// AutoCompleteMarker is emitted by the model when the task is done.
	AutoCompleteMarker = "[AUTO_COMPLETE]"

	// AutoPauseMarker is emitted when the model is blocked and needs
	// human input.
	AutoPauseMarker = "[AUTO_PAUSE]"
)

// AutoPromptSuffix is appended to the system prompt when auto mode is
// active. It tells the model how to signal termination and prohibits
// the usual "ask the user a question" fallback that would otherwise
// stall the loop.
const AutoPromptSuffix = `

You are in AUTONOMOUS MODE. Complete the task independently:
- Use tools to read, write, and test code
- Run diagnostics to verify your work
- When the task is fully complete, output [AUTO_COMPLETE]
- If you are blocked and need human input, output [AUTO_PAUSE]
- Do NOT ask questions — make decisions and proceed`

// CheckAutoMarkers scans a response for the two control markers and
// returns which (if any) were present. Both can be true only if the
// model emits both strings in the same response; callers typically
// treat complete as having precedence.
func CheckAutoMarkers(content string) (complete bool, pause bool) {
	complete = strings.Contains(content, AutoCompleteMarker)
	pause = strings.Contains(content, AutoPauseMarker)
	return
}
