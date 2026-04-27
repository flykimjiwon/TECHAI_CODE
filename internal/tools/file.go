package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/kimjiwon/tgc/internal/config"
)

func FileRead(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}
	return string(data), nil
}

func FileWrite(path, content string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}
	return SnapshotAndWrite(absPath, []byte(content))
}

// FileEdit performs a search-and-replace edit on a file with 4-stage fuzzy matching.
// Stage 1: ExactMatch — exact string match (fastest)
// Stage 2: LineTrimmed — ignore leading/trailing whitespace per line
// Stage 3: IndentFlex — normalize all indentation differences
// Stage 4: Levenshtein — accept 85%+ similarity match
// Returns the number of replacements made, a unified diff preview, and any error.
func FileEdit(path, oldStr, newStr string) (int, string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, "", fmt.Errorf("invalid path: %w", err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return 0, "", fmt.Errorf("read failed: %w", err)
	}
	content := string(data)

	applyEdit := func(original, matched, replacement, stage string) (int, string, error) {
		newContent := strings.Replace(original, matched, replacement, 1)
		diff := GenerateUnifiedDiff(path, original, newContent)
		if err := SnapshotAndWrite(absPath, []byte(newContent)); err != nil {
			return 0, "", fmt.Errorf("write failed: %w", err)
		}
		config.DebugLog("[FILE-EDIT] stage=%s path=%s", stage, path)
		return 1, diff, nil
	}

	// Stage 1: ExactMatch
	if strings.Contains(content, oldStr) {
		return applyEdit(content, oldStr, newStr, "ExactMatch")
	}

	// Stage 2: LineTrimmed — match ignoring leading/trailing whitespace per line
	if _, matchedOld := lineTrimmedFind(content, oldStr); matchedOld != "" {
		return applyEdit(content, matchedOld, newStr, "LineTrimmed")
	}

	// Stage 3: IndentFlex — normalize all indentation
	if _, matchedOld := indentFlexFind(content, oldStr); matchedOld != "" {
		return applyEdit(content, matchedOld, newStr, "IndentFlex")
	}

	// Stage 4: Levenshtein — 95%+ similarity (high threshold to avoid wrong-block matches)
	if matchedOld, similarity := levenshteinFind(content, oldStr, 0.95); matchedOld != "" {
		n, diff, err := applyEdit(content, matchedOld, newStr, fmt.Sprintf("Levenshtein(%.1f%%)", similarity*100))
		return n, diff, err
	}

	// All stages failed
	preview := content
	previewRunes := []rune(preview)
	if len(previewRunes) > 500 {
		preview = string(previewRunes[:500]) + "..."
	}
	return 0, "", fmt.Errorf("old_string not found in %s (tried 4 fuzzy stages). File preview:\n%s", path, preview)
}

// lineTrimmedFind finds oldStr in content by comparing lines with trimmed whitespace.
// Returns the start index and the actual matched substring from content.
func lineTrimmedFind(content, oldStr string) (int, string) {
	contentLines := strings.Split(content, "\n")
	searchLines := strings.Split(oldStr, "\n")
	if len(searchLines) == 0 {
		return -1, ""
	}

	trimmedSearch := make([]string, len(searchLines))
	for i, l := range searchLines {
		trimmedSearch[i] = strings.TrimSpace(l)
	}

	for i := 0; i <= len(contentLines)-len(searchLines); i++ {
		match := true
		for j, ts := range trimmedSearch {
			if strings.TrimSpace(contentLines[i+j]) != ts {
				match = false
				break
			}
		}
		if match {
			// Reconstruct the actual matched text from content
			matched := strings.Join(contentLines[i:i+len(searchLines)], "\n")
			idx := strings.Index(content, matched)
			return idx, matched
		}
	}
	return -1, ""
}

