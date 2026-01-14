// internal/ui/app.go
// Modernized Professional Application - Ricambi Manager

package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/mongo"

	"ricambi-manager/internal/domain"
	"ricambi-manager/internal/repository"
	"ricambi-manager/internal/usecase"
	"ricambi-manager/pkg/auth"
)

// ViewState defines all possible views in the application
type ViewState int

const (
	ViewLogin ViewState = iota
	ViewMainMenu
	ViewArticleSearch
	ViewArticleList
	ViewArticleDetail
	ViewCustomerSearch
	ViewCustomerList
	ViewCustomerDetail
	ViewPromotions
	ViewCreditVouchers
	ViewBudgets
	ViewKits
	ViewSettings
	ViewHelp
)

// AppModel is the main application model
type AppModel struct {
	db          *mongo.Database
	width       int
	height      int
	currentView ViewState
	viewStack   []ViewState

	operator    *domain.Operator
	authService *auth.AuthService

	articleRepo   *repository.ArticleRepository
	customerRepo  *repository.CustomerRepository
	operatorRepo  *repository.OperatorRepository
	promotionRepo *repository.PromotionRepository
	voucherRepo   *repository.CreditVoucherRepository
	budgetRepo    *repository.BudgetRepository
	kitRepo       *repository.KitRepository

	searchUC   *usecase.SearchArticlesUseCase
	discountUC *usecase.ManageDiscountsUseCase
	stockUC    *usecase.ManageStockUseCase

	loginView       *LoginView
	mainMenuView    *MainMenuView
	searchView      *ArticleSearchView
	customerSearchView *CustomerSearchView
	promotionsView  *PromotionsView
	vouchersView    *VouchersView
	budgetsView     *BudgetsView
	kitsView        *KitsView
	settingsView    *SettingsView

	error      string
	message    string
	loading    bool
	lastActivity time.Time
	sessionTimeout time.Duration
}

// LoginView handles user authentication
type LoginView struct {
	username    string
	password    string
	focusIndex  int
	error       string
	showPassword bool
}

// MainMenuView displays the main navigation menu
type MainMenuView struct {
	selectedIndex int
	menuItems     []MenuItem
	stats         DashboardStats
}

// DashboardStats shows key metrics on main screen
type DashboardStats struct {
	totalArticles   int
	lowStockItems   int
	activeCustomers int
	pendingOrders   int
}

// MenuItem represents a menu option
type MenuItem struct {
	Label       string
	Description string
	Shortcut    string
	View        ViewState
	Enabled     bool
	Icon        string
}

// ArticleSearchView handles article search functionality
type ArticleSearchView struct {
	query          string
	searchType     string
	results        []*domain.Article
	selectedIndex  int
	loading        bool
	searchTypes    []string
	filtersActive  bool
}

// CustomerSearchView handles customer search
type CustomerSearchView struct {
	query          string
	results        []*domain.Customer
	selectedIndex  int
	loading        bool
}

// PromotionsView shows active promotions
type PromotionsView struct {
	promotions   []*domain.Promotion
	selectedIndex int
	loading      bool
}

// VouchersView handles credit vouchers
type VouchersView struct {
	vouchers     []*domain.CreditVoucher
	selectedIndex int
	loading      bool
}

// BudgetsView shows budget information
type BudgetsView struct {
	budgets     []*domain.Budget
	selectedIndex int
	loading     bool
}

// KitsView manages kit operations
type KitsView struct {
	kits        []*domain.Kit
	selectedIndex int
	loading     bool
}

// SettingsView handles system configuration
type SettingsView struct {
	activeTab   int
}

// Message types
type loginResultMsg struct {
	operator *domain.Operator
	err      error
}

type searchResultMsg struct {
	results []*domain.Article
	err     error
}

type customerSearchResultMsg struct {
	results []*domain.Customer
	err     error
}

