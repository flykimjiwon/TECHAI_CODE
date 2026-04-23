package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	openai "github.com/sashabaranov/go-openai"
)

// ChatMessage is the frontend-facing message type.
type ChatMessage struct {
	Role    string `json:"role"` // "user", "ai", "tool"
	Content string `json:"content"`
}

// chatEngine manages conversation state and LLM streaming.
type chatEngine struct {
	client  *openai.Client
	model   string
	history []openai.ChatCompletionMessage
	app     *App // back-reference for Wails events + tool execution
}

func newChatEngine(cfg TGCConfig, app *App) *chatEngine {
	ocfg := openai.DefaultConfig(cfg.API.APIKey)
	baseURL := strings.TrimRight(cfg.API.BaseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/chat/completions")
	ocfg.BaseURL = strings.TrimRight(baseURL, "/")

	return &chatEngine{
		client: openai.NewClientWithConfig(ocfg),
		model:  cfg.Models.Super,
		app:    app,
		history: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt(),
			},
		},
	}
}

func systemPrompt() string {
	base := `You are TECHAI, an AI coding assistant embedded in an IDE.
You can read and write files, search code, and run shell commands.
Answer in the same language the user uses. Be concise and direct.
When editing files, show the changes clearly.`

	// Load project context (.techai.md + project type detection)
	cwd, _ := os.Getwd()
	if ctx := loadKnowledgeContext(cwd); ctx != "" {
		base += ctx
	}

	return base
}

// toolDefs returns the OpenAI tool definitions — mirrors TUI's 15 tools.
func toolDefs() []openai.Tool {
	return []openai.Tool{
		makeTool("file_read", "Read a file and return its content", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":   map[string]string{"type": "string", "description": "File path to read"},
				"offset": map[string]string{"type": "string", "description": "Start line (1-indexed, optional)"},
				"limit":  map[string]string{"type": "string", "description": "Max lines to read (optional)"},
			},
			"required": []string{"path"},
		}),
		makeTool("file_write", "Write content to a file (creates or overwrites)", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":    map[string]string{"type": "string", "description": "File path to write"},
				"content": map[string]string{"type": "string", "description": "File content"},
			},
			"required": []string{"path", "content"},
		}),
		makeTool("shell_exec", "Execute a shell command and return output", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]string{"type": "string", "description": "Shell command to execute"},
				"timeout": map[string]string{"type": "string", "description": "Timeout in seconds (default: 30)"},
			},
			"required": []string{"command"},
		}),
		makeTool("grep_search", "Search file contents using grep pattern", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]string{"type": "string", "description": "Search pattern (regex supported)"},
				"path":    map[string]string{"type": "string", "description": "Directory to search in (default: .)"},
				"include": map[string]string{"type": "string", "description": "File glob filter (e.g. *.go)"},
			},
			"required": []string{"pattern"},
		}),
		makeTool("glob_search", "Find files matching a glob pattern", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]string{"type": "string", "description": "Glob pattern (e.g. **/*.go)"},
				"path":    map[string]string{"type": "string", "description": "Base directory (default: .)"},
			},
			"required": []string{"pattern"},
		}),
		makeTool("list_files", "List files in a directory", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]string{"type": "string", "description": "Directory path (default: .)"},
			},
		}),
		makeTool("git_status", "Get git repository status (branch, changes)", map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{},
		}),
		makeTool("git_diff", "Get git diff of current changes", map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{},
		}),
		makeTool("apply_patch", "Apply a unified diff patch to modify a file", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":  map[string]string{"type": "string", "description": "File path to patch"},
				"patch": map[string]string{"type": "string", "description": "Unified diff content"},
			},
			"required": []string{"path", "patch"},
		}),
		makeTool("git_log", "Get recent git commit history", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"n": map[string]string{"type": "string", "description": "Number of commits (default: 10)"},
			},
		}),
	}
}

func makeTool(name, desc string, params interface{}) openai.Tool {
	raw, _ := json.Marshal(params)
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        name,
			Description: desc,
			Parameters:  json.RawMessage(raw),
		},
	}
}

// SendMessage is called from the frontend. It streams the response via Wails events.
func (a *App) SendMessage(prompt string) {
	if a.chat == nil {
		return
	}

	// Add user message to history
	a.chat.history = append(a.chat.history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	// Stream response
	a.chat.streamResponse()
}

// GetChatHistory returns the conversation for display/export.
func (a *App) GetChatHistory() []ChatMessage {
	if a.chat == nil {
		return nil
	}
	var msgs []ChatMessage
	for _, m := range a.chat.history {
		if m.Role == "system" {
			continue
		}
		role := "ai"
		if m.Role == "user" {
			role = "user"
		} else if m.Role == "tool" {
			role = "tool"
		}
		msgs = append(msgs, ChatMessage{Role: role, Content: m.Content})
	}
	return msgs
}

// ExportChat saves the conversation to a markdown file.
func (a *App) ExportChat() (string, error) {
	msgs := a.GetChatHistory()
	if len(msgs) == 0 {
		return "", fmt.Errorf("no messages to export")
	}
	var sb strings.Builder
	sb.WriteString("# TECHAI Chat Export\n\n")
	for _, m := range msgs {
		switch m.Role {
		case "user":
			sb.WriteString("## You\n" + m.Content + "\n\n")
		case "ai":
			sb.WriteString("## AI\n" + m.Content + "\n\n")
		case "tool":
			sb.WriteString("> " + m.Content + "\n\n")
		}
	}
	path := filepath.Join(a.cwd, "techai-chat-export.md")
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return "", err
	}
	return path, nil
}

