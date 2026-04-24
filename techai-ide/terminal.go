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

type termSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	mu     sync.Mutex
	closed bool
}

func (a *App) GetAvailableShells() []string {
	if runtime.GOOS == "windows" {
		var found []string
		for _, s := range []string{"powershell.exe", "cmd.exe"} {
			if p, err := exec.LookPath(s); err == nil {
				found = append(found, p)
			}
		}
		return found
	}
	var found []string
	for _, s := range []string{"/bin/zsh", "/bin/bash", "/bin/sh"} {
		if _, err := os.Stat(s); err == nil {
			found = append(found, s)
		}
	}
	return found
}

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
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/bash"
}

func (a *App) SetShell(shell string) error {
	a.shellPath = shell
	a.StopTerminal()
	return a.StartTerminal()
}

func (a *App) StartTerminal() error {
	if a.term != nil && !a.term.closed {
		return nil
	}

	shell := a.GetCurrentShell()
	cmd := exec.Command(shell)
	cmd.Dir = a.cwd

	if runtime.GOOS != "windows" {
		cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	}

	// Platform-specific: hide console window on Windows
	hideConsoleWindow(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	term := &termSession{cmd: cmd, stdin: stdin}
	a.term = term

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

func (a *App) ResizeTerminal(rows, cols int) {}

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
