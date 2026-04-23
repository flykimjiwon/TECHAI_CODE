package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parsedToolCall holds a tool call parsed from text content.
type parsedToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// parseTextToolCalls extracts tool calls from text content when the API proxy
// doesn't convert model-native format to OpenAI tool_calls.
// Ported from techai TUI's parseToolCallsFromContent.
//
// Supported formats:
//   - <tool_call>{"name":"...","arguments":{...}}</tool_call>
//   - <function=name>{"key":"val"}</function>
//   - <function=name> <parameter=key> value </function>
func parseTextToolCalls(content string) []parsedToolCall {
	// Strip <think>...</think> blocks
	content = stripThinkTags(content)
	if content == "" {
		return nil
	}

	var calls []parsedToolCall

	// Pattern 1: <tool_call>...</tool_call>
	content = parseTagPair(content, "<tool_call>", "</tool_call>", &calls)

	// Pattern 2: <function=name>...</function> or </tool_call>
	remaining := content
	for {
		funcStart := strings.Index(remaining, "<function=")
		if funcStart == -1 {
			break
		}
		nameEnd := strings.Index(remaining[funcStart:], ">")
		if nameEnd == -1 {
			break
		}
		name := strings.TrimSpace(remaining[funcStart+len("<function=") : funcStart+nameEnd])
		if name == "" {
			remaining = remaining[funcStart+nameEnd+1:]
			continue
		}

		closeTag := "</function>"
		funcEnd := strings.Index(remaining[funcStart:], closeTag)
		if funcEnd == -1 {
			closeTag = "</tool_call>"
			funcEnd = strings.Index(remaining[funcStart:], closeTag)
			if funcEnd == -1 {
				break
			}
		}
		argsStr := strings.TrimSpace(remaining[funcStart+nameEnd+1 : funcStart+funcEnd])
		remaining = remaining[funcStart+funcEnd+len(closeTag):]

		if len(argsStr) == 0 {
			argsStr = "{}"
		}

		// Try JSON first
		if argsStr[0] == '{' {
			if json.Valid([]byte(argsStr)) {
				calls = append(calls, parsedToolCall{
					ID: fmt.Sprintf("text-tc-%d", len(calls)), Name: name, Arguments: argsStr,
				})
				continue
			}
		}

		// Try <parameter=key> value format
		if params := parseParameterTags(argsStr); len(params) > 0 {
			pj, _ := json.Marshal(params)
			calls = append(calls, parsedToolCall{
				ID: fmt.Sprintf("text-tc-%d", len(calls)), Name: name, Arguments: string(pj),
			})
		}
	}

	return calls
}

func parseTagPair(content, openTag, closeTag string, calls *[]parsedToolCall) string {
	for {
		start := strings.Index(content, openTag)
		if start == -1 {
			break
		}
		end := strings.Index(content[start:], closeTag)
		if end == -1 {
			// Unclosed — try to recover
			jsonStr := strings.TrimSpace(content[start+len(openTag):])
			if idx := strings.Index(jsonStr, "{"); idx >= 0 {
				jsonStr = jsonStr[idx:]
			}
			if tc, ok := tryParseToolJSON(jsonStr, len(*calls)); ok {
				*calls = append(*calls, tc)
			}
			content = ""
			break
		}
		jsonStr := strings.TrimSpace(content[start+len(openTag) : start+end])
		content = content[start+end+len(closeTag):]

		// Skip preamble text before JSON
		if idx := strings.Index(jsonStr, "{"); idx > 0 {
			jsonStr = jsonStr[idx:]
		}

		if tc, ok := tryParseToolJSON(jsonStr, len(*calls)); ok {
			*calls = append(*calls, tc)
		}
	}
	return content
}

func tryParseToolJSON(jsonStr string, idx int) (parsedToolCall, bool) {
	if len(jsonStr) == 0 || jsonStr[0] != '{' {
		return parsedToolCall{}, false
	}
	var parsed struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil || parsed.Name == "" {
		return parsedToolCall{}, false
	}
	args := string(parsed.Arguments)
	if args == "" || args == "null" {
		args = "{}"
	}
	// Handle escaped string arguments
	if len(args) >= 2 && args[0] == '"' {
		var unescaped string
		if err := json.Unmarshal([]byte(args), &unescaped); err == nil && len(unescaped) > 0 && unescaped[0] == '{' {
			args = unescaped
		}
	}
	return parsedToolCall{
		ID: fmt.Sprintf("text-tc-%d", idx), Name: parsed.Name, Arguments: args,
	}, true
}

func parseParameterTags(s string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(s, "<parameter=")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		closeIdx := strings.Index(part, ">")
		if closeIdx == -1 {
			continue
		}
		key := strings.TrimSpace(part[:closeIdx])
		value := strings.TrimSpace(part[closeIdx+1:])
		if idx := strings.Index(value, "</parameter>"); idx >= 0 {
			value = strings.TrimSpace(value[:idx])
		}
		if key != "" && value != "" {
			params[key] = value
		}
	}
	return params
}

func stripThinkTags(s string) string {
	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "</think>")
		if end == -1 {
			s = s[:start]
			break
		}
		s = s[:start] + s[start+end+len("</think>"):]
	}
	return s
}
