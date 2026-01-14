// internal/ui/app.go

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
)

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

	loginView    *LoginView
	mainMenuView *MainMenuView
	searchView   *ArticleSearchView

	error   string
	message string
	loading bool
	quit    bool

	lastActivity   time.Time
	sessionTimeout time.Duration
	quitCh         chan struct{}
}

type LoginView struct {
	username   string
	password   string
	focusIndex int
	error      string
}

type MainMenuView struct {
	selectedIndex int
	menuItems     []MenuItem
}

type MenuItem struct {
	Label       string
	Description string
	View        ViewState
	Enabled     bool
}

type ArticleSearchView struct {
	query         string
	searchType    string
	results       []*domain.Article
	selectedIndex int
	loading       bool
	scrollOffset  int
}

type loginResultMsg struct {
	operator *domain.Operator
	err      error
}

type searchResultMsg struct {
	results []*domain.Article
	err     error
}

type tickMsg struct {
	time.Time
}

type sessionExpiredMsg struct{}

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
		sessionTimeout: 480 * time.Minute,
		lastActivity:   time.Now(),
		quitCh:         make(chan struct{}),
	}
}

func (m *AppModel) Init() tea.Cmd {
	return m.tickCmd()
}

func (m *AppModel) tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{t}
	})
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quit {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.operator != nil && time.Since(m.lastActivity) > m.sessionTimeout {
			m.setError("Sessione scaduta. Effettua nuovamente il login.")
			m.operator = nil
			m.currentView = ViewLogin
			m.viewStack = []ViewState{}
			m.loginView = &LoginView{}
		}
		return m, m.tickCmd()

	case loginResultMsg:
		return m.handleLoginResult(msg)

	case searchResultMsg:
		return m.handleSearchResult(msg)

	case sessionExpiredMsg:
		m.setError("Sessione scaduta per inattività.")
		m.operator = nil
		m.currentView = ViewLogin
		m.viewStack = []ViewState{}
		m.loginView = &LoginView{}
		return m, nil

	case tea.KeyMsg:
		m.lastActivity = time.Now()
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if m.currentView == ViewLogin || m.currentView == ViewMainMenu {
				return m, tea.Quit
			}
			return m.navigateBack(), nil

		case "esc":
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
	default:
		return m, nil
	}
}

func (m *AppModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	var content string
	switch m.currentView {
	case ViewLogin:
		content = m.viewLogin()
	case ViewMainMenu:
		content = m.viewMainMenu()
	case ViewArticleSearch:
		content = m.viewArticleSearch()
	default:
		content = "View not implemented"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

func (m *AppModel) renderHeader() string {
	title := "🚗 Ricambi Manager"
	if m.operator != nil {
		title += fmt.Sprintf(" • %s", m.operator.FullName)
	}

	now := time.Now()
	datetime := now.Format("15:04:05 • 02/01/2006")

	leftPart := TitleStyle.Render(title)
	rightPart := DateTimeStyle.Render(datetime)

	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(rightPart)
	spacer := m.width - leftWidth - rightWidth - 2
	if spacer < 0 {
		spacer = 0
	}

	headerLine := leftPart + lipgloss.NewStyle().Width(spacer).Render("") + rightPart

	separator := lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(lipgloss.NewStyle().Width(m.width).Render("─"))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		separator,
	)
}

func (m *AppModel) renderFooter() string {
	var sections []string

	breadcrumb := m.renderBreadcrumb()
	if breadcrumb != "" {
		sections = append(sections, breadcrumb)
	}

	status := m.renderStatus()
	if status != "" {
		sections = append(sections, status)
	}

	help := m.renderHelp()
	if help != "" {
		sections = append(sections, help)
	}

	if len(sections) == 0 {
		return ""
	}

	footer := lipgloss.JoinVertical(lipgloss.Left, sections...)

	separator := lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(lipgloss.NewStyle().Width(m.width).Render("─"))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		separator,
		footer,
	)
}

func (m *AppModel) renderBreadcrumb() string {
	if m.currentView == ViewLogin {
		return ""
	}

	var parts []string
	if len(m.viewStack) > 0 {
		parts = append(parts, "Home")
		for _, view := range m.viewStack {
			parts = append(parts, m.getViewName(view))
		}
	}
	parts = append(parts, m.getViewName(m.currentView))

	breadcrumb := ""
	for i, part := range parts {
		if i > 0 {
			breadcrumb += " › "
		}
		breadcrumb += part
	}

	return BreadcrumbStyle.Render(breadcrumb)
}

func (m *AppModel) renderStatus() string {
	if m.loading {
		return InfoStyle.Render("⏳ Caricamento in corso...")
	}
	if m.message != "" {
		return SuccessStyle.Render("✓ " + m.message)
	}
	if m.error != "" {
		return ErrorStyle.Render("⚠ " + m.error)
	}
	return ""
}

