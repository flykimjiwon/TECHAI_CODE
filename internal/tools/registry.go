package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
)

type paramSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]propertySchema `json:"properties"`
	Required   []string                  `json:"required"`
}

type propertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// AllTools returns tool definitions for modes with full access (super, dev).
func AllTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "file_read",
				Description: "Read the contents of a file. Use this to understand existing code before making changes.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path": {Type: "string", Description: "File path (relative to cwd or absolute)"},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "file_write",
				Description: "Create a new file or completely overwrite an existing file. Use for new files only. For modifying existing files, prefer file_edit.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":    {Type: "string", Description: "File path to write"},
						"content": {Type: "string", Description: "Complete file content"},
					},
					Required: []string{"path", "content"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "file_edit",
				Description: "Edit an existing file by replacing a specific string. The old_string must match exactly (including whitespace/indentation). Only the first occurrence is replaced.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":       {Type: "string", Description: "File path to edit"},
						"old_string": {Type: "string", Description: "Exact string to find (must match file content exactly)"},
						"new_string": {Type: "string", Description: "Replacement string"},
					},
					Required: []string{"path", "old_string", "new_string"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "list_files",
				Description: "List files in a directory. Use recursive=true to see the full project tree (skips node_modules, .git, dist).",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":      {Type: "string", Description: "Directory path (default: current directory)"},
						"recursive": {Type: "string", Description: "Set to 'true' for recursive listing"},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "shell_exec",
				Description: "Execute a shell command. Use for: git, npm, build, test, lint, etc. Dangerous commands (rm -rf /, sudo) are blocked. Prefer grep_search/glob_search over shell grep/find.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"command": {Type: "string", Description: "Shell command to execute"},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "grep_search",
				Description: "Search file contents by regex pattern. Returns file:line:content matches. Faster and safer than shell grep. Skips binary files, .git, node_modules, dist.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"pattern":       {Type: "string", Description: "Regex pattern to search for (required)"},
						"path":          {Type: "string", Description: "Directory to search in (default: current directory)"},
						"glob":          {Type: "string", Description: "File filter glob (e.g. '*.go', '*.ts')"},
						"ignore_case":   {Type: "string", Description: "Set to 'true' for case-insensitive search"},
						"context_lines": {Type: "string", Description: "Number of context lines around matches (default: 0)"},
					},
					Required: []string{"pattern"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "glob_search",
				Description: "Find files by glob pattern (supports **). Returns matching file paths. Use instead of shell find. Skips .git, node_modules, dist.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"pattern": {Type: "string", Description: "Glob pattern (e.g. '**/*.go', 'src/**/*.ts', '*.json')"},
						"path":    {Type: "string", Description: "Base directory to search in (default: current directory)"},
					},
					Required: []string{"pattern"},
				},
			},
		},
	}
}

// ReadOnlyTools returns tool definitions for plan mode (no file writes).
func ReadOnlyTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "file_read",
				Description: "Read the contents of a file.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path": {Type: "string", Description: "File path to read"},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "list_files",
				Description: "List files in a directory.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":      {Type: "string", Description: "Directory path"},
						"recursive": {Type: "string", Description: "Set to 'true' for recursive listing"},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "shell_exec",
				Description: "Execute read-only shell commands: ls, cat, git log, git status, etc. Prefer grep_search/glob_search over shell grep/find.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"command": {Type: "string", Description: "Shell command (read-only operations only)"},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "grep_search",
				Description: "Search file contents by regex pattern. Returns file:line:content matches.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"pattern":       {Type: "string", Description: "Regex pattern to search for (required)"},
						"path":          {Type: "string", Description: "Directory to search in (default: current directory)"},
						"glob":          {Type: "string", Description: "File filter glob (e.g. '*.go', '*.ts')"},
						"ignore_case":   {Type: "string", Description: "Set to 'true' for case-insensitive search"},
						"context_lines": {Type: "string", Description: "Number of context lines around matches (default: 0)"},
					},
					Required: []string{"pattern"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "glob_search",
				Description: "Find files by glob pattern (supports **). Returns matching file paths.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"pattern": {Type: "string", Description: "Glob pattern (e.g. '**/*.go', 'src/**/*.ts')"},
						"path":    {Type: "string", Description: "Base directory (default: current directory)"},
					},
					Required: []string{"pattern"},
				},
			},
		},
	}
}