// ClearChat resets the conversation.
func (a *App) ClearChat() {
	if a.chat == nil {
		return
	}
	a.chat.history = []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt()},
	}
	runtime.EventsEmit(a.ctx, "chat:cleared")
}

// GetModel returns the current model name.
func (a *App) GetModel() string {
	if a.chat == nil {
		return ""
	}
	// Strip provider prefix for display
	m := a.chat.model
	if idx := strings.LastIndex(m, "/"); idx >= 0 {
		m = m[idx+1:]
	}
	return m
}

// streamResponse handles the streaming + tool loop.
func (ce *chatEngine) streamResponse() {
	for iter := 0; iter < 15; iter++ {
		req := openai.ChatCompletionRequest{
			Model:    ce.model,
			Messages: ce.history,
			Stream:   true,
			Tools:    toolDefs(),
		}

		stream, err := ce.client.CreateChatCompletionStream(context.Background(), req)
		if err != nil {
			runtime.EventsEmit(ce.app.ctx, "chat:error", err.Error())
			return
		}

		var contentBuf strings.Builder
		toolCalls := make(map[int]*openai.ToolCall)

		// Signal stream start
		runtime.EventsEmit(ce.app.ctx, "chat:stream_start")

		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				runtime.EventsEmit(ce.app.ctx, "chat:error", err.Error())
				stream.Close()
				return
			}

			if len(resp.Choices) == 0 {
				continue
			}

			delta := resp.Choices[0].Delta

			// Stream content chunks
			if delta.Content != "" {
				contentBuf.WriteString(delta.Content)
				runtime.EventsEmit(ce.app.ctx, "chat:chunk", delta.Content)
			}

			// Accumulate tool calls
			for _, tc := range delta.ToolCalls {
				idx := 0
				if tc.Index != nil {
					idx = *tc.Index
				}
				if _, ok := toolCalls[idx]; !ok {
					toolCalls[idx] = &openai.ToolCall{
						ID:   tc.ID,
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				} else {
					if tc.ID != "" {
						toolCalls[idx].ID = tc.ID
					}
					if tc.Function.Name != "" {
						toolCalls[idx].Function.Name = tc.Function.Name
					}
					toolCalls[idx].Function.Arguments += tc.Function.Arguments
				}
			}
		}
		stream.Close()

		// Signal stream done
		runtime.EventsEmit(ce.app.ctx, "chat:stream_done")

		// Text-based tool_call parsing fallback (Qwen3 proxy compat).
		// If the proxy didn't convert tool calls, parse from content.
		if len(toolCalls) == 0 && contentBuf.Len() > 0 {
			if parsed := parseTextToolCalls(contentBuf.String()); len(parsed) > 0 {
				for i, p := range parsed {
					idx := i
					toolCalls[idx] = &openai.ToolCall{
						ID:   p.ID,
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      p.Name,
							Arguments: p.Arguments,
						},
					}
				}
			}
		}

		// If tool calls, execute them and loop
		if len(toolCalls) > 0 {
			// Add assistant message with tool calls to history
			assistantMsg := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: contentBuf.String(),
			}
			for i := 0; i < len(toolCalls); i++ {
				if tc, ok := toolCalls[i]; ok {
					assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, *tc)
				}
			}
			ce.history = append(ce.history, assistantMsg)

			// Execute each tool
			for i := 0; i < len(toolCalls); i++ {
				tc, ok := toolCalls[i]
				if !ok {
					continue
				}

				runtime.EventsEmit(ce.app.ctx, "chat:tool_start", map[string]string{
					"name": tc.Function.Name,
					"args": tc.Function.Arguments,
				})

				result := ce.app.executeTool(tc.Function.Name, tc.Function.Arguments)

				runtime.EventsEmit(ce.app.ctx, "chat:tool_done", map[string]string{
					"name":   tc.Function.Name,
					"result": truncate(result, 200),
				})

				ce.history = append(ce.history, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    result,
					ToolCallID: tc.ID,
				})
			}

			// Continue the loop — model will see tool results
			continue
		}

		// No tool calls — normal completion
		if contentBuf.Len() > 0 {
			ce.history = append(ce.history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: contentBuf.String(),
			})
		}
		return
	}
}

