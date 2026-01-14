// internal/ui/styles.go
// Modernized Professional UI Styles for Ricambi Manager

package ui

import (
    "fmt"
    "time"

    "github.com/charmbracelet/lipgloss"
)

var (
    // Modern Professional Color Palette
    ColorPrimary      = lipgloss.Color("#3B82F6")   // Modern blue
    ColorPrimaryDark  = lipgloss.Color("#2563EB")
    ColorPrimaryLight = lipgloss.Color("#60A5FA")
    
    ColorSecondary    = lipgloss.Color("#8B5CF6")   // Modern purple
    ColorSecondaryDark = lipgloss.Color("#7C3AED")
    
    ColorSuccess      = lipgloss.Color("#10B981")   // Modern green
    ColorSuccessLight = lipgloss.Color("#34D399")
    ColorSuccessDark  = lipgloss.Color("#059669")
    
    ColorWarning      = lipgloss.Color("#F59E0B")   // Modern amber
    ColorWarningLight = lipgloss.Color("#FBBF24")
    
    ColorDanger       = lipgloss.Color("#EF4444")   // Modern red
    ColorDangerLight  = lipgloss.Color("#F87171")
    
    ColorInfo         = lipgloss.Color("#06B6D4")   // Modern cyan
    
    ColorMuted        = lipgloss.Color("#9CA3AF")   // Gray for muted text
    ColorMutedDark    = lipgloss.Color("#6B7280")
    ColorMutedLight   = lipgloss.Color("#D1D5DB")
    
    ColorBorder       = lipgloss.Color("#374151")   // Dark gray for borders
    ColorBorderLight  = lipgloss.Color("#4B5563")
    
    ColorBg           = lipgloss.Color("#111827")   // Dark background
    ColorBgLight      = lipgloss.Color("#1F2937")   // Slightly lighter
    ColorBgLighter    = lipgloss.Color("#374151")   // Even lighter
    ColorBgCard       = lipgloss.Color("#1E293B")   // Card background
    
    ColorFg           = lipgloss.Color("#F9FAFB")   // Primary foreground
    ColorFgMuted      = lipgloss.Color("#D1D5DB")   // Muted foreground
    ColorFgDim        = lipgloss.Color("#9CA3AF")   // Dim foreground
    
    // Gradient colors
    ColorGradientStart = lipgloss.Color("#3B82F6")
    ColorGradientEnd   = lipgloss.Color("#8B5CF6")
)

