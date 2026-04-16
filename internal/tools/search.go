package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kimjiwon/tgc/internal/config"
)

// Directories to skip during search traversal.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"dist":         true,
	"__pycache__":  true,
	".next":        true,
	"vendor":       true,
	".omc":         true,
}

// Binary file extensions to skip.
var binaryExts = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true,
	".bin": true, ".o": true, ".a": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".ico": true,
	".zip": true, ".tar": true, ".gz": true, ".7z": true, ".rar": true,
	".pdf": true, ".doc": true, ".docx": true,
	".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
	".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
	".wasm": true,
}

const (
	maxGrepMatches = 100
	maxGrepBytes   = 30000
	maxGlobFiles   = 2000
)

// GrepSearch searches file contents by regex pattern.
// Returns matches in "file:line:content" format.
// Respects .gitignore if present, falls back to hardcoded skipDirs.
func GrepSearch(pattern, basePath, glob string, ignoreCase bool, contextLines int) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	flags := ""
	if ignoreCase {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex: %w", err)
	}

	if basePath == "" {
		basePath = "."
	}
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	gi := LoadGitIgnore(absBase)

	var globRe *regexp.Regexp
	if glob != "" {
		globPattern := globToRegex(glob)
		globRe, err = regexp.Compile(globPattern)
		if err != nil {
			return "", fmt.Errorf("invalid glob filter %q: %w", glob, err)
		}
	}

	var results strings.Builder
	matchCount := 0

	walkErr := filepath.WalkDir(absBase, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(absBase, path)
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if shouldSkip(gi, rel, true, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip by gitignore or fallback
		if shouldSkip(gi, rel, false, d.Name()) {
			return nil
		}

		// Skip binary files
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if binaryExts[ext] {
			return nil
		}

		// Apply glob filter
		if globRe != nil && !globRe.MatchString(rel) && !globRe.MatchString(d.Name()) {
			return nil
		}

		// Skip large files (>1MB)
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > 1024*1024 {
			return nil
		}

		matches, err := searchFile(path, rel, re, contextLines)
		if err != nil {
			return nil
		}

		for _, m := range matches {
			if matchCount >= maxGrepMatches || results.Len() >= maxGrepBytes {
				results.WriteString(fmt.Sprintf("\n... (truncated, %d+ matches)\n", matchCount))
				return fmt.Errorf("limit reached")
			}
			results.WriteString(m)
			results.WriteString("\n")
			matchCount++
		}

		return nil
	})

	if walkErr != nil && walkErr.Error() != "limit reached" {
		config.DebugLog("[GREP] walk error: %v", walkErr)
	}

	if matchCount == 0 {
		return "No matches found.", nil
	}

	config.DebugLog("[GREP] pattern=%q path=%s matches=%d bytes=%d gitignore=%v", pattern, basePath, matchCount, results.Len(), gi != nil)
	return results.String(), nil
}

// searchFile scans a single file for regex matches.
func searchFile(absPath, relPath string, re *regexp.Regexp, contextLines int) ([]string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var allLines []string
	var matchLineNums []int

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		allLines = append(allLines, line)
		if re.MatchString(line) {
			matchLineNums = append(matchLineNums, lineNum)
		}
	}

	if len(matchLineNums) == 0 {
		return nil, nil
	}

	var results []string
	if contextLines <= 0 {
		for _, ln := range matchLineNums {
			results = append(results, fmt.Sprintf("%s:%d:%s", relPath, ln, allLines[ln-1]))
		}
	} else {
		// With context lines, group nearby matches
		shown := make(map[int]bool)
		for _, ln := range matchLineNums {
			start := ln - contextLines
			if start < 1 {
				start = 1
			}
			end := ln + contextLines
			if end > len(allLines) {
				end = len(allLines)
			}
			for i := start; i <= end; i++ {
				if !shown[i] {
					prefix := " "
					if i == ln {
						prefix = ">"
					}
					results = append(results, fmt.Sprintf("%s:%d:%s %s", relPath, i, prefix, allLines[i-1]))
					shown[i] = true
				}
			}
		}
	}

	return results, nil
}

// GlobSearch finds files matching a glob pattern (supports **).
// Respects .gitignore if present, falls back to hardcoded skipDirs.
func GlobSearch(pattern, basePath string) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	if basePath == "" {
		basePath = "."
	}
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	gi := LoadGitIgnore(absBase)

	// Convert glob pattern to regex for ** support
	globRe, err := regexp.Compile(globToRegex(pattern))
	if err != nil {
		return "", fmt.Errorf("invalid glob pattern: %w", err)
	}

	var matches []string

	walkErr := filepath.WalkDir(absBase, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(absBase, path)
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if shouldSkip(gi, rel, true, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldSkip(gi, rel, false, d.Name()) {
			return nil
		}

		if globRe.MatchString(rel) {
			matches = append(matches, rel)
			if len(matches) >= maxGlobFiles {
				return fmt.Errorf("limit reached")
			}
		}
		return nil
	})

	if walkErr != nil && walkErr.Error() != "limit reached" {
		config.DebugLog("[GLOB] walk error: %v", walkErr)
	}

	if len(matches) == 0 {
		return "No files matched.", nil
	}

	result := strings.Join(matches, "\n")
	if len(matches) >= maxGlobFiles {
		result += fmt.Sprintf("\n... (truncated at %d files)", maxGlobFiles)
	}

	config.DebugLog("[GLOB] pattern=%q path=%s matches=%d gitignore=%v", pattern, basePath, len(matches), gi != nil)
	return result, nil
}

// globToRegex converts a glob pattern to a regex string.
// Supports: ** (any path), * (any non-separator), ? (single char), {a,b} (alternation)
func globToRegex(pattern string) string {
	var b strings.Builder
	b.WriteString("^")

	braceDepth := 0
	i := 0
	for i < len(pattern) {
		ch := pattern[i]
		switch ch {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				// ** matches any number of path segments
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					b.WriteString("(.*/)?")
					i += 3
				} else {
					b.WriteString(".*")
					i += 2
				}
			} else {
				// * matches anything except /
				b.WriteString("[^/]*")
				i++
			}
		case '?':
			b.WriteString("[^/]")
			i++
		case '.':
			b.WriteString("\\.")
			i++
		case '\\':
			b.WriteString("\\\\")
			i++
		case '{':
			// {a,b,c} → (a|b|c)
			b.WriteString("(")
			braceDepth++
			i++
		case '}':
			b.WriteString(")")
			if braceDepth > 0 {
				braceDepth--
			}
			i++
		case ',':
			if braceDepth > 0 {
				b.WriteString("|")
			} else {
				b.WriteByte(',') // literal comma outside braces
			}
			i++
		default:
			b.WriteByte(ch)
			i++
		}
	}

	b.WriteString("$")
	return b.String()
}