// executeTool runs a tool and returns the result string.
func (a *App) executeTool(name, argsJSON string) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return fmt.Sprintf("Error: invalid arguments: %v", err)
	}

	switch name {
	case "file_read":
		path, _ := args["path"].(string)
		if path == "" {
			return "Error: path is required"
		}
		content, err := a.ReadFile(path)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return content

	case "file_write":
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		if path == "" {
			return "Error: path is required"
		}
		if err := a.WriteFile(path, content); err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		// Notify frontend to refresh
		runtime.EventsEmit(a.ctx, "file:changed", path)
		return fmt.Sprintf("OK: wrote %d bytes to %s", len(content), path)

	case "grep_search", "search_files":
		pattern, _ := args["pattern"].(string)
		dir, _ := args["path"].(string)
		include, _ := args["include"].(string)
		if pattern == "" {
			return "Error: pattern is required"
		}
		// Use grep for better results when available
		grepArgs := []string{"-rn", "--color=never"}
		if include != "" {
			grepArgs = append(grepArgs, "--include="+include)
		}
		grepArgs = append(grepArgs, pattern)
		if dir == "" {
			dir = "."
		}
		grepArgs = append(grepArgs, dir)
		out, _ := a.runShell("grep", grepArgs...)
		if out == "" {
			return "No matches found."
		}
		// Limit output
		lines := strings.Split(out, "\n")
		if len(lines) > 50 {
			return strings.Join(lines[:50], "\n") + fmt.Sprintf("\n... (%d more lines)", len(lines)-50)
		}
		return out

	case "glob_search":
		pattern, _ := args["pattern"].(string)
		dir, _ := args["path"].(string)
		if pattern == "" {
			return "Error: pattern is required"
		}
		if dir == "" {
			dir = "."
		}
		out, _ := a.runShell("find", dir, "-name", pattern, "-not", "-path", "*/node_modules/*", "-not", "-path", "*/.git/*")
		if out == "" {
			return "No files matched."
		}
		return out

	case "list_files":
		dir, _ := args["path"].(string)
		entries, err := a.ListFiles(dir, 1)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		var sb strings.Builder
		for _, e := range entries {
			if e.IsDir {
				fmt.Fprintf(&sb, "%s/\n", e.Name)
			} else {
				fmt.Fprintf(&sb, "%s\n", e.Name)
			}
		}
		return sb.String()

	case "shell_exec":
		command, _ := args["command"].(string)
		if command == "" {
			return "Error: command is required"
		}
		out, err := a.runShell("bash", "-c", command)
		if err != nil {
			return fmt.Sprintf("Error: %v\nOutput: %s", err, out)
		}
		// Limit output to 10KB
		if len(out) > 10240 {
			out = out[:10240] + "\n... [truncated]"
		}
		return out

	case "apply_patch":
		path, _ := args["path"].(string)
		patch, _ := args["patch"].(string)
		if path == "" || patch == "" {
			return "Error: path and patch are required"
		}
		// Simple approach: write patch to temp file, apply with git apply
		tmpFile := filepath.Join(os.TempDir(), "techai-patch.diff")
		if err := os.WriteFile(tmpFile, []byte(patch), 0644); err != nil {
			return fmt.Sprintf("Error writing patch: %v", err)
		}
		defer os.Remove(tmpFile)
		out, err := a.runShell("git", "apply", "--verbose", tmpFile)
		if err != nil {
			// Fallback: try patch command
			out2, err2 := a.runShell("patch", "-p0", "-i", tmpFile)
			if err2 != nil {
				return fmt.Sprintf("Error applying patch:\ngit apply: %s\npatch: %s", out, out2)
			}
			out = out2
		}
		runtime.EventsEmit(a.ctx, "file:changed", path)
		return fmt.Sprintf("OK: patch applied to %s\n%s", path, out)

	case "git_status":
		info := a.GetGitInfo()
		var sb strings.Builder
		fmt.Fprintf(&sb, "Branch: %s\n", info.Branch)
		if len(info.Changes) == 0 {
			sb.WriteString("Clean — no changes\n")
		} else {
			fmt.Fprintf(&sb, "Changes (%d):\n", len(info.Changes))
			for _, c := range info.Changes {
				fmt.Fprintf(&sb, "  %s %s\n", c.Status, c.File)
			}
		}
		return sb.String()

	case "git_diff":
		diff := a.GitDiff()
		if diff == "" {
			return "No changes."
		}
		if len(diff) > 10240 {
			diff = diff[:10240] + "\n... [truncated]"
		}
		return diff

	case "git_log":
		return a.GitLog(10)

	default:
		return fmt.Sprintf("Error: unknown tool '%s'", name)
	}
}

// runShell executes a command and returns combined output.
func (a *App) runShell(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = a.cwd
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
