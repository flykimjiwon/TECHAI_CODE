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

// StatusBarData holds all data needed to render the status bar.
type StatusBarData struct {
	Model        string
	Tokens       int
	Elapsed      time.Duration
	Mode         int
	CWD          string
	Width        int
	Debug        bool
	ToolEnabled  *bool
	OS           string // "darwin/arm64"
	ToolCount    int
	HistoryLen   int // history message count
	HistoryChars int // total character count across history (for token estimation)
	MaxContext   int // model max context window
	SessionStart time.Time
	ToolIter     int
}

func RenderStatusBar(d StatusBarData) string {
	modeStyle := lipgloss.NewStyle().
		Foreground(ModeColor(d.Mode)).
		Bold(true)

	modeName := Tabs[d.Mode].Name
	displayName := d.Model
	// Use short display name from model registry if available
	if len(d.Model) > 20 {
		parts := strings.Split(d.Model, "/")
		if len(parts) > 1 {
			displayName = parts[len(parts)-1]
		}
	}

	left := modeStyle.Render("  "+modeName) +
		Subtle.Render("  "+strings.ToUpper(displayName)) +
		Subtle.Render("  "+d.OS) +
		Subtle.Render("  ./"+d.CWD)

	// Tool status indicator
	if d.ToolEnabled != nil {
		if *d.ToolEnabled {
			toolOn := lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
			left += toolOn.Render(fmt.Sprintf("  Tool:ON  %dtools", d.ToolCount))
		} else {
			toolOff := lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
			left += toolOff.Render("  Tool:OFF")
		}
	} else {
		toolUnknown := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
		left += toolUnknown.Render("  Tool:--")
	}

	if d.Tokens > 0 {
		left += Subtle.Render(fmt.Sprintf("  %dtok", d.Tokens))
	}
	if d.Elapsed > 0 {
		left += Subtle.Render(fmt.Sprintf("  %.1fs", d.Elapsed.Seconds()))
	}

	right := Subtle.Render("Shift+Enter 줄바꿈  Tab 전환  /clear  Ctrl+C ")

	gap := d.Width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	line1 := lipgloss.NewStyle().
		Background(lipgloss.Color("#0F172A")).
		Width(d.Width).
		Render(left + strings.Repeat(" ", gap) + right)

	// HUD line 2: context meter, session time, tool iteration, debug
	hudStyle := lipgloss.NewStyle().Foreground(ColorHUD)
	ctxStyle := lipgloss.NewStyle().Foreground(ColorHUDCtx)
	timeStyle := lipgloss.NewStyle().Foreground(ColorHUDTime)
	debugStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)

	// Estimate tokens from history (~4 chars per token avg)
	estTokens := d.HistoryChars / 4
	if estTokens == 0 && d.HistoryLen > 0 {
		estTokens = d.HistoryLen * 100 // fallback if chars not tracked
	}
	ctxPercent := 0
	if d.MaxContext > 0 {
		ctxPercent = (estTokens * 100) / d.MaxContext
	}
	ctxStr := ctxStyle.Render(fmt.Sprintf("ctx:~%dK/%dK(%d%%)", estTokens/1000, d.MaxContext/1000, ctxPercent))

	sessionDur := time.Since(d.SessionStart)
	sessionStr := timeStyle.Render(fmt.Sprintf("session:%s", formatDuration(sessionDur)))

	iterStr := hudStyle.Render(fmt.Sprintf("toolIter:%d/20", d.ToolIter))

	hudLeft := "  " + ctxStr + "  " + sessionStr + "  " + iterStr
	if d.Debug {
		hudLeft += "  " + debugStyle.Render("[DEBUG]")
	}

	hudGap := d.Width - lipgloss.Width(hudLeft) - 2
	if hudGap < 1 {
		hudGap = 1
	}

	line2 := lipgloss.NewStyle().
		Background(lipgloss.Color("#0F172A")).
		Width(d.Width).
		Render(hudLeft + strings.Repeat(" ", hudGap))

	return line1 + "\n" + line2
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
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
