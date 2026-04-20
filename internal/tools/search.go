package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

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
	maxGrepMatches  = 300
	maxGrepBytes    = 60000
	maxGlobFiles    = 5000
	maxLineChars    = 2000 // truncate long lines (minified JS etc.)
)

// truncateLine limits a line to maxLineChars runes.
func truncateLine(line string) string {
	runes := []rune(line)
	if len(runes) > maxLineChars {
		return string(runes[:maxLineChars]) + "..."
	}
	return line
}

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

	// Phase 1: Collect candidate files (fast walk, no I/O)
	type candidate struct {
		absPath string
		relPath string
	}
	var candidates []candidate

	_ = filepath.WalkDir(absBase, func(path string, d os.DirEntry, err error) error {
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
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if binaryExts[ext] {
			return nil
		}
		if globRe != nil && !globRe.MatchString(rel) && !globRe.MatchString(d.Name()) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > 5*1024*1024 {
			return nil
		}
		candidates = append(candidates, candidate{absPath: path, relPath: rel})
		return nil
	})

	// Phase 2: Parallel file scanning with semaphore (gofis pattern)
	type fileResult struct {
		matches []string
	}
	resultsCh := make(chan fileResult, len(candidates))
	sem := make(chan struct{}, 8) // max 8 concurrent goroutines
	var wg sync.WaitGroup

	for _, c := range candidates {
		wg.Add(1)
		go func(abs, rel string) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			matches, err := searchFile(abs, rel, re, contextLines)
			if err != nil || len(matches) == 0 {
				resultsCh <- fileResult{}
				return
			}
			resultsCh <- fileResult{matches: matches}
		}(c.absPath, c.relPath)
	}

	// Close channel when all goroutines done
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Phase 3: Collect and group results by file
	type fileMatches struct {
		file    string
		lines   []string
	}
	fileMap := make(map[string]*fileMatches)
	var fileOrder []string
	matchCount := 0
	filesScanned := len(candidates)

	for fr := range resultsCh {
		for _, m := range fr.matches {
			if matchCount >= maxGrepMatches {
				goto done
			}
			// Parse "file:line:content" format
			parts := strings.SplitN(m, ":", 3)
			if len(parts) < 3 {
				continue
			}
			file := parts[0]
			lineInfo := parts[1] + ": " + strings.TrimSpace(parts[2])

			if _, exists := fileMap[file]; !exists {
				fileMap[file] = &fileMatches{file: file}
				fileOrder = append(fileOrder, file)
			}
			fm := fileMap[file]
			if len(fm.lines) < 10 { // max 10 matches per file
				fm.lines = append(fm.lines, lineInfo)
			}
			matchCount++
		}
	}
done:
	// Drain remaining results to unblock goroutines on early exit
	go func() {
		for range resultsCh {
		}
	}()

	if matchCount == 0 {
		return "No matches found.", nil
	}

	// Format grouped output
	var results strings.Builder
	matchingFiles := len(fileOrder)
	for _, file := range fileOrder {
		if results.Len() >= maxGrepBytes {
			results.WriteString("\n... (truncated)\n")
			break
		}
		fm := fileMap[file]
		results.WriteString(fmt.Sprintf("\n%s (%d matches):\n", fm.file, len(fm.lines)))
		for _, line := range fm.lines {
			results.WriteString(fmt.Sprintf("  :%s\n", line))
		}
	}

	config.DebugLog("[GREP] pattern=%q path=%s matches=%d files=%d/%d parallel=8", pattern, basePath, matchCount, matchingFiles, filesScanned)
	results.WriteString(fmt.Sprintf("\n(%d matches in %d files, %d files scanned)", matchCount, matchingFiles, filesScanned))
	return results.String(), nil
}

// searchFile scans a single file for regex matches.
// Times out after 3 seconds to prevent large files from blocking.
func searchFile(absPath, relPath string, re *regexp.Regexp, contextLines int) ([]string, error) {
	start := time.Now()
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
		// Timeout check every 1000 lines (3 second limit per file)
		if lineNum%1000 == 0 && time.Since(start) > 3*time.Second {
			config.DebugLog("[GREP] file timeout: %s (%d lines, %.1fs)", relPath, lineNum, time.Since(start).Seconds())
			break
		}
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
			results = append(results, fmt.Sprintf("%s:%d:%s", relPath, ln, truncateLine(allLines[ln-1])))
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
					results = append(results, fmt.Sprintf("%s:%d:%s %s", relPath, i, prefix, truncateLine(allLines[i-1])))
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

	type fileMatch struct {
		rel   string
		mtime time.Time
	}
	var matches []fileMatch

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
			mt := time.Time{}
			if info, err := d.Info(); err == nil {
				mt = info.ModTime()
			}
			matches = append(matches, fileMatch{rel: rel, mtime: mt})
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

	// Sort by mtime descending (newest first) — like OpenCode
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].mtime.After(matches[j].mtime)
	})

	var names []string
	for _, m := range matches {
		names = append(names, m.rel)
	}
	result := strings.Join(names, "\n")
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
