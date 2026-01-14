// internal/ui/view_article_search.go

package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ricambi-manager/internal/domain"
)

func (m *AppModel) viewArticleSearch() string {
	title := TitleStyle.Render("🔍 Ricerca Articoli")

	searchTypeLabel := "Tipo: "
	searchTypes := map[string]string{
		"code":          "Codice",
		"description":   "Descrizione",
		"barcode":       "Barcode",
		"applicability": "Applicabilità",
	}
	currentType := searchTypes[m.searchView.searchType]
	searchTypeDisplay := searchTypeLabel + BadgeStyle.Render(currentType)

	queryLabel := "Query:"
	queryField := m.searchView.query
	if len(queryField) == 0 {
		queryField = "digita per cercare..."
	}
	queryDisplay := InputFocusedStyle.Render(queryField + "█")

	searchBox := CardStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		searchTypeDisplay,
		"",
		queryLabel,
		queryDisplay,
	))

	resultsTitle := SubtitleStyle.Render(fmt.Sprintf("Risultati (%d)", len(m.searchView.results)))

	var resultsList string
	if m.searchView.loading {
		resultsList = InfoStyle.Render("⏳ Caricamento in corso...")
	} else if len(m.searchView.results) == 0 {
		if len(m.searchView.query) > 0 {
			resultsList = InfoStyle.Render("🔍 Nessun risultato trovato")
		} else {
			resultsList = InfoStyle.Render("💡 Inizia a digitare per cercare")
		}
	} else {
		var items []string
		maxVisible := m.height - 20
		if maxVisible < 5 {
			maxVisible = 5
		}

		start := m.searchView.scrollOffset
		end := start + maxVisible
		if end > len(m.searchView.results) {
			end = len(m.searchView.results)
		}

		for i := start; i < end; i++ {
			article := m.searchView.results[i]
			stockBadge := ""
			if article.Stock.Available > 0 {
				stockBadge = BadgeSuccessStyle.Render(fmt.Sprintf("%.0f", article.Stock.Available))
			} else {
				stockBadge = BadgeDangerStyle.Render("0")
			}

			priceBadge := BadgeStyle.Render(fmt.Sprintf("€ %.2f", article.Pricing.ListPrice))

			itemText := fmt.Sprintf("%s - %s %s %s",
				article.Code,
				truncateString(article.Description, 50),
				stockBadge,
				priceBadge,
			)

			if i == m.searchView.selectedIndex {
				items = append(items, SelectedItemStyle.Render(fmt.Sprintf("  %s", itemText)))
			} else {
				items = append(items, UnselectedItemStyle.Render(fmt.Sprintf("  %s", itemText)))
			}
		}

		if len(m.searchView.results) > maxVisible {
			scrollInfo := fmt.Sprintf("%d/%d", m.searchView.selectedIndex+1, len(m.searchView.results))
			items = append(items, lipgloss.NewStyle().
				Foreground(ColorMuted).
				MarginTop(1).
				Render(fmt.Sprintf("  ── %s ──", scrollInfo)))
		}

		resultsList = lipgloss.JoinVertical(lipgloss.Left, items...)
	}

	resultsBox := ContentStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		resultsTitle,
		"",
		resultsList,
	))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		searchBox,
		"",
		resultsBox,
	)

	availableHeight := m.height - 6

	return lipgloss.Place(
		m.width,
		availableHeight,
		lipgloss.Left,
		lipgloss.Top,
		lipgloss.NewStyle().Padding(1, 2).Render(content),
	)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (m *AppModel) updateArticleSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			types := []string{"code", "description", "barcode", "applicability"}
			for i, t := range types {
				if t == m.searchView.searchType {
					m.searchView.searchType = types[(i+1)%len(types)]
					break
				}
			}
			m.searchView.query = ""
			m.searchView.results = []*domain.Article{}
			m.searchView.selectedIndex = 0
			m.searchView.scrollOffset = 0
			return m, nil

		case "up":
			if m.searchView.selectedIndex > 0 {
				m.searchView.selectedIndex--
				m.updateScrollOffset()
			}
			return m, nil

		case "down":
			if m.searchView.selectedIndex < len(m.searchView.results)-1 {
				m.searchView.selectedIndex++
				m.updateScrollOffset()
			}
			return m, nil

		case "pageup":
			maxVisible := m.height - 20
			if maxVisible < 5 {
				maxVisible = 5
			}
			m.searchView.selectedIndex -= maxVisible
			if m.searchView.selectedIndex < 0 {
				m.searchView.selectedIndex = 0
			}
			m.updateScrollOffset()
			return m, nil

		case "pagedown":
			maxVisible := m.height - 20
			if maxVisible < 5 {
				maxVisible = 5
			}
			m.searchView.selectedIndex += maxVisible
			if m.searchView.selectedIndex >= len(m.searchView.results) {
				m.searchView.selectedIndex = len(m.searchView.results) - 1
			}
			m.updateScrollOffset()
			return m, nil

		case "home":
			m.searchView.selectedIndex = 0
			m.searchView.scrollOffset = 0
			return m, nil

		case "end":
			m.searchView.selectedIndex = len(m.searchView.results) - 1
			if m.searchView.selectedIndex < 0 {
				m.searchView.selectedIndex = 0
			}
			m.updateScrollOffset()
			return m, nil

		case "enter":
			if len(m.searchView.results) > 0 {
				return m, nil
			}
			return m, nil

		case "backspace":
			if len(m.searchView.query) > 0 {
				m.searchView.query = m.searchView.query[:len(m.searchView.query)-1]
				if len(m.searchView.query) >= 2 {
					return m, m.performSearch()
				} else {
					m.searchView.results = []*domain.Article{}
				}
			}
			return m, nil

		case "esc":
			return m.navigateBack(), nil

		default:
			if len(msg.String()) == 1 {
				m.searchView.query += msg.String()
				if len(m.searchView.query) >= 2 {
					return m, m.performSearch()
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *AppModel) updateScrollOffset() {
	maxVisible := m.height - 20
	if maxVisible < 5 {
		maxVisible = 5
	}

	if m.searchView.selectedIndex < m.searchView.scrollOffset {
		m.searchView.scrollOffset = m.searchView.selectedIndex
	} else if m.searchView.selectedIndex >= m.searchView.scrollOffset+maxVisible {
		m.searchView.scrollOffset = m.searchView.selectedIndex - maxVisible + 1
	}
}

func (m *AppModel) performSearch() tea.Cmd {
	m.searchView.loading = true

	return func() tea.Msg {
		ctx := context.Background()

		var results []*domain.Article
		var err error

		switch m.searchView.searchType {
		case "code":
			searchResult, searchErr := m.searchUC.SearchByCode(ctx, m.searchView.query, 20)
			if searchErr == nil && searchResult != nil {
				results = searchResult.Articles
			}
			err = searchErr

		case "description":
			searchResult, searchErr := m.searchUC.SearchByDescription(ctx, m.searchView.query, 20)
			if searchErr == nil && searchResult != nil {
				results = searchResult.Articles
			}
			err = searchErr

		case "barcode":
			article, searchErr := m.searchUC.SearchByBarcode(ctx, m.searchView.query)
			if searchErr == nil && article != nil {
				results = []*domain.Article{article}
			}
			err = searchErr

		default:
			searchResult, searchErr := m.searchUC.FuzzySearch(ctx, m.searchView.query, 20)
			if searchErr == nil && searchResult != nil {
				results = searchResult.Articles
			}
			err = searchErr
		}

		return searchResultMsg{results: results, err: err}
	}
}
