package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Adityanrhm/wallet-twin/internal/tui"
)

// dashboardCmd membuka TUI dashboard.
var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"dash", "d"},
	Short:   "🖥️ Open interactive TUI dashboard",
	Long:    "Launch the interactive terminal UI dashboard with real-time updates.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create dashboard model
		model := tui.NewDashboard(application)

		// Create and run Bubble Tea program
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return err
		}

		return nil
	},
}

