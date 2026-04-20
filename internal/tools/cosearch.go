package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/kimjiwon/tgc/internal/config"
)

// coResult holds co-occurrence search results for a single file.
type coResult struct {
	relPath    string
	matchLines map[string][]int // term → line numbers
}

// CoSearch finds files containing ALL given terms (co-occurrence search).
// Solves the multi-line SQL problem: "find files where both TABLE_NAME and COLUMN_NAME appear".
func CoSearch(terms []string, basePath, glob string, ignoreCase bool) (string, error) {
	if len(terms) < 2 {
		return "", fmt.Errorf("at least 2 terms required")
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
		globRe, err = regexp.Compile(globToRegex(glob))
		if err != nil {
			return "", fmt.Errorf("invalid glob: %w", err)
		}
	}

	// Compile term patterns
	patterns := make([]*regexp.Regexp, len(terms))
	for i, term := range terms {
		flags := ""
		if ignoreCase {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + regexp.QuoteMeta(term))
		if err != nil {
			return "", fmt.Errorf("invalid term %q: %w", term, err)
		}
		patterns[i] = re
	}

	// Collect candidate files
	type candidate struct {
		absPath, relPath string
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
		info, _ := d.Info()
		if info != nil && info.Size() > 5*1024*1024 {
			return nil
		}
		candidates = append(candidates, candidate{absPath: path, relPath: rel})
		return nil
	})

	// Parallel scan for co-occurrence
	resultsCh := make(chan *coResult, len(candidates))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup

	for _, c := range candidates {
		wg.Add(1)
		go func(abs, rel string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := checkCoOccurrence(abs, rel, patterns, terms)
			if result != nil {
				resultsCh <- result
			} else {
				resultsCh <- nil
			}
		}(c.absPath, c.relPath)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect matching files
	var matches []coResult
	for r := range resultsCh {
		if r != nil {
			matches = append(matches, *r)
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No files contain all terms: %s", strings.Join(terms, " + ")), nil
	}

	// Format results
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d file(s) containing all terms [%s]:\n\n", len(matches), strings.Join(terms, " + ")))

	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("  %s\n", m.relPath))
		for _, term := range terms {
			lines := m.matchLines[term]
			if len(lines) > 3 {
				sb.WriteString(fmt.Sprintf("    %q → lines %v ... (%d total)\n", term, lines[:3], len(lines)))
			} else {
				sb.WriteString(fmt.Sprintf("    %q → lines %v\n", term, lines))
			}
		}
		sb.WriteString("\n")
	}

	config.DebugLog("[CO-SEARCH] terms=%v files=%d matches=%d", terms, len(candidates), len(matches))
	return sb.String(), nil
}

func checkCoOccurrence(absPath, relPath string, patterns []*regexp.Regexp, terms []string) *coResult {
	f, err := os.Open(absPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	matchLines := make(map[string][]int)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		for i, pat := range patterns {
			if pat.MatchString(line) {
				matchLines[terms[i]] = append(matchLines[terms[i]], lineNum)
			}
		}
	}

	// All terms must be present
	for _, term := range terms {
		if len(matchLines[term]) == 0 {
			return nil
		}
	}

	return &coResult{relPath: relPath, matchLines: matchLines}
}
