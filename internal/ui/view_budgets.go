// internal/ui/view_budgets.go
// Budget Management View

package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) viewBudgets() string {
	title := TitleStyle.Render("Budget e Obiettivi")

	// Stats row
	stats := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderBudgetStatCard("📊", "Target Mensile", "€ 45.000", ColorPrimary),
		m.renderBudgetStatCard("✅", "Realizzato", "€ 32.450", ColorSuccess),
		m.renderBudgetStatCard("📈", "Progresso", "72%", ColorInfo),
	)

	// Budget progress
	budgetProgress := m.renderBudgetProgress()

	// Budgets list
	budgetsList := m.renderBudgetsList()

	// Content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		stats,
		"",
		budgetProgress,
		"",
		budgetsList,
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

func (m *AppModel) renderBudgetStatCard(icon, label, value string, color lipgloss.Color) string {
	iconStyled := lipgloss.NewStyle().FontSize(20).Foreground(color).Render(icon)
	valueStyled := lipgloss.NewStyle().FontSize(24).Bold(true).Foreground(color).Render(value)
	labelStyled := StatsLabelStyle.Render(label)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Background(ColorBgCard).
		Padding(1, 2).
		Width(20).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				iconStyled,
				valueStyled,
				labelStyled,
			),
		)
}

func (m *AppModel) renderBudgetProgress() string {
	title := SubtitleStyle.Render("Progresso Mensile")

	// Overall progress bar
	progress := RenderProgressBarWithValue(72, 50)

	// Details
	details := lipgloss.JoinHorizontal(
		lipgloss.Left,
		TableCellStyle.Render("Target: € 45.000"),
		"  ",
		TableCellStyle.Render("Realizzato: € 32.450"),
		"  ",
		TableCellStyle.Render("Rimanente: € 12.550"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		ProgressBarContainerStyle.Render(progress),
		"",
		details,
	)
}

func (m *AppModel) renderBudgetsList() string {
	header := SubtitleStyle.Render("Dettaglio per Agente")

	var budgets []string
	if len(m.budgetsView.budgets) == 0 {
		budgets = append(budgets, RenderEmptyState("Nessun budget configurato"))
	} else {
		for i := range m.budgetsView.budgets {
			budgetRow := m.renderBudgetRow(i)
			budgets = append(budgets, budgetRow)
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, budgets...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		ContentStyle.Render(list),
	)
}

func (m *AppModel) renderBudgetRow(index int) string {
	agent := fmt.Sprintf("Agente %s", string(rune('A'+index)))
	target := fmt.Sprintf("€ %.0f", float64(15000-index*2000))
	actual := fmt.Sprintf("€ %.0f", float64(10000-index*1500))
	percent := fmt.Sprintf("%d%%", 60+index*5)
	progressBar := RenderProgressBar(float64(60+index*5), 20)

	rowNum := fmt.Sprintf("%2d", index+1)

	if index == m.budgetsView.selectedIndex {
		rowNum = SelectedItemStyle.Render(rowNum)
		agent = SelectedItemStyle.Render(agent)
	} else {
		rowNum = TableCellStyle.Render(rowNum)
		agent = TableCellStyle.Render(agent)
	}

	percentCol := TableCellStyle.Render(percent)
	progressCol := lipgloss.NewStyle().Foreground(ColorSuccess).Render(progressBar)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		rowNum,
		"  ",
		agent,
		"  ",
		TableCellStyle.Render(target),
		"  ",
		TableCellStyle.Render(actual),
		"  ",
		percentCol,
		"  ",
		progressCol,
	)
}

func (m *AppModel) updateBudgets(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.budgetsView.selectedIndex > 0 {
				m.budgetsView.selectedIndex--
			}
			return m, nil

		case "down":
			if m.budgetsView.selectedIndex < len(m.budgetsView.budgets)-1 {
				m.budgetsView.selectedIndex++
			}
			return m, nil

		case "enter":
			if len(m.budgetsView.budgets) > 0 {
				m.setMessage("Dettaglio budget")
			}
			return m, nil

		case "esc":
			return m.navigateBack(), nil
		}
	}
	return m, nil
}
