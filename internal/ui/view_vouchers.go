// internal/ui/view_vouchers.go
// Credit Vouchers Management View

package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) viewVouchers() string {
	title := TitleStyle.Render("Buoni Credito")

	// Stats row
	stats := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderVoucherStatCard("💰", "Totale", "€ 5.240", ColorSuccess),
		m.renderVoucherStatCard("📋", "Attivi", "24", ColorInfo),
		m.renderVoucherStatCard("⏳", "In Scadenza", "3", ColorWarning),
	)

	// Vouchers list
	vouchersList := m.renderVouchersList()

	// Content
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		stats,
		"",
		vouchersList,
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

func (m *AppModel) renderVoucherStatCard(icon, label, value string, color lipgloss.Color) string {
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

func (m *AppModel) renderVouchersList() string {
	header := SubtitleStyle.Render("Buoni Credito")

	var vouchers []string
	if len(m.vouchersView.vouchers) == 0 {
		vouchers = append(vouchers, RenderEmptyState("Nessun buono credito"))
	} else {
		for i := range m.vouchersView.vouchers {
			voucherRow := m.renderVoucherRow(i)
			vouchers = append(vouchers, voucherRow)
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, vouchers...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		ContentStyle.Render(list),
	)
}

func (m *AppModel) renderVoucherRow(index int) string {
	code := fmt.Sprintf("VC-%04d", 1000+index)
	amount := fmt.Sprintf("€ %.2f", float64(50+index*25))
	customer := fmt.Sprintf("Cliente %d", index+1)
	expiry := fmt.Sprintf("%d/%02d/2025", 15+index, 3+index)
	status := "Attivo"

	rowNum := fmt.Sprintf("%2d", index+1)

	if index == m.vouchersView.selectedIndex {
		rowNum = SelectedItemStyle.Render(rowNum)
		code = SelectedItemStyle.Render(code)
		amount = SelectedItemStyle.Render(amount)
		customer = SelectedItemStyle.Render(customer)
		expiry = SelectedItemStyle.Render(expiry)
	} else {
		rowNum = TableCellStyle.Render(rowNum)
		code = TableCellStyle.Render(code)
		amount = TableCellStyle.Render(amount)
		customer = TableCellStyle.Render(customer)
		expiry = TableCellStyle.Render(expiry)
	}

	statusBadge := BadgeSuccessStyle.Render(status)
	if index%3 == 1 {
		statusBadge = BadgeWarningStyle.Render("In scadenza")
	} else if index%3 == 2 {
		statusBadge = BadgeDangerStyle.Render("Utilizzato")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		rowNum,
		"  ",
		code,
		"  ",
		amount,
		"  ",
		customer,
		"  ",
		expiry,
		"  ",
		statusBadge,
	)
}

func (m *AppModel) updateVouchers(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.vouchersView.selectedIndex > 0 {
				m.vouchersView.selectedIndex--
			}
			return m, nil

		case "down":
			if m.vouchersView.selectedIndex < len(m.vouchersView.vouchers)-1 {
				m.vouchersView.selectedIndex++
			}
			return m, nil

		case "enter":
			if len(m.vouchersView.vouchers) > 0 {
				m.setMessage("Utilizzo buono")
			}
			return m, nil

		case "n":
			m.setMessage("Nuovo buono credito")
			return m, nil

		case "esc":
			return m.navigateBack(), nil
		}
	}
	return m, nil
}
