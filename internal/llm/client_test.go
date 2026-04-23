package llm

import (
	"encoding/json"
	"testing"
)

// ── parseToolCallsFromContent tests ──

func TestParseToolCalls_Pattern1_Basic(t *testing.T) {
	content := `<tool_call>{"name":"file_read","arguments":{"path":"main.go"}}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("expected name=file_read, got %q", calls[0].Name)
	}
	if calls[0].Arguments != `{"path":"main.go"}` {
		t.Errorf("unexpected args: %q", calls[0].Arguments)
	}
}

func TestParseToolCalls_Pattern1_Multiple(t *testing.T) {
	content := `<tool_call>{"name":"file_read","arguments":{"path":"a.go"}}</tool_call>
<tool_call>{"name":"file_write","arguments":{"path":"b.go","content":"x"}}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("call[0] name=%q", calls[0].Name)
	}
	if calls[1].Name != "file_write" {
		t.Errorf("call[1] name=%q", calls[1].Name)
	}
}

func TestParseToolCalls_Pattern1_PipeDelimited(t *testing.T) {
	content := `<|tool_call|>{"name":"shell","arguments":{"cmd":"ls"}}<|/tool_call|>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "shell" {
		t.Errorf("expected name=shell, got %q", calls[0].Name)
	}
}

func TestParseToolCalls_Pattern2_Function(t *testing.T) {
	content := `<function=file_read>{"path":"main.go"}</function>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("expected name=file_read, got %q", calls[0].Name)
	}
}

func TestParseToolCalls_Pattern2_CloseToolCall(t *testing.T) {
	content := `<function=grep>{"pattern":"TODO"}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "grep" {
		t.Errorf("expected name=grep, got %q", calls[0].Name)
	}
}

func TestParseToolCalls_ThinkTagStripped(t *testing.T) {
	content := `<think>I should call file_read <tool_call>not real</tool_call></think>
<tool_call>{"name":"file_read","arguments":{"path":"real.go"}}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call (think stripped), got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("name=%q", calls[0].Name)
	}
	if calls[0].Arguments != `{"path":"real.go"}` {
		t.Errorf("args=%q", calls[0].Arguments)
	}
}

func TestParseToolCalls_ThinkPipeVariant(t *testing.T) {
	content := `<|think|>reasoning here with <tool_call> mention<|/think|>
<tool_call>{"name":"shell","arguments":{"cmd":"pwd"}}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "shell" {
		t.Errorf("name=%q", calls[0].Name)
	}
}

func TestParseToolCalls_InvalidJSON_Skipped(t *testing.T) {
	content := `<tool_call>not json at all</tool_call>
<tool_call>{"name":"valid","arguments":{"x":1}}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 valid call, got %d", len(calls))
	}
	if calls[0].Name != "valid" {
		t.Errorf("name=%q", calls[0].Name)
	}
}

func TestParseToolCalls_EmptyName_Skipped(t *testing.T) {
	content := `<tool_call>{"name":"","arguments":{}}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 0 {
		t.Fatalf("expected 0 calls (empty name), got %d", len(calls))
	}
}

func TestParseToolCalls_MissingArguments_DefaultsEmpty(t *testing.T) {
	content := `<tool_call>{"name":"test"}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Arguments != "{}" {
		t.Errorf("expected empty args {}, got %q", calls[0].Arguments)
	}
}

func TestParseToolCalls_NullArguments_DefaultsEmpty(t *testing.T) {
	content := `<tool_call>{"name":"test","arguments":null}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Arguments != "{}" {
		t.Errorf("expected {}, got %q", calls[0].Arguments)
	}
}

func TestParseToolCalls_TextBeforeTag(t *testing.T) {
	content := `I'll read the file for you.

<tool_call>{"name":"file_read","arguments":{"path":"go.mod"}}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
}

func TestParseToolCalls_UnclosedTag_StillParses(t *testing.T) {
	content := `<tool_call>{"name":"file_read","arguments":{"path":"a.go"}}`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call from unclosed tag, got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("name=%q", calls[0].Name)
	}
}

func TestParseToolCalls_EmptyFunction_Skipped(t *testing.T) {
	content := `<function=>{"path":"x"}</function>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 0 {
		t.Fatalf("expected 0 calls (empty function name), got %d", len(calls))
	}
}

func TestParseToolCalls_FunctionInvalidJSON_Skipped(t *testing.T) {
	content := `<function=test>not json</function>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 0 {
		t.Fatalf("expected 0 calls (invalid JSON), got %d", len(calls))
	}
}

