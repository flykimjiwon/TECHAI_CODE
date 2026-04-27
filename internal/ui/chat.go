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

// Q/A accent styling (ported from hanimo). User messages get a subtle
// tinted block; assistant replies get only a left accent bar (no
// background) so long answers don't become a giant colored wall.
var userBlockBg = lipgloss.Color("#0D1520") // very subtle blue tint on dark navy

// renderUserBlock wraps a user message in a left-bar accented block:
// blue bar + subtle background.
func renderUserBlock(content string, width int) string {
	barStyle := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Background(userBlockBg).
		Bold(true)
	textStyle := lipgloss.NewStyle().
		Foreground(ColorText).
		Background(userBlockBg)
	blockWidth := width - 2
	if blockWidth < 20 {
		blockWidth = 20
	}
	wrapped := wrapText(content, blockWidth-4)
	var out []string
	for _, line := range strings.Split(wrapped, "\n") {
		bar := "  ▌ "
		padded := line
		displayW := lipgloss.Width(line)
		if displayW < blockWidth-4 {
			padded = line + strings.Repeat(" ", blockWidth-4-displayW)
		}
		out = append(out, barStyle.Render(bar)+textStyle.Render(padded))
	}
	return strings.Join(out, "\n")
}

// renderAssistantBlock prefixes markdown-rendered assistant content
// with a muted left accent bar — no background fill, so long answers
// stay readable.
func renderAssistantBlock(rendered string) string {
	barStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
	var out []string
	for _, line := range strings.Split(rendered, "\n") {
		out = append(out, barStyle.Render("  ▎ ")+line)
	}
	return strings.Join(out, "\n")
}

