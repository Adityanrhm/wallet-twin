package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/app"
	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
	"github.com/Adityanrhm/wallet-twin/internal/repository/postgres"
	"github.com/Adityanrhm/wallet-twin/internal/service"
)

// Tab represents the current active tab
type Tab int

const (
	TabOverview Tab = iota
	TabWallets
	TabTransactions
	TabBudgets
	TabGoals
)

func (t Tab) String() string {
	return []string{"📊 Overview", "💼 Wallets", "📝 Transactions", "📊 Budgets", "🎯 Goals"}[t]
}

// DashboardModel adalah state utama untuk TUI dashboard.
type DashboardModel struct {
	app       *app.App
	activeTab Tab
	width     int
	height    int

	// Data
	wallets          []*models.Wallet
	totalBalance     decimal.Decimal
	recentTxs        []*models.Transaction
	monthlySummary   *repository.TransactionSummary
	budgetStatuses   []*repository.BudgetStatus
	goals            []*models.Goal

	// Loading state
	loading bool
	err     error
}

// NewDashboard membuat dashboard model baru.
func NewDashboard(application *app.App) *DashboardModel {
	return &DashboardModel{
		app:       application,
		activeTab: TabOverview,
		width:     80,
		height:    24,
		loading:   true,
	}
}

// Init adalah Bubble Tea lifecycle method.
func (m *DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadData,
		tea.SetWindowTitle("💰 Wallet Twin Dashboard"),
	)
}

// Message types
type dataLoadedMsg struct {
	wallets        []*models.Wallet
	totalBalance   decimal.Decimal
	recentTxs      []*models.Transaction
	summary        *repository.TransactionSummary
	budgetStatuses []*repository.BudgetStatus
	goals          []*models.Goal
}

type errMsg struct{ err error }

// loadData mengambil semua data yang diperlukan.
func (m *DashboardModel) loadData() tea.Msg {
	ctx := context.Background()

	txManager := postgres.NewTransactionManager(m.app.DB.Pool)

	// Services
	walletSvc := service.NewWalletService(m.app.Repos.Wallet)
	txSvc := service.NewTransactionService(m.app.Repos.Transaction, m.app.Repos.Wallet, txManager)
	budgetSvc := service.NewBudgetService(m.app.Repos.Budget, m.app.Repos.Transaction)
	goalSvc := service.NewGoalService(m.app.Repos.Goal)

	// Get wallets
	wallets, err := walletSvc.ListActive(ctx)
	if err != nil {
		return errMsg{err}
	}

	// Get total balance
	totalBalance, err := walletSvc.GetTotalBalance(ctx)
	if err != nil {
		return errMsg{err}
	}

	// Get recent transactions
	recentTxs, err := txSvc.GetRecent(ctx, 5)
	if err != nil {
		return errMsg{err}
	}

	// Get monthly summary
	now := time.Now()
	summary, err := txSvc.GetMonthlySummary(ctx, now.Year(), now.Month())
	if err != nil {
		return errMsg{err}
	}

	// Get budget statuses
	budgetStatuses, err := budgetSvc.GetAllStatus(ctx)
	if err != nil {
		// Non-critical, continue
		budgetStatuses = nil
	}

	// Get goals
	goals, err := goalSvc.ListActive(ctx)
	if err != nil {
		// Non-critical, continue
		goals = nil
	}

	return dataLoadedMsg{
		wallets:        wallets,
		totalBalance:   totalBalance,
		recentTxs:      recentTxs,
		summary:        summary,
		budgetStatuses: budgetStatuses,
		goals:          goals,
	}
}

// Update handles messages (Elm Architecture).
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left", "h":
			if m.activeTab > TabOverview {
				m.activeTab--
			}
		case "right", "l":
			if m.activeTab < TabGoals {
				m.activeTab++
			}
		case "r":
			m.loading = true
			return m, m.loadData
		case "1":
			m.activeTab = TabOverview
		case "2":
			m.activeTab = TabWallets
		case "3":
			m.activeTab = TabTransactions
		case "4":
			m.activeTab = TabBudgets
		case "5":
			m.activeTab = TabGoals
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dataLoadedMsg:
		m.loading = false
		m.wallets = msg.wallets
		m.totalBalance = msg.totalBalance
		m.recentTxs = msg.recentTxs
		m.monthlySummary = msg.summary
		m.budgetStatuses = msg.budgetStatuses
		m.goals = msg.goals

	case errMsg:
		m.loading = false
		m.err = msg.err
	}

	return m, nil
}

// View renders the UI (Elm Architecture).
func (m *DashboardModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	// Build layout
	header := m.renderHeader()
	tabs := m.renderTabs()
	content := m.renderContent()
	help := m.renderHelp()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		help,
	)
}

func (m *DashboardModel) renderLoading() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(primaryColor).Render("⏳ Loading..."),
	)
}

func (m *DashboardModel) renderError() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(dangerColor).Render("❌ Error: "+m.err.Error()),
	)
}

func (m *DashboardModel) renderHeader() string {
	title := "💰 Wallet Twin Dashboard"
	return headerStyle.Render(title)
}

