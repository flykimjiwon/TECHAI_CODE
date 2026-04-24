package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct — Wails binds these methods to the frontend.
type App struct {
	ctx         context.Context
	cwd         string
	chat        *chatEngine
	chatMu      sync.Mutex
	term        *termSession
	shellPath   string
	watcherDone chan struct{}
}

func NewApp() *App {
	cwd, _ := os.Getwd()
	return &App{cwd: cwd}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg := LoadTGCConfig()
	a.chat = newChatEngine(cfg, a)

	// Set window title
	wailsRuntime.WindowSetTitle(ctx, "TECHAI IDE — "+a.cwd)

	// Auto-start terminal
	_ = a.StartTerminal()

	// Save to recent projects
	a.saveRecentProject(a.cwd)

	// File watcher with proper shutdown
	a.watcherDone = make(chan struct{})
	go a.watchFiles()
}

func (a *App) watchFiles() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	modTimes := map[string]time.Time{}

	for {
		select {
		case <-a.watcherDone:
			return
		case <-ticker.C:
			entries, _ := os.ReadDir(a.cwd)
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				info, err := e.Info()
				if err != nil {
					continue
				}
				path := filepath.Join(a.cwd, e.Name())
				mt := info.ModTime()
				if prev, ok := modTimes[path]; ok && mt.After(prev) {
					wailsRuntime.EventsEmit(a.ctx, "file:changed", path)
				}
				modTimes[path] = mt
			}
		}
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.watcherDone != nil {
		close(a.watcherDone)
	}
	a.StopTerminal()
}

// OpenFolder opens a native folder picker and switches the project.
func (a *App) OpenFolder() (string, error) {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Open Project Folder",
	})
	if err != nil || dir == "" {
		return "", err
	}
	a.cwd = dir
	wailsRuntime.WindowSetTitle(a.ctx, "TECHAI IDE — "+dir)
	return dir, nil
}

// ── File System API ──

// FileEntry represents a file or directory in the tree.
type FileEntry struct {
	Name  string      `json:"name"`
	Path  string      `json:"path"`
	IsDir bool        `json:"isDir"`
	Size  int64       `json:"size"`
	Kids  []FileEntry `json:"kids,omitempty"`
}

// ListFiles returns directory contents. If depth > 0, recurses.
func (a *App) ListFiles(dir string, depth int) ([]FileEntry, error) {
	if dir == "" || dir == "." {
		dir = a.cwd
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(a.cwd, dir)
	}
	return listDir(dir, depth, 0)
}

func listDir(dir string, maxDepth, curDepth int) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []FileEntry
	for _, e := range entries {
		name := e.Name()
		// Skip hidden files and common noise
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "__pycache__" || name == "dist" {
			continue
		}
		info, _ := e.Info()
		fe := FileEntry{
			Name:  name,
			Path:  filepath.Join(dir, name),
			IsDir: e.IsDir(),
		}
		if info != nil {
			fe.Size = info.Size()
		}
		if e.IsDir() && curDepth < maxDepth {
			kids, _ := listDir(filepath.Join(dir, name), maxDepth, curDepth+1)
			fe.Kids = kids
		}
		result = append(result, fe)
	}

	// Sort: directories first, then alphabetical
	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}

// safePath resolves a path and ensures it stays within the project directory.
func (a *App) safePath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(a.cwd, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	cwdAbs, _ := filepath.Abs(a.cwd)
	if !strings.HasPrefix(abs, cwdAbs) {
		return "", fmt.Errorf("access denied: path outside project")
	}
	return abs, nil
}

