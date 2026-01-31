package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dashboardCmd membuka TUI dashboard.
var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"dash", "d"},
	Short:   "🖥️ Open interactive TUI dashboard",
	Long:    "Launch the interactive terminal UI dashboard with real-time updates.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement TUI with Bubble Tea
		// Ini akan di-implement di Phase 7 (TUI Dashboard)

		fmt.Println(titleStyle.Render("\n🖥️  Dashboard\n"))
		fmt.Println("Interactive TUI dashboard coming soon!")
		fmt.Println()
		fmt.Println("For now, use these commands:")
		fmt.Println("  wallet wallet list    - List wallets")
		fmt.Println("  wallet tx list        - List transactions")
		fmt.Println("  wallet tx summary     - Monthly summary")
		fmt.Println("  wallet budget list    - Budget status")
		fmt.Println("  wallet goal list      - Goal progress")

		return nil
	},
}