var (
    // Main title style with gradient effect simulation
    TitleStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorPrimary).
            FontSize(20).
            MarginBottom(2)

    SubtitleStyle = lipgloss.NewStyle().
            Foreground(ColorFgMuted).
            FontSize(14).
            MarginBottom(2)

    // Header styles
    HeaderStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorFg).
            Background(ColorBgLight).
            Padding(0, 2).
            Width(100)

    // DateTime with modern border
    DateTimeStyle = lipgloss.NewStyle().
            Foreground(ColorWarning).
            Bold(true).
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorBorder).
            Padding(0, 1).
            Background(ColorBgLight)

    // Card containers
    CardStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorBorder).
            Background(ColorBgCard).
            Padding(1, 2).
            Shadow(5)

    ContentStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorBorder).
            Padding(1, 2).
            Background(ColorBgCard)

    // Status bar
    StatusBarStyle = lipgloss.NewStyle().
            Background(ColorBgLight).
            Foreground(ColorFgMuted).
            Padding(0, 1).
            Bold(true)

    StatusBarItemStyle = lipgloss.NewStyle().
            Foreground(ColorPrimary).
            Padding(0, 1)

    // Help bar
    HelpStyle = lipgloss.NewStyle().
            Foreground(ColorMuted).
            MarginTop(1).
            FontSize(11)

    // Menu styles
    MenuStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorBorder).
            Padding(1, 2).
            MarginRight(2).
            Background(ColorBgCard)

    MenuTitleStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorPrimary).
            FontSize(16).
            MarginBottom(1)

    // List items
    UnselectedItemStyle = lipgloss.NewStyle().
            Foreground(ColorFgMuted).
            Padding(0, 1)

    SelectedItemStyle = lipgloss.NewStyle().
            Foreground(ColorPrimary).
            Bold(true).
            Padding(0, 1).
            Background(ColorBgLight).
            BorderLeft(true).
            BorderForeground(ColorPrimary)

    // Shortcut keys
    ShortcutStyle = lipgloss.NewStyle().
            Foreground(ColorWarning).
            Bold(true).
            Background(ColorBgLighter).
            Padding(0, 1).
            MarginRight(1).
            Border(true).
            BorderForeground(ColorBorder)

    ShortcutKeyStyle = lipgloss.NewStyle().
            Foreground(ColorFg).
            Bold(true).
            Background(ColorPrimaryDark).
            Padding(0, 1).
            MarginRight(1).
            Border(true).
            BorderForeground(ColorPrimary)

    // Table styles
    TableHeaderStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorPrimary).
            BorderBottom(true).
            BorderStyle(lipgloss.NormalBorder()).
            BorderForeground(ColorBorder).
            Background(ColorBgLight).
            Padding(0, 1).
            MarginBottom(1)

    TableCellStyle = lipgloss.NewStyle().
            Padding(0, 1).
            Foreground(ColorFgMuted)

    TableSelectedStyle = lipgloss.NewStyle().
            Foreground(ColorPrimary).
            Bold(true).
            Padding(0, 1).
            Background(ColorBgLight).
            BorderLeft(true).
            BorderForeground(ColorPrimary)

    TableRowEvenStyle = lipgloss.NewStyle().
            Background(ColorBgCard)

    TableRowOddStyle = lipgloss.NewStyle().
            Background(ColorBgLight)

    // Input styles
    InputStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorBorder).
            Padding(0, 1).
            Background(ColorBg).
            Foreground(ColorFg)

    InputFocusedStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorPrimary).
            Border(true).
            Padding(0, 1).
            Background(ColorBg).
            Foreground(ColorFg).
            Underline(true)

    InputLabelStyle = lipgloss.NewStyle().
            Foreground(ColorMuted).
            FontSize(12).
            MarginBottom(1)

    // Button styles
    ButtonStyle = lipgloss.NewStyle().
            Background(ColorBgLighter).
            Foreground(ColorFgMuted).
            Padding(0, 2).
            MarginRight(1).
            Bold(true).
            Border(true).
            BorderForeground(ColorBorder)

    ButtonActiveStyle = lipgloss.NewStyle().
            Background(ColorPrimary).
            Foreground(ColorFg).
            Padding(0, 2).
            MarginRight(1).
            Bold(true)

    ButtonPrimaryStyle = lipgloss.NewStyle().
            Background(ColorPrimary).
            Foreground(ColorFg).
            Padding(0, 3).
            MarginRight(1).
            Bold(true).
            Border(true).
            BorderForeground(ColorPrimaryDark)

    ButtonSuccessStyle = lipgloss.NewStyle().
            Background(ColorSuccess).
            Foreground(ColorFg).
            Padding(0, 2).
            MarginRight(1).
            Bold(true)

    // Alert styles
    ErrorStyle = lipgloss.NewStyle().
            Foreground(ColorDanger).
            Bold(true).
            Padding(1).
            Background("#7F1D1D").
            Border(true).
            BorderForeground(ColorDanger)

    WarningStyle = lipgloss.NewStyle().
            Foreground(ColorWarning).
            Bold(true).
            padding(1).
            Background("#78350F").
            Border(true).
            BorderForeground(ColorWarning)

    SuccessStyle = lipgloss.NewStyle().
            Foreground(ColorSuccess).
            Bold(true).
            Padding(1).
            Background("#064E3B").
            Border(true).
            BorderForeground(ColorSuccess)

    InfoStyle = lipgloss.NewStyle().
            Foreground(ColorInfo).
            Padding(1).
            Background("#0E7490").
            Border(true).
            BorderForeground(ColorInfo)

    // Breadcrumb
    BreadcrumbStyle = lipgloss.NewStyle().
            Foreground(ColorMuted).
            MarginBottom(1).
            FontSize(11)

    BreadcrumbItemStyle = lipgloss.NewStyle().
            Foreground(ColorMuted).
            Padding(0, 1)

    BreadcrumbActiveStyle = lipgloss.NewStyle().
            Foreground(ColorPrimary).
            Bold(true).
            Padding(0, 1)

    BreadcrumbSeparatorStyle = lipgloss.NewStyle().
            Foreground(ColorBorder).
            Padding(0, 1)

    // Badge styles - Modern pill-shaped
    BadgeStyle = lipgloss.NewStyle().
            Background(ColorBgLighter).
            Foreground(ColorFgMuted).
            Padding(0, 2).
            MarginLeft(1).
            Border(true).
            BorderForeground(ColorBorder).
            BorderRadius(true)

    BadgePrimaryStyle = lipgloss.NewStyle().
            Background(ColorPrimary).
            Foreground(ColorFg).
            Padding(0, 2).
            MarginLeft(1).
            Bold(true).
            BorderRadius(true)

    BadgeSuccessStyle = lipgloss.NewStyle().
            Background(ColorSuccess).
            Foreground(ColorFg).
            Padding(0, 2).
            MarginLeft(1).
            Bold(true).
            BorderRadius(true)

    BadgeWarningStyle = lipgloss.NewStyle().
            Background(ColorWarning).
            Foreground(ColorBg).
            Padding(0, 2).
            MarginLeft(1).
            Bold(true).
            BorderRadius(true)

    BadgeDangerStyle = lipgloss.NewStyle().
            Background(ColorDanger).
            Foreground(ColorFg).
            Padding(0, 2).
            MarginLeft(1).
            Bold(true).
            BorderRadius(true)

    BadgeInfoStyle = lipgloss.NewStyle().
            Background(ColorInfo).
            Foreground(ColorFg).
            Padding(0, 2).
            MarginLeft(1).
            Bold(true).
            BorderRadius(true)

    // Progress bar
    ProgressBarStyle = lipgloss.NewStyle().
            Foreground(ColorSuccess)

    ProgressBarEmptyStyle = lipgloss.NewStyle().
            Foreground(ColorMutedDark)

    ProgressBarContainerStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorBorder).
            Padding(0, 1)

    // Stats card
    StatsCardStyle = lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(ColorBorder).
            Background(ColorBgCard).
            Padding(1, 2).
            Margin(1)

    StatsValueStyle = lipgloss.NewStyle().
            Bold(true).
            FontSize(24).
            Foreground(ColorPrimary)

    StatsLabelStyle = lipgloss.NewStyle().
            Foreground(ColorMuted).
            FontSize(11)

    // Section header
    SectionStyle = lipgloss.NewStyle().
            Bold(true).
            Foreground(ColorFg).
            Background(ColorBgLight).
            Padding(0, 2).
            Width(100)

    SectionContentStyle = lipgloss.NewStyle().
            Foreground(ColorFgMuted).
            MarginBottom(1)

    // Loading spinner
    LoadingStyle = lipgloss.NewStyle().
            Foreground(ColorPrimary)

    // Empty state
    EmptyStateStyle = lipgloss.NewStyle().
            Foreground(ColorMuted).
            Italic(true).
            Padding(2)

    // Tab styles
    TabStyle = lipgloss.NewStyle().
            Foreground(ColorMuted).
            Padding(0, 2).
            MarginRight(1)

    TabActiveStyle = lipgloss.NewStyle().
            Foreground(ColorPrimary).
            Bold(true).
            Padding(0, 2).
            MarginRight(1).
            BorderBottom(true).
            BorderForeground(ColorPrimary)

    // Search highlight
    HighlightStyle = lipgloss.NewStyle().
            Foreground(ColorWarning).
            Bold(true)
)

