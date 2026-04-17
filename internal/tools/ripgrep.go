package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kimjiwon/tgc/internal/config"
)

// rgAvailable caches whether ripgrep is installed.
var rgAvailable *bool

// IsRipgrepAvailable checks if 'rg' is on PATH.
func IsRipgrepAvailable() bool {
	if rgAvailable != nil {
		return *rgAvailable
	}
	_, err := exec.LookPath("rg")
	result := err == nil
	rgAvailable = &result
	if result {
		config.DebugLog("[RG] ripgrep found on PATH")
	} else {
		config.DebugLog("[RG] ripgrep not found, using Go fallback")
	}
	return result
}

// rgMatch represents a single ripgrep JSON match.
type rgMatch struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		LineNumber int `json:"line_number"`
		Lines      struct {
			Text string `json:"text"`
		} `json:"lines"`
		Submatches []struct {
			Match struct {
				Text string `json:"text"`
			} `json:"match"`
		} `json:"submatches"`
	} `json:"data"`
}

// RipgrepSearch uses the rg binary for fast regex search.
// Falls back to Go implementation if rg is not available.
func RipgrepSearch(pattern, basePath, glob string, ignoreCase bool, contextLines int) (string, error) {
	if !IsRipgrepAvailable() {
		return "", fmt.Errorf("ripgrep not available")
	}

	if basePath == "" {
		basePath = "."
	}

	args := []string{
		"--json",
		"--hidden",
		"--no-messages",
		"--max-count", fmt.Sprintf("%d", maxGrepMatches),
		"--max-filesize", "5M",
	}

	if ignoreCase {
		args = append(args, "-i")
	}
	if contextLines > 0 {
		args = append(args, "-C", fmt.Sprintf("%d", contextLines))
	}
	if glob != "" {
		args = append(args, "--glob", glob)
	}

	// Exclude common noise directories
	args = append(args, "--glob", "!.git/*")
	args = append(args, "--glob", "!node_modules/*")
	args = append(args, "--glob", "!dist/*")
	args = append(args, "--glob", "!__pycache__/*")
	args = append(args, "--glob", "!vendor/*")

	args = append(args, pattern, basePath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rg", args...)
	output, err := cmd.Output()

	// rg exits with 1 when no matches found (not an error)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matches found.", nil
		}
		// Timeout or other error — fall back
		config.DebugLog("[RG] error: %v, falling back to Go", err)
		return "", fmt.Errorf("ripgrep error: %w", err)
	}

	// Parse JSON output
	var results strings.Builder
	matchCount := 0

	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		var m rgMatch
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if m.Type != "match" {
			continue
		}

		text := strings.TrimRight(m.Data.Lines.Text, "\n\r")
		text = truncateLine(text)
		results.WriteString(fmt.Sprintf("%s:%d:%s\n", m.Data.Path.Text, m.Data.LineNumber, text))
		matchCount++

		if matchCount >= maxGrepMatches || results.Len() >= maxGrepBytes {
			results.WriteString(fmt.Sprintf("\n... (truncated, %d+ matches)\n", matchCount))
			break
		}
	}

	if matchCount == 0 {
		return "No matches found.", nil
	}

	config.DebugLog("[RG] pattern=%q path=%s matches=%d bytes=%d", pattern, basePath, matchCount, results.Len())
	return results.String(), nil
}

// RipgrepFiles uses rg --files for fast file listing with glob.
func RipgrepFiles(pattern, basePath string) (string, error) {
	if !IsRipgrepAvailable() {
		return "", fmt.Errorf("ripgrep not available")
	}

	if basePath == "" {
		basePath = "."
	}

	args := []string{
		"--files",
		"--hidden",
		"--glob", "!.git/*",
		"--glob", "!node_modules/*",
		"--glob", "!dist/*",
		"--glob", pattern,
		basePath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rg", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ripgrep files error: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > maxGlobFiles {
		lines = lines[:maxGlobFiles]
		lines = append(lines, fmt.Sprintf("\n... (truncated at %d files)", maxGlobFiles))
	}

	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return "No files matched.", nil
	}

	config.DebugLog("[RG-FILES] pattern=%q path=%s matches=%d", pattern, basePath, len(lines))
	return strings.Join(lines, "\n"), nil
}
