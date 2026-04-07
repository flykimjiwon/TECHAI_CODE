package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type TabItem struct {
	Name string
}

var Tabs = []TabItem{
	{Name: "슈퍼택가이"},
	{Name: "개발"},
	{Name: "플랜"},
}

func RenderTabBar(activeIdx int, width int) string {
	var tabs []string

	for i, tab := range Tabs {
		label := " " + tab.Name + " "
		if i == activeIdx {
			style := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#E2E8F0")).
				Background(lipgloss.Color("#334155")).
				Padding(0, 1)
			tabs = append(tabs, style.Render(label))
		} else {
			style := lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 1)
			tabs = append(tabs, style.Render(label))
		}
	}

	hint := Subtle.Render("  Tab")
	row := strings.Join(tabs, "") + hint

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#0F172A")).
		Width(width).
		Render(row)
}
