// internal/ui/view_promotions.go
// Promotions Management View

package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) viewPromotions() string {
	title := TitleStyle.Render("Promozioni Attive")

	// Stats row
	stats := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderPromoStatCard("🎉", "Attive", len(m.promotionsView.promotions), ColorSuccess),
		m.renderPromoStatCard("⏰", "In Scadenza", 2, ColorWarning),
		m.renderPromoStatCard("📈", "Sconti Medi", "12%", ColorInfo),
	)

	// Promotions list
	promosList := m.renderPromotionsList()

	// Content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		stats,
		"",
		promosList,
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

func (m *AppModel) renderPromoStatCard(icon, label string, value interface{}, color lipgloss.Color) string {
	iconStyled := lipgloss.NewStyle().FontSize(20).Foreground(color).Render(icon)
	
	var valueStr string
	switch v := value.(type) {
	case int:
		valueStr = fmt.Sprintf("%d", v)
	case string:
		valueStr = v
	}
	
	valueStyled := lipgloss.NewStyle().FontSize(24).Bold(true).Foreground(color).Render(valueStr)
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

func (m *AppModel) renderPromotionsList() string {
	header := SubtitleStyle.Render("Dettaglio Promozioni")

	var promos []string
	if len(m.promotionsView.promotions) == 0 {
		promos = append(promos, RenderEmptyState("Nessuna promozione attiva"))
	} else {
		for i := range m.promotionsView.promotions {
			promoRow := m.renderPromoRow(i)
			promos = append(promos, promoRow)
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, promos...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		ContentStyle.Render(list),
	)
}

func (m *AppModel) renderPromoRow(index int) string {
	// Placeholder for demo
	name := fmt.Sprintf("Promozione %d", index+1)
	discount := fmt.Sprintf("%d%%", 5+index*2)
	duration := fmt.Sprintf("%d giorni", 7-index)
	
	rowNum := fmt.Sprintf("%2d", index+1)
	
	if index == m.promotionsView.selectedIndex {
		rowNum = SelectedItemStyle.Render(rowNum)
		name = SelectedItemStyle.Render(name)
		discount = SelectedItemStyle.Render(discount)
		duration = SelectedItemStyle.Render(duration)
	} else {
		rowNum = TableCellStyle.Render(rowNum)
		name = TableCellStyle.Render(name)
		discount = BadgePrimaryStyle.Render(discount)
		duration = TableCellStyle.Render(duration)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		rowNum,
		"  ",
		name,
		"  ",
		discount,
		"  ",
		duration,
	)
}

func (m *AppModel) updatePromotions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.promotionsView.selectedIndex > 0 {
				m.promotionsView.selectedIndex--
			}
			return m, nil

		case "down":
			if m.promotionsView.selectedIndex < len(m.promotionsView.promotions)-1 {
				m.promotionsView.selectedIndex++
			}
			return m, nil

		case "enter":
			if len(m.promotionsView.promotions) > 0 {
				m.setMessage("Dettaglio promozione")
			}
			return m, nil

		case "n":
			m.setMessage("Nuova promozione")
			return m, nil

		case "esc":
			return m.navigateBack(), nil
		}
	}
	return m, nil
}
