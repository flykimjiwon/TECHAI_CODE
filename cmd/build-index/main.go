package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	Path     string   `json:"path"`
	Title    string   `json:"title"`
	Tier     int      `json:"tier"`
	OS       string   `json:"os,omitempty"`
	Keywords []string `json:"keywords"`
	Tokens   int      `json:"tokens"`
}

type IndexFile struct {
	Documents []Document `json:"documents"`
}

func main() {
	knowledgeDir := "knowledge"
	if len(os.Args) > 1 {
		knowledgeDir = os.Args[1]
	}

	var idx IndexFile

	err := filepath.Walk(knowledgeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "WARN: skip %s: %v\n", path, readErr)
			return nil
		}

		// Use forward slashes for consistency
		relPath := filepath.ToSlash(strings.TrimPrefix(path, knowledgeDir+string(filepath.Separator)))
		// Prepend "knowledge/" to match embed.FS paths
		embeddedPath := "knowledge/" + relPath

		doc := Document{
			Path:   embeddedPath,
			Tokens: (len(content) + 3) / 4, // ~4 chars per token, ceiling
		}

		// Extract title from first # heading
		for _, line := range strings.SplitN(string(content), "\n", 20) {
			if strings.HasPrefix(line, "# ") {
				doc.Title = strings.TrimPrefix(line, "# ")
				break
			}
		}

		// Parse path for metadata: knowledge/{category}/{subcategory}/filename.md
		parts := strings.Split(relPath, "/")

		// Determine tier, OS, keywords from path structure
		if len(parts) >= 2 {
			category := parts[0] // "docs" or "skills"
			if category == "docs" && len(parts) >= 3 {
				subcategory := parts[1]
				doc.Keywords = append(doc.Keywords, subcategory)

				switch subcategory {
				case "bxm":
					doc.Tier = 0
				case "go", "javascript", "typescript", "react", "css", "charts":
					doc.Tier = 1
				case "vue", "java":
					doc.Tier = 2
				case "python":
					doc.Tier = 3
				case "terminal":
					doc.Tier = 1
					fname := strings.TrimSuffix(parts[len(parts)-1], ".md")
					switch fname {
					case "windows":
						doc.OS = "windows"
					case "linux":
						doc.OS = "linux"
					case "macos":
						doc.OS = "darwin"
					}
				default:
					doc.Tier = 2
				}
			} else if category == "skills" {
				doc.Tier = 1
			}
		}

		// Add filename (without .md) as keyword
		fname := strings.TrimSuffix(parts[len(parts)-1], ".md")
		doc.Keywords = append(doc.Keywords, fname)

		// Split hyphenated filenames into extra keywords
		for _, part := range strings.Split(fname, "-") {
			lp := strings.ToLower(part)
			if lp != "" && !contains(doc.Keywords, lp) {
				doc.Keywords = append(doc.Keywords, lp)
			}
		}

		idx.Documents = append(idx.Documents, doc)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join(knowledgeDir, "index.json")
	data, _ := json.MarshalIndent(idx, "", "  ")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s: %d documents indexed\n", outPath, len(idx.Documents))
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
