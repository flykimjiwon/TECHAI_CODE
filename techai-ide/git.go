package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitInfo holds basic git repository state.
type GitInfo struct {
	Branch   string   `json:"branch"`
	IsDirty  bool     `json:"isDirty"`
	Changes  []GitChange `json:"changes"`
}

// GitChange represents a single changed file.
type GitChange struct {
	Status string `json:"status"` // "M", "A", "D", "??"
	File   string `json:"file"`
}

// GetGitInfo returns the current git status for the working directory.
func (a *App) GetGitInfo() GitInfo {
	info := GitInfo{}

	// Branch
	out, err := a.runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return info
	}
	info.Branch = strings.TrimSpace(out)

	// Status
	out, err = a.runGit("status", "--porcelain")
	if err != nil {
		return info
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		status := strings.TrimSpace(line[:2])
		file := strings.TrimSpace(line[3:])
		info.Changes = append(info.Changes, GitChange{Status: status, File: file})
	}
	info.IsDirty = len(info.Changes) > 0

	return info
}

// GitDiff returns the diff output for staged + unstaged changes.
func (a *App) GitDiff() string {
	out, err := a.runGit("diff")
	if err != nil {
		return ""
	}
	staged, _ := a.runGit("diff", "--cached")
	if staged != "" {
		out = "=== Staged ===\n" + staged + "\n=== Unstaged ===\n" + out
	}
	return out
}

// GitLog returns recent commit log.
func (a *App) GitLog(n int) string {
	if n <= 0 {
		n = 10
	}
	out, _ := a.runGit("log", "--oneline", "-n", fmt.Sprintf("%d", n))
	return out
}

// GitDiffFile returns the diff for a specific file.
func (a *App) GitDiffFile(path string) string {
	out, _ := a.runGit("diff", "--", path)
	staged, _ := a.runGit("diff", "--cached", "--", path)
	if staged != "" {
		out = staged + "\n" + out
	}
	return out
}

// GitStage stages a file.
func (a *App) GitStage(path string) error {
	_, err := a.runGit("add", path)
	return err
}

// GitUnstage unstages a file.
func (a *App) GitUnstage(path string) error {
	_, err := a.runGit("reset", "HEAD", "--", path)
	return err
}

// GitCommit creates a commit with the given message.
func (a *App) GitCommit(message string) (string, error) {
	out, err := a.runGit("commit", "-m", message)
	return out, err
}

// GitLogEntry represents a single commit.
type GitLogEntry struct {
	Hash    string `json:"hash"`
	Short   string `json:"short"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	Message string `json:"message"`
	Branch  string `json:"branch"`
	Refs    string `json:"refs"`
}

// GetGitGraph returns formatted commit log for visualization.
func (a *App) GetGitGraph(count int) []GitLogEntry {
	if count <= 0 {
		count = 30
	}
	format := "%H|%h|%an|%ar|%s|%D"
	out, err := a.runGit("log", "--all", "--oneline", "--graph",
		"--format="+format, "-n", fmt.Sprintf("%d", count))
	if err != nil {
		return nil
	}

	var entries []GitLogEntry
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove graph characters (*, |, /, \)
		clean := line
		for len(clean) > 0 && (clean[0] == '*' || clean[0] == '|' || clean[0] == '/' || clean[0] == '\\' || clean[0] == ' ') {
			clean = clean[1:]
		}
		parts := strings.SplitN(clean, "|", 6)
		if len(parts) < 5 {
			continue
		}
		entry := GitLogEntry{
			Hash:    strings.TrimSpace(parts[0]),
			Short:   strings.TrimSpace(parts[1]),
			Author:  strings.TrimSpace(parts[2]),
			Date:    strings.TrimSpace(parts[3]),
			Message: strings.TrimSpace(parts[4]),
		}
		if len(parts) > 5 {
			entry.Refs = strings.TrimSpace(parts[5])
		}
		entries = append(entries, entry)
	}
	return entries
}

// GetGitBranches returns all branches.
func (a *App) GetGitBranches() []string {
	out, err := a.runGit("branch", "-a", "--format=%(refname:short)")
	if err != nil {
		return nil
	}
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches
}

// GitCheckout switches to a branch.
func (a *App) GitCheckout(branch string) (string, error) {
	return a.runGit("checkout", branch)
}

// GitCreateBranch creates and switches to a new branch.
func (a *App) GitCreateBranch(name string) (string, error) {
	return a.runGit("checkout", "-b", name)
}

// GitPull pulls from remote.
func (a *App) GitPull() (string, error) {
	return a.runGit("pull")
}

// GitPush pushes to remote.
func (a *App) GitPush() (string, error) {
	return a.runGit("push")
}

func (a *App) runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = a.cwd
	out, err := cmd.CombinedOutput()
	return string(out), err
}
