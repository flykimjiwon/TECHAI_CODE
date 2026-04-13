package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kimjiwon/tgc/internal/session"
)

// MainMenuItems is the default list of menu entries.
var MainMenuItems = []string{
	"세션 목록",
	"새 세션",
	"Git 상태",
	"설정 초기화",
	"화면 정리",
	"도움말",
}

// MenuActionFromIndex returns the slash command for a given menu item index.
func MenuActionFromIndex(idx int) string {
	actions := []string{"/sessions", "/new", "/git", "/setup", "/clear", "/help"}
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

	lines = append(lines, titleStyle.Render("택가이코드"))
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

// RenderSessionPicker renders the session picker overlay.
func RenderSessionPicker(sessions []session.SessionMeta, selected int, currentID int64, width int) string {
	pickerWidth := 55
	if width < 65 {
		pickerWidth = width - 10
	}
	if pickerWidth < 35 {
		pickerWidth = 35
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	var lines []string
	lines = append(lines, titleStyle.Render("세션 목록"))
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("─", pickerWidth-4)))

	maxVisible := 10
	if len(sessions) < maxVisible {
		maxVisible = len(sessions)
	}

	if len(sessions) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("  저장된 세션이 없습니다"))
	} else {
		offset := 0
		if selected >= maxVisible {
			offset = selected - maxVisible + 1
		}

		for i := offset; i < offset+maxVisible && i < len(sessions); i++ {
			s := sessions[i]
			title := s.Title
			titleRunes := []rune(title)
			if len(titleRunes) > 25 {
				title = string(titleRunes[:25]) + ".."
			}
			date := s.UpdatedAt.Format("01/02 15:04")
			marker := "  "
			if s.ID == currentID {
				marker = "* "
			}
			label := fmt.Sprintf("%s#%d %s  %s", marker, s.ID, title, date)

			if i == selected {
				sel := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#1E1E2E")).
					Background(ColorPrimary).
					Bold(true)
				pad := pickerWidth - 4 - lipgloss.Width(label)
				if pad > 0 {
					label += strings.Repeat(" ", pad)
				}
				lines = append(lines, sel.Render(label))
			} else {
				style := lipgloss.NewStyle().Foreground(ColorText)
				if s.ID == currentID {
					style = style.Foreground(ColorAccent)
				}
				lines = append(lines, style.Render(label))
			}
		}

		if len(sessions) > maxVisible {
			countStyle := lipgloss.NewStyle().Foreground(ColorMuted)
			lines = append(lines, countStyle.Render(fmt.Sprintf(" [%d/%d]", selected+1, len(sessions))))
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
		Width(pickerWidth).
		Background(lipgloss.Color("#1E1E2E"))

	return box.Render(content)
}
