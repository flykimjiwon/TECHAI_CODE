package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Diagnostic represents a single code issue found by a linter or compiler.
type Diagnostic struct {
	File     string
	Line     int
	Column   int
	Severity string // "error" or "warning"
	Message  string
	Source   string // "go", "tsc", "eslint", "ruff"
}

// RunDiagnostics auto-detects project type and runs appropriate linters.
func RunDiagnostics(dir string, targetFile string) (string, error) {
	projectType := detectProjectType(dir)
	var diagnostics []Diagnostic

	switch projectType {
	case "go":
		d, _ := runGoVet(dir)
		diagnostics = append(diagnostics, d...)
	case "typescript":
		d, _ := runTsc(dir)
		diagnostics = append(diagnostics, d...)
	case "javascript":
		d, _ := runEslint(dir, targetFile)
		diagnostics = append(diagnostics, d...)
	case "python":
		d, _ := runRuff(dir, targetFile)
		diagnostics = append(diagnostics, d...)
	default:
		return "Unknown project type. No diagnostics available.", nil
	}

	// Filter by target file if specified
	if targetFile != "" {
		var filtered []Diagnostic
		for _, d := range diagnostics {
			if strings.HasSuffix(d.File, targetFile) || d.File == targetFile {
				filtered = append(filtered, d)
			}
		}
		diagnostics = filtered
	}

	return formatDiagnostics(diagnostics), nil
}

func detectProjectType(dir string) string {
	if diagFileExists(filepath.Join(dir, "go.mod")) {
		return "go"
	}
	if diagFileExists(filepath.Join(dir, "tsconfig.json")) {
		return "typescript"
	}
	if diagFileExists(filepath.Join(dir, "package.json")) {
		return "javascript"
	}
	if diagFileExists(filepath.Join(dir, "pyproject.toml")) {
		return "python"
	}
	if diagFileExists(filepath.Join(dir, "requirements.txt")) {
		return "python"
	}
	return "unknown"
}

func diagFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// runGoVet runs "go vet ./..." and parses output.
func runGoVet(dir string) ([]Diagnostic, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "vet", "./...")
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return parseGoDiagnostics(string(out)), nil
}

// parseGoDiagnostics parses "file.go:line:col: message" format.
func parseGoDiagnostics(output string) []Diagnostic {
	var diags []Diagnostic
	re := regexp.MustCompile(`^(.+\.go):(\d+):(\d+):\s*(.+)$`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		lineNum, _ := strconv.Atoi(matches[2])
		col, _ := strconv.Atoi(matches[3])
		msg := matches[4]
		severity := "warning"
		if strings.Contains(strings.ToLower(msg), "error") {
			severity = "error"
		}
		diags = append(diags, Diagnostic{
			File:     matches[1],
			Line:     lineNum,
			Column:   col,
			Severity: severity,
			Message:  msg,
			Source:   "go",
		})
	}
	return diags
}

// runTsc runs "npx tsc --noEmit --pretty false" and parses output.
func runTsc(dir string) ([]Diagnostic, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "npx", "tsc", "--noEmit", "--pretty", "false")
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return parseTscDiagnostics(string(out)), nil
}

// parseTscDiagnostics parses "file(line,col): error TSxxxx: message" format.
func parseTscDiagnostics(output string) []Diagnostic {
	var diags []Diagnostic
	re := regexp.MustCompile(`^(.+)\((\d+),(\d+)\):\s*(error|warning)\s+TS\d+:\s*(.+)$`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		lineNum, _ := strconv.Atoi(matches[2])
		col, _ := strconv.Atoi(matches[3])
		diags = append(diags, Diagnostic{
			File:     matches[1],
			Line:     lineNum,
			Column:   col,
			Severity: matches[4],
			Message:  matches[5],
			Source:   "tsc",
		})
	}
	return diags
}

// runEslint runs "npx eslint --format compact" and parses output.
func runEslint(dir, file string) ([]Diagnostic, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	target := "."
	if file != "" {
		target = file
	}
	cmd := exec.CommandContext(ctx, "npx", "eslint", "--format", "compact", target)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return parseEslintDiagnostics(string(out)), nil
}

// parseEslintDiagnostics parses eslint compact format:
// "file: line X, col Y, Severity - Message (rule)"
func parseEslintDiagnostics(output string) []Diagnostic {
	var diags []Diagnostic
	re := regexp.MustCompile(`^(.+):\s*line\s+(\d+),\s*col\s+(\d+),\s*(Error|Warning)\s*-\s*(.+)$`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		lineNum, _ := strconv.Atoi(matches[2])
		col, _ := strconv.Atoi(matches[3])
		severity := strings.ToLower(matches[4])
		diags = append(diags, Diagnostic{
			File:     matches[1],
			Line:     lineNum,
			Column:   col,
			Severity: severity,
			Message:  matches[5],
			Source:   "eslint",
		})
	}
	return diags
}

// runRuff runs "ruff check --output-format text" and parses output.
func runRuff(dir, file string) ([]Diagnostic, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	target := "."
	if file != "" {
		target = file
	}
	cmd := exec.CommandContext(ctx, "ruff", "check", "--output-format", "text", target)
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	return parseRuffDiagnostics(string(out)), nil
}

// parseRuffDiagnostics parses "file:line:col: CODE message" format.
func parseRuffDiagnostics(output string) []Diagnostic {
	var diags []Diagnostic
	re := regexp.MustCompile(`^(.+):(\d+):(\d+):\s*([A-Z]\d+)\s+(.+)$`)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		lineNum, _ := strconv.Atoi(matches[2])
		col, _ := strconv.Atoi(matches[3])
		// Ruff codes starting with E or F are errors, others are warnings
		severity := "warning"
		code := matches[4]
		if strings.HasPrefix(code, "E") || strings.HasPrefix(code, "F") {
			severity = "error"
		}
		diags = append(diags, Diagnostic{
			File:     matches[1],
			Line:     lineNum,
			Column:   col,
			Severity: severity,
			Message:  fmt.Sprintf("[%s] %s", code, matches[5]),
			Source:   "ruff",
		})
	}
	return diags
}

func formatDiagnostics(diags []Diagnostic) string {
	if len(diags) == 0 {
		return "No issues found."
	}
	var sb strings.Builder
	errors, warnings := 0, 0
	for _, d := range diags {
		icon := "warning"
		if d.Severity == "error" {
			icon = "error"
			errors++
		} else {
			warnings++
		}
		sb.WriteString(fmt.Sprintf("[%s] %s:%d:%d [%s] %s\n", icon, d.File, d.Line, d.Column, d.Source, d.Message))
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d errors, %d warnings", errors, warnings))
	return sb.String()
}
