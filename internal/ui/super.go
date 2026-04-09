package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var logoLines = []string{
	" ████████╗███████╗ ██████╗██╗  ██╗ █████╗ ██╗",
	" ╚══██╔══╝██╔════╝██╔════╝██║  ██║██╔══██╗██║",
	"    ██║   █████╗  ██║     ███████║███████║██║",
	"    ██║   ██╔══╝  ██║     ██╔══██║██╔══██║██║",
	"    ██║   ███████╗╚██████╗██║  ██║██║  ██║██║",
	"    ╚═╝   ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝",
	" ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	"    ██████╗  ██████╗  ██████╗  ███████╗",
	"   ██╔════╝ ██╔═══██╗ ██╔══██╗ ██╔════╝",
	"   ██║      ██║   ██║ ██║  ██║ █████╗",
	"   ██║      ██║   ██║ ██║  ██║ ██╔══╝",
	"   ╚██████╗ ╚██████╔╝ ██████╔╝ ███████╗",
	"    ╚═════╝  ╚═════╝  ╚═════╝  ╚══════╝",
}

func RenderLogo() string {
	bright := lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Bold(true) // blue-400
	mid := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))              // blue-500
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#1D4ED8"))              // blue-700
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("#475569"))        // slate-600
	codeBright := lipgloss.NewStyle().Foreground(lipgloss.Color("#93C5FD")).Bold(true) // blue-300
	codeMid := lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))          // blue-400
	codeDim := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))          // blue-500

	var b strings.Builder

	for i, line := range logoLines {
		switch {
		case i <= 1:
			b.WriteString(bright.Render(line))
		case i <= 3:
			b.WriteString(mid.Render(line))
		case i <= 5:
			b.WriteString(dim.Render(line))
		case i == 6:
			b.WriteString(separator.Render(line))
		case i <= 8:
			b.WriteString(codeBright.Render(line))
		case i <= 10:
			b.WriteString(codeMid.Render(line))
		default:
			b.WriteString(codeDim.Render(line))
		}
		if i < len(logoLines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func ModeWelcome(mode int, modelID string) string {
	var b strings.Builder

	b.WriteString(RenderLogo())
	b.WriteString("\n\n")

	b.WriteString(modeInfoBoxInner(mode, modelID))

	return b.String()
}

// ModeInfoBox renders just the mode description box (no logo).
func ModeInfoBox(mode int, modelID string) string {
	return modeInfoBoxInner(mode, modelID)
}

func modeInfoBoxInner(mode int, modelID string) string {
	// Strip provider prefix (e.g. "google/gemma-4-31b-it" → "gemma-4-31b-it")
	shortModel := modelID
	if idx := strings.LastIndex(modelID, "/"); idx >= 0 {
		shortModel = modelID[idx+1:]
	}
	modeClr := ModeColor(mode)
	modeName := lipgloss.NewStyle().Foreground(modeClr).Bold(true)
	desc := lipgloss.NewStyle().Foreground(ColorText)

	tipStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#9CA3AF")).
		Padding(0, 1).
		Width(55)

	var tips string
	switch mode {
	case 0:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render(fmt.Sprintf("슈퍼택가이 — %s", shortModel)),
			desc.Render("만능 모드. 코드 CRUD, 분석, 대화 자동 감지"),
		)
	case 1:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render(fmt.Sprintf("개발 — %s", shortModel)),
			desc.Render("코딩 특화. 파일 생성/읽기/수정/삭제"),
		)
	case 2:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render(fmt.Sprintf("플랜 — %s", shortModel)),
			desc.Render("분석/계획. 읽기 전용, 구조 파악, 리뷰"),
		)
	}
	return tipStyle.Render(tips)
}