func NewAppModel(db *mongo.Database) *AppModel {
	articleRepo := repository.NewArticleRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	operatorRepo := repository.NewOperatorRepository(db)
	promotionRepo := repository.NewPromotionRepository(db)
	voucherRepo := repository.NewCreditVoucherRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)
	kitRepo := repository.NewKitRepository(db)

	return &AppModel{
		db:             db,
		currentView:    ViewLogin,
		viewStack:      []ViewState{},
		authService:    auth.NewAuthService(480),
		articleRepo:    articleRepo,
		customerRepo:   customerRepo,
		operatorRepo:   operatorRepo,
		promotionRepo:  promotionRepo,
		voucherRepo:    voucherRepo,
		budgetRepo:     budgetRepo,
		kitRepo:        kitRepo,
		searchUC:       usecase.NewSearchArticlesUseCase(articleRepo),
		discountUC:     usecase.NewManageDiscountsUseCase(customerRepo, articleRepo, promotionRepo),
		stockUC:        usecase.NewManageStockUseCase(articleRepo, kitRepo),
		loginView:      &LoginView{},
		mainMenuView:   &MainMenuView{selectedIndex: 0},
		searchView:     &ArticleSearchView{},
		customerSearchView: &CustomerSearchView{},
		promotionsView:  &PromotionsView{},
		vouchersView:    &VouchersView{},
		budgetsView:     &BudgetsView{},
		kitsView:        &KitsView{},
		settingsView:    &SettingsView{},
		sessionTimeout: 480 * time.Minute,
		lastActivity:   time.Now(),
	}
}

func (m *AppModel) Init() tea.Cmd {
	return nil
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.lastActivity = time.Now()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case loginResultMsg:
		return m.handleLoginResult(msg)

	case searchResultMsg:
		return m.handleSearchResult(msg)

	case customerSearchResultMsg:
		return m.handleCustomerSearchResult(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == ViewLogin || m.currentView == ViewMainMenu {
				return m, tea.Quit
			}
			return m.navigateBack(), nil

		case "esc":
			return m.navigateBack(), nil

		case "?":
			if m.currentView != ViewHelp {
				return m.navigateTo(ViewHelp), nil
			}
			return m.navigateBack(), nil
		}
	}

	switch m.currentView {
	case ViewLogin:
		return m.updateLogin(msg)
	case ViewMainMenu:
		return m.updateMainMenu(msg)
	case ViewArticleSearch:
		return m.updateArticleSearch(msg)
	case ViewCustomerSearch:
		return m.updateCustomerSearch(msg)
	case ViewPromotions:
		return m.updatePromotions(msg)
	case ViewCreditVouchers:
		return m.updateVouchers(msg)
	case ViewBudgets:
		return m.updateBudgets(msg)
	case ViewKits:
		return m.updateKits(msg)
	case ViewSettings:
		return m.updateSettings(msg)
	default:
		return m, nil
	}
}

func (m *AppModel) View() string {
	if m.width == 0 {
		return "Inizializzazione..."
	}

	header := m.renderHeader()
	sidebar := m.renderSidebar()
	statusBar := m.renderStatusBar()
	help := m.renderHelp()

	var content string
	switch m.currentView {
	case ViewLogin:
		content = m.viewLogin()
	case ViewMainMenu:
		content = m.viewMainMenu()
	case ViewArticleSearch:
		content = m.viewArticleSearch()
	case ViewCustomerSearch:
		content = m.viewCustomerSearch()
	case ViewPromotions:
		content = m.viewPromotions()
	case ViewCreditVouchers:
		content = m.viewVouchers()
	case ViewBudgets:
		content = m.viewBudgets()
	case ViewKits:
		content = m.viewKits()
	case ViewSettings:
		content = m.viewSettings()
	case ViewHelp:
		content = m.viewHelp()
	default:
		content = "Vista non implementata"
	}

	mainArea := lipgloss.JoinHorizontal(
		lipgloss.Left,
		sidebar,
		lipgloss.NewStyle().Width(m.width - lipgloss.Width(sidebar)).Render(content),
	)

	footerArea := lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		help,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainArea,
		footerArea,
	)
}

func (m *AppModel) renderHeader() string {
	now := time.Now()
	datetime := now.Format("15:04 • 02/01/2006")

	var title string
	if m.operator != nil {
		title = fmt.Sprintf(" %s %s", m.operator.FullName, m.getRoleBadge())
	} else {
		title = " Ricambi Manager"
	}

	titleStyled := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg).
		Render(title)

	datetimeStyled := DateTimeStyle.Render(datetime)

	leftWidth := lipgloss.Width(titleStyled)
	rightWidth := lipgloss.Width(datetimeStyled)
	spacer := m.width - leftWidth - rightWidth - 2
	if spacer < 0 {
		spacer = 0
	}

	headerLine := titleStyled + lipgloss.NewStyle().Width(spacer).Render("") + datetimeStyled

	separator := lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(lipgloss.NewStyle().Width(m.width).Render("━"))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		separator,
	)
}

