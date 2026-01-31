package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/service"
)

// budgetCmd adalah parent command untuk budget operations.
var budgetCmd = &cobra.Command{
	Use:     "budget",
	Aliases: []string{"b"},
	Short:   "📊 Manage budgets",
	Long:    "Create and track spending budgets per category.",
}

// budgetListCmd menampilkan semua budgets dengan status.
var budgetListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all active budgets with status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		budgetService := service.NewBudgetService(
			application.Repos.Budget,
			application.Repos.Transaction,
		)

		statuses, err := budgetService.GetAllStatus(ctx)
		if err != nil {
			return err
		}

		if len(statuses) == 0 {
			fmt.Println("No active budgets. Create one with: wallet budget add")
			return nil
		}

		fmt.Println(titleStyle.Render("\n📊 Budget Status\n"))

		table := tablewriter.NewTable(os.Stdout)
		table.Header("Category", "Budget", "Spent", "Remaining", "Progress")

		for _, s := range statuses {
			// Progress bar
			progressBar := renderProgressBar(s.Progress, 10)

			// Color based on status
			remaining := formatMoney(s.Remaining)
			if s.IsOverBudget {
				remaining = "⚠️ OVER"
			}

			table.Append([]string{
				s.CategoryIcon + " " + s.CategoryName,
				formatMoney(s.Budget.Amount),
				formatMoney(s.Spent),
				remaining,
				progressBar,
			})
		}

		table.Render()
		return nil
	},
}

// budgetAddCmd menambah budget baru.
var budgetAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new budget",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		budgetService := service.NewBudgetService(
			application.Repos.Budget,
			application.Repos.Transaction,
		)

		categoryID, _ := cmd.Flags().GetString("category")
		amountStr, _ := cmd.Flags().GetString("amount")
		period, _ := cmd.Flags().GetString("period")

		// Parse category ID
		catID, err := parseUUID(categoryID)
		if err != nil {
			return fmt.Errorf("invalid category ID: %w", err)
		}

		// Parse amount
		amount, err := decimal.NewFromString(amountStr)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}

		// Set start date (first of current month for monthly)
		now := time.Now()
		startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

		budget, err := budgetService.Create(ctx, service.CreateBudgetInput{
			CategoryID: catID,
			Amount:     amount,
			Period:     models.BudgetPeriod(period),
			StartDate:  startDate,
		})

		if err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Budget created!"))
		fmt.Printf("   💰 Amount: %s\n", formatMoney(budget.Amount))
		fmt.Printf("   📅 Period: %s\n", budget.Period)

		return nil
	},
}

// budgetDeleteCmd menghapus budget.
var budgetDeleteCmd = &cobra.Command{
	Use:   "delete [budget-id]",
	Short: "Delete a budget",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		budgetService := service.NewBudgetService(
			application.Repos.Budget,
			application.Repos.Transaction,
		)

		id, err := parseUUID(args[0])
		if err != nil {
			return err
		}

		if err := budgetService.Delete(ctx, id); err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Budget deleted!"))
		return nil
	},
}

func init() {
	// budget list
	budgetCmd.AddCommand(budgetListCmd)

	// budget add
	budgetAddCmd.Flags().StringP("category", "c", "", "Category ID (required)")
	budgetAddCmd.Flags().StringP("amount", "a", "", "Budget amount (required)")
	budgetAddCmd.Flags().StringP("period", "p", "monthly", "Budget period: weekly, monthly, yearly")
	_ = budgetAddCmd.MarkFlagRequired("category")
	_ = budgetAddCmd.MarkFlagRequired("amount")
	budgetCmd.AddCommand(budgetAddCmd)

	// budget delete
	budgetCmd.AddCommand(budgetDeleteCmd)
}

// renderProgressBar membuat visual progress bar.
func renderProgressBar(progress float64, width int) string {
	filled := int(progress / 100.0 * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			if progress > 100 {
				bar += "🟥"
			} else if progress > 80 {
				bar += "🟨"
			} else {
				bar += "🟩"
			}
		} else {
			bar += "⬜"
		}
	}

	return fmt.Sprintf("%s %.0f%%", bar, progress)
}
