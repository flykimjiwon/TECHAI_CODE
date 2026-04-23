package main

import (
	"os"
	"path/filepath"
	"strings"
)

// KnowledgePack represents a knowledge document the user can toggle.
type KnowledgePack struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Enabled  bool   `json:"enabled"`
	Path     string `json:"path"`
}

// knowledgeState tracks which packs are enabled for this session.
var enabledPacks = map[string]bool{}

// knowledgeBasePath is the path to the TUI's knowledge docs.
// Can be set to the embedded location or a local override.
var knowledgeBasePath string

func init() {
	// Try to find knowledge docs relative to common locations
	candidates := []string{
		// Same parent as techai-ide binary
		"../택가이코드/knowledge/docs",
		// Dev location
		filepath.Join(os.Getenv("HOME"), "Desktop/kimjiwon/택가이코드/knowledge/docs"),
		// Relative to cwd
		"knowledge/docs",
	}
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			knowledgeBasePath = abs
			break
		}
	}
}

// GetKnowledgePacks scans the actual knowledge docs directory.
func (a *App) GetKnowledgePacks() []KnowledgePack {
	if knowledgeBasePath == "" {
		return nil
	}

	var packs []KnowledgePack

	categories, _ := os.ReadDir(knowledgeBasePath)
	for _, cat := range categories {
		if !cat.IsDir() {
			continue
		}
		catName := cat.Name()
		catPath := filepath.Join(knowledgeBasePath, catName)

		docs, _ := os.ReadDir(catPath)
		for _, doc := range docs {
			if doc.IsDir() || !strings.HasSuffix(doc.Name(), ".md") {
				continue
			}
			name := strings.TrimSuffix(doc.Name(), ".md")
			id := catName + "/" + name
			displayName := strings.ReplaceAll(name, "-", " ")
			// Title case
			words := strings.Split(displayName, " ")
			for i, w := range words {
				if len(w) > 0 {
					words[i] = strings.ToUpper(w[:1]) + w[1:]
				}
			}
			displayName = strings.Join(words, " ")

			packs = append(packs, KnowledgePack{
				ID:       id,
				Name:     displayName,
				Category: strings.ToUpper(catName[:1]) + catName[1:],
				Enabled:  enabledPacks[id],
				Path:     filepath.Join(catPath, doc.Name()),
			})
		}
	}

	// Auto-enable based on project type
	cwd := a.cwd
	if fileExists(filepath.Join(cwd, "go.mod")) {
		for i := range packs {
			if packs[i].Category == "Go" && !hasExplicitChoice(packs[i].ID) {
				packs[i].Enabled = true
			}
		}
	}
	if fileExists(filepath.Join(cwd, "package.json")) {
		for i := range packs {
			cat := packs[i].Category
			if (cat == "React" || cat == "Css" || cat == "Typescript" || cat == "Javascript") && !hasExplicitChoice(packs[i].ID) {
				packs[i].Enabled = true
			}
		}
	}

	return packs
}

var explicitChoices = map[string]bool{}

func hasExplicitChoice(id string) bool {
	_, ok := explicitChoices[id]
	return ok
}

// ToggleKnowledgePack enables/disables a knowledge pack.
func (a *App) ToggleKnowledgePack(id string, enabled bool) {
	enabledPacks[id] = enabled
	explicitChoices[id] = true
}

// getEnabledKnowledgeContext loads content of enabled packs into context.
func getEnabledKnowledgeContext() string {
	if knowledgeBasePath == "" {
		return ""
	}

	var sb strings.Builder
	totalSize := 0
	maxSize := 8000 // ~2000 tokens budget

	for id, on := range enabledPacks {
		if !on {
			continue
		}
		parts := strings.SplitN(id, "/", 2)
		if len(parts) != 2 {
			continue
		}
		docPath := filepath.Join(knowledgeBasePath, parts[0], parts[1]+".md")
		data, err := os.ReadFile(docPath)
		if err != nil {
			continue
		}
		content := string(data)
		if totalSize+len(content) > maxSize {
			// Truncate to fit budget
			remaining := maxSize - totalSize
			if remaining > 200 {
				content = content[:remaining] + "\n...[truncated]"
			} else {
				break
			}
		}
		sb.WriteString("\n\n## " + id + "\n")
		sb.WriteString(content)
		totalSize += len(content)
	}

	if sb.Len() == 0 {
		return ""
	}
	return "\n\n# Knowledge Reference\n" + sb.String()
}

// loadKnowledgeContext loads .techai.md + enabled packs.
func loadKnowledgeContext(cwd string) string {
	var sb strings.Builder

	if data, err := os.ReadFile(filepath.Join(cwd, ".techai.md")); err == nil {
		sb.WriteString("\n\n# Project Context (.techai.md)\n")
		sb.WriteString(string(data))
	}

	if fileExists(filepath.Join(cwd, "go.mod")) {
		sb.WriteString("\n\nProject type: Go")
	} else if fileExists(filepath.Join(cwd, "package.json")) {
		sb.WriteString("\n\nProject type: Node.js")
	} else if fileExists(filepath.Join(cwd, "requirements.txt")) {
		sb.WriteString("\n\nProject type: Python")
	}

	sb.WriteString(getEnabledKnowledgeContext())

	return sb.String()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
