package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager_NoFile(t *testing.T) {
	// Work in temp dir with no hooks.json
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldCwd)

	m := NewManager()
	if m.HasHooks() {
		t.Error("expected no hooks when no hooks.json exists")
	}
}

func TestNewManager_LoadsProjectLocal(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldCwd)

	// Create .tgc/hooks.json
	os.MkdirAll(filepath.Join(tmpDir, ".tgc"), 0755)
	cfg := HooksConfig{
		PostToolUse: []HookDef{
			{Command: []string{"echo", "test"}, TimeoutMs: 1000},
		},
	}
	data, _ := json.Marshal(cfg)
	os.WriteFile(filepath.Join(tmpDir, ".tgc", "hooks.json"), data, 0644)

	m := NewManager()
	if !m.HasHooks() {
		t.Error("expected hooks to be loaded")
	}
	if len(m.cfg.PostToolUse) != 1 {
		t.Errorf("expected 1 post_tool_use hook, got %d", len(m.cfg.PostToolUse))
	}
}

func TestMatchesTool(t *testing.T) {
	tests := []struct {
		matcher  ToolMatcher
		toolName string
		want     bool
	}{
		{ToolMatcher{}, "file_edit", true},              // empty = match all
		{ToolMatcher{ToolName: "file_edit"}, "file_edit", true},
		{ToolMatcher{ToolName: "file_edit"}, "shell_exec", false},
		{ToolMatcher{ToolName: ""}, "anything", true},
	}

	for _, tt := range tests {
		got := matchesTool(tt.matcher, tt.toolName)
		if got != tt.want {
			t.Errorf("matchesTool(%+v, %q) = %v, want %v", tt.matcher, tt.toolName, got, tt.want)
		}
	}
}

func TestRunPreToolUse_Echo(t *testing.T) {
	m := &Manager{
		cfg: HooksConfig{
			PreToolUse: []HookDef{
				{
					Matcher:   ToolMatcher{ToolName: "shell_exec"},
					Command:   []string{"echo", "ok"},
					TimeoutMs: 2000,
				},
			},
		},
	}

	// Should succeed (echo returns 0)
	result := m.RunPreToolUse("shell_exec", `{"command":"ls"}`)
	if result != HookSuccess {
		t.Errorf("expected HookSuccess, got %v", result)
	}

	// Should not match (different tool)
	result = m.RunPreToolUse("file_edit", `{}`)
	if result != HookSuccess {
		t.Errorf("expected HookSuccess for unmatched tool, got %v", result)
	}
}

func TestRunPreToolUse_Abort(t *testing.T) {
	m := &Manager{
		cfg: HooksConfig{
			PreToolUse: []HookDef{
				{
					Matcher:   ToolMatcher{},
					Command:   []string{"bash", "-c", "exit 2"},
					TimeoutMs: 2000,
				},
			},
		},
	}

	result := m.RunPreToolUse("any_tool", `{}`)
	if result != HookFailAbort {
		t.Errorf("expected HookFailAbort for exit 2, got %v", result)
	}
}

func TestRunPostToolUse(t *testing.T) {
	m := &Manager{
		cfg: HooksConfig{
			PostToolUse: []HookDef{
				{
					Matcher:   ToolMatcher{ToolName: "file_edit"},
					Command:   []string{"echo", "linted"},
					TimeoutMs: 2000,
				},
			},
		},
	}

	// Should not panic
	m.RunPostToolUse("file_edit", `{}`, "ok")
	m.RunPostToolUse("shell_exec", `{}`, "ok") // unmatched, noop
}

func TestHookPayloadJSON(t *testing.T) {
	payload := HookPayload{
		Event:    EventPreToolUse,
		ToolName: "shell_exec",
		ToolArgs: `{"command":"ls"}`,
		Cwd:      "/tmp",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded HookPayload
	json.Unmarshal(data, &decoded)
	if decoded.Event != EventPreToolUse {
		t.Errorf("expected event=%s, got %s", EventPreToolUse, decoded.Event)
	}
	if decoded.ToolName != "shell_exec" {
		t.Errorf("expected tool_name=shell_exec, got %s", decoded.ToolName)
	}
}