func (m *AppModel) renderSidebar() string {
	if m.currentView == ViewLogin {
		return ""
	}

	items := []string{}
	for i, item := range m.mainMenuView.menuItems {
		if !item.Enabled {
			continue
		}

		shortcut := ShortcutKeyStyle.Render(item.Shortcut)
		icon := lipgloss.NewStyle().Foreground(ColorMuted).Render(item.Icon)
		label := item.Label

		if i == m.mainMenuView.selectedIndex && m.currentView == ViewMainMenu {
			row := lipgloss.JoinHorizontal(
				lipgloss.Left,
				shortcut,
				" ",
				lipgloss.NewStyle().Foreground(ColorPrimary).Render(icon),
				" ",
				lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(label),
			)
			items = append(items, SelectedItemStyle.Render(row))
		} else if m.currentView == item.View {
			row := lipgloss.JoinHorizontal(
				lipgloss.Left,
				shortcut,
				" ",
				icon,
				" ",
				lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(label),
			)
			items = append(items, SelectedItemStyle.Render(row))
		} else {
			row := lipgloss.JoinHorizontal(
				lipgloss.Left,
				shortcut,
				" ",
				icon,
				" ",
				UnselectedItemStyle.Render(label),
			)
			items = append(items, row)
		}
	}

	menuContent := lipgloss.JoinVertical(lipgloss.Left, items...)

	sidebarWidth := 24
	return lipgloss.NewStyle().
		Width(sidebarWidth).
		BorderRight(true).
		BorderForeground(ColorBorder).
		Padding(1, 0).
		Render(menuContent)
}

func (m *AppModel) renderStatusBar() string {
	if m.currentView == ViewLogin {
		return ""
	}

	var leftItems []string
	var rightItems []string

	// Session info
	if m.operator != nil {
		leftItems = append(leftItems, StatusBarItemStyle.Render("👤 "+m.operator.Username))
	}

	// View context info
	rightItems = append(rightItems, StatusBarItemStyle.Render(m.getViewName(m.currentView)))

	// Loading indicator
	if m.loading {
		rightItems = append(rightItems, StatusBarItemStyle.Render("⏳"))
	}

	leftContent := lipgloss.JoinHorizontal(lipgloss.Left, leftItems...)
	rightContent := lipgloss.JoinHorizontal(lipgloss.Right, rightItems...)

	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(rightContent)
	spacer := m.width - leftWidth - rightWidth - 2
	if spacer < 0 {
		spacer = 0
	}

	return StatusBarStyle.Render(leftContent + lipgloss.NewStyle().Width(spacer).Render("") + rightContent)
}

func (m *AppModel) renderHelp() string {
	if m.currentView == ViewLogin {
		return HelpStyle.Render(" TAB: campo successivo  •  ENTER: login  •  CTRL+C: esci")
	}

	var help string
	switch m.currentView {
	case ViewMainMenu:
		help = " ↑/↓: naviga  •  ENTER: seleziona  •  [1-7]: accesso rapido  •  ?: aiuto  •  Q: esci"
	case ViewArticleSearch:
		help = " TAB: tipo ricerca  •  digita: cerca  •  ↑/↓: naviga  •  ENTER: dettagli  •  ESC: indietro"
	case ViewCustomerSearch:
		help = " digita: cerca  •  ↑/↓: naviga  •  ENTER: dettagli cliente  •  ESC: indietro"
	case ViewPromotions:
		help = " ↑/↓: naviga  •  ENTER: dettagli  •  N: nuova promozione  •  ESC: indietro"
	case ViewCreditVouchers:
		help = " ↑/↓: naviga  •  N: nuovo buono  •  ENTER: utilizza  •  ESC: indietro"
	case ViewBudgets:
		help = " ↑/↓: naviga  •  ENTER: dettagli  •  ESC: indietro"
	case ViewKits:
		help = " ↑/↓: naviga  •  N: nuovo kit  •  ENTER: modifica  •  ESC: indietro"
	case ViewSettings:
		help = " ↑/↓: naviga  •  TAB: schede  •  ENTER: modifica  •  ESC: indietro"
	case ViewHelp:
		help = " ESC: chiudi aiuto"
	default:
		help = " ESC: indietro  •  Q: esci"
	}

	return HelpStyle.Render(help)
}

