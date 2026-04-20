package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kimjiwon/tgc/internal/config"
)

// FuzzyMatch represents a fuzzy-matched file with a score.
type FuzzyMatch struct {
	Path    string
	Score   int
	ModTime time.Time
}

// FuzzyFind searches for files whose names fuzzy-match the query.
// "apgo" matches "app.go", "regst" matches "registry.go".
func FuzzyFind(query, basePath string, maxResults int) (string, error) {
	if query == "" {
		return "", nil
	}
	if basePath == "" {
		basePath = "."
	}
	if maxResults <= 0 {
		maxResults = 20
	}

	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", err
	}

	gi := LoadGitIgnore(absBase)
	queryLower := strings.ToLower(query)

	var matches []FuzzyMatch

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

		// Skip binary files
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if binaryExts[ext] {
			return nil
		}

		score := fuzzyScore(queryLower, strings.ToLower(rel))
		if score > 0 {
			mt := time.Time{}
			if info, err := d.Info(); err == nil {
				mt = info.ModTime()
			}
			matches = append(matches, FuzzyMatch{Path: rel, Score: score, ModTime: mt})
		}

		// Collect enough candidates then stop
		if len(matches) > maxResults*10 {
			return fmt.Errorf("enough candidates")
		}
		return nil
	})

	if len(matches) == 0 {
		return "No files matched.", nil
	}

	// Sort by score descending, then mtime descending
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].ModTime.After(matches[j].ModTime)
	})

	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	var sb strings.Builder
	for _, m := range matches {
		sb.WriteString(m.Path)
		sb.WriteString("\n")
	}

	config.DebugLog("[FUZZY] query=%q matches=%d", query, len(matches))
	return sb.String(), nil
}

// fuzzyScore returns a score for how well needle matches haystack.
// Higher is better. 0 means no match.
// Consecutive character matches and word-boundary matches score higher.
func fuzzyScore(needle, haystack string) int {
	if len(needle) == 0 {
		return 0
	}

	// Exact substring match — highest score
	if strings.Contains(haystack, needle) {
		return 1000 + len(needle)*10
	}

	// Basename match scores higher than full path match
	base := filepath.Base(haystack)
	if strings.Contains(base, needle) {
		return 900 + len(needle)*10
	}

	// Fuzzy character-by-character matching
	needleRunes := []rune(needle)
	score := 0
	needleIdx := 0
	consecutive := 0
	prevMatched := false

	for i, ch := range haystack {
		if needleIdx >= len(needleRunes) {
			break
		}

		if ch == needleRunes[needleIdx] {
			needleIdx++
			if prevMatched {
				consecutive++
				score += 5 + consecutive*3 // consecutive bonus
			} else {
				consecutive = 0
				score += 5
			}
			// Word boundary bonus (after /, ., _, -)
			if i > 0 {
				prev := rune(haystack[i-1])
				if prev == '/' || prev == '.' || prev == '_' || prev == '-' {
					score += 10
				}
			}
			if i == 0 {
				score += 15 // first char match
			}
			prevMatched = true
		} else {
			prevMatched = false
			consecutive = 0
		}
	}

	// All characters must match
	if needleIdx < len(needleRunes) {
		return 0
	}

	return score
}
