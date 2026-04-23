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

	// Mode-specific accent colors (ported from hanimo)
	SuperColor = ColorPrimary // blue — Super
	DeepColor  = ColorSuccess // emerald — Deep Agent
	DevColor   = DeepColor   // backwards-compat alias
	PlanColor  = ColorOrange  // orange — Plan
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

// Theme represents a color theme preset.
type Theme struct {
	Name      string
	Primary   color.Color
	Secondary color.Color
	Accent    color.Color
	Bg        color.Color
	Text      color.Color
}

// Themes contains all available theme presets.
var Themes = map[string]Theme{
	"navy": {
		Name:      "Navy (기본)",
		Primary:   lipgloss.Color("#60A5FA"),
		Secondary: lipgloss.Color("#38BDF8"),
		Accent:    lipgloss.Color("#FBBF24"),
		Bg:        lipgloss.Color("#0F172A"),
		Text:      lipgloss.Color("#E2E8F0"),
	},
	"honey": {
		Name:      "Honey Gold",
		Primary:   lipgloss.Color("#F9E2AF"),
		Secondary: lipgloss.Color("#FAB387"),
		Accent:    lipgloss.Color("#CBA6F7"),
		Bg:        lipgloss.Color("#1E1E2E"),
		Text:      lipgloss.Color("#CDD6F4"),
	},
	"ocean": {
		Name:      "Ocean",
		Primary:   lipgloss.Color("#7AA2F7"),
		Secondary: lipgloss.Color("#7DCFFF"),
		Accent:    lipgloss.Color("#BB9AF7"),
		Bg:        lipgloss.Color("#1A1B26"),
		Text:      lipgloss.Color("#C0CAF5"),
	},
	"dracula": {
		Name:      "Dracula",
		Primary:   lipgloss.Color("#FF79C6"),
		Secondary: lipgloss.Color("#BD93F9"),
		Accent:    lipgloss.Color("#50FA7B"),
		Bg:        lipgloss.Color("#282A36"),
		Text:      lipgloss.Color("#F8F8F2"),
	},
	"nord": {
		Name:      "Nord",
		Primary:   lipgloss.Color("#88C0D0"),
		Secondary: lipgloss.Color("#81A1C1"),
		Accent:    lipgloss.Color("#B48EAD"),
		Bg:        lipgloss.Color("#2E3440"),
		Text:      lipgloss.Color("#ECEFF4"),
	},
	"forest": {
		Name:      "Forest",
		Primary:   lipgloss.Color("#A5D6A7"),
		Secondary: lipgloss.Color("#81C784"),
		Accent:    lipgloss.Color("#FFD54F"),
		Bg:        lipgloss.Color("#1B2A1B"),
		Text:      lipgloss.Color("#E0E0E0"),
	},
}

// CurrentTheme holds the active theme name.
var CurrentTheme = "navy"

// ApplyTheme switches the global color variables to the given theme.
func ApplyTheme(name string) bool {
	t, ok := Themes[name]
	if !ok {
		return false
	}
	CurrentTheme = name
	ColorPrimary = t.Primary
	ColorSecondary = t.Secondary
	ColorAccent = t.Accent
	ColorBg = t.Bg
	ColorText = t.Text
	return true
}
