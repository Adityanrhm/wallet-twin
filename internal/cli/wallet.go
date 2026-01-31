package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
	"github.com/Adityanrhm/wallet-twin/internal/service"
)

// Styles untuk output berwarna
var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	moneyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
)

// walletCmd adalah parent command untuk wallet operations.
var walletCmd = &cobra.Command{
	Use:     "wallet",
	Aliases: []string{"w"},
	Short:   "💼 Manage your wallets",
	Long:    "Add, list, update, and delete wallets (accounts).",
}

// walletListCmd menampilkan semua wallets.
var walletListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all wallets",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		walletService := service.NewWalletService(application.Repos.Wallet)

		showAll, _ := cmd.Flags().GetBool("all")

		filter := repository.WalletFilter{}
		if !showAll {
			isActive := true
			filter.IsActive = &isActive
		}

		wallets, err := walletService.List(ctx, filter)
		if err != nil {
			return err
		}

		if len(wallets) == 0 {
			fmt.Println("No wallets found. Create one with: wallet wallet add")
			return nil
		}

		// Print table
		fmt.Println(titleStyle.Render("\n💼 Your Wallets\n"))

		table := tablewriter.NewTable(os.Stdout)
		table.Header("Name", "Type", "Balance", "Currency", "Status")

		for _, w := range wallets {
			status := "✅"
			if !w.IsActive {
				status = "❌"
			}

			table.Append([]string{
				w.Icon + " " + w.Name,
				string(w.Type),
				formatMoney(w.Balance),
				w.Currency,
				status,
			})
		}

		table.Render()

		// Total
		total, _ := walletService.GetTotalBalance(ctx)
		fmt.Printf("\n💰 Total Balance: %s\n\n", moneyStyle.Render(formatMoney(total)))

		return nil
	},
}

// walletAddCmd menambah wallet baru.
var walletAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		walletService := service.NewWalletService(application.Repos.Wallet)

		name, _ := cmd.Flags().GetString("name")
		walletType, _ := cmd.Flags().GetString("type")
		currency, _ := cmd.Flags().GetString("currency")
		balance, _ := cmd.Flags().GetString("balance")
		icon, _ := cmd.Flags().GetString("icon")

		// Parse balance
		initialBalance := decimal.Zero
		if balance != "" {
			var err error
			initialBalance, err = decimal.NewFromString(balance)
			if err != nil {
				return fmt.Errorf("invalid balance: %w", err)
			}
		}

		// Create wallet
		wallet, err := walletService.Create(ctx, service.CreateWalletInput{
			Name:           name,
			Type:           models.WalletType(walletType),
			Currency:       currency,
			InitialBalance: initialBalance,
			Icon:           icon,
		})

		if err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Wallet created successfully!"))
		fmt.Printf("   ID: %s\n", wallet.ID)
		fmt.Printf("   Name: %s %s\n", wallet.Icon, wallet.Name)
		fmt.Printf("   Balance: %s %s\n", wallet.Currency, formatMoney(wallet.Balance))

		return nil
	},
}

// walletDeleteCmd menghapus wallet.
var walletDeleteCmd = &cobra.Command{
	Use:   "delete [wallet-id]",
	Short: "Delete a wallet (soft delete)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		walletService := service.NewWalletService(application.Repos.Wallet)

		id, err := parseUUID(args[0])
		if err != nil {
			return err
		}

		if err := walletService.Delete(ctx, id); err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Wallet deleted successfully!"))
		return nil
	},
}

// walletBalanceCmd menampilkan total balance.
var walletBalanceCmd = &cobra.Command{
	Use:     "balance",
	Aliases: []string{"bal"},
	Short:   "Show total balance across all wallets",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		walletService := service.NewWalletService(application.Repos.Wallet)

		total, err := walletService.GetTotalBalance(ctx)
		if err != nil {
			return err
		}

		fmt.Println(titleStyle.Render("\n💰 Total Balance"))
		fmt.Printf("%s %s\n\n", application.Config.App.Currency, moneyStyle.Render(formatMoney(total)))

		return nil
	},
}

func init() {
	// wallet list
	walletListCmd.Flags().BoolP("all", "a", false, "Show all wallets including inactive")
	walletCmd.AddCommand(walletListCmd)

	// wallet add
	walletAddCmd.Flags().StringP("name", "n", "", "Wallet name (required)")
	walletAddCmd.Flags().StringP("type", "t", "cash", "Wallet type: cash, bank, ewallet")
	walletAddCmd.Flags().StringP("currency", "c", "IDR", "Currency code")
	walletAddCmd.Flags().StringP("balance", "b", "0", "Initial balance")
	walletAddCmd.Flags().StringP("icon", "i", "💰", "Wallet icon")
	_ = walletAddCmd.MarkFlagRequired("name")
	walletCmd.AddCommand(walletAddCmd)

	// wallet delete
	walletCmd.AddCommand(walletDeleteCmd)

	// wallet balance
	walletCmd.AddCommand(walletBalanceCmd)
}

// formatMoney memformat decimal sebagai string dengan thousand separator.
func formatMoney(d decimal.Decimal) string {
	return d.StringFixed(0)
}
