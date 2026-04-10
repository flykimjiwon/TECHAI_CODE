package session

import (
	"path/filepath"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestOpen_CreatesSchema(t *testing.T) {
	store := newTestStore(t)
	// Listing an empty store should succeed, not error, and return no rows.
	sessions, err := store.ListSessions(10)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("len(sessions) = %d, want 0", len(sessions))
	}
}

func TestCreateSession_ReturnsID(t *testing.T) {
	store := newTestStore(t)
	id, err := store.CreateSession("first chat", 0, "gpt-oss-120b")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if id <= 0 {
		t.Errorf("id = %d, want > 0", id)
	}
}

func TestAppendAndLoadSession(t *testing.T) {
	store := newTestStore(t)
	id, err := store.CreateSession("chat", 0, "gpt-oss-120b")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful assistant."},
		{Role: openai.ChatMessageRoleUser, Content: "안녕?"},
		{Role: openai.ChatMessageRoleAssistant, Content: "안녕하세요!"},
	}
	for _, m := range msgs {
		if err := store.AppendMessage(id, m); err != nil {
			t.Fatalf("AppendMessage: %v", err)
		}
	}

	meta, loaded, err := store.LoadSession(id)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if meta.ID != id {
		t.Errorf("meta.ID = %d, want %d", meta.ID, id)
	}
	if meta.Title != "chat" {
		t.Errorf("meta.Title = %q, want %q", meta.Title, "chat")
	}
	if meta.Model != "gpt-oss-120b" {
		t.Errorf("meta.Model = %q, want %q", meta.Model, "gpt-oss-120b")
	}
	if len(loaded) != len(msgs) {
		t.Fatalf("len(loaded) = %d, want %d", len(loaded), len(msgs))
	}
	for i, m := range msgs {
		if loaded[i].Role != m.Role {
			t.Errorf("loaded[%d].Role = %q, want %q", i, loaded[i].Role, m.Role)
		}
		if loaded[i].Content != m.Content {
			t.Errorf("loaded[%d].Content = %q, want %q", i, loaded[i].Content, m.Content)
		}
	}
}

func TestAppendMessage_PreservesToolCalls(t *testing.T) {
	store := newTestStore(t)
	id, _ := store.CreateSession("tools", 0, "gpt-oss-120b")

	assistantWithTools := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: "",
		ToolCalls: []openai.ToolCall{
			{
				ID:   "call_abc",
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      "list_files",
					Arguments: `{"path":"."}`,
				},
			},
		},
	}
	toolResp := openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    "file1.go\nfile2.go",
		ToolCallID: "call_abc",
	}

	if err := store.AppendMessage(id, assistantWithTools); err != nil {
		t.Fatalf("AppendMessage assistant: %v", err)
	}
	if err := store.AppendMessage(id, toolResp); err != nil {
		t.Fatalf("AppendMessage tool: %v", err)
	}

	_, loaded, err := store.LoadSession(id)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("len(loaded) = %d, want 2", len(loaded))
	}
	if len(loaded[0].ToolCalls) != 1 {
		t.Fatalf("assistant ToolCalls len = %d, want 1", len(loaded[0].ToolCalls))
	}
	if got := loaded[0].ToolCalls[0].ID; got != "call_abc" {
		t.Errorf("ToolCalls[0].ID = %q, want call_abc", got)
	}
	if got := loaded[0].ToolCalls[0].Function.Name; got != "list_files" {
		t.Errorf("ToolCalls[0].Function.Name = %q, want list_files", got)
	}
	if got := loaded[1].ToolCallID; got != "call_abc" {
		t.Errorf("tool ToolCallID = %q, want call_abc", got)
	}
}

func TestListSessions_MostRecentFirst(t *testing.T) {
	store := newTestStore(t)
	id1, _ := store.CreateSession("first", 0, "model-a")
	id2, _ := store.CreateSession("second", 1, "model-b")
	id3, _ := store.CreateSession("third", 2, "model-c")

	got, err := store.ListSessions(10)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	// Most recent (id3) first
	if got[0].ID != id3 {
		t.Errorf("got[0].ID = %d, want %d", got[0].ID, id3)
	}
	if got[1].ID != id2 {
		t.Errorf("got[1].ID = %d, want %d", got[1].ID, id2)
	}
	if got[2].ID != id1 {
		t.Errorf("got[2].ID = %d, want %d", got[2].ID, id1)
	}
}

func TestListSessions_RespectsLimit(t *testing.T) {
	store := newTestStore(t)
	for i := 0; i < 5; i++ {
		_, _ = store.CreateSession("chat", 0, "m")
	}
	got, err := store.ListSessions(3)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("len = %d, want 3 (limit)", len(got))
	}
}

func TestDeleteSession_RemovesMessages(t *testing.T) {
	store := newTestStore(t)
	id, _ := store.CreateSession("gone", 0, "m")
	_ = store.AppendMessage(id, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser, Content: "bye",
	})

	if err := store.DeleteSession(id); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if _, _, err := store.LoadSession(id); err == nil {
		t.Error("LoadSession after delete returned no error, want ErrNotFound")
	}
}

func TestUpdateSessionTitle(t *testing.T) {
	store := newTestStore(t)
	id, _ := store.CreateSession("untitled", 0, "m")
	if err := store.UpdateSessionTitle(id, "renamed"); err != nil {
		t.Fatalf("UpdateSessionTitle: %v", err)
	}
	meta, _, err := store.LoadSession(id)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if meta.Title != "renamed" {
		t.Errorf("Title = %q, want renamed", meta.Title)
	}
}

func TestAppendMessage_BumpsUpdatedAt(t *testing.T) {
	store := newTestStore(t)
	id, _ := store.CreateSession("bump", 0, "m")

	metaBefore, _, _ := store.LoadSession(id)
	// Append a message — UpdatedAt should move forward (or stay equal on
	// second-resolution clocks, but never go backward).
	_ = store.AppendMessage(id, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser, Content: "hi",
	})
	metaAfter, _, _ := store.LoadSession(id)

	if metaAfter.UpdatedAt.Before(metaBefore.UpdatedAt) {
		t.Errorf("UpdatedAt went backward: before=%v after=%v",
			metaBefore.UpdatedAt, metaAfter.UpdatedAt)
	}
}
