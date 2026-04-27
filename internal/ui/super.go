// Copyright 2025-2026 Kim Jiwon (flykimjiwon). All rights reserved.
// TECHAI CODE TUI — github.com/flykimjiwon/TECHAI_CODE
// Forked from Hanimo Code: github.com/flykimjiwon/hanimo

package ui

import (
	"fmt"
	"os/user"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kimjiwon/tgc/internal/config"
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
	bright := lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Bold(true)      // blue-400
	mid := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))                    // blue-500
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#1D4ED8"))                    // blue-700
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("#475569"))              // slate-600
	codeBright := lipgloss.NewStyle().Foreground(lipgloss.Color("#93C5FD")).Bold(true)  // blue-300
	codeMid := lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))                // blue-400
	codeDim := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6"))                // blue-500
	versionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))           // slate-400

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

	// Version line below logo — show only vX.Y.Z (strip git describe suffix)
	ver := config.AppVersion
	if i := strings.Index(ver, "-"); i > 0 {
		ver = ver[:i]
	}
	b.WriteString("\n")
	b.WriteString(versionStyle.Render(fmt.Sprintf("   %s", ver)))

	// Username greeting — prefer display name (Name), fall back to login (Username)
	if u, err := user.Current(); err == nil {
		displayName := u.Name
		if displayName == "" {
			displayName = u.Username
		}
		greetStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8")).Bold(true)
		b.WriteString("  ")
		b.WriteString(greetStyle.Render(fmt.Sprintf("👋 반갑습니다 %s님", displayName)))
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
			modeName.Render(fmt.Sprintf("Super — %s", shortModel)),
			desc.Render("All-purpose. Code, analysis, conversation auto-detect"),
		)
	case 1:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render(fmt.Sprintf("Deep Agent — %s", shortModel)),
			desc.Render("Autonomous coding. Up to 100 iterations, self-verify"),
		)
	case 2:
		tips = fmt.Sprintf("%s\n%s",
			modeName.Render(fmt.Sprintf("Plan — %s", shortModel)),
			desc.Render("Plan first. Step-by-step plan → approve → execute"),
		)
	}
	return tipStyle.Render(tips)
}
