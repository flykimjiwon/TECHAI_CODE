package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	return os.WriteFile(absPath, []byte(content), 0644)
}

// FileEdit performs a search-and-replace edit on a file.
// Returns the number of replacements made.
func FileEdit(path, oldStr, newStr string) (int, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, fmt.Errorf("invalid path: %w", err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return 0, fmt.Errorf("read failed: %w", err)
	}
	content := string(data)
	count := strings.Count(content, oldStr)
	if count == 0 {
		// Show a snippet of the file for context
		preview := content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return 0, fmt.Errorf("old_string not found in %s. File preview:\n%s", path, preview)
	}
	newContent := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
		return 0, fmt.Errorf("write failed: %w", err)
	}
	return 1, nil
}

// ListFiles lists files in a directory, optionally recursive.
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
			if d.IsDir() && (strings.HasPrefix(name, ".") || name == "node_modules" || name == "dist" || name == "__pycache__") {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(absDir, path)
			if d.IsDir() {
				files = append(files, rel+"/")
			} else {
				files = append(files, rel)
			}
			if len(files) > 500 {
				return fmt.Errorf("too many files, showing first 500")
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
	return files, err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
