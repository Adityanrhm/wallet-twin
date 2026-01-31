// Package cli berisi CLI commands menggunakan Cobra.
//
// Cobra adalah library populer untuk membuat CLI di Go.
// Digunakan oleh: kubectl, hugo, gh, docker, dll.
//
// Struktur command Cobra:
//
//	wallet                    # Root command
//	├── wallet list           # Subcommand
//	├── wallet add            # Subcommand
//	├── transaction           # Command group
//	│   ├── transaction add   # Nested subcommand
//	│   └── transaction list
//	└── dashboard             # Start TUI
//
// Setiap command adalah cobra.Command struct yang memiliki:
// - Use: nama command
// - Short: deskripsi singkat (ditampilkan di help)
// - Long: deskripsi panjang
// - Run: function yang dijalankan
// - RunE: sama seperti Run, tapi return error
package cli

import (
	"github.com/spf13/cobra"

	"github.com/Adityanrhm/wallet-twin/internal/app"
)

// rootCmd adalah command utama.
// Semua subcommand di-attach ke rootCmd.
var rootCmd = &cobra.Command{
	Use:   "wallet",
	Short: "💰 Personal Finance CLI - Track your money with style",
	Long: `
 _    _       _ _      _     _____          _       
| |  | |     | | |    | |   |_   _|        (_)      
| |  | | __ _| | | ___| |_    | |_      ___ _ _ __  
| |/\| |/ _' | | |/ _ \ __|   | \ \ /\ / / | '_ \ 
\  /\  / (_| | | |  __/ |_    | |\ V  V /| | | | |
 \/  \/ \__,_|_|_|\___|\__|   \_/ \_/\_/ |_|_| |_|

A CLI personal finance application for tracking income, 
expenses, transfers, budgets, and savings goals.

Get started:
  wallet wallet add      Add a new wallet
  wallet tx add          Add a new transaction
  wallet dashboard       Open interactive TUI dashboard
`,
}

// application adalah pointer ke app.App yang di-set saat Execute.
var application *app.App

// Execute menjalankan root command.
//
// Ini adalah satu-satunya "public" function di package cli.
// Dipanggil dari main.go:
//
//	if err := cli.Execute(application); err != nil {
//	    os.Exit(1)
//	}
func Execute(app *app.App) error {
	application = app
	return rootCmd.Execute()
}

// init adalah special function Go yang dipanggil otomatis.
// Di sini kita add semua subcommands ke root.
func init() {
	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add subcommands
	rootCmd.AddCommand(walletCmd)
	rootCmd.AddCommand(transactionCmd)
	rootCmd.AddCommand(transferCmd)
	rootCmd.AddCommand(budgetCmd)
	rootCmd.AddCommand(goalCmd)
	rootCmd.AddCommand(dashboardCmd)
}
