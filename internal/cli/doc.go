// Package cli berisi command-line interface menggunakan Cobra.
//
// Cobra adalah library standard untuk CLI di Go.
// Digunakan oleh Docker, Kubernetes, Hugo, dan banyak tool populer lainnya.
//
// Struktur command:
//
//	wallet                    # root command - tampilkan help atau dashboard
//	├── init                  # inisialisasi database
//	├── wallet                # sub-command untuk manage wallets
//	│   ├── add              # tambah wallet baru
//	│   ├── list             # list semua wallets
//	│   └── ...
//	├── transaction (tx)      # sub-command untuk transactions
//	│   ├── add              # tambah transaksi
//	│   └── ...
//	└── ...
//
// Setiap command didefinisikan sebagai *cobra.Command:
//
//	var addWalletCmd = &cobra.Command{
//	    Use:   "add",
//	    Short: "Add a new wallet",
//	    Long: `Add a new wallet to track your finances.
//	           Example: wallet wallet add`,
//	    RunE: func(cmd *cobra.Command, args []string) error {
//	        // Interactive form dengan Huh library
//	        return runAddWallet()
//	    },
//	}
//
//	func init() {
//	    walletCmd.AddCommand(addWalletCmd)
//	}
//
// Untuk user experience yang baik, kita menggunakan:
// - Huh: Interactive forms dan prompts
// - Lipgloss: Styling dan colors
// - Table: Pretty table output
package cli

// TODO: Add command implementations
