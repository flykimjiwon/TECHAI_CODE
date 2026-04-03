package tools

import (
	"fmt"
	"os"
	"path/filepath"
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

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
