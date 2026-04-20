package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/knowledge"
	"github.com/kimjiwon/tgc/internal/mcp"
)

// MCPTools holds openai.Tool definitions registered from MCP servers.
var MCPTools []openai.Tool

// MCPManager is the global MCP manager used for routing tool calls.
var MCPManager *mcp.Manager

// RegisterMCPTools converts MCP tools to openai.Tool format and appends to MCPTools.
func RegisterMCPTools(tools []mcp.MCPTool) {
	for _, t := range tools {
		schema := t.InputSchema
		if schema == nil {
			schema = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
		}
		MCPTools = append(MCPTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  schema,
			},
		})
	}
	config.DebugLog("[TOOLS] registered %d MCP tools", len(tools))
}

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
// MCP tools are appended at the end when registered.
func AllTools() []openai.Tool {
	base := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "file_read",
				Description: "Read file contents. Use offset and limit to read specific sections — e.g. after grep_search finds a match at line 150, use offset=140 limit=30 to see context around it.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":   {Type: "string", Description: "File path (relative to cwd or absolute)"},
						"offset": {Type: "string", Description: "Line number to start from (1-indexed, default: 1)"},
						"limit":  {Type: "string", Description: "Max lines to read (default: all, max: 2000)"},
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
				Description: "Execute a shell command. Use for: git, npm, build, test, lint, etc. Dangerous commands (rm -rf /, sudo) are blocked. Prefer grep_search/glob_search over shell grep/find. Set timeout for long-running commands (builds, tests).",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"command": {Type: "string", Description: "Shell command to execute"},
						"timeout": {Type: "string", Description: "Timeout in seconds (default: 30, max: 300). Use higher values for builds/tests."},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "grep_search",
				Description: "Fast content search by regex. Returns matches grouped by file with context lines. When the user mentions multiple terms (e.g. TABLE + COLUMN), search EACH term separately and cross-reference the files. Use file_read with offset to examine matches. If pattern A.*B returns no matches, the system automatically finds files containing both A and B on different lines.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"pattern":       {Type: "string", Description: "Regex pattern to search for (required)"},
						"path":          {Type: "string", Description: "Directory to search in (default: current directory)"},
						"include":       {Type: "string", Description: "File pattern to include (e.g. '*.go', '*.sh', '*.{ts,tsx}')"},
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
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "hashline_read",
				Description: "Read a file with hash-anchored line numbers (e.g. '1#a3f1| code'). Each line gets a 4-char MD5 hash. Use with hashline_edit for safe edits.",
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
				Name:        "hashline_edit",
				Description: "Edit a file using hash anchors for stale-edit protection. Anchors are 'N#hash' format from hashline_read. Verifies hash before replacing.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":          {Type: "string", Description: "File path to edit"},
						"start_anchor":  {Type: "string", Description: "Start anchor (e.g. '3#e4d9')"},
						"end_anchor":    {Type: "string", Description: "End anchor (e.g. '5#b2c1')"},
						"new_content":   {Type: "string", Description: "Replacement content for the line range"},
					},
					Required: []string{"path", "start_anchor", "end_anchor", "new_content"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "git_status",
				Description: "Show git status (short format) for the working directory.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path": {Type: "string", Description: "Directory path (default: current directory)"},
					},
					Required: []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "git_diff",
				Description: "Show git diff. Set staged='true' for staged changes only.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":   {Type: "string", Description: "Directory path (default: current directory)"},
						"staged": {Type: "string", Description: "Set to 'true' for staged diff"},
					},
					Required: []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "git_log",
				Description: "Show recent git commits in oneline format.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path": {Type: "string", Description: "Directory path (default: current directory)"},
						"n":    {Type: "string", Description: "Number of commits to show (default: 10)"},
					},
					Required: []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "diagnostics",
				Description: "Auto-detect project type (Go/TS/JS/Python) and run linters. Returns structured diagnostic output with file, line, severity.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path": {Type: "string", Description: "Project directory (default: current directory)"},
						"file": {Type: "string", Description: "Filter results to a specific file (optional)"},
					},
					Required: []string{},
				},
			},
		},
		// NOTE: symbol_search, co_search, fuzzy_find are kept as internal tools
		// (execution handlers remain) but hidden from the model to reduce
		// decision paralysis. co_search is auto-triggered by grep fallback.
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "knowledge_search",
				Description: "Search user knowledge docs (.tgc/knowledge/ or ~/.tgc/knowledge/). Returns matching documents with excerpts. Use when the user's question touches project-specific or framework-specific topics listed in the User Knowledge section of the system prompt.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"query":       {Type: "string", Description: "Search query (keywords, AND match)"},
						"max_results": {Type: "string", Description: "Max results to return (default: 3)"},
					},
					Required: []string{"query"},
				},
			},
		},
	}
	return append(base, MCPTools...)
}

