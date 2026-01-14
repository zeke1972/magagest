// internal/ui/view_customer_search.go
// Modernized Customer Search View

package ui

import (
    "context"
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"

    "ricambi-manager/internal/domain"
)

func (m *AppModel) viewCustomerSearch() string {
    title := TitleStyle.Render("Ricerca Clienti")

    // Search input
    searchInput := m.renderCustomerSearchInput()

    // Results section
    resultsSection := m.renderCustomerResults()

    // Content layout
    content := lipgloss.JoinVertical(
        lipgloss.Left,
        title,
        "",
        searchInput,
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

    availableWidth := m.width - 28

    return lipgloss.NewStyle().
        Width(availableWidth).
        Padding(1, 2).
        Render(content)
}

func (m *AppModel) renderCustomerSearchInput() string {
    label := InputLabelStyle.Render("Cerca cliente:")

    queryField := m.customerSearchView.query
    if len(queryField) == 0 {
        queryField = "ragione sociale, partita IVA, codice..."
        placeholderStyle := lipgloss.NewStyle().
            Foreground(ColorMuted).
            Italic(true)
        queryField = placeholderStyle.Render(queryField)
    }

    queryDisplay := InputFocusedStyle.Render(queryField + "█")

    searchIcon := lipgloss.NewStyle().
        Foreground(ColorPrimary).
        Render("👥")

    hint := lipgloss.NewStyle().
        Foreground(ColorMuted).
        FontSize(10).
        Render("cerca per ragione sociale, P.IVA, email o telefono")

    return lipgloss.JoinVertical(
        lipgloss.Left,
        label,
        lipgloss.JoinHorizontal(lipgloss.Left, searchIcon, " ", queryDisplay),
        hint,
    )
}

func (m *AppModel) renderCustomerResults() string {
    resultsTitle := SubtitleStyle.Render(
        fmt.Sprintf("Clienti trovati (%d)", len(m.customerSearchView.results)),
    )

    var resultsList string
    if m.customerSearchView.loading {
        resultsList = lipgloss.JoinVertical(
            lipgloss.Center,
            RenderLoadingSpinner(),
        )
    } else if len(m.customerSearchView.results) == 0 {
        if len(m.customerSearchView.query) > 0 {
            resultsList = RenderEmptyState("Nessun cliente trovato")
        } else {
            resultsList = RenderEmptyState("Inizia a digitare per cercare clienti")
        }
    } else {
        resultsList = m.renderCustomerTable()
    }

    return lipgloss.JoinVertical(
        lipgloss.Left,
        resultsTitle,
        "",
        resultsList,
    )
}

func (m *AppModel) renderCustomerTable() string {
    headerCode := TableHeaderStyle.Render("Codice")
    headerName := TableHeaderStyle.Render("Ragione Sociale")
    headerVAT := TableHeaderStyle.Render("P.IVA")
    headerCity := TableHeaderStyle.Render("Città")
    headerCredit := TableHeaderStyle.Render("Credito")
    headerStatus := TableHeaderStyle.Render("Stato")

    headerRow := lipgloss.JoinHorizontal(
        lipgloss.Left,
        headerCode,
        headerName,
        headerVAT,
        headerCity,
        headerCredit,
        headerStatus,
    )

    var rows []string
    for i, customer := range m.customerSearchView.results {
        row := m.renderCustomerRow(customer, i == m.customerSearchView.selectedIndex)
        rows = append(rows, row)
    }

    body := lipgloss.JoinVertical(lipgloss.Left, rows...)

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

func (m *AppModel) renderCustomerRow(customer *domain.Customer, selected bool) string {
    code := fmt.Sprintf("%-8s", customer.Code)
    name := truncateString(customer.CompanyName, 35)
    vat := fmt.Sprintf("%-12s", customer.VATNumber)
    city := truncateString(customer.BillingAddress.City, 15)
    
    credit := fmt.Sprintf("€ %.2f", customer.CreditInfo.CurrentExposure)
    creditStatus := "Attivo"
    
    if customer.CreditInfo.FidoLimit > 0 {
        usage := (customer.CreditInfo.CurrentExposure / customer.CreditInfo.FidoLimit) * 100
        if usage >= 100 {
            creditStatus = "Bloccato"
            credit = lipgloss.NewStyle().Foreground(ColorDanger).Render(credit)
        } else if usage >= 80 {
            creditStatus = "Attenzione"
            credit = lipgloss.NewStyle().Foreground(ColorWarning).Render(credit)
        }
    }

    if selected {
        code = SelectedItemStyle.Render(code)
        name = SelectedItemStyle.Render(name)
        vat = SelectedItemStyle.Render(vat)
        city = SelectedItemStyle.Render(city)
        credit = SelectedItemStyle.Render(credit)
    } else {
        code = TableCellStyle.Render(code)
        name = TableCellStyle.Render(name)
        vat = TableCellStyle.Render(vat)
        city = TableCellStyle.Render(city)
        credit = TableCellStyle.Render(credit)
    }

    statusBadge := BadgeSuccessStyle.Render(creditStatus)
    if creditStatus == "Attenzione" {
        statusBadge = BadgeWarningStyle.Render(creditStatus)
    } else if creditStatus == "Bloccato" {
        statusBadge = BadgeDangerStyle.Render(creditStatus)
    }

    return lipgloss.JoinHorizontal(
        lipgloss.Left,
        code,
        name,
        vat,
        city,
        credit,
        statusBadge,
    )
}

func (m *AppModel) updateCustomerSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up":
            if m.customerSearchView.selectedIndex > 0 {
                m.customerSearchView.selectedIndex--
            }
            return m, nil

        case "down":
            if m.customerSearchView.selectedIndex < len(m.customerSearchView.results)-1 {
                m.customerSearchView.selectedIndex++
            }
            return m, nil

        case "enter":
            if len(m.customerSearchView.results) > 0 {
                customer := m.customerSearchView.results[m.customerSearchView.selectedIndex]
                m.setMessage(fmt.Sprintf("Cliente: %s", customer.CompanyName))
                return m, nil
            }
            return m, nil

        case "backspace":
            if len(m.customerSearchView.query) > 0 {
                m.customerSearchView.query = m.customerSearchView.query[:len(m.customerSearchView.query)-1]
                if len(m.customerSearchView.query) >= 2 {
                    m.customerSearchView.loading = true
                    return m, m.performCustomerSearch()
                } else {
                    m.customerSearchView.results = []*domain.Customer{}
                }
            }
            return m, nil

        case "n":
            m.setMessage("Funzione: nuovo cliente")
            return m, nil

        case "esc":
            return m.navigateBack(), nil

        default:
            if len(msg.String()) == 1 {
                m.customerSearchView.query += msg.String()
                if len(m.customerSearchView.query) >= 2 {
                    m.customerSearchView.loading = true
                    return m, m.performCustomerSearch()
                }
            }
            return m, nil
        }
    }

    return m, nil
}
