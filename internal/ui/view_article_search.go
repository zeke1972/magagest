// internal/ui/view_article_search.go
// Modernized Article Search View

package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ricambi-manager/internal/domain"
)

func (m *AppModel) viewArticleSearch() string {
	title := TitleStyle.Render("Ricerca Articoli")

	// Search type tabs
	searchTypes := []string{"Codice", "Descrizione", "Barcode", "Applicabilità", "Cross-Reference"}
	searchTypeLabels := []string{"code", "description", "barcode", "applicability", "crossref"}

	tabsRow := m.renderSearchTabs(searchTypes, searchTypeLabels)

	// Search input
	searchInput := m.renderSearchInput()

	// Filters indicator
	filters := m.renderSearchFilters()

	// Results section
	resultsSection := m.renderSearchResults()

	// Content layout
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		tabsRow,
		"",
		searchInput,
		filters,
		"",
		resultsSection,
	)

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

func (m *AppModel) renderSearchTabs(types []string, labels []string) string {
	var tabs []string
	for i, t := range types {
		tabLabel := labels[i]
		isActive := m.searchView.searchType == tabLabel

		var tab string
		if isActive {
			tab = TabActiveStyle.Render(" "+t+" ")
		} else {
			tab = TabStyle.Render(" "+t+" ")
		}
		tabs = append(tabs, tab)
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m *AppModel) renderSearchInput() string {
	label := InputLabelStyle.Render("Cerca:")

	queryField := m.searchView.query
	if len(queryField) == 0 {
		queryField = "digita per cercare..."
		placeholderStyle := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)
		queryField = placeholderStyle.Render(queryField)
	}

	queryDisplay := InputFocusedStyle.Render(queryField + "█")

	searchIcon := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Render("🔍")

	hint := lipgloss.NewStyle().
		Foreground(ColorMuted).
		FontSize(10).
		Render("scrivi almeno 2 caratteri per cercare")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		label,
		lipgloss.JoinHorizontal(lipgloss.Left, searchIcon, " ", queryDisplay),
		hint,
	)
}

func (m *AppModel) renderSearchFilters() string {
	if !m.searchView.filtersActive {
		return ""
	}

	filters := []string{
		"✓ Solo disponibili",
		"✓ Prezzo > 0",
	}

	var filterItems []string
	for _, f := range filters {
		filterItems = append(filterItems, BadgeStyle.Render(f))
	}

	return lipgloss.NewStyle().
		MarginTop(1).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				filterItems...,
			),
		)
}

func (m *AppModel) renderSearchResults() string {
	resultsTitle := SubtitleStyle.Render(
		fmt.Sprintf("Risultati (%d articoli trovati)", len(m.searchView.results)),
	)

	var resultsList string
	if m.searchView.loading {
		resultsList = lipgloss.JoinVertical(
			lipgloss.Center,
			RenderLoadingSpinner(),
		)
	} else if len(m.searchView.results) == 0 {
		if len(m.searchView.query) > 0 {
			resultsList = RenderEmptyState("Nessun articolo trovato")
		} else {
			resultsList = RenderEmptyState("Inizia a digitare per cercare articoli")
		}
	} else {
		resultsList = m.renderResultsTable()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		resultsTitle,
		"",
		resultsList,
	)
}

func (m *AppModel) renderResultsTable() string {
	// Table header
	headerCode := TableHeaderStyle.Render("Codice")
	headerDesc := TableHeaderStyle.Render("Descrizione")
	headerStock := TableHeaderStyle.Render("Giac.")
	headerPrice := TableHeaderStyle.Render("Prezzo")
	headerBrand := TableHeaderStyle.Render("Brand")

	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerCode,
		headerDesc,
		headerStock,
		headerPrice,
		headerBrand,
	)

	// Table rows
	var rows []string
	for i, article := range m.searchView.results {
		row := m.renderArticleRow(article, i == m.searchView.selectedIndex)
		rows = append(rows, row)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Container
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				headerRow,
				lipgloss.NewStyle().
					Foreground(ColorBorder).
					Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"),
				body,
			),
		)
}

