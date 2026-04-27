// Copyright 2025-2026 Kim Jiwon (flykimjiwon). All rights reserved.
// TECHAI CODE — github.com/flykimjiwon/TECHAI_CODE
// Forked from Hanimo Code: github.com/flykimjiwon/hanimo

package main

import (
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

// runBrowserMode starts an HTTP server.
// Used as Electron backend or standalone browser mode.
func runBrowserMode(app *App, assets fs.FS) error {
	// Use TECHAI_PORT env or find available port
	port := 8080
	if p := os.Getenv("TECHAI_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	} else {
		for port <= 8100 {
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				ln.Close()
				break
			}
			port++
		}
	}

	// Serve embedded frontend
	webFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		return fmt.Errorf("failed to access embedded assets: %w", err)
	}

	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	// API endpoints — mirror Wails bindings as REST API
	registerAPI(mux, app)

	addr := fmt.Sprintf("localhost:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	fmt.Printf("TECHAI IDE running at %s\n", url)
	fmt.Println("Press Ctrl+C to stop.")

	// Only open browser if not running as Electron backend
	if os.Getenv("TECHAI_PORT") == "" {
		go openBrowser(url)
	}

	// Initialize app
	cfg := LoadTGCConfig()
	app.chatMu.Lock()
	app.chat = newChatEngine(cfg, app)
	app.chatMu.Unlock()

	// File watcher
	app.watcherDone = make(chan struct{})
	go app.watchFiles()

	return http.ListenAndServe(addr, mux)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