func RenderMessages(messages []Message, streaming string, width int) string {
	var lines []string
	contentWidth := width - 6

	for _, msg := range messages {
		switch msg.Role {
		case RoleUser:
			block := renderUserBlock(msg.Content, contentWidth)
			lines = append(lines, block)
		case RoleAssistant:
			rendered := renderMarkdown(cleanAskUser(msg.Content), contentWidth-4)
			msgLines := strings.Split(rendered, "\n")
			// Show line count for long messages
			if len(msgLines) > 20 {
				countStyle := lipgloss.NewStyle().Foreground(ColorMuted)
				lines = append(lines, countStyle.Render(fmt.Sprintf("  [%d lines]", len(msgLines))))
			}
			lines = append(lines, renderAssistantBlock(rendered))
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
			for _, rawLine := range strings.Split(msg.Content, "\n") {
				wrapped := wrapText(rawLine, contentWidth-2)
				for _, l := range strings.Split(wrapped, "\n") {
					lines = append(lines, toolStyle.Render("  "+l))
				}
			}
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

func RenderStatusBar(model string, tokens int, contextWindow int, elapsed time.Duration, mode int, cwd string, width int, debug bool, toolCount int, gitLabel string, multiOn ...bool) string {
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

	// Git branch + dirty indicator. gitLabel is empty when the cwd is
	// not a git repo, and ends with "*" when the working tree is dirty.
	if gitLabel != "" {
		var gitStyle lipgloss.Style
		if strings.HasSuffix(gitLabel, "*") {
			gitStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Bold(true)
		} else {
			gitStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
		}
		left += gitStyle.Render("  \u2387 " + gitLabel)
	}

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

	// Multi-agent indicator
	if len(multiOn) > 0 && multiOn[0] {
		multiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Bold(true)
		left += multiStyle.Render("  Multi:ON")
	}

	if tokens > 0 {
		left += Subtle.Render(fmt.Sprintf("  %dtok", tokens))
		// Estimated cost: ~$0.30/1M input tokens (gpt-oss-120b tier)
		cost := float64(tokens) * 0.30 / 1_000_000
		if cost >= 0.01 {
			left += Subtle.Render(fmt.Sprintf("  $%.2f", cost))
		} else if cost >= 0.001 {
			left += Subtle.Render(fmt.Sprintf("  $%.3f", cost))
		}
	}
	// Context window usage: ctx:XX% colored by severity so the user can
	// see how close they are to the model's limit at a glance. Shown
	// whenever both tokens and a known window are available, even under
	// 1% — otherwise large-window models (e.g. 128K) would hide the
	// indicator for the whole early session.
	if tokens > 0 && contextWindow > 0 {
		pct := ContextPercent(tokens, contextWindow)
		var ctxStyle lipgloss.Style
		switch ContextLevel(pct) {
		case ContextCritical:
			ctxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Bold(true)
		case ContextWarn:
			ctxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Bold(true)
		default:
			ctxStyle = Subtle
		}
		var label string
		if pct == 0 {
			label = "  ctx:<1%"
		} else {
			label = fmt.Sprintf("  ctx:%d%%", pct)
		}
		left += ctxStyle.Render(label)
	}
	if elapsed > 0 {
		left += Subtle.Render(fmt.Sprintf("  %.1fs", elapsed.Seconds()))
	}

	right := Subtle.Render("Ctrl+K Palette  Esc Menu  Tab Switch  Ctrl+C")

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#0F172A")).
		Width(width).
		Render(left + strings.Repeat(" ", gap) + right)
}

// customDarkStyle is a modified glamour "dark" theme that replaces raw
// markdown heading prefixes (##, ###) with clean visual indicators so
// headings don't look like unprocessed markdown in the TUI.
var customDarkStyle = []byte(`{
  "document": {
    "block_prefix": "\n",
    "block_suffix": "\n",
    "color": "252",
    "margin": 2
  },
  "block_quote": {
    "indent": 1,
    "indent_token": "│ "
  },
  "paragraph": {},
  "list": {
    "level_indent": 2
  },
  "heading": {
    "block_suffix": "\n",
    "color": "39",
    "bold": true
  },
  "h1": {
    "prefix": " ",
    "suffix": " ",
    "color": "228",
    "background_color": "63",
    "bold": true
  },
  "h2": {
    "prefix": "▌ ",
    "color": "39",
    "bold": true
  },
  "h3": {
    "prefix": "▎ ",
    "color": "39",
    "bold": true
  },
  "h4": {
    "prefix": "▪ ",
    "color": "39",
    "bold": true
  },
  "h5": {
    "prefix": "· "
  },
  "h6": {
    "prefix": "· ",
    "color": "35",
    "bold": false
  },
  "text": {},
  "strikethrough": {
    "crossed_out": true
  },
  "emph": {
    "italic": true
  },
  "strong": {
    "bold": true
  },
  "hr": {
    "color": "240",
    "format": "\n--------\n"
  },
  "item": {
    "block_prefix": "• "
  },
  "enumeration": {
    "block_prefix": ". "
  },
  "task": {
    "ticked": "[✓] ",
    "unticked": "[ ] "
  },
  "link": {
    "color": "30",
    "underline": true
  },
  "link_text": {
    "color": "35",
    "bold": true
  },
  "image": {
    "color": "212",
    "underline": true
  },
  "image_text": {
    "color": "243",
    "format": "Image: {{.text}} →"
  },
  "code": {
    "prefix": " ",
    "suffix": " ",
    "color": "203",
    "background_color": "236"
  },
  "code_block": {
    "color": "244",
    "margin": 2,
    "chroma": {
      "text": { "color": "#C4C4C4" },
      "error": { "color": "#F1F1F1", "background_color": "#F05B5B" },
      "comment": { "color": "#676767" },
      "comment_preproc": { "color": "#FF875F" },
      "keyword": { "color": "#00AAFF" },
      "keyword_reserved": { "color": "#FF5FD2" },
      "keyword_namespace": { "color": "#FF5F87" },
      "keyword_type": { "color": "#6E6ED8" },
      "operator": { "color": "#EF8080" },
      "punctuation": { "color": "#E8E8A8" },
      "name": { "color": "#C4C4C4" },
      "name_builtin": { "color": "#FF8EC7" },
      "name_tag": { "color": "#B083EA" },
      "name_attribute": { "color": "#7A7AE6" },
      "name_class": { "color": "#F1F1F1", "underline": true, "bold": true },
      "name_constant": {},
      "name_decorator": { "color": "#FFFF87" },
      "name_exception": {},
      "name_function": { "color": "#00D787" },
      "name_other": {},
      "literal": {},
      "literal_number": { "color": "#6EEFC0" },
      "literal_date": {},
      "literal_string": { "color": "#C69669" },
      "literal_string_escape": { "color": "#AFFFD7" },
      "generic_deleted": { "color": "#FD5B5B" },
      "generic_emph": { "italic": true },
      "generic_inserted": { "color": "#00D787" },
      "generic_strong": { "bold": true },
      "generic_subheading": { "color": "#777777" },
      "background": { "background_color": "#373737" }
    }
  },
  "table": {},
  "definition_list": {},
  "definition_term": {},
  "definition_description": {
    "block_prefix": "\n🠶 "
  },
  "html_block": {},
  "html_span": {}
}`)

// renderMarkdown renders markdown content using glamour (custom dark theme).
func renderMarkdown(content string, width int) string {
	if width < 20 {
		width = 20
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes(customDarkStyle),
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

// cleanAskUser parses [ASK_USER]...[/ASK_USER] blocks into clean question format.
func cleanAskUser(content string) string {
	for {
		start := strings.Index(content, "[ASK_USER]")
		if start < 0 {
			break
		}
		end := strings.Index(content, "[/ASK_USER]")
		if end < 0 {
			break
		}

		block := content[start+len("[ASK_USER]") : end]
		var question string
		var options []string

		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "question:") {
				question = strings.TrimSpace(strings.TrimPrefix(line, "question:"))
			} else if strings.HasPrefix(line, "- ") {
				options = append(options, line)
			}
		}

		// Format as clean question
		var replacement string
		if question != "" {
			replacement = "\n❓ " + question
			if len(options) > 0 {
				replacement += "\n" + strings.Join(options, "\n")
			}
			replacement += "\n"
		}

		content = content[:start] + replacement + content[end+len("[/ASK_USER]"):]
	}
	return content
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
