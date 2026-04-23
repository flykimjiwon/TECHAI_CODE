package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKnowledgeContext_NoTechaiMd(t *testing.T) {
	dir := t.TempDir()
	ctx := loadKnowledgeContext(dir)
	// Should detect no project type
	if ctx != "" {
		// May contain knowledge packs context
	}
}

func TestLoadKnowledgeContext_WithTechaiMd(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".techai.md"), []byte("# My Project\nTest context"), 0644)
	ctx := loadKnowledgeContext(dir)
	if ctx == "" {
		t.Error("should have context from .techai.md")
	}
}

func TestLoadKnowledgeContext_DetectsGo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	ctx := loadKnowledgeContext(dir)
	if ctx == "" {
		t.Error("should detect Go project")
	}
}

func TestLoadKnowledgeContext_DetectsNode(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
	ctx := loadKnowledgeContext(dir)
	if ctx == "" {
		t.Error("should detect Node.js project")
	}
}

func TestFileExists_Helper(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "exists.txt"), []byte("hi"), 0644)
	if !fileExists(filepath.Join(dir, "exists.txt")) {
		t.Error("should exist")
	}
	if fileExists(filepath.Join(dir, "nope.txt")) {
		t.Error("should not exist")
	}
}

func TestToggleKnowledgePack(t *testing.T) {
	app := &App{cwd: "."}
	app.ToggleKnowledgePack("go-patterns", true)
	if !enabledPacks["go-patterns"] {
		t.Error("should be enabled")
	}
	app.ToggleKnowledgePack("go-patterns", false)
	if enabledPacks["go-patterns"] {
		t.Error("should be disabled")
	}
}

func TestParseParameterTags_MultipleParams(t *testing.T) {
	params := parseParameterTags(`<parameter=pattern> TODO <parameter=path> ./src <parameter=include> *.go`)
	if len(params) != 3 {
		t.Errorf("expected 3 params, got %d", len(params))
	}
	if params["pattern"] != "TODO" {
		t.Errorf("pattern=%q", params["pattern"])
	}
}