// ToolsForMode returns the appropriate tool definitions based on mode.
func ToolsForMode(mode int) []openai.Tool {
	switch mode {
	case 2: // Plan — read-only
		return ReadOnlyTools()
	default: // Super, Dev — full access
		return AllTools()
	}
}

// Execute runs a tool by name with the given JSON arguments and returns the result.
func Execute(name string, argsJSON string) string {
	config.DebugLog("[TOOL-CALL] %s | args=%s", name, argsJSON)

	result := executeInner(name, argsJSON)

	truncated := len(result) > 30000
	config.DebugLog("[TOOL-RESULT] %s | resultLen=%d | truncated=%v", name, len(result), truncated)
	return result
}

func executeInner(name string, argsJSON string) string {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		config.DebugLog("[TOOL-ERR] %s | invalid JSON: %v", name, err)
		return fmt.Sprintf("Error: invalid arguments: %v", err)
	}

	switch name {
	case "file_read":
		path, _ := args["path"].(string)
		if path == "" {
			return "Error: path is required"
		}
		content, err := FileRead(path)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		if len(content) > 50000 {
			return content[:50000] + "\n\n... [truncated, file too large]"
		}
		return content

	case "file_write":
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		if path == "" {
			return "Error: path is required"
		}
		if err := FileWrite(path, content); err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("OK: written %d bytes to %s", len(content), path)

	case "file_edit":
		path, _ := args["path"].(string)
		oldStr, _ := args["old_string"].(string)
		newStr, _ := args["new_string"].(string)
		if path == "" || oldStr == "" {
			return "Error: path and old_string are required"
		}
		count, err := FileEdit(path, oldStr, newStr)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return fmt.Sprintf("OK: replaced %d occurrence(s) in %s", count, path)

	case "list_files":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		recursive := false
		if r, ok := args["recursive"].(string); ok && r == "true" {
			recursive = true
		}
		if r, ok := args["recursive"].(bool); ok && r {
			recursive = true
		}
		files, err := ListFiles(path, recursive)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return strings.Join(files, "\n")

	case "shell_exec":
		command, _ := args["command"].(string)
		if command == "" {
			return "Error: command is required"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err := ShellExec(ctx, command)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		output := result.Stdout
		if result.Stderr != "" {
			output += "\nSTDERR: " + result.Stderr
		}
		if result.ExitCode != 0 {
			output += fmt.Sprintf("\nExit code: %d", result.ExitCode)
		}
		if len(output) > 30000 {
			output = output[:30000] + "\n\n... [truncated]"
		}
		return output

	case "grep_search":
		pattern, _ := args["pattern"].(string)
		if pattern == "" {
			return "Error: pattern is required"
		}
		searchPath, _ := args["path"].(string)
		glob, _ := args["glob"].(string)
		ignoreCase := false
		if ic, ok := args["ignore_case"].(string); ok && ic == "true" {
			ignoreCase = true
		}
		if ic, ok := args["ignore_case"].(bool); ok && ic {
			ignoreCase = true
		}
		contextLines := 0
		if cl, ok := args["context_lines"].(string); ok {
			fmt.Sscanf(cl, "%d", &contextLines)
		}
		if cl, ok := args["context_lines"].(float64); ok {
			contextLines = int(cl)
		}
		result, err := GrepSearch(pattern, searchPath, glob, ignoreCase, contextLines)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	case "glob_search":
		pattern, _ := args["pattern"].(string)
		if pattern == "" {
			return "Error: pattern is required"
		}
		searchPath, _ := args["path"].(string)
		result, err := GlobSearch(pattern, searchPath)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	default:
		config.DebugLog("[TOOL-ERR] unknown tool '%s'", name)
		return fmt.Sprintf("Error: unknown tool '%s'", name)
	}
}
