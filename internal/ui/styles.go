// internal/ui/styles.go

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorPrimary   = lipgloss.Color("#00ADD8")
	ColorSecondary = lipgloss.Color("#5DC9E2")
	ColorSuccess   = lipgloss.Color("#00C851")
	ColorWarning   = lipgloss.Color("#FFD700")
	ColorDanger    = lipgloss.Color("#FF4444")
	ColorInfo      = lipgloss.Color("#33B5E5")
	ColorMuted     = lipgloss.Color("#666666")
	ColorBorder    = lipgloss.Color("#444444")
	ColorBg        = lipgloss.Color("#1a1a1a")
	ColorFg        = lipgloss.Color("#ffffff")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			MarginBottom(1)

	DateTimeStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	MenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			MarginRight(2)

	ContentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorBg).
			Padding(0, 1).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Padding(0, 1)

	ShortcutStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorBorder)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Padding(0, 1)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Padding(0, 1)

	ButtonStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorBg).
			Padding(0, 2).
			MarginRight(1).
			Bold(true)

	ButtonActiveStyle = lipgloss.NewStyle().
				Background(ColorSuccess).
				Foreground(ColorBg).
				Padding(0, 2).
				MarginRight(1).
				Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true).
			Padding(1)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true).
			Padding(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true).
			Padding(1)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Padding(1)

	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginBottom(1)

	BadgeStyle = lipgloss.NewStyle().
			Background(ColorInfo).
			Foreground(ColorBg).
			Padding(0, 1).
			MarginLeft(1)

	BadgeSuccessStyle = lipgloss.NewStyle().
				Background(ColorSuccess).
				Foreground(ColorBg).
				Padding(0, 1).
				MarginLeft(1)

	BadgeWarningStyle = lipgloss.NewStyle().
				Background(ColorWarning).
				Foreground(ColorBg).
				Padding(0, 1).
				MarginLeft(1)

	BadgeDangerStyle = lipgloss.NewStyle().
				Background(ColorDanger).
				Foreground(ColorBg).
				Padding(0, 1).
				MarginLeft(1)

	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ProgressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)
)

func RenderProgressBar(percent float64, width int) string {
	filled := int(float64(width) * (percent / 100))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < width; i++ {
		bar += "░"
	}

	return ProgressBarStyle.Render(bar[:filled]) + ProgressBarEmptyStyle.Render(bar[filled:])
}

func RenderBadge(text string, style lipgloss.Style) string {
	return style.Render(text)
}

func RenderStatusBadge(status string) string {
	switch status {
	case "active", "ok", "success":
		return BadgeSuccessStyle.Render(status)
	case "warning", "pending":
		return BadgeWarningStyle.Render(status)
	case "error", "blocked", "danger":
		return BadgeDangerStyle.Render(status)
	default:
		return BadgeStyle.Render(status)
	}
}
