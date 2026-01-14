// internal/ui/view_settings.go
// System Settings View

package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) viewSettings() string {
	title := TitleStyle.Render("Impostazioni di Sistema")

	// Tabs
	tabs := []string{"Generali", "Database", "Stampa", "Utenti", "Logs"}
	var tabRows []string
	for i, t := range tabs {
		if i == m.settingsView.activeTab {
			tabRows = append(tabRows, TabActiveStyle.Render(" "+t+" "))
		} else {
			tabRows = append(tabRows, TabStyle.Render(" "+t+" "))
		}
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Left, tabRows...)

	// Content based on active tab
	var content string
	switch m.settingsView.activeTab {
	case 0:
		content = m.renderGeneralSettings()
	case 1:
		content = m.renderDatabaseSettings()
	case 2:
		content = m.renderPrintSettings()
	case 3:
		content = m.renderUsersSettings()
	case 4:
		content = m.renderLogsSettings()
	}

	// Full content
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		tabsRow,
		"",
		content,
	)

	availableWidth := m.width - 28

	return lipgloss.NewStyle().
		Width(availableWidth).
		Padding(1, 2).
		Render(fullContent)
}

func (m *AppModel) renderGeneralSettings() string {
	header := SubtitleStyle.Render("Impostazioni Generali")

	settings := []struct {
		label  string
		value  string
		status string
	}{
		{"Nome Applicazione", "Ricambi Manager", "✓"},
		{"Versione", "1.0.0", "✓"},
		{"Ambiente", "Production", "✓"},
		{"Timeout Sessione", "480 minuti", "✓"},
		{"Tema", "Dark Modern", "✓"},
		{"Lingua", "Italiano", "✓"},
	}

	var rows []string
	for _, s := range settings {
		label := TableCellStyle.Render(s.label)
		value := lipgloss.NewStyle().Foreground(ColorPrimary).Render(s.value)
		status := BadgeSuccessStyle.Render(s.status)
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, label, "  ", value, "  ", status))
	}

	list := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		ContentStyle.Render(list),
	)
}

func (m *AppModel) renderDatabaseSettings() string {
	header := SubtitleStyle.Render("Configurazione Database MongoDB")

	settings := []struct {
		label  string
		value  string
		status string
	}{
		{"Host", "localhost:27017", "✓"},
		{"Database", "ricambi_db", "✓"},
		{"Pool Min", "5 connessioni", "✓"},
		{"Pool Max", "50 connessioni", "✓"},
		{"Timeout", "10 secondi", "✓"},
		{"Stato Connessione", "Attiva", "✓"},
	}

	var rows []string
	for _, s := range settings {
		label := TableCellStyle.Render(s.label)
		value := lipgloss.NewStyle().Foreground(ColorPrimary).Render(s.value)
		status := BadgeSuccessStyle.Render(s.status)
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, label, "  ", value, "  ", status))
	}

	list := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		ContentStyle.Render(list),
	)
}

func (m *AppModel) renderPrintSettings() string {
	header := SubtitleStyle.Render("Configurazione Stampa")

	settings := []struct {
		label  string
		value  string
		status string
	}{
		{"Formato Barcode", "EAN13", "✓"},
		{"Formato Etichetta", "ZPL", "✓"},
		{"Larghezza Etichetta", "50mm", "✓"},
		{"Altezza Etichetta", "30mm", "✓"},
		{"Stampante Default", "Zebra ZPL", "✓"},
	}

	var rows []string
	for _, s := range settings {
		label := TableCellStyle.Render(s.label)
		value := lipgloss.NewStyle().Foreground(ColorPrimary).Render(s.value)
		status := BadgeSuccessStyle.Render(s.status)
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, label, "  ", value, "  ", status))
	}

	list := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		ContentStyle.Render(list),
	)
}

func (m *AppModel) renderUsersSettings() string {
	header := SubtitleStyle.Render("Gestione Utenti")

	stats := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderSettingStat("Totali", "12"),
		m.renderSettingStat("Attivi", "8"),
		m.renderSettingStat("Amministratori", "2"),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		stats,
	)
}

func (m *AppModel) renderSettingStat(label, value string) string {
	v := lipgloss.NewStyle().FontSize(18).Bold(true).Foreground(ColorPrimary).Render(value)
	l := StatsLabelStyle.Render(label)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Background(ColorBgCard).
		Padding(1, 2).
		Width(12).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				v,
				l,
			),
		)
}

func (m *AppModel) renderLogsSettings() string {
	header := SubtitleStyle.Render("Log e Audit")

	settings := []struct {
		label  string
		value  string
		status string
	}{
		{"File Log", "logs/app.log", "✓"},
		{"File Audit", "logs/audit.log", "✓"},
		{"Livello Log", "Info", "✓"},
		{"Backup Log", "10 file", "✓"},
		{"Rotazione", "100MB", "✓"},
		{"Conservazione", "30 giorni", "✓"},
	}

	var rows []string
	for _, s := range settings {
		label := TableCellStyle.Render(s.label)
		value := lipgloss.NewStyle().Foreground(ColorPrimary).Render(s.value)
		status := BadgeSuccessStyle.Render(s.status)
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, label, "  ", value, "  ", status))
	}

	list := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		ContentStyle.Render(list),
	)
}

func (m *AppModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.settingsView.activeTab = (m.settingsView.activeTab + 1) % 5
			return m, nil

		case "shift+tab":
			m.settingsView.activeTab--
			if m.settingsView.activeTab < 0 {
				m.settingsView.activeTab = 4
			}
			return m, nil

		case "up", "down":
			return m, nil

		case "enter":
			m.setMessage("Modifica impostazione")
			return m, nil

		case "esc":
			return m.navigateBack(), nil
		}
	}
	return m, nil
}
