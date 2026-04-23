package main

import (
	"encoding/json"
	"testing"
)

func TestParseTextToolCalls_ToolCallTag(t *testing.T) {
	content := `<tool_call>{"name":"file_read","arguments":{"path":"main.go"}}</tool_call>`
	calls := parseTextToolCalls(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1, got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("name=%q", calls[0].Name)
	}
}

func TestParseTextToolCalls_FunctionTag(t *testing.T) {
	content := `<function=file_read>{"path":"main.go"}</function>`
	calls := parseTextToolCalls(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1, got %d", len(calls))
	}
	if calls[0].Name != "file_read" {
		t.Errorf("name=%q", calls[0].Name)
	}
}

func TestParseTextToolCalls_ParameterFormat(t *testing.T) {
	content := `<function=list_files> <parameter=path> . <parameter=recursive> false </tool_call>`
	calls := parseTextToolCalls(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1, got %d", len(calls))
	}
	if calls[0].Name != "list_files" {
		t.Errorf("name=%q", calls[0].Name)
	}
	var args map[string]string
	json.Unmarshal([]byte(calls[0].Arguments), &args)
	if args["path"] != "." {
		t.Errorf("path=%q", args["path"])
	}
}

func TestParseTextToolCalls_ThinkTagStripped(t *testing.T) {
	content := `<think>I should call <tool_call>fake</tool_call></think>
<tool_call>{"name":"real","arguments":{"x":1}}</tool_call>`
	calls := parseTextToolCalls(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1 (think stripped), got %d", len(calls))
	}
	if calls[0].Name != "real" {
		t.Errorf("name=%q", calls[0].Name)
	}
}

func TestParseTextToolCalls_EscapedArguments(t *testing.T) {
	content := `<tool_call>{"name":"file_read","arguments":"{\"path\":\"main.go\"}"}</tool_call>`
	calls := parseTextToolCalls(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1, got %d", len(calls))
	}
	if calls[0].Arguments != `{"path":"main.go"}` {
		t.Errorf("args=%q", calls[0].Arguments)
	}
}

func TestParseTextToolCalls_Multiple(t *testing.T) {
	content := `<tool_call>{"name":"a","arguments":{}}</tool_call>
<tool_call>{"name":"b","arguments":{}}</tool_call>`
	calls := parseTextToolCalls(content)
	if len(calls) != 2 {
		t.Fatalf("expected 2, got %d", len(calls))
	}
}

func TestParseTextToolCalls_Empty(t *testing.T) {
	calls := parseTextToolCalls("")
	if len(calls) != 0 {
		t.Fatalf("expected 0, got %d", len(calls))
	}
}

func TestParseTextToolCalls_NoTags(t *testing.T) {
	calls := parseTextToolCalls("just regular text")
	if len(calls) != 0 {
		t.Fatalf("expected 0, got %d", len(calls))
	}
}

func TestParseTextToolCalls_TextBeforeJSON(t *testing.T) {
	content := `<tool_call>Sure!
{"name":"test","arguments":{}}</tool_call>`
	calls := parseTextToolCalls(content)
	if len(calls) != 1 {
		t.Fatalf("expected 1, got %d", len(calls))
	}
}

func TestParseParameterTags(t *testing.T) {
	params := parseParameterTags(`<parameter=path> . <parameter=recursive> false`)
	if params["path"] != "." {
		t.Errorf("path=%q", params["path"])
	}
	if params["recursive"] != "false" {
		t.Errorf("recursive=%q", params["recursive"])
	}
}

func TestParseParameterTags_Empty(t *testing.T) {
	params := parseParameterTags("")
	if len(params) != 0 {
		t.Errorf("expected empty, got %v", params)
	}
}

func TestStripThinkTags(t *testing.T) {
	tests := []struct{ in, want string }{
		{"<think>inner</think>outer", "outer"},
		{"before<think>mid</think>after", "beforeafter"},
		{"no tags", "no tags"},
		{"<think>unclosed", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := stripThinkTags(tt.in)
		if got != tt.want {
			t.Errorf("stripThinkTags(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
