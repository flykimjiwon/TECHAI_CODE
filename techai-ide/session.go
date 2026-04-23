package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ChatSession represents a saved chat session.
type ChatSession struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	Messages  int    `json:"messages"`
}

func sessionsDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".tgc", "ide-sessions")
	os.MkdirAll(dir, 0755)
	return dir
}

// SaveSession saves the current chat to a file.
func (a *App) SaveSession(title string) (string, error) {
	if a.chat == nil {
		return "", fmt.Errorf("no chat engine")
	}
	id := time.Now().Format("20060102-150405")
	if title == "" {
		title = "Chat " + id
	}

	data := struct {
		ID        string                    `json:"id"`
		Title     string                    `json:"title"`
		CreatedAt string                    `json:"createdAt"`
		History   []map[string]string       `json:"history"`
	}{
		ID:        id,
		Title:     title,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	for _, m := range a.chat.history {
		if m.Role == "system" {
			continue
		}
		data.History = append(data.History, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	path := filepath.Join(sessionsDir(), id+".json")
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return "", err
	}
	return id, nil
}

// ListSessions returns saved sessions sorted by newest first.
func (a *App) ListSessions() []ChatSession {
	dir := sessionsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var sessions []ChatSession
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var s struct {
			ID        string              `json:"id"`
			Title     string              `json:"title"`
			CreatedAt string              `json:"createdAt"`
			History   []map[string]string `json:"history"`
		}
		if json.Unmarshal(data, &s) != nil {
			continue
		}
		sessions = append(sessions, ChatSession{
			ID:        s.ID,
			Title:     s.Title,
			CreatedAt: s.CreatedAt,
			Messages:  len(s.History),
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt > sessions[j].CreatedAt
	})
	return sessions
}

// LoadSession loads a saved session into the chat engine.
func (a *App) LoadSession(id string) error {
	path := filepath.Join(sessionsDir(), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var s struct {
		History []map[string]string `json:"history"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if a.chat == nil {
		return fmt.Errorf("no chat engine")
	}

	// Reset and reload
	a.chat.history = a.chat.history[:1] // keep system prompt
	for _, m := range s.History {
		a.chat.history = append(a.chat.history, openaiMsg(m["role"], m["content"]))
	}
	return nil
}

// DeleteSession removes a saved session.
func (a *App) DeleteSession(id string) error {
	return os.Remove(filepath.Join(sessionsDir(), id+".json"))
}
