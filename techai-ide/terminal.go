package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// termSession holds a running terminal session.
type termSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	mu     sync.Mutex
	closed bool
}

// GetAvailableShells returns shells found on the system.
func (a *App) GetAvailableShells() []string {
	if runtime.GOOS == "windows" {
		var found []string
		for _, s := range []string{"powershell.exe", "cmd.exe", "pwsh.exe"} {
			if p, err := exec.LookPath(s); err == nil {
				found = append(found, p)
			}
		}
		return found
	}
	candidates := []string{"/bin/zsh", "/bin/bash", "/bin/sh"}
	var found []string
	for _, s := range candidates {
		if _, err := os.Stat(s); err == nil {
			found = append(found, s)
		}
	}
	return found
}

// GetCurrentShell returns the current shell path.
func (a *App) GetCurrentShell() string {
	if a.shellPath != "" {
		return a.shellPath
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("powershell.exe"); err == nil {
			return p
		}
		return "cmd.exe"
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

// StartTerminal spawns a new shell session.
func (a *App) StartTerminal() error {
	if a.term != nil && !a.term.closed {
		return nil
	}

	shell := a.GetCurrentShell()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(shell)
	} else {
		cmd = exec.Command(shell)
		cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	}
	cmd.Dir = a.cwd

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	term := &termSession{cmd: cmd, stdin: stdin}
	a.term = term

	// Read output
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			term.mu.Lock()
			if !term.closed {
				wailsRuntime.EventsEmit(a.ctx, "term:output", scanner.Text()+"\r\n")
			}
			term.mu.Unlock()
		}
		if !term.closed {
			wailsRuntime.EventsEmit(a.ctx, "term:output", "\r\n[Process exited]\r\n")
		}
	}()

	return nil
}

// WriteTerminal sends input to the terminal.
func (a *App) WriteTerminal(input string) {
	if a.term == nil {
		return
	}
	a.term.mu.Lock()
	defer a.term.mu.Unlock()
	if a.term.closed || a.term.stdin == nil {
		return
	}
	_, _ = a.term.stdin.Write([]byte(input))
}

// ResizeTerminal — no-op without PTY (pipe-based).
func (a *App) ResizeTerminal(rows, cols int) {}

// StopTerminal kills the terminal session.
func (a *App) StopTerminal() {
	if a.term == nil {
		return
	}
	a.term.mu.Lock()
	a.term.closed = true
	a.term.mu.Unlock()

	if a.term.stdin != nil {
		a.term.stdin.Close()
	}
	if a.term.cmd != nil && a.term.cmd.Process != nil {
		a.term.cmd.Process.Kill()
		a.term.cmd.Wait()
	}
	a.term = nil
}
