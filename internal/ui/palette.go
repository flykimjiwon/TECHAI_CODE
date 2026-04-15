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
	{Label: "새 세션", Description: "새 대화 세션 시작", Action: "/new"},
	{Label: "세션 목록", Description: "최근 세션 보기", Action: "/sessions"},
	{Label: "자율 모드", Description: "AI 자율 작업 수행", Action: "/auto"},
	{Label: "멀티 에이전트", Description: "두 모델 병렬 실행", Action: "/multi"},
	{Label: "진단", Description: "프로젝트 코드 진단", Action: "/diagnostics"},
	{Label: "Git 상태", Description: "저장소 상태 확인", Action: "/git"},
	{Label: "화면 정리", Description: "대화 초기화", Action: "/clear"},
	{Label: "설정 초기화", Description: "API 키 재설정", Action: "/setup"},
	{Label: "도움말", Description: "키보드 단축키 안내", Action: "/help"},
	{Label: "컴패니언", Description: "브라우저 대시보드", Action: "/companion"},
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
		searchDisplay = lipgloss.NewStyle().Foreground(ColorMuted).Render("검색...")
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
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Italic(true).Render("  일치하는 명령 없음"))
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
	lines = append(lines, hintStyle.Render("↑↓ 이동  Enter 선택  Esc 닫기"))

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
