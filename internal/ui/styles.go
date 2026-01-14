// internal/ui/styles.go

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorPrimary   = lipgloss.Color("#7aa2f7")
	ColorSecondary = lipgloss.Color("#9aa5ce")
	ColorSuccess   = lipgloss.Color("#9ece6a")
	ColorWarning   = lipgloss.Color("#e0af68")
	ColorDanger    = lipgloss.Color("#f7768e")
	ColorInfo      = lipgloss.Color("#0db9d7")
	ColorMuted     = lipgloss.Color("#565f89")
	ColorBorder    = lipgloss.Color("#414868")
	ColorBg        = lipgloss.Color("#1a1b26")
	ColorFg        = lipgloss.Color("#c0caf5")
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
			Background(lipgloss.Color("#1f2335")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	MenuStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			MarginRight(2).
			Background(lipgloss.Color("#24283b"))

	ContentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			Background(lipgloss.Color("#24283b"))

	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorBg).
			Padding(0, 1).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorFg).
				Padding(0, 1).
				Background(lipgloss.Color("#24283b"))

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorBg).
				Background(ColorPrimary).
				Bold(true).
				Padding(0, 1).
				MarginLeft(1).
				MarginRight(1)

	ShortcutStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true).
			Padding(0, 1)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorBorder).
				Padding(0, 1).
				Background(lipgloss.Color("#1f2335"))

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(ColorFg)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorBg).
				Background(ColorPrimary).
				Bold(true).
				Padding(0, 1)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			Background(lipgloss.Color("#1f2335")).
			Foreground(ColorFg).
			Width(40)

	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess).
				Padding(0, 1).
				Background(lipgloss.Color("#1f2335")).
				Foreground(ColorFg).
				Width(40)

	ButtonStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorBg).
			Padding(0, 2).
			MarginRight(1).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary)

	ButtonActiveStyle = lipgloss.NewStyle().
				Background(ColorSuccess).
				Foreground(ColorBg).
				Padding(0, 2).
				MarginRight(1).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorSuccess)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorBg).
			Background(ColorDanger).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorBg).
			Background(ColorWarning).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorBg).
			Background(ColorSuccess).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorBg).
			Background(ColorInfo).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorInfo)

	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginBottom(1)

	BadgeStyle = lipgloss.NewStyle().
			Background(ColorInfo).
			Foreground(ColorBg).
			Padding(0, 1).
			MarginLeft(1).
			Bold(true)

	BadgeSuccessStyle = lipgloss.NewStyle().
				Background(ColorSuccess).
				Foreground(ColorBg).
				Padding(0, 1).
				MarginLeft(1).
				Bold(true)

	BadgeWarningStyle = lipgloss.NewStyle().
				Background(ColorWarning).
				Foreground(ColorBg).
				Padding(0, 1).
				MarginLeft(1).
				Bold(true)

	BadgeDangerStyle = lipgloss.NewStyle().
				Background(ColorDanger).
				Foreground(ColorBg).
				Padding(0, 1).
				MarginLeft(1).
				Bold(true)

	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	ProgressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			Margin(1, 0).
			Background(lipgloss.Color("#24283b"))

	ScrollThumbStyle = lipgloss.NewStyle().
				Background(ColorPrimary)

	ScrollTrackStyle = lipgloss.NewStyle().
				Background(ColorBorder)
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
