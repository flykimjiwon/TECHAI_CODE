package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePatch_AddFile(t *testing.T) {
	patch := `*** Begin Patch
*** Add File: test_new.go
+package main
+
+func hello() {}
*** End Patch`

	ops, err := ParsePatch(patch)
	if err != nil {
		t.Fatalf("ParsePatch failed: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
	if ops[0].Type != "add" {
		t.Errorf("expected type=add, got %s", ops[0].Type)
	}
	if ops[0].Path != "test_new.go" {
		t.Errorf("expected path=test_new.go, got %s", ops[0].Path)
	}
	if !strings.Contains(ops[0].Content, "package main") {
		t.Errorf("expected content to contain 'package main', got %q", ops[0].Content)
	}
}

func TestParsePatch_UpdateFile(t *testing.T) {
	patch := `*** Begin Patch
*** Update File: main.go
@@ func Run() {
-    old line
+    new line
*** End Patch`

	ops, err := ParsePatch(patch)
	if err != nil {
		t.Fatalf("ParsePatch failed: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
	if ops[0].Type != "update" {
		t.Errorf("expected type=update, got %s", ops[0].Type)
	}
	if len(ops[0].Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(ops[0].Hunks))
	}
	hunk := ops[0].Hunks[0]
	if hunk.Context != "func Run() {" {
		t.Errorf("expected context='func Run() {', got %q", hunk.Context)
	}
	if len(hunk.RemoveLines) != 1 || !strings.Contains(hunk.RemoveLines[0], "old line") {
		t.Errorf("unexpected remove lines: %v", hunk.RemoveLines)
	}
	if len(hunk.AddLines) != 1 || !strings.Contains(hunk.AddLines[0], "new line") {
		t.Errorf("unexpected add lines: %v", hunk.AddLines)
	}
}

func TestParsePatch_DeleteFile(t *testing.T) {
	patch := `*** Begin Patch
*** Delete File: old.go
*** End Patch`

	ops, err := ParsePatch(patch)
	if err != nil {
		t.Fatalf("ParsePatch failed: %v", err)
	}
	if len(ops) != 1 || ops[0].Type != "delete" {
		t.Fatalf("expected 1 delete op, got %v", ops)
	}
}

func TestParsePatch_MultipleOps(t *testing.T) {
	patch := `*** Begin Patch
*** Add File: new.go
+package new
*** Update File: existing.go
@@ func Foo() {
-old
+new
*** Delete File: dead.go
*** End Patch`

	ops, err := ParsePatch(patch)
	if err != nil {
		t.Fatalf("ParsePatch failed: %v", err)
	}
	if len(ops) != 3 {
		t.Fatalf("expected 3 ops, got %d", len(ops))
	}
	if ops[0].Type != "add" || ops[1].Type != "update" || ops[2].Type != "delete" {
		t.Errorf("wrong op types: %s, %s, %s", ops[0].Type, ops[1].Type, ops[2].Type)
	}
}

func TestParsePatch_NoEndPatch(t *testing.T) {
	patch := `*** Begin Patch
*** Add File: test.go
+content`

	ops, err := ParsePatch(patch)
	if err != nil {
		t.Fatalf("should not error on missing End Patch: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("expected 1 op, got %d", len(ops))
	}
}

func TestParsePatch_Empty(t *testing.T) {
	_, err := ParsePatch("nothing here")
	if err == nil {
		t.Fatal("expected error for empty patch")
	}
}

func TestParsePatch_MoveFile(t *testing.T) {
	patch := `*** Begin Patch
*** Update File: old/path.go
*** Move to: new/path.go
@@ func X() {
-a
+b
*** End Patch`

	ops, err := ParsePatch(patch)
	if err != nil {
		t.Fatalf("ParsePatch failed: %v", err)
	}
	if ops[0].MoveTo != "new/path.go" {
		t.Errorf("expected MoveTo=new/path.go, got %q", ops[0].MoveTo)
	}
}

func TestApplyHunk_WithContext(t *testing.T) {
	content := `package main

func Run() {
    old line
    keep this
}`

	hunk := PatchHunk{
		Context:     "func Run() {",
		RemoveLines: []string{"    old line"},
		AddLines:    []string{"    new line"},
	}

	result, ok := applyHunk(content, hunk)
	if !ok {
		t.Fatal("applyHunk failed")
	}
	if !strings.Contains(result, "new line") {
		t.Errorf("expected 'new line' in result, got:\n%s", result)
	}
	if strings.Contains(result, "old line") {
		t.Errorf("'old line' should be removed, got:\n%s", result)
	}
	if !strings.Contains(result, "keep this") {
		t.Errorf("'keep this' should be preserved, got:\n%s", result)
	}
}

func TestApplyHunk_PureInsertion(t *testing.T) {
	content := `line1
line2
line3`

	hunk := PatchHunk{
		Context:  "line1",
		AddLines: []string{"inserted"},
	}

	result, ok := applyHunk(content, hunk)
	if !ok {
		t.Fatal("applyHunk failed")
	}
	lines := strings.Split(result, "\n")
	if len(lines) != 4 || lines[1] != "inserted" {
		t.Errorf("expected insertion after line1, got:\n%s", result)
	}
}

func TestApplyHunk_FuzzyWhitespace(t *testing.T) {
	content := "func Foo() {\n\t\told line  \n}"
	hunk := PatchHunk{
		Context:     "func Foo() {",
		RemoveLines: []string{"  old line"},
		AddLines:    []string{"\t\tnew line"},
	}

	result, ok := applyHunk(content, hunk)
	if !ok {
		t.Fatal("applyHunk should fuzzy-match whitespace")
	}
	if !strings.Contains(result, "new line") {
		t.Errorf("expected 'new line', got:\n%s", result)
	}
}

func TestValidatePatchPath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"src/main.go", false},
		{"./internal/app.go", false},
		{"/etc/passwd", true},
		{"../../../etc/passwd", true},
		{"../../secret.go", true},
	}

	for _, tt := range tests {
		err := validatePatchPath(tt.path)
		if (err != nil) != tt.wantErr {
			t.Errorf("validatePatchPath(%q): err=%v, wantErr=%v", tt.path, err, tt.wantErr)
		}
	}
}

func TestApplyPatch_Integration(t *testing.T) {
	// Create a temp directory and work in it
	tmpDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldCwd)

	// Create an existing file
	os.WriteFile("existing.go", []byte("package main\n\nfunc Hello() {\n\tprintln(\"hello\")\n}\n"), 0644)

	patch := `*** Begin Patch
*** Add File: new_file.go
+package main
+
+func World() {}

*** Update File: existing.go
@@ func Hello() {
-	println("hello")
+	println("hello world")

*** End Patch`

	result, err := ApplyPatch(patch)
	if err != nil {
		t.Fatalf("ApplyPatch failed: %v", err)
	}

	// Check new file created
	if _, err := os.Stat(filepath.Join(tmpDir, "new_file.go")); os.IsNotExist(err) {
		t.Error("new_file.go should have been created")
	}

	// Check existing file updated
	data, _ := os.ReadFile("existing.go")
	if !strings.Contains(string(data), "hello world") {
		t.Errorf("existing.go should contain 'hello world', got:\n%s", string(data))
	}

	if !strings.Contains(result, "Created new_file.go") {
		t.Errorf("result should mention created file: %s", result)
	}
	if !strings.Contains(result, "Updated existing.go") {
		t.Errorf("result should mention updated file: %s", result)
	}
}
