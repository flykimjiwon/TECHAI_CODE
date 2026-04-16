package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kimjiwon/tgc/internal/config"
)

// MemoryEntry represents a single remembered fact.
type MemoryEntry struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	HitCount  int       `json:"hit_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoryStore manages project-local and global memories.
type MemoryStore struct {
	LocalPath  string
	GlobalPath string
	Local      []MemoryEntry `json:"local"`
	Global     []MemoryEntry `json:"global"`
}

const maxMemories = 100

// NewMemoryStore loads memories from project-local and global paths.
func NewMemoryStore() *MemoryStore {
	globalPath := filepath.Join(config.ConfigDir(), "memories.json")
	localPath := filepath.Join(".tgc", "memories.json")

	ms := &MemoryStore{
		LocalPath:  localPath,
		GlobalPath: globalPath,
	}
	ms.load()
	config.DebugLog("[MEMORY] loaded local=%d global=%d", len(ms.Local), len(ms.Global))
	return ms
}

func (ms *MemoryStore) load() {
	ms.Local = loadEntries(ms.LocalPath)
	ms.Global = loadEntries(ms.GlobalPath)
}

func loadEntries(path string) []MemoryEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []MemoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		config.DebugLog("[MEMORY] parse error %s: %v", path, err)
		return nil
	}
	return entries
}

func saveEntries(path string, entries []MemoryEntry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Add adds a new memory. Use global=true for cross-project memories.
func (ms *MemoryStore) Add(content string, global bool) string {
	entry := MemoryEntry{
		Content:   content,
		HitCount:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if global {
		entry.ID = ms.nextID(ms.Global)
		ms.Global = append(ms.Global, entry)
		if len(ms.Global) > maxMemories {
			ms.Global = ms.Global[len(ms.Global)-maxMemories:]
		}
		if err := saveEntries(ms.GlobalPath, ms.Global); err != nil {
			return fmt.Sprintf("Failed to save global memory: %v", err)
		}
		config.DebugLog("[MEMORY] added global #%d: %s", entry.ID, truncateStr(content, 50))
		return fmt.Sprintf("Global memory #%d saved: %s", entry.ID, truncateStr(content, 60))
	}

	entry.ID = ms.nextID(ms.Local)
	ms.Local = append(ms.Local, entry)
	if len(ms.Local) > maxMemories {
		ms.Local = ms.Local[len(ms.Local)-maxMemories:]
	}
	if err := saveEntries(ms.LocalPath, ms.Local); err != nil {
		return fmt.Sprintf("Failed to save local memory: %v", err)
	}
	config.DebugLog("[MEMORY] added local #%d: %s", entry.ID, truncateStr(content, 50))
	return fmt.Sprintf("Project memory #%d saved: %s", entry.ID, truncateStr(content, 60))
}

// Edit updates an existing memory by ID.
func (ms *MemoryStore) Edit(id int, newContent string, global bool) string {
	entries, path, label := ms.targetList(global)
	for i, e := range entries {
		if e.ID == id {
			entries[i].Content = newContent
			entries[i].UpdatedAt = time.Now()
			ms.setList(global, entries)
			if err := saveEntries(path, entries); err != nil {
				return fmt.Sprintf("Failed to save: %v", err)
			}
			config.DebugLog("[MEMORY] edited %s #%d", label, id)
			return fmt.Sprintf("%s memory #%d updated: %s", label, id, truncateStr(newContent, 60))
		}
	}
	return fmt.Sprintf("%s memory #%d not found", label, id)
}

// Delete removes a memory by ID.
func (ms *MemoryStore) Delete(id int, global bool) string {
	entries, path, label := ms.targetList(global)
	for i, e := range entries {
		if e.ID == id {
			entries = append(entries[:i], entries[i+1:]...)
			ms.setList(global, entries)
			if err := saveEntries(path, entries); err != nil {
				return fmt.Sprintf("Failed to save: %v", err)
			}
			config.DebugLog("[MEMORY] deleted %s #%d", label, id)
			return fmt.Sprintf("%s memory #%d deleted", label, id)
		}
	}
	return fmt.Sprintf("%s memory #%d not found", label, id)
}

// List returns a formatted list of all memories.
func (ms *MemoryStore) List() string {
	var lines []string

	if len(ms.Local) > 0 {
		lines = append(lines, "  [Project memories]")
		for _, e := range ms.Local {
			lines = append(lines, fmt.Sprintf("    #%d %s (hits: %d)", e.ID, truncateStr(e.Content, 50), e.HitCount))
		}
	}

	if len(ms.Global) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "  [Global memories]")
		for _, e := range ms.Global {
			lines = append(lines, fmt.Sprintf("    #%d %s (hits: %d)", e.ID, truncateStr(e.Content, 50), e.HitCount))
		}
	}

	if len(lines) == 0 {
		return "No memories saved. Use /remember <text> to add."
	}

	return strings.Join(lines, "\n")
}

// ForContext returns memories formatted for system prompt injection.
// Sorted by hit_count descending. Increments hit_count for returned entries.
func (ms *MemoryStore) ForContext() string {
	all := make([]MemoryEntry, 0, len(ms.Local)+len(ms.Global))

	for i := range ms.Local {
		ms.Local[i].HitCount++
		all = append(all, ms.Local[i])
	}
	for i := range ms.Global {
		ms.Global[i].HitCount++
		all = append(all, ms.Global[i])
	}

	if len(all) == 0 {
		return ""
	}

	// Sort by hit_count descending
	sort.Slice(all, func(i, j int) bool {
		return all[i].HitCount > all[j].HitCount
	})

	// Save updated hit counts (best-effort)
	_ = saveEntries(ms.LocalPath, ms.Local)
	_ = saveEntries(ms.GlobalPath, ms.Global)

	var sb strings.Builder
	sb.WriteString("\n\n## User Memories\n")
	for _, e := range all {
		sb.WriteString(fmt.Sprintf("- %s\n", e.Content))
	}
	return sb.String()
}

// Search returns memories matching the query (case-insensitive).
func (ms *MemoryStore) Search(query string) string {
	query = strings.ToLower(query)
	var matches []string

	for _, e := range ms.Local {
		if strings.Contains(strings.ToLower(e.Content), query) {
			matches = append(matches, fmt.Sprintf("  [Project] #%d %s", e.ID, e.Content))
		}
	}
	for _, e := range ms.Global {
		if strings.Contains(strings.ToLower(e.Content), query) {
			matches = append(matches, fmt.Sprintf("  [Global] #%d %s", e.ID, e.Content))
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No memories matching '%s'", query)
	}
	return strings.Join(matches, "\n")
}

func (ms *MemoryStore) nextID(entries []MemoryEntry) int {
	maxID := 0
	for _, e := range entries {
		if e.ID > maxID {
			maxID = e.ID
		}
	}
	return maxID + 1
}

func (ms *MemoryStore) targetList(global bool) ([]MemoryEntry, string, string) {
	if global {
		return ms.Global, ms.GlobalPath, "Global"
	}
	return ms.Local, ms.LocalPath, "Project"
}

func (ms *MemoryStore) setList(global bool, entries []MemoryEntry) {
	if global {
		ms.Global = entries
	} else {
		ms.Local = entries
	}
}

func truncateStr(s string, max int) string {
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max]) + "..."
	}
	return s
}