// ReadFile returns file content as string.
func (a *App) ReadFile(path string) (string, error) {
	path, err := a.safePath(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Limit to 1MB for safety
	if len(data) > 1024*1024 {
		return string(data[:1024*1024]) + "\n\n... [truncated at 1MB]", nil
	}
	return string(data), nil
}

// WriteFile saves content to a file.
func (a *App) WriteFile(path, content string) error {
	var err error
	path, err = a.safePath(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// GetCwd returns the current working directory.
func (a *App) GetCwd() string {
	return a.cwd
}

// SetCwd changes the working directory.
func (a *App) SetCwd(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}
	a.cwd = dir
	return nil
}

// DeleteFile removes a file or empty directory.
func (a *App) DeleteFile(path string) error {
	if !filepath.IsAbs(path) {
		path = filepath.Join(a.cwd, path)
	}
	return os.Remove(path)
}

// RenameFile renames/moves a file.
func (a *App) RenameFile(oldPath, newPath string) error {
	if !filepath.IsAbs(oldPath) {
		oldPath = filepath.Join(a.cwd, oldPath)
	}
	if !filepath.IsAbs(newPath) {
		newPath = filepath.Join(a.cwd, newPath)
	}
	return os.Rename(oldPath, newPath)
}

// GetRecentProjects returns recently opened project paths.
func (a *App) GetRecentProjects() []string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".tgc", "ide-recent.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var projects []string
	json.Unmarshal(data, &projects)
	return projects
}

func (a *App) saveRecentProject(dir string) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".tgc", "ide-recent.json")
	projects := a.GetRecentProjects()
	filtered := []string{dir}
	for _, p := range projects {
		if p != dir {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) > 10 {
		filtered = filtered[:10]
	}
	data, _ := json.Marshal(filtered)
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, data, 0644)
}

// OpenInBrowser opens a file in the default browser.
func (a *App) OpenInBrowser(path string) error {
	if !filepath.IsAbs(path) {
		path = filepath.Join(a.cwd, path)
	}
	return openURL("file://" + path)
}

// StartLiveServer starts a simple HTTP server for the given directory.
func (a *App) StartLiveServer(dir string) (string, error) {
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(a.cwd, dir)
	}
	// Find available port
	port := 5500
	for port <= 5600 {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			break
		}
		port++
	}
	if port > 5600 {
		return "", fmt.Errorf("no available ports (5500-5600)")
	}
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", port), http.FileServer(http.Dir(dir)))
	}()
	url := fmt.Sprintf("http://localhost:%d", port)
	// Open in IDE preview panel instead of external browser
	wailsRuntime.EventsEmit(a.ctx, "preview:open", url)
	return url, nil
}

func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// FileExists checks if a file/dir exists.
func (a *App) FileExists(path string) bool {
	if !filepath.IsAbs(path) {
		path = filepath.Join(a.cwd, path)
	}
	_, err := os.Stat(path)
	return err == nil
}

// GetFileIcon returns a suggested icon name based on file extension.
func (a *App) GetFileIcon(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return "file-code"
	case ".ts", ".tsx":
		return "file-type"
	case ".js", ".jsx":
		return "file-json"
	case ".md":
		return "file-text"
	case ".yaml", ".yml", ".json", ".toml":
		return "file-cog"
	case ".css", ".scss":
		return "palette"
	case ".html":
		return "globe"
	case ".py":
		return "file-code"
	case ".sql":
		return "database"
	case ".sh", ".bash", ".zsh":
		return "terminal"
	default:
		return "file"
	}
}

// WalkProject returns the full project tree up to depth 3.
func (a *App) WalkProject() ([]FileEntry, error) {
	return a.ListFiles(".", 3)
}

// SearchInFiles searches for a pattern across files using simple string matching.
func (a *App) SearchInFiles(pattern, dir string) ([]SearchResult, error) {
	if dir == "" || dir == "." {
		dir = a.cwd
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(a.cwd, dir)
	}

	var results []SearchResult
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip binary files
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".exe" || ext == ".bin" || ext == ".so" || ext == ".dylib" || ext == ".png" || ext == ".jpg" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil || len(data) > 512*1024 {
			return nil
		}

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(line, pattern) {
				rel, _ := filepath.Rel(a.cwd, path)
				results = append(results, SearchResult{
					File:    rel,
					Line:    i + 1,
					Content: strings.TrimSpace(line),
				})
				if len(results) >= 100 {
					return fmt.Errorf("limit")
				}
			}
		}
		return nil
	})

	return results, nil
}

// GetSearchLimit returns the max results limit (for UI display).
func (a *App) GetSearchLimit() int { return 100 }

// SearchResult holds a single search match.
type SearchResult struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}
