package tools

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjiwon/tgc/internal/config"
)

// LoadCustomCommands scans ~/.tgc/commands/ (global) and .tgc/commands/ (project-local)
// for .md files. Returns map[commandName]templateContent.
// Project-local commands override global ones with the same name.
func LoadCustomCommands() map[string]string {
	commands := make(map[string]string)

	// Load global commands first (~/.tgc/commands/)
	globalDir := filepath.Join(config.ConfigDir(), "commands")
	loadCommandsFromDir(globalDir, commands)

	// Load project-local commands (.tgc/commands/), overriding globals
	localDir := filepath.Join(".tgc", "commands")
	loadCommandsFromDir(localDir, commands)

	config.DebugLog("[COMMANDS] loaded %d custom commands", len(commands))
	return commands
}

// loadCommandsFromDir reads all .md files from dir and adds them to commands.
func loadCommandsFromDir(dir string, commands map[string]string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// Directory not existing is normal — not an error
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		cmdName := strings.ToLower(strings.TrimSuffix(name, ".md"))
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			config.DebugLog("[COMMANDS] failed to read %s: %v", path, err)
			continue
		}
		commands[cmdName] = string(data)
		config.DebugLog("[COMMANDS] loaded /%s from %s", cmdName, path)
	}
}
