// internal/ui/view_main_menu.go
// Modernized Main Menu View

package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ricambi-manager/internal/domain"
)

func (m *AppModel) viewMainMenu() string {
	title := TitleStyle.Render("Dashboard")
	welcome := SubtitleStyle.Render("Benvenuto, " + m.operator.FullName)

	// Dashboard stats cards
	statsRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderStatCard("📦", "Articoli", fmt.Sprintf("%d", m.mainMenuView.stats.totalArticles)),
		m.renderStatCard("⚠️", "Scorta Bassa", fmt.Sprintf("%d", m.mainMenuView.stats.lowStockItems)),
		m.renderStatCard("👥", "Clienti", fmt.Sprintf("%d", m.mainMenuView.stats.activeCustomers)),
		m.renderStatCard("📋", "Ordini", fmt.Sprintf("%d", m.mainMenuView.stats.pendingOrders)),
	)

	// Quick actions section
	quickActions := m.renderQuickActions()

	// Menu items with enhanced styling
	menuTitle := MenuTitleStyle.Render("Menu Navigazione")

	var menuItems []string
	for i, item := range m.mainMenuView.menuItems {
		if !item.Enabled {
			continue
		}

		shortcut := ShortcutKeyStyle.Render(item.Shortcut)
		icon := lipgloss.NewStyle().Foreground(ColorMuted).Render(item.Icon)
		label := item.Label
		desc := lipgloss.NewStyle().Foreground(ColorMuted).FontSize(10).Render(item.Description)

		if i == m.mainMenuView.selectedIndex {
			row := lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.JoinHorizontal(lipgloss.Left, "  ► ", shortcut, " ", icon, " ", label),
				desc,
			)
			menuItems = append(menuItems, SelectedItemStyle.Render(row))
		} else {
			row := lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.JoinHorizontal(lipgloss.Left, "   ", shortcut, " ", icon, " ", label),
				desc,
			)
			menuItems = append(menuItems, UnselectedItemStyle.Render(row))
		}
	}

	menuContent := lipgloss.JoinVertical(lipgloss.Left, menuItems...)
	menuBox := ContentStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			menuTitle,
			"",
			menuContent,
		),
	)

	// Main content layout
	leftPanel := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		welcome,
		"",
		statsRow,
		"",
		quickActions,
	)

	rightPanel := lipgloss.JoinVertical(
		lipgloss.Right,
		menuBox,
	)

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		rightPanel,
	)

	// Messages
	if m.message != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			SuccessStyle.Render("✓ "+m.message),
		)
	}

	if m.error != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			ErrorStyle.Render("✗ "+m.error),
		)
	}

	availableWidth := m.width - 28 // Account for sidebar

	return lipgloss.NewStyle().
		Width(availableWidth).
		Padding(1, 2).
		Render(content)
}

func (m *AppModel) renderStatCard(icon, label, value string) string {
	iconStyled := lipgloss.NewStyle().FontSize(20).Foreground(ColorPrimary).Render(icon)
	valueStyled := StatsValueStyle.Render(value)
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

func (m *AppModel) renderQuickActions() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg).
		Render("Azioni Rapide")

	actions := []struct {
		key  string
		desc string
	}{
		{"F1", "Nuovo Articolo"},
		{"F2", "Nuovo Cliente"},
		{"F3", "Cerca Articolo"},
		{"F4", "Stampa Etichetta"},
		{"F5", "Aggiorna Lista"},
		{"F12", "Esci"},
	}

	var actionRows []string
	for _, a := range actions {
		key := ShortcutKeyStyle.Render(a.key)
		desc := UnselectedItemStyle.Render(a.desc)
		actionRows = append(actionRows, lipgloss.JoinHorizontal(lipgloss.Left, key, " ", desc))
	}

	actionsContent := lipgloss.JoinVertical(lipgloss.Left, actionRows...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		actionsContent,
	)
}

func (m *AppModel) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5", "6", "7":
			num := int(msg.String()[0] - '0')

			enabledIndex := 0
			for i, item := range m.mainMenuView.menuItems {
				if !item.Enabled {
					continue
				}
				enabledIndex++
				if enabledIndex == num {
					m.mainMenuView.selectedIndex = i
					m.clearMessages()

					switch item.View {
					case ViewArticleSearch:
						m.searchView = &ArticleSearchView{
							searchType:  "code",
							searchTypes: []string{"code", "description", "barcode", "applicability"},
							results:     []*domain.Article{},
						}
					case ViewCustomerSearch:
						m.customerSearchView = &CustomerSearchView{
							results: []*domain.Customer{},
						}
					}

					return m.navigateTo(item.View), nil
				}
			}
			return m, nil

		case "up", "k":
			if m.mainMenuView.selectedIndex > 0 {
				m.mainMenuView.selectedIndex--
				for m.mainMenuView.selectedIndex > 0 &&
					!m.mainMenuView.menuItems[m.mainMenuView.selectedIndex].Enabled {
					m.mainMenuView.selectedIndex--
				}
			}
			return m, nil

		case "down", "j":
			if m.mainMenuView.selectedIndex < len(m.mainMenuView.menuItems)-1 {
				m.mainMenuView.selectedIndex++
				for m.mainMenuView.selectedIndex < len(m.mainMenuView.menuItems) &&
					!m.mainMenuView.menuItems[m.mainMenuView.selectedIndex].Enabled {
					m.mainMenuView.selectedIndex++
				}
				if m.mainMenuView.selectedIndex >= len(m.mainMenuView.menuItems) {
					m.mainMenuView.selectedIndex = len(m.mainMenuView.menuItems) - 1
				}
			}
			return m, nil

		case "enter":
			selectedItem := m.mainMenuView.menuItems[m.mainMenuView.selectedIndex]
			if selectedItem.Enabled {
				m.clearMessages()

				switch selectedItem.View {
				case ViewArticleSearch:
					m.searchView = &ArticleSearchView{
						searchType:  "code",
						searchTypes: []string{"code", "description", "barcode", "applicability"},
						results:     []*domain.Article{},
					}
				case ViewCustomerSearch:
					m.customerSearchView = &CustomerSearchView{
						results: []*domain.Customer{},
					}
				}

				return m.navigateTo(selectedItem.View), nil
			}
			return m, nil

		case "f1":
			// Quick action: New article
			m.setMessage("Funzione non ancora implementata")
			return m, nil

		case "f2":
			// Quick action: New customer
			m.setMessage("Funzione non ancora implementata")
			return m, nil

		case "f3":
			// Quick action: Search article
			m.searchView = &ArticleSearchView{
				searchType:  "code",
				searchTypes: []string{"code", "description", "barcode", "applicability"},
				results:     []*domain.Article{},
			}
			return m.navigateTo(ViewArticleSearch), nil

		case "f4":
			// Quick action: Print label
			m.setMessage("Funzione non ancora implementata")
			return m, nil

		case "f5":
			// Quick action: Refresh
			return m, nil

		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

// getCurrentTime returns formatted current time
func getCurrentTime() string {
	return time.Now().Format("15:04:05")
}
