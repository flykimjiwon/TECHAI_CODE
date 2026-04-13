package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// MainMenuItems is the default list of menu entries.
var MainMenuItems = []string{
	"자율 모드 (/auto)",
	"진단 실행 (/diagnostics)",
	"Git 상태 (/git)",
	"새 세션 (/new)",
	"세션 목록 (/sessions)",
	"설정 (/setup)",
	"화면 정리 (/clear)",
	"도움말 (/help)",
}

// MenuActionFromIndex returns the slash command for a given menu item index.
func MenuActionFromIndex(idx int) string {
	actions := []string{"/auto", "/diagnostics", "/git", "/new", "/sessions", "/setup", "/clear", "/help"}
	if idx >= 0 && idx < len(actions) {
		return actions[idx]
	}
	return ""
}

// RenderMenu renders the interactive menu as a floating overlay.
func RenderMenu(items []string, selected int, width int) string {
	menuWidth := 40
	if width < 50 {
		menuWidth = width - 10
	}
	if menuWidth < 25 {
		menuWidth = 25
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	var lines []string

	lines = append(lines, titleStyle.Render("택갈이코드"))
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", menuWidth-4)))

	// Items (max 12 visible)
	maxVisible := 12
	if len(items) < maxVisible {
		maxVisible = len(items)
	}

	if len(items) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("  항목 없음"))
	} else {
		offset := 0
		if selected >= maxVisible {
			offset = selected - maxVisible + 1
		}

		for i := offset; i < offset+maxVisible && i < len(items); i++ {
			item := items[i]

			if i == selected {
				sel := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#1E1E2E")).
					Background(ColorPrimary).
					Bold(true)

				line := fmt.Sprintf(" %s ", item)
				pad := menuWidth - 4 - lipgloss.Width(line)
				if pad > 0 {
					line += strings.Repeat(" ", pad)
				}
				lines = append(lines, sel.Render(line))
			} else {
				labelStyle := lipgloss.NewStyle().Foreground(ColorText)
				lines = append(lines, labelStyle.Render(fmt.Sprintf(" %s", item)))
			}
		}

		// Show scroll indicator
		if len(items) > maxVisible {
			countStyle := lipgloss.NewStyle().Foreground(ColorMuted)
			lines = append(lines, countStyle.Render(fmt.Sprintf(" [%d/%d]", selected+1, len(items))))
		}
	}

	lines = append(lines, "")
	hintStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	lines = append(lines, hintStyle.Render("↑↓ 이동  Enter 선택  Esc 닫기"))

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(menuWidth).
		Background(lipgloss.Color("#1E1E2E"))

	return box.Render(content)
}