func (m *AppModel) renderHelp() string {
	help := ""
	switch m.currentView {
	case ViewLogin:
		help = "tab: campo successivo • enter: login • ctrl+c: esci"
	case ViewMainMenu:
		help = "1-7: selezione rapida • ↑/↓/j/k: naviga • enter: conferma • q: esci"
	case ViewArticleSearch:
		help = "tab: tipo ricerca • digita: cerca • ↑/↓/j/k: naviga • pgup/pgdwn: pagina • home/end: inizio/fine • enter: seleziona • esc: indietro"
	default:
		help = "esc: indietro • q: esci"
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
	case ViewCustomerSearch:
		return "Ricerca Clienti"
	case ViewPromotions:
		return "Promozioni"
	case ViewCreditVouchers:
		return "Buoni Credito"
	case ViewBudgets:
		return "Budget"
	case ViewKits:
		return "Kit"
	default:
		return "Unknown"
	}
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
		{Label: "🔍 Ricerca Articoli", Description: "Cerca e gestisci articoli", View: ViewArticleSearch, Enabled: true},
		{Label: "👥 Gestione Clienti", Description: "Gestisci anagrafica clienti", View: ViewCustomerSearch, Enabled: true},
		{Label: "🎁 Promozioni", Description: "Gestisci promozioni attive", View: ViewPromotions, Enabled: true},
		{Label: "💰 Buoni Credito", Description: "Gestisci buoni a credito", View: ViewCreditVouchers, Enabled: true},
		{Label: "📊 Budget", Description: "Monitora obiettivi di vendita", View: ViewBudgets, Enabled: true},
		{Label: "📦 Kit", Description: "Gestisci kit di vendita", View: ViewKits, Enabled: true},
		{Label: "⚙️  Impostazioni", Description: "Configurazione sistema", View: ViewSettings, Enabled: m.operator.IsAdmin()},
	}
}

func (m *AppModel) viewLogin() string {
	title := TitleStyle.Render("🔐 Login - Ricambi Manager")

	usernameLabel := "Username:"
	passwordLabel := "Password:"

	usernameField := m.loginView.username
	if m.loginView.focusIndex == 0 {
		usernameField = InputFocusedStyle.Render(usernameField + "█")
	} else {
		usernameField = InputStyle.Render(usernameField)
	}

	passwordField := ""
	for range m.loginView.password {
		passwordField += "*"
	}
	if m.loginView.focusIndex == 1 {
		passwordField = InputFocusedStyle.Render(passwordField + "█")
	} else {
		passwordField = InputStyle.Render(passwordField)
	}

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		usernameLabel,
		usernameField,
		"",
		passwordLabel,
		passwordField,
		"",
		ButtonStyle.Render("[ Login ]"),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		form,
	)

	if m.loginView.error != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			content,
			"",
			ErrorStyle.Render("❌ "+m.loginView.error),
		)
	}

	availableHeight := m.height - 6

	return lipgloss.Place(
		m.width,
		availableHeight,
		lipgloss.Center,
		lipgloss.Center,
		ContentStyle.Render(content),
	)
}

func (m *AppModel) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.loginView.focusIndex = (m.loginView.focusIndex + 1) % 2
			return m, nil

		case "shift+tab", "up":
			m.loginView.focusIndex--
			if m.loginView.focusIndex < 0 {
				m.loginView.focusIndex = 1
			}
			return m, nil

		case "enter":
			return m, m.performLogin()

		case "backspace":
			if m.loginView.focusIndex == 0 && len(m.loginView.username) > 0 {
				m.loginView.username = m.loginView.username[:len(m.loginView.username)-1]
			} else if m.loginView.focusIndex == 1 && len(m.loginView.password) > 0 {
				m.loginView.password = m.loginView.password[:len(m.loginView.password)-1]
			}
			return m, nil

		default:
			if len(msg.String()) == 1 {
				if m.loginView.focusIndex == 0 {
					m.loginView.username += msg.String()
				} else {
					m.loginView.password += msg.String()
				}
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *AppModel) performLogin() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		if m.loginView.username == "" {
			return loginResultMsg{err: fmt.Errorf("inserire username")}
		}

		if m.loginView.password == "" {
			return loginResultMsg{err: fmt.Errorf("inserire password")}
		}

		operator, err := m.operatorRepo.FindByUsername(ctx, m.loginView.username)
		if err != nil {
			return loginResultMsg{err: fmt.Errorf("credenziali non valide")}
		}

		if err := operator.CheckPassword(m.loginView.password); err != nil {
			if updateErr := m.operatorRepo.Update(ctx, operator); updateErr != nil {
				return loginResultMsg{err: fmt.Errorf("errore di sistema: %w", updateErr)}
			}
			return loginResultMsg{err: fmt.Errorf("credenziali non valide")}
		}

		session, err := m.authService.CreateSession(operator, "localhost", "TUI")
		if err != nil {
			return loginResultMsg{err: fmt.Errorf("errore di creazione sessione: %w", err)}
		}

		operator.SessionToken = session.Token
		if err := m.operatorRepo.Update(ctx, operator); err != nil {
			return loginResultMsg{err: fmt.Errorf("errore di aggiornamento sessione: %w", err)}
		}

		return loginResultMsg{operator: operator}
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
	m.searchView.loading = false

	if msg.err != nil {
		m.setError("Errore durante la ricerca: " + msg.err.Error())
		m.searchView.results = []*domain.Article{}
		m.searchView.selectedIndex = 0
		m.searchView.scrollOffset = 0
		return m, nil
	}

	m.clearMessages()
	m.searchView.results = msg.results
	m.searchView.selectedIndex = 0
	m.searchView.scrollOffset = 0

	return m, nil
}
