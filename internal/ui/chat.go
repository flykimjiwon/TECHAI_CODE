package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

type Role int

const (
	RoleUser Role = iota
	RoleAssistant
	RoleSystem
	RoleTool
)

type Message struct {
	Role      Role
	Content   string
	Timestamp time.Time
	Tag       string // optional tag for filtering (e.g. "modebox")
}

func RenderMessages(messages []Message, streaming string, width int) string {
	var lines []string
	contentWidth := width - 6

	for _, msg := range messages {
		switch msg.Role {
		case RoleUser:
			prefix := UserMsg.Render("  > ")
			content := wrapText(msg.Content, contentWidth-4)
			lines = append(lines, prefix+content)
		case RoleAssistant:
			rendered := renderMarkdown(msg.Content, contentWidth)
			msgLines := strings.Split(rendered, "\n")
			// Show line count for long messages
			if len(msgLines) > 20 {
				countStyle := lipgloss.NewStyle().Foreground(ColorMuted)
				lines = append(lines, countStyle.Render(fmt.Sprintf("  [%d lines]", len(msgLines))))
			}
			for _, line := range msgLines {
				lines = append(lines, "  "+line)
			}
			lines = append(lines, "")
			continue
		case RoleSystem:
			wrapped := wrapText(msg.Content, contentWidth)
			for _, line := range strings.Split(wrapped, "\n") {
				lines = append(lines, SystemMsg.Render("  "+line))
			}
			lines = append(lines, "")
			continue
		case RoleTool:
			toolStyle := lipgloss.NewStyle().Foreground(ColorAccent)
			wrapped := wrapText(msg.Content, contentWidth-2)
			lines = append(lines, toolStyle.Render("  "+wrapped))
		}
		lines = append(lines, "")
	}

	if streaming != "" {
		thinkStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
		lines = append(lines, "")
		lines = append(lines, thinkStyle.Render("  "+streaming))
	}

	return strings.Join(lines, "\n")
}

func RenderStatusBar(model string, tokens int, elapsed time.Duration, mode int, cwd string, width int, debug bool, toolCount int) string {
	modeStyle := lipgloss.NewStyle().
		Foreground(ModeColor(mode)).
		Bold(true)

	modeName := Tabs[mode].Name
	// Strip provider prefix (e.g. "google/gemma-4-31b-it" → "gemma-4-31b-it")
	shortModel := model
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		shortModel = model[idx+1:]
	}
	left := modeStyle.Render("  "+modeName) +
		Subtle.Render("  "+shortModel) +
		Subtle.Render("  ./"+cwd)

	if debug {
		debugStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
		left += debugStyle.Render("  [DEBUG]")
	}

	if toolCount > 0 {
		toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
		left += toolStyle.Render(fmt.Sprintf("  Tool:ON(%d)", toolCount))
	} else {
		toolOffStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
		left += toolOffStyle.Render("  Tool:OFF")
	}

	if tokens > 0 {
		left += Subtle.Render(fmt.Sprintf("  %dtok", tokens))
	}
	if elapsed > 0 {
		left += Subtle.Render(fmt.Sprintf("  %.1fs", elapsed.Seconds()))
	}

	right := Subtle.Render("Shift+Enter 줄바꿈  Tab 전환  /clear  Ctrl+C ")

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#0F172A")).
		Width(width).
		Render(left + strings.Repeat(" ", gap) + right)
}

// renderMarkdown renders markdown content using glamour (dark theme).
func renderMarkdown(content string, width int) string {
	if width < 20 {
		width = 20
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return wrapText(content, width)
	}
	out, err := r.Render(content)
	if err != nil {
		return wrapText(content, width)
	}
	return strings.TrimRight(out, "\n")
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var result strings.Builder
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		// Use display width (handles CJK double-width characters correctly)
		if lipgloss.Width(line) <= width {
			result.WriteString(line)
			continue
		}
		runes := []rune(line)
		cur := 0 // current display width
		start := 0
		for j, r := range runes {
			rw := lipgloss.Width(string(r))
			if cur+rw > width {
				result.WriteString(string(runes[start:j]))
				result.WriteString("\n")
				start = j
				cur = rw
			} else {
				cur += rw
			}
		}
		if start < len(runes) {
			result.WriteString(string(runes[start:]))
		}
	}
	return result.String()
}