func TestParseToolCalls_EmptyContent(t *testing.T) {
	calls := parseToolCallsFromContent("")
	if len(calls) != 0 {
		t.Fatalf("expected 0 calls, got %d", len(calls))
	}
}

func TestParseToolCalls_NoTags(t *testing.T) {
	calls := parseToolCallsFromContent("Just normal text without any tool calls.")
	if len(calls) != 0 {
		t.Fatalf("expected 0 calls, got %d", len(calls))
	}
}

func TestParseToolCalls_MultilineJSON(t *testing.T) {
	content := `<tool_call>
{
  "name": "file_write",
  "arguments": {
    "path": "test.go",
    "content": "package main\n\nfunc main() {}\n"
  }
}
</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "file_write" {
		t.Errorf("name=%q", calls[0].Name)
	}
}

// ── StripToolCallTags tests ──

func TestStripToolCallTags_Basic(t *testing.T) {
	input := `Here is the result.<tool_call>{"name":"test","arguments":{}}</tool_call>`
	got := StripToolCallTags(input)
	if got != "Here is the result." {
		t.Errorf("expected clean text, got %q", got)
	}
}

func TestStripToolCallTags_WithThink(t *testing.T) {
	input := `<think>reasoning</think>Text here.<tool_call>{"name":"x","arguments":{}}</tool_call>`
	got := StripToolCallTags(input)
	if got != "Text here." {
		t.Errorf("expected 'Text here.', got %q", got)
	}
}

func TestStripToolCallTags_FunctionTag(t *testing.T) {
	input := `Reading file.<function=file_read>{"path":"a.go"}</function>`
	got := StripToolCallTags(input)
	if got != "Reading file." {
		t.Errorf("expected 'Reading file.', got %q", got)
	}
}

func TestStripToolCallTags_NoTags(t *testing.T) {
	input := "Just text"
	got := StripToolCallTags(input)
	if got != "Just text" {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestStripToolCallTags_Empty(t *testing.T) {
	got := StripToolCallTags("")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// ── stripThinkTags tests ──

func TestStripThinkTags_Basic(t *testing.T) {
	got := stripThinkTags("<think>inner</think>outer")
	if got != "outer" {
		t.Errorf("expected 'outer', got %q", got)
	}
}

func TestStripThinkTags_Unclosed(t *testing.T) {
	got := stripThinkTags("before<think>after")
	if got != "before" {
		t.Errorf("expected 'before', got %q", got)
	}
}

func TestStripThinkTags_Multiple(t *testing.T) {
	got := stripThinkTags("<think>a</think>mid<think>b</think>end")
	if got != "midend" {
		t.Errorf("expected 'midend', got %q", got)
	}
}

func TestStripThinkTags_PipeVariant(t *testing.T) {
	got := stripThinkTags("<|think|>inner<|/think|>outer")
	if got != "outer" {
		t.Errorf("expected 'outer', got %q", got)
	}
}

// ── findToolCallTagStart tests ──

func TestFindToolCallTagStart_Found(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"abc<tool_call>rest", 3},
		{"abc<function=test>rest", 3},
		{"abc<|tool_call|>rest", 3},
		{"no tags here", -1},
		{"", -1},
	}
	for _, tt := range tests {
		got := findToolCallTagStart(tt.input)
		if got != tt.want {
			t.Errorf("findToolCallTagStart(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestFindToolCallTagStart_EarliestMatch(t *testing.T) {
	// <function= appears before <tool_call>
	input := "x<function=a>y<tool_call>z"
	got := findToolCallTagStart(input)
	if got != 1 {
		t.Errorf("expected 1 (function= is earlier), got %d", got)
	}
}

// ── partialToolTagSuffix tests ──

func TestPartialToolTagSuffix_NoMatch(t *testing.T) {
	if got := partialToolTagSuffix("hello world"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestPartialToolTagSuffix_FullTag_NoHold(t *testing.T) {
	// Full tag is detected by findToolCallTagStart, not by suffix hold
	if got := partialToolTagSuffix("abc<tool_call>"); got != 0 {
		t.Errorf("full tag should not be held, got %d", got)
	}
}

func TestPartialToolTagSuffix_PartialToolCall(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"some text<", 1},
		{"some text<t", 2},
		{"some text<to", 3},
		{"some text<too", 4},
		{"some text<tool", 5},
		{"some text<tool_", 6},
		{"some text<tool_c", 7},
		{"some text<tool_ca", 8},
		{"some text<tool_cal", 9},
		{"some text<tool_call", 10},
	}
	for _, tt := range tests {
		got := partialToolTagSuffix(tt.input)
		if got != tt.want {
			t.Errorf("partialToolTagSuffix(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestPartialToolTagSuffix_PartialFunction(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"text<f", 2},
		{"text<fu", 3},
		{"text<fun", 4},
		{"text<func", 5},
		{"text<funct", 6},
		{"text<functi", 7},
		{"text<functio", 8},
		{"text<function", 9},
	}
	for _, tt := range tests {
		got := partialToolTagSuffix(tt.input)
		if got != tt.want {
			t.Errorf("partialToolTagSuffix(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestPartialToolTagSuffix_PipeVariant(t *testing.T) {
	got := partialToolTagSuffix("text<|tool")
	if got != 6 { // "<|tool" = 6 bytes
		t.Errorf("expected 6, got %d", got)
	}
}

func TestPartialToolTagSuffix_ThinkTag(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"text</thi", 5},   // partial </think>
		{"text<thin", 5},   // partial <think>
		{"text</think", 7}, // partial </think> (missing >)
	}
	for _, tt := range tests {
		got := partialToolTagSuffix(tt.input)
		if got != tt.want {
			t.Errorf("partialToolTagSuffix(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

// ── arguments as escaped string ──

func TestParseToolCalls_ArgumentsAsEscapedString(t *testing.T) {
	// Some models output arguments as a JSON string instead of object
	content := `<tool_call>{"name":"file_read","arguments":"{\"path\":\"main.go\"}"}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Arguments != `{"path":"main.go"}` {
		t.Errorf("expected unescaped args, got %q", calls[0].Arguments)
	}
}

