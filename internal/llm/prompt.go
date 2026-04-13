package llm

import _ "embed"

type Mode int

const (
	ModeSuper Mode = iota // 0 — Super (smart all-in-one)
	ModeDev               // 1 — Deep Agent (long-running autonomous)
	ModePlan              // 2 — Plan (plan-first explicit)
)

// ModeDeep is a semantic alias for ModeDev — the rename kept the
// internal identifier stable to avoid a sweeping refactor of switch-by-index
// call sites, while the user-facing name is now "Deep Agent".
const ModeDeep = ModeDev

const ModeCount = 3

type ModeInfo struct {
	ID          string
	Name        string
	Description string
	Model       string // "super" or "dev" — config key
	Tools       []string
}

var Modes = [ModeCount]ModeInfo{
	ModeSuper: {
		ID:          "super",
		Name:        "Super",
		Description: "Smart all-in-one. Auto-detects intent and handles chat/plan/deep tasks",
		Model:       "super",
		Tools:       []string{"grep_search", "glob_search", "file_read", "file_write", "file_edit", "list_files", "shell_exec"},
	},
	ModeDev: {
		ID:          "deep",
		Name:        "Deep Agent",
		Description: "Long-running autonomous coding. Up to 100 iterations with self-verification",
		Model:       "super",
		Tools:       []string{"grep_search", "glob_search", "file_read", "file_write", "file_edit", "list_files", "shell_exec"},
	},
	ModePlan: {
		ID:          "plan",
		Name:        "Plan",
		Description: "Plan-first. Creates a step-by-step plan and executes on approval",
		Model:       "super",
		Tools:       []string{"grep_search", "glob_search", "file_read", "file_write", "file_edit", "list_files", "shell_exec"},
	},
}

func (m Mode) String() string {
	if int(m) < ModeCount {
		return Modes[m].Name
	}
	return "unknown"
}

func (m Mode) Info() ModeInfo {
	if int(m) < ModeCount {
		return Modes[m]
	}
	return ModeInfo{}
}

// System prompts are embedded from prompts/*.md at compile time.
// Benefits: prompt caching, diff hygiene, human-editable Markdown.

//go:embed prompts/core.md
var clarifyFirstDirective string

//go:embed prompts/super.md
var promptSuper string

//go:embed prompts/dev.md
var promptDev string

//go:embed prompts/plan.md
var promptPlan string

//go:embed prompts/askuser.md
var askUserPromptDoc string

func SystemPrompt(mode Mode) string {
	var body string
	switch mode {
	case ModeSuper:
		body = promptSuper
	case ModeDev:
		body = promptDev
	case ModePlan:
		body = promptPlan
	default:
		return ""
	}
	return clarifyFirstDirective + body + askUserPromptDoc
}
