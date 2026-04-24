package main

import (
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// termSession holds a running PTY terminal session.
type termSession struct {
	cmd    *exec.Cmd
	ptmx   *os.File
	mu     sync.Mutex
	closed bool
}

// GetAvailableShells returns shells found on the system.
func (a *App) GetAvailableShells() []string {
	candidates := []string{"/bin/zsh", "/bin/bash", "/bin/sh", "/usr/bin/fish", "/opt/homebrew/bin/fish"}
	var found []string
	for _, s := range candidates {
		if _, err := os.Stat(s); err == nil {
			found = append(found, s)
		}
	}
	winShells := []string{"powershell.exe", "cmd.exe", "pwsh.exe"}
	for _, s := range winShells {
		if p, err := exec.LookPath(s); err == nil {
			found = append(found, p)
		}
	}
	return found
}

// GetCurrentShell returns the current shell path.
func (a *App) GetCurrentShell() string {
	if a.shellPath != "" {
		return a.shellPath
	}
	s := os.Getenv("SHELL")
	if s == "" {
		s = "/bin/bash"
	}
	return s
}

// SetShell changes the shell and restarts the terminal.
func (a *App) SetShell(shell string) error {
	a.shellPath = shell
	a.StopTerminal()
	return a.StartTerminal()
}

// StartTerminal spawns a new shell PTY session.
func (a *App) StartTerminal() error {
	if a.term != nil && !a.term.closed {
		return nil
	}

	shell := a.GetCurrentShell()
	cmd := exec.Command(shell)
	cmd.Dir = a.cwd
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	_ = pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 120})

	term := &termSession{cmd: cmd, ptmx: ptmx}
	a.term = term

	// Read output — goroutine exits when ptmx is closed
	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := ptmx.Read(buf)
			if n > 0 {
				term.mu.Lock()
				if !term.closed {
					wailsRuntime.EventsEmit(a.ctx, "term:output", string(buf[:n]))
				}
				term.mu.Unlock()
			}
			if readErr != nil {
				if readErr != io.EOF && !term.closed {
					wailsRuntime.EventsEmit(a.ctx, "term:output", "\r\n[Process exited]\r\n")
				}
				return
			}
		}
	}()

	return nil
}

// WriteTerminal sends input to the terminal PTY.
func (a *App) WriteTerminal(input string) {
	if a.term == nil {
		return
	}
	a.term.mu.Lock()
	defer a.term.mu.Unlock()
	if a.term.closed || a.term.ptmx == nil {
		return
	}
	_, _ = a.term.ptmx.Write([]byte(input))
}

// ResizeTerminal updates the PTY window size.
func (a *App) ResizeTerminal(rows, cols int) {
	if a.term == nil || a.term.ptmx == nil || a.term.closed {
		return
	}
	_ = pty.Setsize(a.term.ptmx, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
}

// StopTerminal kills the terminal session.
func (a *App) StopTerminal() {
	if a.term == nil {
		return
	}
	a.term.mu.Lock()
	a.term.closed = true
	a.term.mu.Unlock()

	if a.term.ptmx != nil {
		a.term.ptmx.Close()
	}
	if a.term.cmd != nil && a.term.cmd.Process != nil {
		a.term.cmd.Process.Kill()
		a.term.cmd.Wait()
	}
	a.term = nil
}
