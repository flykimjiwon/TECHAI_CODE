// Author: Kim Jiwon (github.com/flykimjiwon) — forked from hanimo-code
package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimjiwon/tgc/internal/config"
)

// Event types for lifecycle hooks.
const (
	EventSessionStart    = "session_start"
	EventPreToolUse      = "pre_tool_use"
	EventPostToolUse     = "post_tool_use"
	EventStop            = "stop"
	EventUserPrompt      = "user_prompt"
)

// HookResult represents the outcome of a hook execution.
type HookResult int

const (
	HookSuccess       HookResult = iota // hook ran OK
	HookFailContinue                     // hook failed but continue
	HookFailAbort                        // hook failed — abort the operation
)

// ToolMatcher filters which tools a hook applies to.
type ToolMatcher struct {
	ToolName string `json:"tool_name,omitempty"` // exact match, empty = all tools
}

// HookDef defines a single hook from hooks.json.
type HookDef struct {
	Matcher   ToolMatcher `json:"matcher"`
	Command   []string    `json:"command"`
	TimeoutMs int         `json:"timeout_ms,omitempty"` // default 10000
}

// HooksConfig is the top-level hooks.json structure.
type HooksConfig struct {
	SessionStart []HookDef `json:"session_start,omitempty"`
	PreToolUse   []HookDef `json:"pre_tool_use,omitempty"`
	PostToolUse  []HookDef `json:"post_tool_use,omitempty"`
	Stop         []HookDef `json:"stop,omitempty"`
	UserPrompt   []HookDef `json:"user_prompt,omitempty"`
}

// HookPayload is passed to hook commands via stdin as JSON.
type HookPayload struct {
	Event     string `json:"event"`
	SessionID string `json:"session_id,omitempty"`
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name,omitempty"`
	ToolArgs  string `json:"tool_args,omitempty"`
	ToolResult string `json:"tool_result,omitempty"`
	Prompt    string `json:"prompt,omitempty"`
}

// Manager loads and executes lifecycle hooks.
type Manager struct {
	cfg HooksConfig
}

// NewManager loads hooks from the project-local .tgc/hooks.json,
// falling back to ~/.tgc/hooks.json.
func NewManager() *Manager {
	m := &Manager{}

	// Try project-local first
	paths := []string{
		filepath.Join(".tgc", "hooks.json"),
		filepath.Join(config.ConfigDir(), "hooks.json"),
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &m.cfg); err != nil {
			config.DebugLog("[HOOKS] failed to parse %s: %v", p, err)
			continue
		}
		total := len(m.cfg.SessionStart) + len(m.cfg.PreToolUse) + len(m.cfg.PostToolUse) + len(m.cfg.Stop) + len(m.cfg.UserPrompt)
		config.DebugLog("[HOOKS] loaded %d hooks from %s", total, p)
		return m
	}

	config.DebugLog("[HOOKS] no hooks.json found")
	return m
}

// HasHooks returns true if any hooks are configured.
func (m *Manager) HasHooks() bool {
	return len(m.cfg.SessionStart)+len(m.cfg.PreToolUse)+len(m.cfg.PostToolUse)+len(m.cfg.Stop)+len(m.cfg.UserPrompt) > 0
}

// RunSessionStart fires session_start hooks.
func (m *Manager) RunSessionStart(sessionID string) {
	payload := HookPayload{
		Event:     EventSessionStart,
		SessionID: sessionID,
		Cwd:       mustCwd(),
	}
	for _, h := range m.cfg.SessionStart {
		m.runHook(h, payload)
	}
}

// RunPreToolUse fires pre_tool_use hooks. Returns HookFailAbort if any hook aborts.
func (m *Manager) RunPreToolUse(toolName, toolArgs string) HookResult {
	payload := HookPayload{
		Event:    EventPreToolUse,
		ToolName: toolName,
		ToolArgs: toolArgs,
		Cwd:      mustCwd(),
	}
	for _, h := range m.cfg.PreToolUse {
		if !matchesTool(h.Matcher, toolName) {
			continue
		}
		result := m.runHook(h, payload)
		if result == HookFailAbort {
			return HookFailAbort
		}
	}
	return HookSuccess
}

// RunPostToolUse fires post_tool_use hooks.
func (m *Manager) RunPostToolUse(toolName, toolArgs, toolResult string) {
	payload := HookPayload{
		Event:      EventPostToolUse,
		ToolName:   toolName,
		ToolArgs:   toolArgs,
		ToolResult: truncateForHook(toolResult, 4096),
		Cwd:        mustCwd(),
	}
	for _, h := range m.cfg.PostToolUse {
		if !matchesTool(h.Matcher, toolName) {
			continue
		}
		m.runHook(h, payload)
	}
}

// RunStop fires stop hooks (agent turn finished).
func (m *Manager) RunStop(sessionID string) {
	payload := HookPayload{
		Event:     EventStop,
		SessionID: sessionID,
		Cwd:       mustCwd(),
	}
	for _, h := range m.cfg.Stop {
		m.runHook(h, payload)
	}
}

// RunUserPrompt fires user_prompt hooks.
func (m *Manager) RunUserPrompt(prompt string) {
	payload := HookPayload{
		Event:  EventUserPrompt,
		Prompt: prompt,
		Cwd:    mustCwd(),
	}
	for _, h := range m.cfg.UserPrompt {
		m.runHook(h, payload)
	}
}

func (m *Manager) runHook(h HookDef, payload HookPayload) HookResult {
	if len(h.Command) == 0 {
		return HookSuccess
	}

	timeout := time.Duration(h.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, h.Command[0], h.Command[1:]...)
	cmd.Dir = mustCwd()

	// Pass payload as JSON via stdin
	payloadJSON, _ := json.Marshal(payload)
	cmd.Stdin = strings.NewReader(string(payloadJSON))

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		config.DebugLog("[HOOKS] %s hook failed: cmd=%v err=%v output=%s", payload.Event, h.Command, err, truncateForHook(string(output), 500))
		// Check exit code — exit 2 means abort
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
			return HookFailAbort
		}
		return HookFailContinue
	}

	config.DebugLog("[HOOKS] %s hook OK: cmd=%v output=%s", payload.Event, h.Command, truncateForHook(string(output), 200))
	return HookSuccess
}

func matchesTool(matcher ToolMatcher, toolName string) bool {
	if matcher.ToolName == "" {
		return true // empty matcher = match all
	}
	return matcher.ToolName == toolName
}

func mustCwd() string {
	cwd, _ := os.Getwd()
	return cwd
}

func truncateForHook(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// CreateDefaultHooksFile creates a template hooks.json for the user.
func CreateDefaultHooksFile(projectLocal bool) (string, error) {
	template := HooksConfig{
		PostToolUse: []HookDef{
			{
				Matcher:   ToolMatcher{ToolName: "file_edit"},
				Command:   []string{"echo", "file edited"},
				TimeoutMs: 5000,
			},
		},
		Stop: []HookDef{
			{
				Command:   []string{"echo", "turn completed"},
				TimeoutMs: 3000,
			},
		},
	}

	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return "", err
	}

	var dir string
	if projectLocal {
		dir = ".tgc"
	} else {
		dir = config.ConfigDir()
	}
	_ = os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "hooks.json")

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write hooks.json: %w", err)
	}
	return path, nil
}