// paddingHelper is a helper function for padding
func paddingHelper(n int) lipgloss.Style {
    return lipgloss.NewStyle().Padding(n)
}

// RenderProgressBar creates a modern progress bar
func RenderProgressBar(percent float64, width int) string {
    filled := int(float64(width) * (percent / 100))
    if filled > width {
        filled = width
    }
    if filled < 0 {
        filled = 0
    }

    // Use block characters for smoother look
    blocks := []string{"░", "▒", "▓", "█"}
    
    bar := ""
    for i := 0; i < filled; i++ {
        bar += blocks[3]
    }
    for i := filled; i < width; i++ {
        bar += blocks[0]
    }

    return ProgressBarStyle.Render(bar)
}

// RenderProgressBarWithValue shows progress with percentage
func RenderProgressBarWithValue(percent float64, width int) string {
    bar := RenderProgressBar(percent, width-8)
    value := fmt.Sprintf(" %.0f%% ", percent)
    return bar + ProgressBarStyle.Render(value)
}

// RenderBadge creates a styled badge
func RenderBadge(text string, style lipgloss.Style) string {
    return style.Render(" " + text + " ")
}

// RenderStatusBadge returns a badge based on status
func RenderStatusBadge(status string) string {
    switch status {
    case "active", "ok", "success", "disponibile":
        return BadgeSuccessStyle.Render(status)
    case "warning", "pending", "limitato":
        return BadgeWarningStyle.Render(status)
    case "error", "blocked", "danger", "esaurito":
        return BadgeDangerStyle.Render(status)
    case "info", "info":
        return BadgeInfoStyle.Render(status)
    default:
        return BadgeStyle.Render(status)
    }
}

// RenderStatCard creates a statistics card
func RenderStatCard(label string, value string, icon string) string {
    iconStyled := lipgloss.NewStyle().Foreground(ColorPrimary).Render(icon)
    valueStyled := StatsValueStyle.Render(value)
    labelStyled := StatsLabelStyle.Render(label)
    
    return lipgloss.JoinVertical(
        lipgloss.Center,
        iconStyled,
        valueStyled,
        labelStyled,
    )
}

// RenderLoadingSpinner shows an animated loading indicator
func RenderLoadingSpinner() string {
    spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    idx := int(time.Now().UnixNano() / 100000000) % len(spinners)
    return LoadingStyle.Render(spinners[idx] + " Caricamento in corso...")
}

// RenderEmptyState shows an empty state message
func RenderEmptyState(message string) string {
    icon := "○"
    return lipgloss.JoinVertical(
        lipgloss.Center,
        EmptyStateStyle.Render(icon),
        EmptyStateStyle.Render(message),
    )
}

// RenderSection creates a section with header
func RenderSection(title string, content string) string {
    header := SectionStyle.Render(title)
    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        SectionContentStyle.Render(content),
    )
}

// FormatCurrency formats a value as currency
func FormatCurrency(amount float64) string {
    return fmt.Sprintf("€ %.2f", amount)
}

// FormatNumber formats a number with thousand separators
func FormatNumber(n int) string {
    return fmt.Sprintf("%d", n)
}

// FormatPercentage formats a percentage value
func FormatPercentage(p float64) string {
    return fmt.Sprintf("%.1f%%", p)
}
