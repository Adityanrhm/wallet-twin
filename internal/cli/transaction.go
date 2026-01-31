package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
	"github.com/Adityanrhm/wallet-twin/internal/repository/postgres"
	"github.com/Adityanrhm/wallet-twin/internal/service"
)

// transactionCmd adalah parent command untuk transactions.
var transactionCmd = &cobra.Command{
	Use:     "transaction",
	Aliases: []string{"tx", "t"},
	Short:   "📝 Manage transactions",
	Long:    "Add, list, and delete income/expense transactions.",
}

// txListCmd menampilkan transactions.
var txListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		txManager := postgres.NewTransactionManager(application.DB.Pool)
		txService := service.NewTransactionService(
			application.Repos.Transaction,
			application.Repos.Wallet,
			txManager,
		)

		limit, _ := cmd.Flags().GetInt("limit")
		txType, _ := cmd.Flags().GetString("type")

		filter := repository.TransactionFilter{}
		if txType != "" {
			t := models.TransactionType(txType)
			filter.Type = &t
		}

		params := repository.ListParams{Limit: limit, Offset: 0}
		transactions, err := txService.List(ctx, filter, params)
		if err != nil {
			return err
		}

		if len(transactions) == 0 {
			fmt.Println("No transactions found. Add one with: wallet tx add")
			return nil
		}

		fmt.Println(titleStyle.Render("\n📝 Recent Transactions\n"))

		table := tablewriter.NewTable(os.Stdout)
		table.Header("Date", "Type", "Amount", "Description")

		for _, tx := range transactions {
			typeIcon := "📈"
			if tx.Type == models.TransactionTypeExpense {
				typeIcon = "📉"
			}

			table.Append([]string{
				tx.TransactionDate.Format("02 Jan"),
				typeIcon + " " + string(tx.Type),
				formatMoney(tx.Amount),
				truncate(tx.Description, 30),
			})
		}

		table.Render()
		return nil
	},
}

// txAddCmd menambah transaction baru.
var txAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		txManager := postgres.NewTransactionManager(application.DB.Pool)
		txService := service.NewTransactionService(
			application.Repos.Transaction,
			application.Repos.Wallet,
			txManager,
		)

		walletID, _ := cmd.Flags().GetString("wallet")
		txType, _ := cmd.Flags().GetString("type")
		amountStr, _ := cmd.Flags().GetString("amount")
		desc, _ := cmd.Flags().GetString("description")
		dateStr, _ := cmd.Flags().GetString("date")

		// Parse wallet ID
		wID, err := parseUUID(walletID)
		if err != nil {
			return fmt.Errorf("invalid wallet ID: %w", err)
		}

		// Parse amount
		amount, err := decimal.NewFromString(amountStr)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}

		// Parse date
		date := time.Now()
		if dateStr != "" {
			date, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
			}
		}

		// Create transaction
		tx, err := txService.Create(ctx, service.CreateTransactionInput{
			WalletID:    wID,
			Type:        models.TransactionType(txType),
			Amount:      amount,
			Description: desc,
			Date:        date,
		})

		if err != nil {
			return err
		}

		typeIcon := "📈"
		if tx.Type == models.TransactionTypeExpense {
			typeIcon = "📉"
		}

		fmt.Println(successStyle.Render("✅ Transaction added!"))
		fmt.Printf("   %s %s: %s\n", typeIcon, tx.Type, formatMoney(tx.Amount))
		fmt.Printf("   📝 %s\n", tx.Description)

		return nil
	},
}

// txDeleteCmd menghapus transaction.
var txDeleteCmd = &cobra.Command{
	Use:   "delete [transaction-id]",
	Short: "Delete a transaction (and rollback wallet balance)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		txManager := postgres.NewTransactionManager(application.DB.Pool)
		txService := service.NewTransactionService(
			application.Repos.Transaction,
			application.Repos.Wallet,
			txManager,
		)

		id, err := parseUUID(args[0])
		if err != nil {
			return err
		}

		if err := txService.Delete(ctx, id); err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Transaction deleted and balance rolled back!"))
		return nil
	},
}

// txSummaryCmd menampilkan ringkasan transaksi.
var txSummaryCmd = &cobra.Command{
	Use:     "summary",
	Aliases: []string{"sum"},
	Short:   "Show transaction summary for current month",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		txManager := postgres.NewTransactionManager(application.DB.Pool)
		txService := service.NewTransactionService(
			application.Repos.Transaction,
			application.Repos.Wallet,
			txManager,
		)

		now := time.Now()
		summary, err := txService.GetMonthlySummary(ctx, now.Year(), now.Month())
		if err != nil {
			return err
		}

		fmt.Println(titleStyle.Render("\n📊 Monthly Summary - " + now.Format("January 2006") + "\n"))

		incomeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		expenseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

		fmt.Printf("📈 Income:  %s\n", incomeStyle.Render(formatMoney(summary.TotalIncome)))
		fmt.Printf("📉 Expense: %s\n", expenseStyle.Render(formatMoney(summary.TotalExpense)))
		fmt.Printf("💰 Net:     %s\n", moneyStyle.Render(formatMoney(summary.Net)))
		fmt.Printf("📝 Total transactions: %d\n\n", summary.Count)

		return nil
	},
}

func init() {
	// tx list
	txListCmd.Flags().IntP("limit", "l", 10, "Number of transactions to show")
	txListCmd.Flags().StringP("type", "t", "", "Filter by type: income or expense")
	transactionCmd.AddCommand(txListCmd)

	// tx add
	txAddCmd.Flags().StringP("wallet", "w", "", "Wallet ID (required)")
	txAddCmd.Flags().StringP("type", "t", "expense", "Transaction type: income or expense")
	txAddCmd.Flags().StringP("amount", "a", "", "Amount (required)")
	txAddCmd.Flags().StringP("description", "d", "", "Description")
	txAddCmd.Flags().StringP("date", "D", "", "Transaction date (YYYY-MM-DD)")
	_ = txAddCmd.MarkFlagRequired("wallet")
	_ = txAddCmd.MarkFlagRequired("amount")
	transactionCmd.AddCommand(txAddCmd)

	// tx delete
	transactionCmd.AddCommand(txDeleteCmd)

	// tx summary
	transactionCmd.AddCommand(txSummaryCmd)
}

// truncate memotong string jika terlalu panjang.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