func (m *DashboardModel) renderTabs() string {
	tabs := []Tab{TabOverview, TabWallets, TabTransactions, TabBudgets, TabGoals}
	var renderedTabs []string

	for _, tab := range tabs {
		style := inactiveTabStyle
		if tab == m.activeTab {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(tab.String()))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m *DashboardModel) renderContent() string {
	switch m.activeTab {
	case TabOverview:
		return m.renderOverview()
	case TabWallets:
		return m.renderWallets()
	case TabTransactions:
		return m.renderTransactions()
	case TabBudgets:
		return m.renderBudgets()
	case TabGoals:
		return m.renderGoals()
	default:
		return ""
	}
}

func (m *DashboardModel) renderOverview() string {
	// Total Balance Card
	balanceCard := cardStyle.Render(
		cardTitleStyle.Render("💰 Total Balance") + "\n\n" +
			moneyStyle.Render(formatMoney(m.totalBalance)),
	)

	// Monthly Summary Card
	var summaryContent string
	if m.monthlySummary != nil {
		summaryContent = fmt.Sprintf(
			"%s\n%s\n%s",
			incomeStyle.Render("📈 Income:  "+formatMoney(m.monthlySummary.TotalIncome)),
			expenseStyle.Render("📉 Expense: "+formatMoney(m.monthlySummary.TotalExpense)),
			moneyStyle.Render("💵 Net:     "+formatMoney(m.monthlySummary.Net)),
		)
	} else {
		summaryContent = "No data"
	}

	summaryCard := cardStyle.Render(
		cardTitleStyle.Render("📊 This Month") + "\n\n" + summaryContent,
	)

	// Goals Preview
	var goalsContent string
	if len(m.goals) > 0 {
		for i, g := range m.goals {
			if i >= 3 { // Show max 3
				break
			}
			progress := g.GetProgress()
			bar := renderProgressBar(progress, 20)
			goalsContent += fmt.Sprintf("%s %s %.0f%%\n", g.Icon, g.Name, progress)
			goalsContent += bar + "\n\n"
		}
	} else {
		goalsContent = "No active goals"
	}

	goalsCard := cardStyle.Render(
		cardTitleStyle.Render("🎯 Goals Progress") + "\n\n" + goalsContent,
	)

	return lipgloss.JoinVertical(lipgloss.Left, balanceCard, summaryCard, goalsCard)
}

func (m *DashboardModel) renderWallets() string {
	if len(m.wallets) == 0 {
		return cardStyle.Render("No wallets found. Add one with: wallet wallet add")
	}

	var content string
	for _, w := range m.wallets {
		status := "✅"
		if !w.IsActive {
			status = "❌"
		}
		content += fmt.Sprintf("%s %s %s\n   %s %s\n\n",
			w.Icon, w.Name, status,
			w.Currency, moneyStyle.Render(formatMoney(w.Balance)),
		)
	}

	return cardStyle.Render(
		cardTitleStyle.Render("💼 Your Wallets") + "\n\n" + content,
	)
}

func (m *DashboardModel) renderTransactions() string {
	if len(m.recentTxs) == 0 {
		return cardStyle.Render("No recent transactions")
	}

	var content string
	for _, tx := range m.recentTxs {
		icon := "📈"
		if tx.Type == models.TransactionTypeExpense {
			icon = "📉"
		}
		content += fmt.Sprintf("%s %s | %s\n   %s\n\n",
			icon,
			tx.TransactionDate.Format("02 Jan"),
			formatMoney(tx.Amount),
			truncate(tx.Description, 40),
		)
	}

	return cardStyle.Render(
		cardTitleStyle.Render("📝 Recent Transactions") + "\n\n" + content,
	)
}

func (m *DashboardModel) renderBudgets() string {
	if len(m.budgetStatuses) == 0 {
		return cardStyle.Render("No active budgets")
	}

	var content string
	for _, s := range m.budgetStatuses {
		bar := renderProgressBar(s.Progress, 20)
		status := ""
		if s.IsOverBudget {
			status = " ⚠️ OVER"
		}

		content += fmt.Sprintf("%s %s%s\n", s.CategoryIcon, s.CategoryName, status)
		content += fmt.Sprintf("%s %.0f%%\n", bar, s.Progress)
		content += fmt.Sprintf("Spent: %s / %s\n\n",
			formatMoney(s.Spent), formatMoney(s.Budget.Amount))
	}

	return cardStyle.Render(
		cardTitleStyle.Render("📊 Budget Status") + "\n\n" + content,
	)
}

func (m *DashboardModel) renderGoals() string {
	if len(m.goals) == 0 {
		return cardStyle.Render("No active goals. Add one with: wallet goal add")
	}

	var content string
	for _, g := range m.goals {
		progress := g.GetProgress()
		bar := renderProgressBar(progress, 25)

		status := "🔄 In Progress"
		if g.IsCompleted() {
			status = "✅ Completed!"
		}

		content += fmt.Sprintf("%s %s\n", g.Icon, g.Name)
		content += fmt.Sprintf("%s %.1f%%\n", bar, progress)
		content += fmt.Sprintf("%s / %s | %s\n\n",
			formatMoney(g.CurrentAmount),
			formatMoney(g.TargetAmount),
			status,
		)
	}

	return cardStyle.Render(
		cardTitleStyle.Render("🎯 Savings Goals") + "\n\n" + content,
	)
}

func (m *DashboardModel) renderHelp() string {
	return helpStyle.Render("← → Navigate | 1-5 Jump | r Refresh | q Quit")
}

// Helper functions
func formatMoney(d decimal.Decimal) string {
	return "Rp " + d.StringFixed(0)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
