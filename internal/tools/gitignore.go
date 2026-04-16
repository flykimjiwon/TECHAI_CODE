package tools

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kimjiwon/tgc/internal/config"
)

// GitIgnore holds compiled patterns from .gitignore files.
type GitIgnore struct {
	patterns []ignorePattern
}

type ignorePattern struct {
	re     *regexp.Regexp
	negate bool // lines starting with '!'
	dirOnly bool // lines ending with '/'
}

// LoadGitIgnore reads .gitignore files from the base directory up to the git root.
// Returns nil if no .gitignore found (caller falls back to skipDirs).
func LoadGitIgnore(basePath string) *GitIgnore {
	gi := &GitIgnore{}
	loaded := 0

	// Walk up from basePath looking for .gitignore files
	dir, err := filepath.Abs(basePath)
	if err != nil {
		return nil
	}

	// Collect .gitignore paths bottom-up, then apply top-down
	var paths []string
	for {
		candidate := filepath.Join(dir, ".gitignore")
		if _, err := os.Stat(candidate); err == nil {
			paths = append(paths, candidate)
		}
		// Stop at git root or filesystem root
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Apply top-down (reverse order) so deeper .gitignore overrides
	for i := len(paths) - 1; i >= 0; i-- {
		if err := gi.loadFile(paths[i]); err == nil {
			loaded++
		}
	}

	if loaded == 0 {
		return nil
	}

	config.DebugLog("[GITIGNORE] loaded %d .gitignore file(s) with %d patterns", loaded, len(gi.patterns))
	return gi
}

func (gi *GitIgnore) loadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		negate := false
		if strings.HasPrefix(line, "!") {
			negate = true
			line = line[1:]
		}

		dirOnly := false
		if strings.HasSuffix(line, "/") {
			dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		re, err := gitignorePatternToRegex(line)
		if err != nil {
			continue
		}

		gi.patterns = append(gi.patterns, ignorePattern{
			re:      re,
			negate:  negate,
			dirOnly: dirOnly,
		})
	}
	return nil
}

// ShouldIgnore checks if a relative path should be ignored.
// isDir indicates whether the path is a directory.
func (gi *GitIgnore) ShouldIgnore(relPath string, isDir bool) bool {
	if gi == nil {
		return false
	}

	// Normalize to forward slashes
	relPath = filepath.ToSlash(relPath)

	ignored := false
	for _, p := range gi.patterns {
		if p.dirOnly && !isDir {
			continue
		}
		// Match against full path and also just the basename
		name := filepath.Base(relPath)
		if p.re.MatchString(relPath) || p.re.MatchString(name) {
			if p.negate {
				ignored = false
			} else {
				ignored = true
			}
		}
	}
	return ignored
}

// gitignorePatternToRegex converts a gitignore glob pattern to a regex.
// Supports: * (non-slash), ** (anything including /), ? (single char)
func gitignorePatternToRegex(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder

	// If pattern contains /, it's anchored to path; otherwise matches basename anywhere
	anchored := strings.Contains(pattern, "/")
	pattern = strings.TrimPrefix(pattern, "/")

	if !anchored {
		b.WriteString("(^|/)")
	} else {
		b.WriteString("^")
	}

	i := 0
	for i < len(pattern) {
		ch := pattern[i]
		switch ch {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					b.WriteString("(.*/)?")
					i += 3
				} else {
					b.WriteString(".*")
					i += 2
				}
			} else {
				b.WriteString("[^/]*")
				i++
			}
		case '?':
			b.WriteString("[^/]")
			i++
		case '.':
			b.WriteString("\\.")
			i++
		case '[':
			// Pass through character classes
			end := strings.IndexByte(pattern[i:], ']')
			if end > 0 {
				b.WriteString(pattern[i : i+end+1])
				i += end + 1
			} else {
				b.WriteString("\\[")
				i++
			}
		default:
			if strings.ContainsRune(`+{}()^$|\\`, rune(ch)) {
				b.WriteByte('\\')
			}
			b.WriteByte(ch)
			i++
		}
	}

	b.WriteString("(/.*)?$")

	return regexp.Compile(b.String())
}

// shouldSkip checks a path against gitignore (if loaded) or falls back to hardcoded skipDirs.
func shouldSkip(gi *GitIgnore, relPath string, isDir bool, name string) bool {
	// .git is always skipped
	if name == ".git" {
		return true
	}

	if gi != nil {
		return gi.ShouldIgnore(relPath, isDir)
	}

	// Fallback to hardcoded skip list
	if isDir {
		return skipDirs[name]
	}
	return false
}
