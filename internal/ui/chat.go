package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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
			content := wrapText(msg.Content, contentWidth)
			for _, line := range strings.Split(content, "\n") {
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
		wrapped := wrapText(streaming, contentWidth)
		for _, line := range strings.Split(wrapped, "\n") {
			lines = append(lines, "  "+line)
		}
		// Add cursor to last line
		if len(lines) > 0 {
			lines[len(lines)-1] += "▊"
		}
	}

	return strings.Join(lines, "\n")
}

func RenderStatusBar(model string, tokens int, elapsed time.Duration, mode int, width int) string {
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		Bold(true)

	modeName := Tabs[mode].Name
	left := modeStyle.Render("  "+modeName) +
		Subtle.Render("  "+model)

	if tokens > 0 {
		left += Subtle.Render(fmt.Sprintf("  %dtok", tokens))
	}
	if elapsed > 0 {
		left += Subtle.Render(fmt.Sprintf("  %.1fs", elapsed.Seconds()))
	}

	right := Subtle.Render("택가이코드  /help")

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#0F172A")).
		Width(width).
		Render(left + strings.Repeat(" ", gap) + right)
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
