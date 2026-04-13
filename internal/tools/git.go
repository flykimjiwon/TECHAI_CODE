package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// runGit executes a git command in the given directory with a 10s timeout.
func runGit(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return result, fmt.Errorf("git %s: %s (%w)", args[0], result, err)
	}
	return result, nil
}

// GitStatus returns short status output for the given path.
func GitStatus(path string) (string, error) {
	return runGit(path, "status", "--short")
}

// GitDiff returns diff output. If staged is true, shows staged changes.
func GitDiff(path string, staged bool) (string, error) {
	if staged {
		return runGit(path, "diff", "--staged")
	}
	return runGit(path, "diff")
}

// GitLog returns the last n commits in oneline format.
func GitLog(path string, n int) (string, error) {
	return runGit(path, "log", fmt.Sprintf("-n%d", n), "--oneline")
}

// GitCommit creates a commit with the given message.
func GitCommit(path, message string) (string, error) {
	return runGit(path, "commit", "-m", message)
}

// GitBranch returns the current branch name.
func GitBranch(path string) (string, error) {
	return runGit(path, "branch", "--show-current")
}
