// internal/ui/view_kits.go
// Kit Management View

package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) viewKits() string {
	title := TitleStyle.Render("Gestione Kit")

	// Stats row
	stats := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderKitStatCard("📦", "Kit Attivi", "18", ColorPrimary),
		m.renderKitStatCard("⚠️", "Parziali", "3", ColorWarning),
		m.renderKitStatCard("✅", "Completi", "15", ColorSuccess),
	)

	// Kits list
	kitsList := m.renderKitsList()

	// Content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		stats,
		"",
		kitsList,
	)

	if m.message != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			SuccessStyle.Render("✓ "+m.message),
		)
	}

	availableWidth := m.width - 28

	return lipgloss.NewStyle().
		Width(availableWidth).
		Padding(1, 2).
		Render(content)
}

func (m *AppModel) renderKitStatCard(icon, label, value string, color lipgloss.Color) string {
	iconStyled := lipgloss.NewStyle().FontSize(20).Foreground(color).Render(icon)
	valueStyled := lipgloss.NewStyle().FontSize(24).Bold(true).Foreground(color).Render(value)
	labelStyled := StatsLabelStyle.Render(label)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Background(ColorBgCard).
		Padding(1, 2).
		Width(18).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				iconStyled,
				valueStyled,
				labelStyled,
			),
		)
}

func (m *AppModel) renderKitsList() string {
	header := SubtitleStyle.Render("Kit di Vendita")

	var kits []string
	if len(m.kitsView.kits) == 0 {
		kits = append(kits, RenderEmptyState("Nessun kit configurato"))
	} else {
		for i := range m.kitsView.kits {
			kitRow := m.renderKitRow(i)
			kits = append(kits, kitRow)
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, kits...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		ContentStyle.Render(list),
	)
}

func (m *AppModel) renderKitRow(index int) string {
	code := fmt.Sprintf("KIT-%03d", 100+index)
	name := fmt.Sprintf("Kit Manutenzione %d", index+1)
	items := fmt.Sprintf("%d art.", 3+index)
	price := fmt.Sprintf("€ %.2f", float64(50+index*25))
	status := "Completo"

	rowNum := fmt.Sprintf("%2d", index+1)

	if index == m.kitsView.selectedIndex {
		rowNum = SelectedItemStyle.Render(rowNum)
		code = SelectedItemStyle.Render(code)
		name = SelectedItemStyle.Render(name)
	} else {
		rowNum = TableCellStyle.Render(rowNum)
		code = TableCellStyle.Render(code)
		name = TableCellStyle.Render(name)
	}

	statusBadge := BadgeSuccessStyle.Render(status)
	if index%3 == 1 {
		statusBadge = BadgeWarningStyle.Render("Parziale")
	} else if index%3 == 2 {
		statusBadge = BadgeInfoStyle.Render("Da assemblare")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		rowNum,
		"  ",
		code,
		"  ",
		name,
		"  ",
		TableCellStyle.Render(items),
		"  ",
		TableCellStyle.Render(price),
		"  ",
		statusBadge,
	)
}

func (m *AppModel) updateKits(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.kitsView.selectedIndex > 0 {
				m.kitsView.selectedIndex--
			}
			return m, nil

		case "down":
			if m.kitsView.selectedIndex < len(m.kitsView.kits)-1 {
				m.kitsView.selectedIndex++
			}
			return m, nil

		case "enter":
			if len(m.kitsView.kits) > 0 {
				m.setMessage("Modifica kit")
			}
			return m, nil

		case "n":
			m.setMessage("Nuovo kit")
			return m, nil

		case "r":
			m.setMessage("Prenotazione kit")
			return m, nil

		case "esc":
			return m.navigateBack(), nil
		}
	}
	return m, nil
}