func (m *AppModel) getViewName(view ViewState) string {
	switch view {
	case ViewLogin:
		return "Login"
	case ViewMainMenu:
		return "Menu Principale"
	case ViewArticleSearch:
		return "Ricerca Articoli"
	case ViewArticleList:
		return "Lista Articoli"
	case ViewArticleDetail:
		return "Dettaglio Articolo"
	case ViewCustomerSearch:
		return "Ricerca Clienti"
	case ViewCustomerList:
		return "Lista Clienti"
	case ViewCustomerDetail:
		return "Dettaglio Cliente"
	case ViewPromotions:
		return "Promozioni"
	case ViewCreditVouchers:
		return "Buoni Credito"
	case ViewBudgets:
		return "Budget"
	case ViewKits:
		return "Kit"
	case ViewSettings:
		return "Impostazioni"
	case ViewHelp:
		return "Aiuto"
	default:
		return "Unknown"
	}
}

func (m *AppModel) getRoleBadge() string {
	if m.operator == nil {
		return ""
	}
	if m.operator.IsAdmin() {
		return BadgePrimaryStyle.Render("ADMIN")
	}
	return BadgeInfoStyle.Render("OPERATORE")
}

func (m *AppModel) navigateTo(view ViewState) *AppModel {
	m.viewStack = append(m.viewStack, m.currentView)
	m.currentView = view
	return m
}

func (m *AppModel) navigateBack() *AppModel {
	if len(m.viewStack) > 0 {
		m.currentView = m.viewStack[len(m.viewStack)-1]
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
	}
	return m
}

func (m *AppModel) setError(err string) {
	m.error = err
	m.message = ""
}

func (m *AppModel) setMessage(msg string) {
	m.message = msg
	m.error = ""
}

func (m *AppModel) clearMessages() {
	m.error = ""
	m.message = ""
}

func (m *AppModel) initMainMenu() {
	m.mainMenuView.menuItems = []MenuItem{
		{"Ricerca Articoli", "Cerca e gestisci articoli", "1", ViewArticleSearch, true, "🔍"},
		{"Clienti", "Gestisci anagrafica clienti", "2", ViewCustomerSearch, true, "👥"},
		{"Promozioni", "Gestisci promozioni attive", "3", ViewPromotions, true, "🎁"},
		{"Buoni Credito", "Gestisci buoni a credito", "4", ViewCreditVouchers, true, "💰"},
		{"Budget", "Monitora obiettivi vendite", "5", ViewBudgets, true, "📊"},
		{"Kit", "Gestisci kit di vendita", "6", ViewKits, true, "📦"},
		{"Impostazioni", "Configurazione sistema", "7", ViewSettings, m.operator.IsAdmin(), "⚙️"},
	}
}

func (m *AppModel) handleLoginResult(msg loginResultMsg) (*AppModel, tea.Cmd) {
	if msg.err != nil {
		m.loginView.error = msg.err.Error()
		m.loginView.password = ""
		return m, nil
	}

	m.operator = msg.operator
	m.loginView.username = ""
	m.loginView.password = ""
	m.loginView.error = ""
	m.currentView = ViewMainMenu
	m.initMainMenu()

	return m, nil
}

func (m *AppModel) handleSearchResult(msg searchResultMsg) (*AppModel, tea.Cmd) {
	if msg.err != nil {
		m.setError(msg.err.Error())
	} else {
		m.searchView.results = msg.results
		m.searchView.loading = false
		m.searchView.selectedIndex = 0
	}
	return m, nil
}

func (m *AppModel) handleCustomerSearchResult(msg customerSearchResultMsg) (*AppModel, tea.Cmd) {
	if msg.err != nil {
		m.setError(msg.err.Error())
	} else {
		m.customerSearchView.results = msg.results
		m.customerSearchView.loading = false
		m.customerSearchView.selectedIndex = 0
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

func (m *AppModel) performCustomerSearch() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var results []*domain.Customer
		var err error

		if len(m.customerSearchView.query) >= 2 {
			results, err = m.customerRepo.Search(ctx, m.customerSearchView.query, 20)
		}

		return customerSearchResultMsg{results: results, err: err}
	}
}
