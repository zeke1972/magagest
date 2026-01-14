// internal/ui/view_main_menu.go

package ui

import (
	"fmt"
	"ricambi-manager/internal/domain"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) viewMainMenu() string {
	title := TitleStyle.Render("📋 Menu Principale")
	welcome := SubtitleStyle.Render("Benvenuto, " + m.operator.FullName)

	var menuItems []string
	shortcutNum := 1

	for i, item := range m.mainMenuView.menuItems {
		if !item.Enabled {
			continue
		}

		shortcut := ShortcutStyle.Render(fmt.Sprintf("[%d]", shortcutNum))
		label := item.Label

		var itemLine string
		if i == m.mainMenuView.selectedIndex {
			itemLine = SelectedItemStyle.Render(fmt.Sprintf("  %s  %s", shortcut, label))
			menuItems = append(menuItems, itemLine)
			if item.Description != "" {
				descLine := lipgloss.NewStyle().
					Foreground(ColorMuted).
					MarginLeft(8).
					Render("  └─ " + item.Description)
				menuItems = append(menuItems, descLine)
			}
		} else {
			itemLine = UnselectedItemStyle.Render(fmt.Sprintf("  %s  %s", shortcut, label))
			menuItems = append(menuItems, itemLine)
		}

		menuItems = append(menuItems, "")

		shortcutNum++
	}

	if len(menuItems) > 0 {
		menuItems = menuItems[:len(menuItems)-1]
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, menuItems...)

	menuBox := ContentStyle.Render(menu)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		welcome,
		"",
		menuBox,
	)

	availableHeight := m.height - 6

	return lipgloss.Place(
		m.width,
		availableHeight,
		lipgloss.Center,
		lipgloss.Center,
		content,
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
							searchType: "code",
							results:    []*domain.Article{},
						}
					}

					return m.navigateTo(item.View), nil
				}
			}
			return m, nil

		case "up", "k":
			m.mainMenuView.selectedIndex--
			if m.mainMenuView.selectedIndex < 0 {
				m.mainMenuView.selectedIndex = len(m.mainMenuView.menuItems) - 1
			}
			for !m.mainMenuView.menuItems[m.mainMenuView.selectedIndex].Enabled {
				m.mainMenuView.selectedIndex--
				if m.mainMenuView.selectedIndex < 0 {
					m.mainMenuView.selectedIndex = len(m.mainMenuView.menuItems) - 1
				}
			}
			return m, nil

		case "down", "j":
			m.mainMenuView.selectedIndex++
			if m.mainMenuView.selectedIndex >= len(m.mainMenuView.menuItems) {
				m.mainMenuView.selectedIndex = 0
			}
			for !m.mainMenuView.menuItems[m.mainMenuView.selectedIndex].Enabled {
				m.mainMenuView.selectedIndex++
				if m.mainMenuView.selectedIndex >= len(m.mainMenuView.menuItems) {
					m.mainMenuView.selectedIndex = 0
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
						searchType: "code",
						results:    []*domain.Article{},
					}
				}

				return m.navigateTo(selectedItem.View), nil
			}
			return m, nil

		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}
