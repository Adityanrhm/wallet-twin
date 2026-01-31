package cli

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/Adityanrhm/wallet-twin/internal/repository"
	"github.com/Adityanrhm/wallet-twin/internal/service"
)

// goalCmd adalah parent command untuk goal operations.
var goalCmd = &cobra.Command{
	Use:     "goal",
	Aliases: []string{"g"},
	Short:   "🎯 Manage savings goals",
	Long:    "Create and track progress toward savings goals.",
}

// goalListCmd menampilkan semua goals.
var goalListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all goals with progress",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		goalService := service.NewGoalService(application.Repos.Goal)

		showAll, _ := cmd.Flags().GetBool("all")

		filter := repository.GoalFilter{}
		if !showAll {
			// Default: show only active
			// Note: we don't filter here for simplicity
		}

		goals, err := goalService.List(ctx, filter)
		if err != nil {
			return err
		}

		if len(goals) == 0 {
			fmt.Println("No goals found. Create one with: wallet goal add")
			return nil
		}

		fmt.Println(titleStyle.Render("\n🎯 Savings Goals\n"))

		table := tablewriter.NewTable(os.Stdout)
		table.Header("Name", "Progress", "Current", "Target", "Status")

		for _, g := range goals {
			progress := g.GetProgress()
			progressBar := renderProgressBar(progress, 8)

			statusIcon := "🔄"
			if g.IsCompleted() {
				statusIcon = "✅"
			}

			table.Append([]string{
				g.Icon + " " + g.Name,
				progressBar,
				formatMoney(g.CurrentAmount),
				formatMoney(g.TargetAmount),
				statusIcon,
			})
		}

		table.Render()
		return nil
	},
}

// goalAddCmd menambah goal baru.
var goalAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new savings goal",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		goalService := service.NewGoalService(application.Repos.Goal)

		name, _ := cmd.Flags().GetString("name")
		targetStr, _ := cmd.Flags().GetString("target")
		desc, _ := cmd.Flags().GetString("description")
		icon, _ := cmd.Flags().GetString("icon")

		// Parse target
		target, err := decimal.NewFromString(targetStr)
		if err != nil {
			return fmt.Errorf("invalid target amount: %w", err)
		}

		goal, err := goalService.Create(ctx, service.CreateGoalInput{
			Name:         name,
			Description:  desc,
			TargetAmount: target,
			Icon:         icon,
		})

		if err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Goal created!"))
		fmt.Printf("   🎯 %s %s\n", goal.Icon, goal.Name)
		fmt.Printf("   💰 Target: %s\n", formatMoney(goal.TargetAmount))

		return nil
	},
}

// goalContributeCmd menambah kontribusi ke goal.
var goalContributeCmd = &cobra.Command{
	Use:     "contribute",
	Aliases: []string{"add-funds", "c"},
	Short:   "Add contribution to a goal",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		goalService := service.NewGoalService(application.Repos.Goal)

		goalID, _ := cmd.Flags().GetString("goal")
		amountStr, _ := cmd.Flags().GetString("amount")
		note, _ := cmd.Flags().GetString("note")

		// Parse goal ID
		gID, err := parseUUID(goalID)
		if err != nil {
			return fmt.Errorf("invalid goal ID: %w", err)
		}

		// Parse amount
		amount, err := decimal.NewFromString(amountStr)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}

		err = goalService.AddContribution(ctx, gID, service.AddContributionInput{
			Amount: amount,
			Note:   note,
		})

		if err != nil {
			return err
		}

		// Get updated progress
		progress, _ := goalService.GetProgress(ctx, gID)

		fmt.Println(successStyle.Render("✅ Contribution added!"))
		fmt.Printf("   💰 Amount: %s\n", formatMoney(amount))
		if progress != nil {
			fmt.Printf("   📊 Progress: %.1f%%\n", progress.Progress)
			if progress.IsCompleted {
				fmt.Println("   🎉 Goal completed!")
			}
		}

		return nil
	},
}

// goalDeleteCmd menghapus goal.
var goalDeleteCmd = &cobra.Command{
	Use:   "delete [goal-id]",
	Short: "Delete a goal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		goalService := service.NewGoalService(application.Repos.Goal)

		id, err := parseUUID(args[0])
		if err != nil {
			return err
		}

		if err := goalService.Delete(ctx, id); err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Goal deleted!"))
		return nil
	},
}

func init() {
	// goal list
	goalListCmd.Flags().BoolP("all", "a", false, "Show all goals including completed")
	goalCmd.AddCommand(goalListCmd)

	// goal add
	goalAddCmd.Flags().StringP("name", "n", "", "Goal name (required)")
	goalAddCmd.Flags().StringP("target", "t", "", "Target amount (required)")
	goalAddCmd.Flags().StringP("description", "d", "", "Description")
	goalAddCmd.Flags().StringP("icon", "i", "🎯", "Goal icon")
	_ = goalAddCmd.MarkFlagRequired("name")
	_ = goalAddCmd.MarkFlagRequired("target")
	goalCmd.AddCommand(goalAddCmd)

	// goal contribute
	goalContributeCmd.Flags().StringP("goal", "g", "", "Goal ID (required)")
	goalContributeCmd.Flags().StringP("amount", "a", "", "Contribution amount (required)")
	goalContributeCmd.Flags().StringP("note", "n", "", "Contribution note")
	_ = goalContributeCmd.MarkFlagRequired("goal")
	_ = goalContributeCmd.MarkFlagRequired("amount")
	goalCmd.AddCommand(goalContributeCmd)

	// goal delete
	goalCmd.AddCommand(goalDeleteCmd)
}
