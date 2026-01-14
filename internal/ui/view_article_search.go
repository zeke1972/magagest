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

	searchTypeLabel := "Tipo di ricerca: "
	searchTypes := []string{"Codice", "Descrizione", "Barcode", "Applicabilità"}
	currentType := searchTypes[0]
	switch m.searchView.searchType {
	case "code":
		currentType = searchTypes[0]
	case "description":
		currentType = searchTypes[1]
	case "barcode":
		currentType = searchTypes[2]
	case "applicability":
		currentType = searchTypes[3]
	}

	searchTypeDisplay := searchTypeLabel + BadgeStyle.Render(currentType)

	queryLabel := "Query:"
	queryField := m.searchView.query
	if len(queryField) == 0 {
		queryField = "digita per cercare..."
	}
	queryDisplay := InputFocusedStyle.Render(queryField + "█")

	searchBox := lipgloss.JoinVertical(
		lipgloss.Left,
		searchTypeDisplay,
		"",
		queryLabel,
		queryDisplay,
	)

	resultsTitle := SubtitleStyle.Render(fmt.Sprintf("Risultati (%d)", len(m.searchView.results)))

	var resultsList string
	if m.searchView.loading {
		resultsList = InfoStyle.Render("⏳ Caricamento...")
	} else if len(m.searchView.results) == 0 {
		if len(m.searchView.query) > 0 {
			resultsList = InfoStyle.Render("Nessun risultato trovato")
		} else {
			resultsList = InfoStyle.Render("Inizia a digitare per cercare")
		}
	} else {
		var items []string
		for i, article := range m.searchView.results {
			stockBadge := ""
			if article.Stock.Available > 0 {
				stockBadge = BadgeSuccessStyle.Render(fmt.Sprintf("%.0f", article.Stock.Available))
			} else {
				stockBadge = BadgeDangerStyle.Render("0")
			}

			priceBadge := BadgeStyle.Render(fmt.Sprintf("€ %.2f", article.Pricing.ListPrice))

			itemText := fmt.Sprintf("%s - %s %s %s",
				article.Code,
				article.Description,
				stockBadge,
				priceBadge,
			)

			if i == m.searchView.selectedIndex {
				items = append(items, SelectedItemStyle.Render("► "+itemText))
			} else {
				items = append(items, UnselectedItemStyle.Render("  "+itemText))
			}
		}
		resultsList = lipgloss.JoinVertical(lipgloss.Left, items...)
	}

	results := lipgloss.JoinVertical(
		lipgloss.Left,
		resultsTitle,
		"",
		resultsList,
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		searchBox,
		"",
		"",
		results,
	)

	if m.error != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			ErrorStyle.Render("❌ "+m.error),
		)
	}

	availableHeight := m.height - 6

	return lipgloss.Place(
		m.width,
		availableHeight,
		lipgloss.Left,
		lipgloss.Top,
		lipgloss.NewStyle().Padding(1, 2).Render(content),
	)
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
			return m, nil

		case "up":
			if m.searchView.selectedIndex > 0 {
				m.searchView.selectedIndex--
			}
			return m, nil

		case "down":
			if m.searchView.selectedIndex < len(m.searchView.results)-1 {
				m.searchView.selectedIndex++
			}
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

func (m *AppModel) performSearch() tea.Cmd {
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
