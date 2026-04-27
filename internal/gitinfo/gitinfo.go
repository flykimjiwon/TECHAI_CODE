// Package gitinfo provides a lightweight read-only snapshot of the git
// state of the current working directory. It is used by the HUD and the
// /git slash command to show branch and dirty-status at a glance.
package gitinfo

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Info is a snapshot of the current git working tree. A zero-valued
// Info (Branch == "") means "not a git repository" and should be
// silently ignored by callers.
type Info struct {
	Branch    string
	Dirty     bool
	Staged    int
	Unstaged  int
	Untracked int
}

// ParseBranch normalizes the output of `git rev-parse --abbrev-ref HEAD`.
// Empty input returns "" so callers can distinguish "not a repo" from
// "detached HEAD" (which returns "HEAD").
func ParseBranch(out string) string {
	return strings.TrimSpace(out)
}

// ApplyPorcelain populates the Staged / Unstaged / Untracked counters
// and the Dirty flag from `git status --porcelain=v1` output. Porcelain
// v1 is one line per entry with a two-char status (XY) followed by a
// space and the path.
//
//	X (staged):   M, A, D, R, C, U, ?
//	Y (unstaged): M, D, U, ?
//
// We follow git's own rule: "?? path" is purely untracked and counted
// separately.
func ApplyPorcelain(info *Info, out string) {
	if info == nil {
		return
	}
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 3 {
			continue
		}
		x, y := line[0], line[1]
		if x == '?' && y == '?' {
			info.Untracked++
			info.Dirty = true
			continue
		}
		if x != ' ' && x != '?' {
			info.Staged++
			info.Dirty = true
		}
		if y != ' ' && y != '?' {
			info.Unstaged++
			info.Dirty = true
		}
	}
}

// Label returns the short HUD form: "branch" when clean, "branch*"
// when dirty, empty when not a repository.
func (i Info) Label() string {
	if i.Branch == "" {
		return ""
	}
	if i.Dirty {
		return i.Branch + "*"
	}
	return i.Branch
}

// Summary returns a multi-line human-readable description used by the
// /git slash command.
func (i Info) Summary() string {
	if i.Branch == "" {
		return "  git: 현재 디렉토리는 git 저장소가 아닙니다."
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "  git: %s\n", i.Branch)
	if !i.Dirty {
		sb.WriteString("  clean working tree")
		return sb.String()
	}
	fmt.Fprintf(&sb, "  staged:    %d\n", i.Staged)
	fmt.Fprintf(&sb, "  unstaged:  %d\n", i.Unstaged)
	fmt.Fprintf(&sb, "  untracked: %d", i.Untracked)
	return sb.String()
}

// Fetch runs `git` in cwd and returns a populated Info. Any error
// (not a repo, git binary missing, timeout) yields a zero-valued Info
// and a nil error so the UI degrades gracefully.
func Fetch(cwd string) Info {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	branchOut, err := run(ctx, cwd, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return Info{}
	}
	info := Info{Branch: ParseBranch(branchOut)}
	if info.Branch == "" {
		return Info{}
	}

	porcOut, err := run(ctx, cwd, "status", "--porcelain=v1")
	if err != nil {
		// We still have a branch, just no dirty info — return what we
		// have rather than zeroing out.
		return info
	}
	ApplyPorcelain(&info, porcOut)
	return info
}

func run(ctx context.Context, cwd string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
