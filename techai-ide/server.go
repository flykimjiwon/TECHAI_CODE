package main

import (
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"runtime"
)

// runBrowserMode starts an HTTP server and opens the default browser.
// Used on Windows where WebView2 may not work properly.
func runBrowserMode(app *App, assets fs.FS) error {
	// Find available port
	port := 8080
	for port <= 8100 {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			break
		}
		port++
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

	// Open browser
	go openBrowser(url)

	// Initialize app
	go func() {
		app.ctx = nil // No Wails context in browser mode
		cfg := LoadTGCConfig()
		app.chatMu.Lock()
		app.chat = newChatEngine(cfg, app)
		app.chatMu.Unlock()
	}()

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
