package companion

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"time"
)

// Server serves the companion dashboard and SSE event stream.
// It runs on its own http.Server in a dedicated goroutine so it
// never interferes with the LLM HTTP client (see DEBUG_TRANSPORT_FREEZE.md).
type Server struct {
	hub *Hub
	srv *http.Server
}

// NewServer creates a companion HTTP server.
// webFS should be the "web" sub-tree of the embedded filesystem.
func NewServer(hub *Hub, webFS fs.FS, port int) *Server {
	mux := http.NewServeMux()
	s := &Server{
		hub: hub,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	mux.HandleFunc("/events", s.handleSSE)
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	return s
}

// Start launches the HTTP server in a background goroutine.
func (s *Server) Start() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[COMPANION] server error: %v\n", err)
		}
	}()
}

// Stop performs a graceful shutdown with a 3-second deadline.
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = s.srv.Shutdown(ctx)
}

// handleSSE implements the Server-Sent Events endpoint.
// It sends catch-up events (via Last-Event-ID or ?after= query param),
// then streams live events with a 15-second heartbeat.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Determine the last event ID the client has seen.
	var afterID int64
	if idStr := r.Header.Get("Last-Event-ID"); idStr != "" {
		afterID, _ = strconv.ParseInt(idStr, 10, 64)
	}
	if q := r.URL.Query().Get("after"); q != "" {
		afterID, _ = strconv.ParseInt(q, 10, 64)
	}

	// 1. Catch-up: replay missed events.
	for _, ev := range s.hub.EventsAfter(afterID) {
		_, _ = w.Write(ev.ToSSE())
	}
	flusher.Flush()

	// 2. Subscribe for live events.
	ch, cancel := s.hub.Subscribe()
	defer cancel()

	// 3. Heartbeat ticker.
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-ch:
			_, _ = w.Write(ev.ToSSE())
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = w.Write([]byte(": heartbeat\n\n"))
			flusher.Flush()
		}
	}
}
