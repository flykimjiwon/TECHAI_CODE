// Package skills implements a SKILL.md loader compatible with the
// agentskills.io open standard. Users drop SKILL.md files into
// ~/.tgc/skills/<name>/ or .tgc/skills/<name>/ and techai
// picks them up on startup. The LLM sees skill descriptions in the
// system prompt and can invoke them via /skill <name>.
//
// Frontmatter format (YAML between --- fences):
//
//	---
//	name: review-pr
//	description: PR 코드 리뷰 수행
//	allowed-tools: [grep_search, file_read, shell_exec]
//	user-invocable: true
//	---
//	<body — the skill prompt injected into context when invoked>
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Skill is a parsed SKILL.md.
type Skill struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	AllowedTools  []string `yaml:"allowed-tools"`
	UserInvocable bool     `yaml:"user-invocable"`
	Body          string   // the prompt content after frontmatter
	Path          string   // filesystem path for debug
}

// Registry holds all loaded skills, keyed by name.
type Registry struct {
	mu     sync.RWMutex
	skills map[string]*Skill
	order  []string // sorted names for deterministic listing
}

// GlobalRegistry is the process-wide skill registry.
var GlobalRegistry = &Registry{skills: map[string]*Skill{}}

// skillDirs returns search paths in priority order (project-local first).
func skillDirs() []string {
	dirs := []string{".tgc/skills"}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".tgc", "skills"))
	}
	return dirs
}

// ScanSkills walks skill directories and loads every SKILL.md found.
// Project-local skills override global ones with the same name.
func ScanSkills() *Registry {
	reg := &Registry{skills: map[string]*Skill{}}
	dirs := skillDirs()
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		abs, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		entries, err := os.ReadDir(abs)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			skillFile := filepath.Join(abs, e.Name(), "SKILL.md")
			data, err := os.ReadFile(skillFile)
			if err != nil {
				continue
			}
			skill := parseSkillMD(string(data))
			if skill.Name == "" {
				skill.Name = e.Name()
			}
			skill.Path = skillFile
			reg.skills[skill.Name] = skill
		}
	}
	for name := range reg.skills {
		reg.order = append(reg.order, name)
	}
	sort.Strings(reg.order)
	return reg
}

// parseSkillMD extracts YAML frontmatter and body from a SKILL.md file.
func parseSkillMD(content string) *Skill {
	s := &Skill{UserInvocable: true}
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		s.Body = content
		return s
	}
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}
	if endIdx < 0 {
		s.Body = content
		return s
	}
	for _, line := range lines[1:endIdx] {
		key, val := splitFrontmatter(line)
		switch key {
		case "name":
			s.Name = val
		case "description":
			s.Description = val
		case "allowed-tools":
			s.AllowedTools = parseYAMLList(val)
		case "user-invocable":
			s.UserInvocable = val != "false"
		}
	}
	s.Body = strings.TrimSpace(strings.Join(lines[endIdx+1:], "\n"))
	return s
}

func splitFrontmatter(line string) (string, string) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", ""
	}
	key := strings.TrimSpace(line[:idx])
	val := strings.TrimSpace(line[idx+1:])
	val = strings.Trim(val, "\"'")
	return key, val
}

func parseYAMLList(val string) []string {
	val = strings.TrimPrefix(val, "[")
	val = strings.TrimSuffix(val, "]")
	parts := strings.Split(val, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"'")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Count returns the number of loaded skills.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}

// Get returns a skill by name, or nil if not found.
func (r *Registry) Get(name string) *Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.skills[name]
}

// List returns all skills in sorted order.
func (r *Registry) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Skill, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.skills[name])
	}
	return out
}

// TableOfContents returns a compact summary for system prompt injection.
func (r *Registry) TableOfContents() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.skills) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n\n## 스킬 (%d개 사용 가능)\n", len(r.skills)))
	b.WriteString("사용자가 `/skill <이름>`으로 호출할 수 있습니다.\n\n")
	for _, name := range r.order {
		s := r.skills[name]
		desc := s.Description
		if desc == "" {
			desc = "(설명 없음)"
		}
		b.WriteString(fmt.Sprintf("- `/skill %s` — %s\n", name, desc))
	}
	return b.String()
}

// FormatSkillBody returns the full skill prompt ready for injection.
func FormatSkillBody(s *Skill, args string) string {
	body := s.Body
	if args != "" {
		body = strings.ReplaceAll(body, "$ARGUMENTS", args)
	}
	return fmt.Sprintf("## 스킬: %s\n\n%s", s.Name, body)
}