// ReadOnlyTools returns tool definitions for plan mode (no file writes).
func ReadOnlyTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "file_read",
				Description: "Read file contents. Use offset and limit to read specific sections.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":   {Type: "string", Description: "File path to read"},
						"offset": {Type: "string", Description: "Line number to start from (1-indexed)"},
						"limit":  {Type: "string", Description: "Max lines to read"},
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
				Description: "Fast content search. Returns file paths and line numbers sorted by modification time. Use file_read with offset to examine matches.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"pattern":       {Type: "string", Description: "Regex pattern to search for (required)"},
						"path":          {Type: "string", Description: "Directory to search in (default: current directory)"},
						"include":       {Type: "string", Description: "File pattern to include (e.g. '*.go', '*.sh', '*.{ts,tsx}')"},
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
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "git_status",
				Description: "Show git status (short format).",
				Parameters: paramSchema{
					Type:       "object",
					Properties: map[string]propertySchema{"path": {Type: "string", Description: "Directory path"}},
					Required:   []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "git_diff",
				Description: "Show git diff.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path":   {Type: "string", Description: "Directory path"},
						"staged": {Type: "string", Description: "Set to 'true' for staged diff"},
					},
					Required: []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "git_log",
				Description: "Show recent git commits.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path": {Type: "string", Description: "Directory path"},
						"n":    {Type: "string", Description: "Number of commits (default: 10)"},
					},
					Required: []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "diagnostics",
				Description: "Auto-detect project type and run linters.",
				Parameters: paramSchema{
					Type: "object",
					Properties: map[string]propertySchema{
						"path": {Type: "string", Description: "Project directory"},
						"file": {Type: "string", Description: "Filter to specific file"},
					},
					Required: []string{},
				},
			},
		},
	}
}

// ToolsForMode returns the appropriate tool definitions based on mode.
// All three modes (Super, Deep Agent, Plan) now share the full tool set,
// matching hanimo's design where Plan mode has write access for execution
// after the user approves the plan.
func ToolsForMode(mode int) []openai.Tool {
	return AllTools()
}

// failedPatterns tracks grep patterns that already returned no matches.
var (
	failedPatterns   = make(map[string]int)
	failedPatternsMu sync.Mutex
	// userKeywords stores extracted keywords from the latest user message.
	// Used to auto-cross-reference when grep only searches 1 of N keywords.
	userKeywords   []string
	userKeywordsMu sync.Mutex
)

// ResetFailedPatterns clears the cache (call on new user message).
func ResetFailedPatterns() {
	failedPatternsMu.Lock()
	failedPatterns = make(map[string]int)
	failedPatternsMu.Unlock()
}

// SetUserContext extracts searchable keywords from the user message.
// Called from app.go sendMessage() so grep can auto-cross-reference.
func SetUserContext(message string) {
	userKeywordsMu.Lock()
	defer userKeywordsMu.Unlock()
	userKeywords = extractSearchTerms(message)
	if len(userKeywords) > 0 {
		config.DebugLog("[USER-CTX] extracted keywords: %v", userKeywords)
	}
}

// identifierRe matches UPPER_CASE_IDENTIFIERS anywhere in text (not just space-separated).
// Handles: "RWA_IBS_DMB_CMM_MAS테이블의 DMB_K컬럼" → ["RWA_IBS_DMB_CMM_MAS", "DMB_K"]
var identifierRe = regexp.MustCompile(`[A-Z][A-Z0-9_]{2,}`)

