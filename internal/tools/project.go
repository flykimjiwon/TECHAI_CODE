package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjiwon/tgc/internal/config"
)

// ProjectInfo describes the detected project type and key files.
type ProjectInfo struct {
	Type       string   // "go", "node", "python", "rust", "java", "unknown"
	Name       string   // project name if detectable
	Framework  string   // "nextjs", "react", "vue", "django", "flask", "gin", etc.
	KeyFiles   []string // important config files found
}

// DetectProject scans the working directory for project markers.
func DetectProject(dir string) ProjectInfo {
	if dir == "" {
		dir = "."
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ProjectInfo{Type: "unknown"}
	}

	info := ProjectInfo{Type: "unknown"}

	// Check for project markers in priority order
	markers := []struct {
		file      string
		projType  string
		framework string
	}{
		{"go.mod", "go", ""},
		{"Cargo.toml", "rust", ""},
		{"package.json", "node", ""},
		{"pyproject.toml", "python", ""},
		{"requirements.txt", "python", ""},
		{"setup.py", "python", ""},
		{"pom.xml", "java", "maven"},
		{"build.gradle", "java", "gradle"},
		{"Gemfile", "ruby", ""},
		{"composer.json", "php", ""},
	}

	for _, m := range markers {
		path := filepath.Join(absDir, m.file)
		if _, err := os.Stat(path); err == nil {
			info.Type = m.projType
			info.KeyFiles = append(info.KeyFiles, m.file)
			if m.framework != "" {
				info.Framework = m.framework
			}
			break
		}
	}

	// Detect framework from contents
	if info.Type == "node" {
		info.Framework = detectNodeFramework(absDir)
		info.Name = detectPackageName(absDir)
	} else if info.Type == "go" {
		info.Name = detectGoModule(absDir)
	} else if info.Type == "python" {
		info.Framework = detectPythonFramework(absDir)
	}

	// Additional key files
	extras := []string{
		"Makefile", "Dockerfile", "docker-compose.yml",
		".env", ".env.local", "tsconfig.json",
		"next.config.js", "next.config.ts", "next.config.mjs",
		"vite.config.ts", "webpack.config.js",
	}
	for _, f := range extras {
		if _, err := os.Stat(filepath.Join(absDir, f)); err == nil {
			info.KeyFiles = append(info.KeyFiles, f)
		}
	}

	config.DebugLog("[PROJECT] type=%s framework=%s name=%s keyFiles=%v", info.Type, info.Framework, info.Name, info.KeyFiles)
	return info
}

// FormatProjectContext returns a string suitable for system prompt injection.
func FormatProjectContext(info ProjectInfo) string {
	if info.Type == "unknown" {
		return ""
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("Project type: %s", info.Type))
	if info.Name != "" {
		parts = append(parts, fmt.Sprintf("Project name: %s", info.Name))
	}
	if info.Framework != "" {
		parts = append(parts, fmt.Sprintf("Framework: %s", info.Framework))
	}
	if len(info.KeyFiles) > 0 {
		parts = append(parts, fmt.Sprintf("Key files: %s", strings.Join(info.KeyFiles, ", ")))
	}
	return strings.Join(parts, " | ")
}

func detectNodeFramework(dir string) string {
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return ""
	}
	content := string(data)
	switch {
	case strings.Contains(content, `"next"`):
		return "nextjs"
	case strings.Contains(content, `"nuxt"`):
		return "nuxt"
	case strings.Contains(content, `"@angular/core"`):
		return "angular"
	case strings.Contains(content, `"vue"`):
		return "vue"
	case strings.Contains(content, `"svelte"`):
		return "svelte"
	case strings.Contains(content, `"react"`):
		return "react"
	case strings.Contains(content, `"express"`):
		return "express"
	case strings.Contains(content, `"fastify"`):
		return "fastify"
	}
	return ""
}

func detectPackageName(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	// Simple extraction without full JSON parse
	content := string(data)
	if idx := strings.Index(content, `"name"`); idx >= 0 {
		rest := content[idx+6:]
		if start := strings.IndexByte(rest, '"'); start >= 0 {
			rest = rest[start+1:]
			if end := strings.IndexByte(rest, '"'); end >= 0 {
				return rest[:end]
			}
		}
	}
	return ""
}

func detectGoModule(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func detectPythonFramework(dir string) string {
	// Check requirements.txt or pyproject.toml
	for _, f := range []string{"requirements.txt", "pyproject.toml"} {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			continue
		}
		content := strings.ToLower(string(data))
		switch {
		case strings.Contains(content, "django"):
			return "django"
		case strings.Contains(content, "flask"):
			return "flask"
		case strings.Contains(content, "fastapi"):
			return "fastapi"
		}
	}
	return ""
}
