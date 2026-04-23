package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupGitRepo(t *testing.T) *App {
	dir := t.TempDir()
	// Initialize git repo
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "test")
	os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\n"), 0644)
	run("add", ".")
	run("commit", "-m", "initial")
	return &App{cwd: dir}
}

func TestGetGitInfo(t *testing.T) {
	app := setupGitRepo(t)
	info := app.GetGitInfo()
	if info.Branch == "" {
		t.Error("branch should not be empty")
	}
	// Should be clean after commit
	if info.IsDirty {
		t.Error("repo should be clean after commit")
	}
}

func TestGetGitInfo_Dirty(t *testing.T) {
	app := setupGitRepo(t)
	os.WriteFile(filepath.Join(app.cwd, "new.txt"), []byte("change"), 0644)
	info := app.GetGitInfo()
	if !info.IsDirty {
		t.Error("repo should be dirty after adding file")
	}
	if len(info.Changes) == 0 {
		t.Error("should have changes")
	}
}

func TestGitDiff(t *testing.T) {
	app := setupGitRepo(t)
	os.WriteFile(filepath.Join(app.cwd, "hello.go"), []byte("package main\n// changed\n"), 0644)
	diff := app.GitDiff()
	if diff == "" {
		t.Error("diff should not be empty after modification")
	}
}

func TestGetGitBranches(t *testing.T) {
	app := setupGitRepo(t)
	branches := app.GetGitBranches()
	if len(branches) == 0 {
		t.Error("should have at least one branch")
	}
}

func TestGetGitGraph(t *testing.T) {
	app := setupGitRepo(t)
	entries := app.GetGitGraph(10)
	if len(entries) == 0 {
		t.Error("should have at least one commit")
	}
	if entries[0].Message != "initial" {
		t.Errorf("first commit message=%q, want 'initial'", entries[0].Message)
	}
}

func TestGitStageAndCommit(t *testing.T) {
	app := setupGitRepo(t)
	os.WriteFile(filepath.Join(app.cwd, "staged.txt"), []byte("test"), 0644)

	err := app.GitStage("staged.txt")
	if err != nil {
		t.Fatalf("GitStage failed: %v", err)
	}

	_, err = app.GitCommit("test commit")
	if err != nil {
		t.Fatalf("GitCommit failed: %v", err)
	}

	info := app.GetGitInfo()
	if info.IsDirty {
		t.Error("should be clean after commit")
	}
}

func TestGitCreateBranch(t *testing.T) {
	app := setupGitRepo(t)
	_, err := app.GitCreateBranch("feature-test")
	if err != nil {
		t.Fatalf("GitCreateBranch failed: %v", err)
	}
	info := app.GetGitInfo()
	if info.Branch != "feature-test" {
		t.Errorf("branch=%q, want 'feature-test'", info.Branch)
	}
}

func TestGitCheckout(t *testing.T) {
	app := setupGitRepo(t)
	// Get initial branch name (main or master)
	initialBranch := app.GetGitInfo().Branch
	app.GitCreateBranch("dev")
	// Now on "dev", checkout back to initial
	app.GitCheckout(initialBranch)
	info := app.GetGitInfo()
	if info.Branch != initialBranch {
		t.Errorf("branch=%q after checkout, want %q", info.Branch, initialBranch)
	}
}