// extractSearchTerms pulls out UPPER_CASE identifiers from anywhere in the message.
// Uses regex so it works even when Korean text is attached without spaces.
func extractSearchTerms(msg string) []string {
	matches := identifierRe.FindAllString(msg, -1)
	var terms []string
	seen := make(map[string]bool)
	for _, m := range matches {
		// Must contain underscore (to distinguish from regular words like "SQL", "API")
		if !strings.Contains(m, "_") {
			continue
		}
		if !seen[m] {
			seen[m] = true
			terms = append(terms, m)
		}
	}
	return terms
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

		// Apply offset/limit for targeted reading (like OpenCode's read tool)
		offset := 0
		limit := 0
		if os, ok := args["offset"].(string); ok {
			fmt.Sscanf(os, "%d", &offset)
		}
		if os, ok := args["offset"].(float64); ok {
			offset = int(os)
		}
		if ls, ok := args["limit"].(string); ok {
			fmt.Sscanf(ls, "%d", &limit)
		}
		if ls, ok := args["limit"].(float64); ok {
			limit = int(ls)
		}

		if offset > 0 || limit > 0 {
			lines := strings.Split(content, "\n")
			start := offset - 1 // 1-indexed to 0-indexed
			if start < 0 {
				start = 0
			}
			if start >= len(lines) {
				return fmt.Sprintf("Error: offset %d exceeds file length (%d lines)", offset, len(lines))
			}
			end := len(lines)
			if limit > 0 {
				end = start + limit
				if end > len(lines) {
					end = len(lines)
				}
			}
			content = strings.Join(lines[start:end], "\n")
			content = fmt.Sprintf("[lines %d-%d of %d]\n%s", start+1, end, len(lines), content)
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
		// Check sensitive file and secrets
		if warn := CheckSensitiveFile(path); warn != "" {
			return warn + " — writing blocked. Use shell_exec if you really need to write this file."
		}
		secretWarn := CheckSecrets(content)
		if err := FileWrite(path, content); err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		result := fmt.Sprintf("OK: written %d bytes to %s", len(content), path)
		if secretWarn != "" {
			result += "\n" + secretWarn
		}
		return result

	case "file_edit":
		path, _ := args["path"].(string)
		oldStr, _ := args["old_string"].(string)
		newStr, _ := args["new_string"].(string)
		if path == "" || oldStr == "" {
			return "Error: path and old_string are required"
		}
		count, diff, err := FileEdit(path, oldStr, newStr)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		result := fmt.Sprintf("OK: replaced %d occurrence(s) in %s", count, path)
		if diff != "" {
			result += "\n\n" + diff
		}
		return result

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
		// Configurable timeout (default 30s, max 300s)
		timeoutSec := 30
		if ts, ok := args["timeout"].(string); ok {
			fmt.Sscanf(ts, "%d", &timeoutSec)
		}
		if tf, ok := args["timeout"].(float64); ok {
			timeoutSec = int(tf)
		}
		if timeoutSec > 300 {
			timeoutSec = 300
		}
		if timeoutSec < 1 {
			timeoutSec = 30
		}
		// Prepend risky warning if applicable
		warning := CheckRisky(command)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
		defer cancel()
		result, err := ShellExec(ctx, command)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		output := ""
		if warning != "" {
			output = warning + "\n\n"
		}
		output += result.Stdout
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

		// Block repeated failed patterns (prevents infinite loop)
		failedPatternsMu.Lock()
		cacheKey := pattern
		if failedPatterns[cacheKey] >= 2 {
			failedPatternsMu.Unlock()
			config.DebugLog("[GREP-DEDUP] blocked repeated pattern: %q (failed %d times)", pattern, failedPatterns[cacheKey])
			return fmt.Sprintf("Already searched %q — no matches found. Try a DIFFERENT keyword or use file_read to examine specific files.", pattern)
		}
		failedPatternsMu.Unlock()

		searchPath, _ := args["path"].(string)
		glob, _ := args["include"].(string)
		if glob == "" {
			glob, _ = args["glob"].(string) // backward compat
		}
		ignoreCase := false
		if ic, ok := args["ignore_case"].(string); ok && ic == "true" {
			ignoreCase = true
		}
		if ic, ok := args["ignore_case"].(bool); ok && ic {
			ignoreCase = true
		}
		contextLines := 2 // default: show 2 surrounding lines for context
		if cl, ok := args["context_lines"].(string); ok {
			fmt.Sscanf(cl, "%d", &contextLines)
		}
		if cl, ok := args["context_lines"].(float64); ok {
			contextLines = int(cl)
		}
		// Force-split complex patterns (A.*B, A|B, A.*B|B.*A) — extract unique terms
		if strings.Contains(pattern, ".*") || strings.Contains(pattern, "|") {
			// Split on both | and .* to extract individual identifiers
			normalized := strings.ReplaceAll(pattern, "|", " ")
			normalized = strings.ReplaceAll(normalized, ".*", " ")
			parts := strings.Fields(normalized)
			seen := make(map[string]bool)
			var validTerms []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" && len(p) > 2 && !seen[strings.ToUpper(p)] {
					seen[strings.ToUpper(p)] = true
					validTerms = append(validTerms, p)
				}
			}
			if len(validTerms) >= 2 {
				config.DebugLog("[GREP-FORCE-SPLIT] extracting %d unique terms from pattern: %v", len(validTerms), validTerms)
				coResult, coErr := CoSearch(validTerms, searchPath, glob, ignoreCase)
				if coErr == nil && !strings.HasPrefix(coResult, "No files") {
					return coResult
				}
				// co_search no intersection — search each individually
				var sb strings.Builder
				sb.WriteString("Individual search results:\n\n")
				for _, term := range validTerms {
					termResult, _ := GrepSearch(term, searchPath, glob, ignoreCase, 2)
					if !strings.HasPrefix(termResult, "No matches") {
						sb.WriteString(fmt.Sprintf("--- %q ---\n%s\n", term, termResult))
					} else {
						sb.WriteString(fmt.Sprintf("--- %q --- No matches\n\n", term))
					}
				}
				return sb.String()
			}
		}

		// Try ripgrep first (100x faster), fall back to Go implementation
		if result, err := RipgrepSearch(pattern, searchPath, glob, ignoreCase, contextLines); err == nil {
			if !strings.HasPrefix(result, "No matches") {
				return result
			}
			// ripgrep found nothing — fall through to smart fallback
		}
		result, err := GrepSearch(pattern, searchPath, glob, ignoreCase, contextLines)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}

		// Track failed patterns to prevent loops
		if strings.HasPrefix(result, "No matches") {
			failedPatternsMu.Lock()
			failedPatterns[cacheKey]++
			failedPatternsMu.Unlock()
		}

		// Auto cross-reference: if user mentioned 2+ keywords but grep only searched 1,
		// auto-search remaining keywords and show intersection
		if !strings.HasPrefix(result, "No matches") {
			userKeywordsMu.Lock()
			remaining := userKeywords
			userKeywordsMu.Unlock()

			if len(remaining) >= 2 {
				// Check if current pattern matches one of the user keywords
				patternUpper := strings.ToUpper(pattern)
				var otherTerms []string
				matchedOne := false
				for _, kw := range remaining {
					if strings.Contains(patternUpper, strings.ToUpper(kw)) || strings.Contains(strings.ToUpper(kw), patternUpper) {
						matchedOne = true
					} else {
						otherTerms = append(otherTerms, kw)
					}
				}
				// If we matched one keyword and there are others unsearched, auto co_search
				if matchedOne && len(otherTerms) > 0 {
					allTerms := append([]string{pattern}, otherTerms...)
					config.DebugLog("[GREP-XREF] auto cross-reference: searched=%q, adding=%v", pattern, otherTerms)
					coResult, coErr := CoSearch(allTerms, searchPath, glob, ignoreCase)
					if coErr == nil && !strings.HasPrefix(coResult, "No files") {
						result += "\n\n[Auto cross-reference with user keywords]\n" + coResult
					}
					// Clear keywords so we don't re-cross-reference on next grep
					userKeywordsMu.Lock()
					userKeywords = nil
					userKeywordsMu.Unlock()
				}
			}
		}

		return result

	case "glob_search":
		pattern, _ := args["pattern"].(string)
		if pattern == "" {
			return "Error: pattern is required"
		}
		searchPath, _ := args["path"].(string)
		// Try ripgrep first for fast file listing
		if result, err := RipgrepFiles(pattern, searchPath); err == nil {
			return result
		}
		result, err := GlobSearch(pattern, searchPath)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	case "hashline_read":
		path, _ := args["path"].(string)
		if path == "" {
			return "Error: path is required"
		}
		content, err := HashlineRead(path)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		if len(content) > 50000 {
			return content[:50000] + "\n\n... [truncated]"
		}
		return content

	case "hashline_edit":
		path, _ := args["path"].(string)
		start, _ := args["start_anchor"].(string)
		end, _ := args["end_anchor"].(string)
		newContent, _ := args["new_content"].(string)
		if path == "" || start == "" || end == "" {
			return "Error: path, start_anchor, and end_anchor are required"
		}
		result, err := HashlineEdit(path, start, end, newContent)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	case "git_status":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		result, err := GitStatus(path)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		if result == "" {
			return "clean working tree"
		}
		return result

	case "git_diff":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		staged := false
		if s, ok := args["staged"].(string); ok && s == "true" {
			staged = true
		}
		result, err := GitDiff(path, staged)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		if result == "" {
			return "no changes"
		}
		if len(result) > 30000 {
			return result[:30000] + "\n\n... [truncated]"
		}
		return result

	case "git_log":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		n := 10
		if ns, ok := args["n"].(string); ok {
			fmt.Sscanf(ns, "%d", &n)
		}
		if nf, ok := args["n"].(float64); ok {
			n = int(nf)
		}
		result, err := GitLog(path, n)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	case "diagnostics":
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		file, _ := args["file"].(string)
		result, err := RunDiagnostics(path, file)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	case "symbol_search":
		query, _ := args["query"].(string)
		if query == "" {
			return "Error: query is required"
		}
		searchPath, _ := args["path"].(string)
		result, err := SymbolSearch(query, searchPath)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	case "co_search":
		termsStr, _ := args["terms"].(string)
		if termsStr == "" {
			return "Error: terms is required (comma-separated, e.g. 'TABLE_NAME,COLUMN_NAME')"
		}
		terms := strings.Split(termsStr, ",")
		for i := range terms {
			terms[i] = strings.TrimSpace(terms[i])
		}
		searchPath, _ := args["path"].(string)
		glob, _ := args["include"].(string)
		if glob == "" {
			glob, _ = args["glob"].(string) // backward compat
		}
		ignoreCase := false
		if ic, ok := args["ignore_case"].(string); ok && ic == "true" {
			ignoreCase = true
		}
		result, err := CoSearch(terms, searchPath, glob, ignoreCase)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	case "fuzzy_find":
		query, _ := args["query"].(string)
		if query == "" {
			return "Error: query is required"
		}
		searchPath, _ := args["path"].(string)
		result, err := FuzzyFind(query, searchPath, 20)
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return result

	case "knowledge_search":
		query, _ := args["query"].(string)
		if query == "" {
			return "Error: query is required"
		}
		maxResults := 3
		if ms, ok := args["max_results"].(string); ok {
			fmt.Sscanf(ms, "%d", &maxResults)
		}
		if mf, ok := args["max_results"].(float64); ok {
			maxResults = int(mf)
		}
		return ExecuteKnowledgeSearch(query, maxResults)

	default:
		if strings.HasPrefix(name, "mcp_") && MCPManager != nil {
			result, err := MCPManager.CallTool(name, args)
			if err != nil {
				config.DebugLog("[TOOL-MCP] call error tool=%s: %v", name, err)
				return fmt.Sprintf("Error: %v", err)
			}
			return result
		}
		config.DebugLog("[TOOL-ERR] unknown tool '%s'", name)
		return fmt.Sprintf("Error: unknown tool '%s'", name)
	}
}

// ExecuteKnowledgeSearch searches the user's knowledge docs index.
func ExecuteKnowledgeSearch(query string, maxResults int) string {
	idx := knowledge.GlobalIndex
	if idx == nil || idx.Count() == 0 {
		return "User knowledge 폴더가 없습니다. .tgc/knowledge/ 또는 ~/.tgc/knowledge/ 에 .md/.txt 파일을 넣으세요."
	}
	docs := idx.Search(query, maxResults)
	return knowledge.FormatSearchResults(docs, query)
}
