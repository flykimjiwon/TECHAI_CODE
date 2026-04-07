package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var (
	// Colors — 시안/틸 기반 모던 팔레트
	ColorPrimary   = lipgloss.Color("#60A5FA") // blue-400 — 메인 컬러
	ColorSecondary = lipgloss.Color("#38BDF8") // sky blue
	ColorAccent    = lipgloss.Color("#FBBF24") // amber
	ColorSuccess   = lipgloss.Color("#34D399") // emerald
	ColorError     = lipgloss.Color("#F87171") // red
	ColorMuted     = lipgloss.Color("#64748B") // slate gray
	ColorBg        = lipgloss.Color("#0F172A") // dark navy
	ColorBgLight   = lipgloss.Color("#1E293B") // slate 800
	ColorText      = lipgloss.Color("#E2E8F0") // slate 200
	ColorTextDim   = lipgloss.Color("#94A3B8") // slate 400
	ColorBlue      = lipgloss.Color("#60A5FA") // blue 400
	ColorOrange    = lipgloss.Color("#FB923C") // orange 400

	// Tab styles
	ActiveTab = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#0F172A")).
			Padding(0, 2)

	InactiveTab = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Background(ColorBgLight).
			Padding(0, 2)

	// Tab bar container
	TabBar = lipgloss.NewStyle().
		Background(ColorBgLight).
		Padding(0, 1)

	// Message styles
	UserMsg = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Bold(true)

	AssistantMsg = lipgloss.NewStyle().
			Foreground(ColorText)

	SystemMsg = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	ErrorMsg = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// Status bar
	StatusBar = lipgloss.NewStyle().
			Background(ColorBgLight).
			Foreground(ColorTextDim).
			Padding(0, 1)

	StatusModel = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StatusTokens = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	// Input area
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	InputPrompt = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// General
	Title = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	Subtle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	// Setup wizard styles
	SetupTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(1, 2)

	SetupLabel = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)

	SetupHint = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	SetupSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// Mode-specific accent colors
	SuperColor = ColorPrimary // teal — 만능
	DevColor   = ColorSuccess // emerald — 개발
	PlanColor  = ColorOrange  // orange — 플랜
)

func ModeColor(mode int) color.Color {
	switch mode {
	case 0:
		return SuperColor
	case 1:
		return DevColor
	case 2:
		return PlanColor
	default:
		return SuperColor
	}
}
