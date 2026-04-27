// Copyright 2025-2026 Kim Jiwon (flykimjiwon). All rights reserved.
// TECHAI CODE — github.com/flykimjiwon/TECHAI_CODE
// Forked from Hanimo Code: github.com/flykimjiwon/hanimo

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
)

// SSE event bus for browser mode (replaces Wails EventsEmit)
var (
	sseClients   = make(map[chan string]bool)
	sseClientsMu sync.Mutex
)

func sseEmit(event, data string) {
	sseClientsMu.Lock()
	defer sseClientsMu.Unlock()
	// SSE spec: multi-line data needs each line prefixed with "data: "
	lines := strings.Split(data, "\n")
	var msg strings.Builder
	msg.WriteString("event: " + event + "\n")
	for _, line := range lines {
		msg.WriteString("data: " + line + "\n")
	}
	msg.WriteString("\n")
	encoded := msg.String()
	for ch := range sseClients {
		select {
		case ch <- encoded:
		default:
		}
	}
}

func registerAPI(mux *http.ServeMux, app *App) {
	// CORS for browser mode
	cors := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				return
			}
			h(w, r)
		}
	}

	// SSE endpoint for real-time events
	mux.HandleFunc("/api/events", cors(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", 500)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := make(chan string, 100)
		sseClientsMu.Lock()
		sseClients[ch] = true
		sseClientsMu.Unlock()

		defer func() {
			sseClientsMu.Lock()
			delete(sseClients, ch)
			sseClientsMu.Unlock()
		}()

		for {
			select {
			case msg := <-ch:
				w.Write([]byte(msg))
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}))

	// File operations
	mux.HandleFunc("/api/listFiles", cors(func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("path")
		depth := 3
		files, err := app.ListFiles(dir, depth)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(files)
	}))

	mux.HandleFunc("/api/readFile", cors(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		content, err := app.ReadFile(path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(content))
	}))

	mux.HandleFunc("/api/writeFile", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.WriteFile(req.Path, req.Content); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		sseEmit("file:changed", req.Path)
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/deleteFile", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Path string `json:"path"` }
		json.NewDecoder(r.Body).Decode(&req)
		app.DeleteFile(req.Path)
		w.Write([]byte("ok"))
	}))

	// Chat
	mux.HandleFunc("/api/sendMessage", cors(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		prompt := strings.TrimSpace(string(body))
		go app.SendMessage(prompt)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))

	mux.HandleFunc("/api/clearChat", cors(func(w http.ResponseWriter, r *http.Request) {
		app.ClearChat()
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/getModel", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(app.GetModel()))
	}))

	// Git
	mux.HandleFunc("/api/gitInfo", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.GetGitInfo())
	}))

	mux.HandleFunc("/api/gitGraph", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.GetGitGraph(50))
	}))

	mux.HandleFunc("/api/gitBranches", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.GetGitBranches())
	}))

	// Terminal
	mux.HandleFunc("/api/startTerminal", cors(func(w http.ResponseWriter, r *http.Request) {
		app.StartTerminal()
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/writeTerminal", cors(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		app.WriteTerminal(string(body))
		w.Write([]byte("ok"))
	}))

	// Settings
	mux.HandleFunc("/api/getSettings", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.GetSettings())
	}))

	mux.HandleFunc("/api/getCwd", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(app.GetCwd()))
	}))

	mux.HandleFunc("/api/search", cors(func(w http.ResponseWriter, r *http.Request) {
		pattern := r.URL.Query().Get("pattern")
		dir := r.URL.Query().Get("path")
		results, _ := app.SearchInFiles(pattern, dir)
		json.NewEncoder(w).Encode(results)
	}))

	mux.HandleFunc("/api/walkProject", cors(func(w http.ResponseWriter, r *http.Request) {
		files, _ := app.WalkProject()
		json.NewEncoder(w).Encode(files)
	}))

	mux.HandleFunc("/api/getShells", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.GetAvailableShells())
	}))

	mux.HandleFunc("/api/getCurrentShell", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(app.GetCurrentShell()))
	}))

	mux.HandleFunc("/api/knowledgePacks", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.GetKnowledgePacks())
	}))

	mux.HandleFunc("/api/recentProjects", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.GetRecentProjects())
	}))

	// ── Missing endpoints (Electron parity) ──

	// File operations
	mux.HandleFunc("/api/renameFile", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OldPath string `json:"oldPath"`
			NewPath string `json:"newPath"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.RenameFile(req.OldPath, req.NewPath); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		sseEmit("tree:refresh", "")
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/setCwd", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Path string `json:"path"` }
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.SetCwd(req.Path); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		sseEmit("tree:refresh", "")
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/fileExists", cors(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		json.NewEncoder(w).Encode(app.FileExists(path))
	}))

	// Settings
	mux.HandleFunc("/api/saveSettings", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			BaseURL string `json:"baseURL"`
			APIKey  string `json:"apiKey"`
			Model   string `json:"model"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.SaveSettings(req.BaseURL, req.APIKey, req.Model); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/setLanguage", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Lang string `json:"lang"` }
		json.NewDecoder(r.Body).Decode(&req)
		app.SetLanguage(req.Lang)
		w.Write([]byte("ok"))
	}))

	// Git operations
	mux.HandleFunc("/api/gitCheckout", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Branch string `json:"branch"` }
		json.NewDecoder(r.Body).Decode(&req)
		result, err := app.GitCheckout(req.Branch)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(result))
	}))

	mux.HandleFunc("/api/gitCreateBranch", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Name string `json:"name"` }
		json.NewDecoder(r.Body).Decode(&req)
		result, err := app.GitCreateBranch(req.Name)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(result))
	}))

	mux.HandleFunc("/api/gitPull", cors(func(w http.ResponseWriter, r *http.Request) {
		result, err := app.GitPull()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(result))
	}))

	mux.HandleFunc("/api/gitPush", cors(func(w http.ResponseWriter, r *http.Request) {
		result, err := app.GitPush()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(result))
	}))

	mux.HandleFunc("/api/gitStage", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Path string `json:"path"` }
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.GitStage(req.Path); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/gitUnstage", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Path string `json:"path"` }
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.GitUnstage(req.Path); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/gitCommit", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Message string `json:"message"` }
		json.NewDecoder(r.Body).Decode(&req)
		result, err := app.GitCommit(req.Message)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(result))
	}))

	mux.HandleFunc("/api/gitDiff", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(app.GitDiff()))
	}))

	mux.HandleFunc("/api/gitDiffFile", cors(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		w.Write([]byte(app.GitDiffFile(path)))
	}))

	mux.HandleFunc("/api/gitLog", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(app.GitLog(50)))
	}))

	// Terminal
	mux.HandleFunc("/api/stopTerminal", cors(func(w http.ResponseWriter, r *http.Request) {
		app.StopTerminal()
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/setShell", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Shell string `json:"shell"` }
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.SetShell(req.Shell); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("ok"))
	}))

	// Knowledge packs
	mux.HandleFunc("/api/toggleKnowledgePack", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID      string `json:"id"`
			Enabled bool   `json:"enabled"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		app.ToggleKnowledgePack(req.ID, req.Enabled)
		w.Write([]byte("ok"))
	}))

	// Chat sessions
	mux.HandleFunc("/api/chatHistory", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.GetChatHistory())
	}))

	mux.HandleFunc("/api/exportChat", cors(func(w http.ResponseWriter, r *http.Request) {
		result, err := app.ExportChat()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(result))
	}))

	mux.HandleFunc("/api/saveSession", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Title string `json:"title"` }
		json.NewDecoder(r.Body).Decode(&req)
		id, err := app.SaveSession(req.Title)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(id))
	}))

	mux.HandleFunc("/api/listSessions", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(app.ListSessions())
	}))

	mux.HandleFunc("/api/loadSession", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ ID string `json:"id"` }
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.LoadSession(req.ID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/deleteSession", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ ID string `json:"id"` }
		json.NewDecoder(r.Body).Decode(&req)
		if err := app.DeleteSession(req.ID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte("ok"))
	}))

	// Live server
	mux.HandleFunc("/api/startLiveServer", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Dir string `json:"dir"` }
		json.NewDecoder(r.Body).Decode(&req)
		url, err := app.StartLiveServer(req.Dir)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(url))
	}))

	// EventsEmit from frontend
	mux.HandleFunc("/api/emit", cors(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string      `json:"name"`
			Data interface{} `json:"data"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.Name != "" {
			payload := ""
			if req.Data != nil {
				if s, ok := req.Data.(string); ok {
					payload = s
				} else {
					b, _ := json.Marshal(req.Data)
					payload = string(b)
				}
			}
			sseEmit(req.Name, payload)
		}
		w.Write([]byte("ok"))
	}))
}
