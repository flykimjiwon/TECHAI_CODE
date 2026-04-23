package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFiles(t *testing.T) {
	app := &App{cwd: "."}
	files, err := app.ListFiles(".", 1)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected files in current directory")
	}
	// Should contain go.mod
	found := false
	for _, f := range files {
		if f.Name == "go.mod" {
			found = true
			break
		}
	}
	if !found {
		t.Error("go.mod not found in listing")
	}
}

func TestListFiles_SortsDirectoriesFirst(t *testing.T) {
	app := &App{cwd: "."}
	files, err := app.ListFiles(".", 0)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	// Find first non-dir after dirs
	seenFile := false
	for _, f := range files {
		if !f.IsDir {
			seenFile = true
		}
		if seenFile && f.IsDir {
			t.Errorf("directory %s appears after file", f.Name)
		}
	}
}

func TestListFiles_HidesHiddenFiles(t *testing.T) {
	app := &App{cwd: "."}
	files, err := app.ListFiles(".", 0)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	for _, f := range files {
		if f.Name[0] == '.' {
			t.Errorf("hidden file should be filtered: %s", f.Name)
		}
	}
}

func TestReadFile(t *testing.T) {
	app := &App{cwd: "."}
	content, err := app.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if content == "" {
		t.Fatal("expected non-empty content")
	}
	if len(content) < 10 {
		t.Error("go.mod content too short")
	}
}

func TestReadFile_NotFound(t *testing.T) {
	app := &App{cwd: "."}
	_, err := app.ReadFile("nonexistent-file-12345.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestWriteAndDeleteFile(t *testing.T) {
	app := &App{cwd: t.TempDir()}
	path := "test-write.txt"
	content := "hello TECHAI"

	err := app.WriteFile(path, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := app.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after write failed: %v", err)
	}
	if got != content {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}

	err = app.DeleteFile(path)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}

	if app.FileExists(path) {
		t.Error("file should not exist after delete")
	}
}

func TestRenameFile(t *testing.T) {
	dir := t.TempDir()
	app := &App{cwd: dir}

	os.WriteFile(filepath.Join(dir, "old.txt"), []byte("data"), 0644)

	err := app.RenameFile("old.txt", "new.txt")
	if err != nil {
		t.Fatalf("RenameFile failed: %v", err)
	}

	if app.FileExists("old.txt") {
		t.Error("old file should not exist")
	}
	if !app.FileExists("new.txt") {
		t.Error("new file should exist")
	}
}

func TestSearchInFiles(t *testing.T) {
	dir := t.TempDir()
	app := &App{cwd: dir}

	os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\nfunc hello() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "world.go"), []byte("package main\nfunc world() {}\n"), 0644)

	results, err := app.SearchInFiles("func", dir)
	if err != nil {
		t.Fatalf("SearchInFiles failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearchInFiles_NoMatch(t *testing.T) {
	dir := t.TempDir()
	app := &App{cwd: dir}

	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("nothing here"), 0644)

	results, err := app.SearchInFiles("zzzzz", dir)
	if err != nil {
		t.Fatalf("SearchInFiles failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestGetCwd(t *testing.T) {
	app := &App{cwd: "/tmp"}
	if app.GetCwd() != "/tmp" {
		t.Errorf("expected /tmp, got %s", app.GetCwd())
	}
}

func TestSetCwd(t *testing.T) {
	app := &App{cwd: "/tmp"}
	err := app.SetCwd(os.TempDir())
	if err != nil {
		t.Fatalf("SetCwd failed: %v", err)
	}
}

func TestSetCwd_InvalidDir(t *testing.T) {
	app := &App{cwd: "/tmp"}
	err := app.SetCwd("/nonexistent-path-12345")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestFileExists(t *testing.T) {
	app := &App{cwd: "."}
	if !app.FileExists("go.mod") {
		t.Error("go.mod should exist")
	}
	if app.FileExists("nonexistent-12345.txt") {
		t.Error("nonexistent file should not exist")
	}
}