func (m *AppModel) renderArticleRow(article *domain.Article, selected bool) string {
	// Stock status and badge
	stockStatus := ""
	stockStyle := TableCellStyle
	if article.Stock.Available > 10 {
		stockStatus = "✓ " + fmt.Sprintf("%.0f", article.Stock.Available)
		stockStyle = lipgloss.NewStyle().Foreground(ColorSuccess)
	} else if article.Stock.Available > 0 {
		stockStatus = "⚠ " + fmt.Sprintf("%.0f", article.Stock.Available)
		stockStyle = lipgloss.NewStyle().Foreground(ColorWarning)
	} else {
		stockStatus = "✗ Esaurito"
		stockStyle = lipgloss.NewStyle().Foreground(ColorDanger)
	}

	code := fmt.Sprintf("%-15s", article.Code)
	desc := truncateString(article.Description, 40)
	stock := fmt.Sprintf("%8s", stockStatus)
	price := fmt.Sprintf("€ %9.2f", article.Pricing.ListPrice)
	brand := fmt.Sprintf("%-10s", article.Brand)

	if selected {
		code = SelectedItemStyle.Render(code)
		desc = SelectedItemStyle.Render(desc)
		stock = stockStyle.Copy().Bold(true).Render(stock)
		price = SelectedItemStyle.Render(price)
		brand = SelectedItemStyle.Render(brand)
	} else {
		code = TableCellStyle.Render(code)
		desc = TableCellStyle.Render(desc)
		stock = stockStyle.Render(stock)
		price = TableCellStyle.Render(price)
		brand = TableCellStyle.Render(brand)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		code,
		desc,
		stock,
		price,
		brand,
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
			types := []string{"code", "description", "barcode", "applicability", "crossref"}
			for i, t := range types {
				if t == m.searchView.searchType {
					m.searchView.searchType = types[(i+1)%len(types)]
					break
				}
			}
			m.searchView.query = ""
			m.searchView.results = []*domain.Article{}
			return m, nil

		case "shift+tab":
			types := []string{"code", "description", "barcode", "applicability", "crossref"}
			for i, t := range types {
				if t == m.searchView.searchType {
					prevIdx := i - 1
					if prevIdx < 0 {
						prevIdx = len(types) - 1
					}
					m.searchView.searchType = types[prevIdx]
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
			if len(m.searchView.results) > 0 && m.searchView.selectedIndex < len(m.searchView.results) {
				selected := m.searchView.results[m.searchView.selectedIndex]
				m.setMessage(fmt.Sprintf("Articolo selezionato: %s - %s", selected.Code, selected.Description))
				return m, nil
			}
			return m, nil

		case "backspace":
			if len(m.searchView.query) > 0 {
				m.searchView.query = m.searchView.query[:len(m.searchView.query)-1]
				if len(m.searchView.query) >= 2 {
					m.searchView.loading = true
					return m, m.performSearch()
				} else {
					m.searchView.results = []*domain.Article{}
				}
			}
			return m, nil

		case "ctrl+u":
			// Clear search
			m.searchView.query = ""
			m.searchView.results = []*domain.Article{}
			m.searchView.selectedIndex = 0
			return m, nil

		case "f":
			// Toggle filters
			m.searchView.filtersActive = !m.searchView.filtersActive
			if len(m.searchView.query) >= 2 {
				return m, m.performSearch()
			}
			return m, nil

		case "r":
			// Refresh results
			if len(m.searchView.query) >= 2 {
				m.searchView.loading = true
				return m, m.performSearch()
			}
			return m, nil

		case "esc":
			return m.navigateBack(), nil

		default:
			if len(msg.String()) == 1 {
				m.searchView.query += msg.String()
				if len(m.searchView.query) >= 2 {
					m.searchView.loading = true
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

		query := strings.TrimSpace(m.searchView.query)
		if len(query) < 2 {
			return searchResultMsg{results: nil, err: nil}
		}

		switch m.searchView.searchType {
		case "code":
			searchResult, searchErr := m.searchUC.SearchByCode(ctx, query, 50)
			if searchErr == nil && searchResult != nil {
				results = searchResult.Articles
			}
			err = searchErr

		case "description":
			searchResult, searchErr := m.searchUC.SearchByDescription(ctx, query, 50)
			if searchErr == nil && searchResult != nil {
				results = searchResult.Articles
			}
			err = searchErr

		case "barcode":
			article, searchErr := m.searchUC.SearchByBarcode(ctx, query)
			if searchErr == nil && article != nil {
				results = []*domain.Article{article}
			}
			err = searchErr

		case "applicability":
			searchResult, searchErr := m.searchUC.SearchByApplicability(ctx, query, 50)
			if searchErr == nil && searchResult != nil {
				results = searchResult.Articles
			}
			err = searchErr

		case "crossref":
			searchResult, searchErr := m.searchUC.SearchByCrossRef(ctx, query, 50)
			if searchErr == nil && searchResult != nil {
				results = searchResult.Articles
			}
			err = searchErr

		default:
			searchResult, searchErr := m.searchUC.FuzzySearch(ctx, query, 50)
			if searchErr == nil && searchResult != nil {
				results = searchResult.Articles
			}
			err = searchErr
		}

		return searchResultMsg{results: results, err: err}
	}
}
