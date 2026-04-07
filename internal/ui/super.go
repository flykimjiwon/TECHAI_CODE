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
	"                     C O D E",
}

func RenderLogo() string {
	bright := lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Bold(true)  // blue-400
	mid := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))               // blue-500
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#1D4ED8"))               // blue-700

	var b strings.Builder
	subtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("#93C5FD")).Bold(true) // blue-300

	for i, line := range logoLines {
		switch {
		case i <= 1:
			b.WriteString(bright.Render(line))
		case i <= 3:
			b.WriteString(mid.Render(line))
		case i == len(logoLines)-1:
			b.WriteString(subtitle.Render(line))
		default:
			b.WriteString(dim.Render(line))
		}
		if i < len(logoLines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func ModeWelcome(mode int) string {
	var b strings.Builder

	b.WriteString(RenderLogo())
	b.WriteString("\n\n")

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
			modeName.Render("슈퍼택가이 — GPT-OSS-120b"),
			desc.Render("만능 모드. 코드 CRUD, 분석, 대화 자동 감지"),
		)
	case 1:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render("개발 — GPT-OSS-120b"),
			desc.Render("코딩 특화. 파일 생성/읽기/수정/삭제"),
		)
	case 2:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render("플랜 — GPT-OSS-120b"),
			desc.Render("분석/계획. 읽기 전용, 구조 파악, 리뷰"),
		)
	}
	b.WriteString(tipStyle.Render(tips))

	return b.String()
}

// ModeInfoBox renders just the mode description box (no logo).
func ModeInfoBox(mode int) string {
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
			modeName.Render("슈퍼택가이 — GPT-OSS-120b"),
			desc.Render("만능 모드. 코드 CRUD, 분석, 대화 자동 감지"),
		)
	case 1:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render("개발 — GPT-OSS-120b"),
			desc.Render("코딩 특화. 파일 생성/읽기/수정/삭제"),
		)
	case 2:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render("플랜 — GPT-OSS-120b"),
			desc.Render("분석/계획. 읽기 전용, 구조 파악, 리뷰"),
		)
	}
	return tipStyle.Render(tips)
}
