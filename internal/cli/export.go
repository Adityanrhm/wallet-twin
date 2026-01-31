package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Adityanrhm/wallet-twin/internal/export"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
	"github.com/Adityanrhm/wallet-twin/internal/repository/postgres"
)

// exportCmd adalah parent command untuk export operations.
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "📤 Export data to CSV/JSON",
	Long:  "Export your financial data to CSV or JSON format.",
}

// exportAllCmd exports semua data ke JSON.
var exportAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Export all data to JSON (full backup)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		exporter := export.NewExporter(
			application.Repos.Wallet,
			application.Repos.Transaction,
			application.Repos.Category,
			application.Repos.Goal,
		)

		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = fmt.Sprintf("wallet-twin-backup-%s.json", time.Now().Format("20060102-150405"))
		}

		if err := exporter.ToJSON(ctx, output); err != nil {
			return err
		}

		absPath, _ := filepath.Abs(output)
		fmt.Println(successStyle.Render("✅ Export successful!"))
		fmt.Printf("   📁 File: %s\n", absPath)

		return nil
	},
}

// exportTransactionsCmd exports transactions.
var exportTransactionsCmd = &cobra.Command{
	Use:     "transactions",
	Aliases: []string{"tx"},
	Short:   "Export transactions to CSV/JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		exporter := export.NewExporter(
			application.Repos.Wallet,
			application.Repos.Transaction,
			application.Repos.Category,
			application.Repos.Goal,
		)

		output, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		if output == "" {
			output = fmt.Sprintf("transactions-%s.%s", time.Now().Format("20060102"), format)
		}

		filter := repository.TransactionFilter{}

		var err error
		if format == "json" {
			err = exporter.TransactionsToJSON(ctx, output, filter)
		} else {
			err = exporter.TransactionsToCSV(ctx, output, filter)
		}

		if err != nil {
			return err
		}

		absPath, _ := filepath.Abs(output)
		fmt.Println(successStyle.Render("✅ Transactions exported!"))
		fmt.Printf("   📁 File: %s\n", absPath)

		return nil
	},
}

// exportWalletsCmd exports wallets.
var exportWalletsCmd = &cobra.Command{
	Use:   "wallets",
	Short: "Export wallets to CSV/JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		exporter := export.NewExporter(
			application.Repos.Wallet,
			application.Repos.Transaction,
			application.Repos.Category,
			application.Repos.Goal,
		)

		output, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		if output == "" {
			output = fmt.Sprintf("wallets-%s.%s", time.Now().Format("20060102"), format)
		}

		var err error
		if format == "json" {
			err = exporter.WalletsToJSON(ctx, output)
		} else {
			err = exporter.WalletsToCSV(ctx, output)
		}

		if err != nil {
			return err
		}

		absPath, _ := filepath.Abs(output)
		fmt.Println(successStyle.Render("✅ Wallets exported!"))
		fmt.Printf("   📁 File: %s\n", absPath)

		return nil
	},
}

// importCmd adalah parent command untuk import operations.
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "📥 Import data from CSV/JSON",
	Long:  "Import financial data from CSV or JSON files.",
}

// importTransactionsCmd imports transactions from CSV.
var importTransactionsCmd = &cobra.Command{
	Use:   "transactions [file]",
	Short: "Import transactions from CSV",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		txManager := postgres.NewTransactionManager(application.DB.Pool)
		importer := export.NewImporter(
			application.Repos.Wallet,
			application.Repos.Transaction,
			application.Repos.Category,
			application.Repos.Goal,
			txManager,
		)

		filename := args[0]
		result, err := importer.TransactionsFromCSV(ctx, filename)
		if err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Import completed!"))
		fmt.Printf("   📊 Total rows: %d\n", result.TotalRows)
		fmt.Printf("   ✅ Imported: %d\n", result.SuccessCount)
		fmt.Printf("   ⏭️ Skipped: %d\n", result.SkippedCount)

		if len(result.Errors) > 0 {
			fmt.Println("\n⚠️ Errors:")
			for _, e := range result.Errors[:min(5, len(result.Errors))] {
				fmt.Printf("   - %s\n", e)
			}
			if len(result.Errors) > 5 {
				fmt.Printf("   ... and %d more\n", len(result.Errors)-5)
			}
		}

		return nil
	},
}

// importBackupCmd imports from JSON backup.
var importBackupCmd = &cobra.Command{
	Use:   "backup [file]",
	Short: "Import from JSON backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		txManager := postgres.NewTransactionManager(application.DB.Pool)
		importer := export.NewImporter(
			application.Repos.Wallet,
			application.Repos.Transaction,
			application.Repos.Category,
			application.Repos.Goal,
			txManager,
		)

		filename := args[0]

		// Validate file extension
		if !strings.HasSuffix(filename, ".json") {
			return fmt.Errorf("backup file must be JSON format")
		}

		result, err := importer.FromJSON(ctx, filename)
		if err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Backup restored!"))
		fmt.Printf("   📊 Total items: %d\n", result.TotalRows)
		fmt.Printf("   ✅ Imported: %d\n", result.SuccessCount)
		fmt.Printf("   ⏭️ Skipped: %d\n", result.SkippedCount)

		return nil
	},
}

func init() {
	// export all
	exportAllCmd.Flags().StringP("output", "o", "", "Output filename")
	exportCmd.AddCommand(exportAllCmd)

	// export transactions
	exportTransactionsCmd.Flags().StringP("output", "o", "", "Output filename")
	exportTransactionsCmd.Flags().StringP("format", "f", "csv", "Output format: csv or json")
	exportCmd.AddCommand(exportTransactionsCmd)

	// export wallets
	exportWalletsCmd.Flags().StringP("output", "o", "", "Output filename")
	exportWalletsCmd.Flags().StringP("format", "f", "csv", "Output format: csv or json")
	exportCmd.AddCommand(exportWalletsCmd)

	// import transactions
	importCmd.AddCommand(importTransactionsCmd)

	// import backup
	importCmd.AddCommand(importBackupCmd)

	// Add to root
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}