// ── text before JSON in tool_call ──

// ── <parameter=key> format (Qwen3 onprem variant) ──

func TestParseToolCalls_ParameterFormat(t *testing.T) {
	content := `<function=list_files> <parameter=path> . <parameter=recursive> false </tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "list_files" {
		t.Errorf("name=%q", calls[0].Name)
	}
	// Verify args were converted to JSON
	var args map[string]string
	if err := json.Unmarshal([]byte(calls[0].Arguments), &args); err != nil {
		t.Fatalf("args not valid JSON: %v | raw=%q", err, calls[0].Arguments)
	}
	if args["path"] != "." {
		t.Errorf("path=%q, want '.'", args["path"])
	}
	if args["recursive"] != "false" {
		t.Errorf("recursive=%q, want 'false'", args["recursive"])
	}
}

func TestParseToolCalls_ParameterFormatWithClosingTags(t *testing.T) {
	content := `<function=file_read><parameter=path>/src/main.go</parameter></function>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("name=%q", calls[0].Name)
	}
	var args map[string]string
	if err := json.Unmarshal([]byte(calls[0].Arguments), &args); err != nil {
		t.Fatalf("args not valid JSON: %v", err)
	}
	if args["path"] != "/src/main.go" {
		t.Errorf("path=%q", args["path"])
	}
}

func TestParseToolCalls_ParameterFormatMultipleParams(t *testing.T) {
	content := `<function=grep_search> <parameter=pattern> TODO <parameter=path> ./src <parameter=include> *.go </function>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	var args map[string]string
	json.Unmarshal([]byte(calls[0].Arguments), &args)
	if args["pattern"] != "TODO" {
		t.Errorf("pattern=%q", args["pattern"])
	}
	if args["path"] != "./src" {
		t.Errorf("path=%q", args["path"])
	}
	if args["include"] != "*.go" {
		t.Errorf("include=%q", args["include"])
	}
}

func TestParseParameterTags_Empty(t *testing.T) {
	got := parseParameterTags("")
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestParseParameterTags_NoParameterTags(t *testing.T) {
	got := parseParameterTags("just plain text")
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestParseToolCalls_TextBeforeJSON(t *testing.T) {
	content := `<tool_call>Sure, I'll read that file for you!
{"name":"file_read","arguments":{"path":"test.go"}}</tool_call>`
	calls := parseToolCallsFromContent(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("name=%q", calls[0].Name)
	}
}
