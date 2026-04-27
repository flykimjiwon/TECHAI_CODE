package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// PaletteItem represents a single command in the palette.
type PaletteItem struct {
	Label       string
	Description string
	Action      string // slash command
}

// PaletteItems is the default list of commands available in the palette.
var PaletteItems = []PaletteItem{
	{Label: "Init", Description: "Scan project → generate .techai.md", Action: "/init"},
	{Label: "Remember", Description: "Save memory for AI context", Action: "/remember"},
	{Label: "Memories", Description: "List all saved memories", Action: "/remember list"},
	{Label: "New Session", Description: "Start new conversation", Action: "/new"},
	{Label: "Sessions", Description: "Browse recent sessions", Action: "/sessions"},
	{Label: "Auto Mode", Description: "AI autonomous execution", Action: "/auto"},
	{Label: "Multi-Agent", Description: "Dual model parallel run", Action: "/multi"},
	{Label: "Diagnostics", Description: "Run project linters", Action: "/diagnostics"},
	{Label: "Git Status", Description: "Repository status", Action: "/git"},
	{Label: "Compact", Description: "Compress conversation history", Action: "/compact"},
	{Label: "Clear", Description: "Clear conversation", Action: "/clear"},
	{Label: "Copy", Description: "Copy last AI response", Action: "/copy"},
	{Label: "Export", Description: "Export session to markdown", Action: "/export"},
	{Label: "Diff", Description: "Show git changes", Action: "/diff"},
	{Label: "Undo", Description: "Undo last file edit", Action: "/undo"},
	{Label: "Setup", Description: "Reset API key", Action: "/setup"},
	{Label: "Companion", Description: "Browser dashboard", Action: "/companion"},
	{Label: "MCP Status", Description: "MCP server connections", Action: "/mcp"},
	{Label: "Help", Description: "Keyboard shortcuts", Action: "/help"},
	{Label: "Exit", Description: "Quit application", Action: "/exit"},
}

// FuzzyFilter returns items matching the query (case-insensitive substring match).
func FuzzyFilter(query string, items []PaletteItem) []PaletteItem {
	if query == "" {
		return items
	}
	query = strings.ToLower(query)
	var matched []PaletteItem
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Label), query) ||
			strings.Contains(strings.ToLower(item.Description), query) ||
			strings.Contains(item.Action, query) {
			matched = append(matched, item)
		}
	}
	return matched
}

// RenderPalette renders the command palette as a floating overlay.
func RenderPalette(items []PaletteItem, selected int, query string, width int) string {
	paletteWidth := 50
	if width < 60 {
		paletteWidth = width - 10
	}
	if paletteWidth < 30 {
		paletteWidth = 30
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	// Search input
	searchStyle := lipgloss.NewStyle().
		Foreground(ColorText)

	promptStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	// Build content
	var lines []string
	lines = append(lines, titleStyle.Render("Command Palette"))
	lines = append(lines, "")

	// Search bar
	searchDisplay := query
	if searchDisplay == "" {
		searchDisplay = lipgloss.NewStyle().Foreground(ColorMuted).Render("Search...")
	} else {
		searchDisplay = searchStyle.Render(query)
	}
	lines = append(lines, promptStyle.Render("> ")+searchDisplay)
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", paletteWidth-4)))

	// Items (max 10 visible)
	maxVisible := 10
	if len(items) < maxVisible {
		maxVisible = len(items)
	}

	if len(items) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("  No matching commands"))
	} else {
		// Calculate scroll offset to keep selected item visible
		offset := 0
		if selected >= maxVisible {
			offset = selected - maxVisible + 1
		}

		for i := offset; i < offset+maxVisible && i < len(items); i++ {
			item := items[i]
			labelWidth := paletteWidth - 8
			label := item.Label
			if lipgloss.Width(label) > labelWidth/2 {
				label = label[:labelWidth/2]
			}

			if i == selected {
				sel := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#1E1E2E")).
					Background(ColorPrimary).
					Bold(true)

				line := fmt.Sprintf(" %s  %s", label, item.Description)
				// Pad to fill width
				pad := paletteWidth - 4 - lipgloss.Width(line)
				if pad > 0 {
					line += strings.Repeat(" ", pad)
				}
				lines = append(lines, sel.Render(line))
			} else {
				labelStyle := lipgloss.NewStyle().Foreground(ColorText)
				descStyle := lipgloss.NewStyle().Foreground(ColorMuted)
				line := fmt.Sprintf(" %s  %s", labelStyle.Render(label), descStyle.Render(item.Description))
				lines = append(lines, line)
			}
		}
	}

	lines = append(lines, "")
	hintStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	lines = append(lines, hintStyle.Render("↑↓ Move  Enter Select  Esc Close"))

	content := strings.Join(lines, "\n")

	// Floating box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(paletteWidth).
		Background(lipgloss.Color("#1E1E2E"))

	return box.Render(content)
}
