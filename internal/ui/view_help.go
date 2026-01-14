// internal/ui/view_help.go
// Help and Shortcuts View

package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) viewHelp() string {
	title := TitleStyle.Render("Guida Rapida")

	// Navigation shortcuts
	navigation := m.renderHelpSection("NAVIGAZIONE", [][]string{
		{"↑ / k", "Elemento precedente"},
		{"↓ / j", "Elemento successivo"},
		{"← / →", "Schede / Campi"},
		{"TAB", "Campo successivo"},
		{"SHIFT+TAB", "Campo precedente"},
		{"ENTER", "Conferma / Seleziona"},
		{"ESC", "Indietro"},
		{"?", "Mostra questa guida"},
	})

	// Main menu shortcuts
	mainMenu := m.renderHelpSection("MENU PRINCIPALE", [][]string{
		{"1-7", "Accesso rapido alle funzioni"},
		{"F1", "Nuovo articolo"},
		{"F2", "Nuovo cliente"},
		{"F3", "Cerca articolo"},
		{"F4", "Stampa etichetta"},
		{"F5", "Aggiorna dati"},
		{"Q", "Esci dall'applicazione"},
	})

	// Search shortcuts
	search := m.renderHelpSection("RICERCA", [][]string{
		{"TAB", "Cambia tipo ricerca"},
		{"CTRL+U", "Pulisci ricerca"},
		{"F", "Attiva filtri"},
		{"R", "Aggiorna risultati"},
		{"↑ / ↓", "Naviga risultati"},
		{"ENTER", "Seleziona articolo"},
	})

	// Actions
	actions := m.renderHelpSection("AZIONI", [][]string{
		{"N", "Nuovo elemento"},
		{"E", "Modifica"},
		{"D", "Elimina"},
		{"P", "Stampa"},
		{"S", "Salva"},
	})

	// Tips
	tips := m.renderHelpTips()

	// Content layout
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		lipgloss.NewStyle().
			Foreground(ColorFgMuted).
			Render("Questa guida mostra i tasti快捷 per navigare e utilizzare Ricambi Manager."),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			navigation,
			mainMenu,
		),
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			search,
			actions,
		),
		"",
		tips,
	)

	availableWidth := m.width - 28

	return lipgloss.NewStyle().
		Width(availableWidth).
		Padding(1, 2).
		Render(content)
}

func (m *AppModel) renderHelpSection(title string, items [][]string) string {
	sectionTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(title)

	var rows []string
	for _, item := range items {
		key := ShortcutKeyStyle.Render(item[0])
		desc := UnselectedItemStyle.Render(item[1])
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, key, "  ", desc))
	}

	list := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return ContentStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			sectionTitle,
			"",
			list,
		),
	)
}

func (m *AppModel) renderHelpTips() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSuccess).
		Render("💡 CONSIGLI")

	tips := []string{
		"• Usa i numeri 1-7 per accesso rapido al menu",
		"• F3 ti porta direttamente alla ricerca articoli",
		"• TAB cicla tra i tipi di ricerca disponibili",
		"• CTRL+C fuori dal menu principale chiude l'app",
		"• ? mostra questa guida in qualsiasi momento",
	}

	var tipRows []string
	for _, t := range tips {
		tipRows = append(tipRows, lipgloss.NewStyle().Foreground(ColorFgMuted).Render(t))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Background(ColorBgCard).
		Padding(1, 2).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				"",
				lipgloss.JoinVertical(lipgloss.Left, tipRows...),
			),
		)
}

func (m *AppModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "?":
			return m.navigateBack(), nil
		}
	}
	return m, nil
}