// indentFlexFind normalizes all whitespace at line starts before comparing.
func indentFlexFind(content, oldStr string) (int, string) {
	contentLines := strings.Split(content, "\n")
	searchLines := strings.Split(oldStr, "\n")
	if len(searchLines) == 0 {
		return -1, ""
	}

	// Normalize: collapse all leading whitespace to single space
	normalizeIndent := func(s string) string {
		trimmed := strings.TrimLeft(s, " \t")
		if len(trimmed) < len(s) {
			return " " + trimmed
		}
		return s
	}

	normSearch := make([]string, len(searchLines))
	for i, l := range searchLines {
		normSearch[i] = normalizeIndent(l)
	}

	for i := 0; i <= len(contentLines)-len(searchLines); i++ {
		match := true
		for j, ns := range normSearch {
			if normalizeIndent(contentLines[i+j]) != ns {
				match = false
				break
			}
		}
		if match {
			matched := strings.Join(contentLines[i:i+len(searchLines)], "\n")
			idx := strings.Index(content, matched)
			return idx, matched
		}
	}
	return -1, ""
}

// levenshteinFind searches content for a block similar to oldStr using Levenshtein distance.
// Only considers blocks of similar line count. Returns matched text and similarity ratio.
func levenshteinFind(content, oldStr string, threshold float64) (string, float64) {
	contentLines := strings.Split(content, "\n")
	searchLines := strings.Split(oldStr, "\n")
	searchLen := len(searchLines)
	if searchLen == 0 || searchLen > len(contentLines) {
		return "", 0
	}

	// For very large files/searches, skip to avoid O(n*m) explosion
	if utf8.RuneCountInString(oldStr) > 2000 {
		return "", 0
	}

	bestSimilarity := 0.0
	bestMatch := ""

	// Slide a window of searchLen lines across content
	for i := 0; i <= len(contentLines)-searchLen; i++ {
		candidate := strings.Join(contentLines[i:i+searchLen], "\n")
		sim := levenshteinSimilarity(candidate, oldStr)
		if sim > bestSimilarity {
			bestSimilarity = sim
			bestMatch = candidate
		}
		// Early exit on near-perfect match
		if sim > 0.98 {
			break
		}
	}

	if bestSimilarity >= threshold {
		return bestMatch, bestSimilarity
	}
	return "", 0
}

// levenshteinSimilarity returns similarity ratio (0.0 to 1.0) between two strings.
func levenshteinSimilarity(a, b string) float64 {
	ra := []rune(a)
	rb := []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 && lb == 0 {
		return 1.0
	}
	maxLen := la
	if lb > maxLen {
		maxLen = lb
	}

	// Optimize: if length difference alone exceeds threshold, skip
	diff := la - lb
	if diff < 0 {
		diff = -diff
	}
	if float64(diff)/float64(maxLen) > 0.3 {
		return 0.0
	}

	// Use two-row optimization for memory efficiency
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			min := del
			if ins < min {
				min = ins
			}
			if sub < min {
				min = sub
			}
			curr[j] = min
		}
		prev, curr = curr, prev
	}
	dist := prev[lb]
	return 1.0 - float64(dist)/float64(maxLen)
}

// ListFiles lists files in a directory, optionally recursive.
// No artificial file count limit — output truncation in registry.go handles size.
func ListFiles(dir string, recursive bool) ([]string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	info, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	var files []string
	if recursive {
		err = filepath.WalkDir(absDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // skip errors
			}
			// Skip hidden dirs and common noise
			name := d.Name()
			if d.IsDir() && (strings.HasPrefix(name, ".") || name == "node_modules" || name == "dist" || name == "__pycache__" || name == "vendor" || name == ".git") {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(absDir, path)
			if d.IsDir() {
				files = append(files, rel+"/")
			} else {
				files = append(files, rel)
			}
			return nil
		})
	} else {
		entries, err2 := os.ReadDir(absDir)
		if err2 != nil {
			return nil, fmt.Errorf("readdir failed: %w", err2)
		}
		for _, e := range entries {
			if e.IsDir() {
				files = append(files, e.Name()+"/")
			} else {
				files = append(files, e.Name())
			}
		}
	}

	// Append total count so LLM knows the full picture even if output gets truncated
	if recursive && len(files) > 500 {
		files = append(files, fmt.Sprintf("\n[Total: %d items]", len(files)))
	}

	return files, err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
