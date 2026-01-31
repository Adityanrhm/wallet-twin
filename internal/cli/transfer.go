package cli

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/Adityanrhm/wallet-twin/internal/repository/postgres"
	"github.com/Adityanrhm/wallet-twin/internal/service"
)

// transferCmd adalah command untuk transfer antar wallet.
var transferCmd = &cobra.Command{
	Use:     "transfer",
	Aliases: []string{"tf"},
	Short:   "🔄 Transfer money between wallets",
	Long:    "Transfer money from one wallet to another, with optional fee.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		txManager := postgres.NewTransactionManager(application.DB.Pool)
		transferService := service.NewTransferService(
			application.Repos.Transfer,
			application.Repos.Wallet,
			txManager,
		)

		fromID, _ := cmd.Flags().GetString("from")
		toID, _ := cmd.Flags().GetString("to")
		amountStr, _ := cmd.Flags().GetString("amount")
		feeStr, _ := cmd.Flags().GetString("fee")
		note, _ := cmd.Flags().GetString("note")

		// Parse IDs
		fromUUID, err := parseUUID(fromID)
		if err != nil {
			return fmt.Errorf("invalid source wallet ID: %w", err)
		}

		toUUID, err := parseUUID(toID)
		if err != nil {
			return fmt.Errorf("invalid destination wallet ID: %w", err)
		}

		// Parse amount
		amount, err := decimal.NewFromString(amountStr)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}

		// Parse fee
		fee := decimal.Zero
		if feeStr != "" {
			fee, err = decimal.NewFromString(feeStr)
			if err != nil {
				return fmt.Errorf("invalid fee: %w", err)
			}
		}

		// Create transfer
		transfer, err := transferService.Create(ctx, service.CreateTransferInput{
			FromWalletID: fromUUID,
			ToWalletID:   toUUID,
			Amount:       amount,
			Fee:          fee,
			Note:         note,
		})

		if err != nil {
			return err
		}

		fmt.Println(successStyle.Render("✅ Transfer successful!"))
		fmt.Printf("   💸 Amount: %s\n", formatMoney(transfer.Amount))
		if !transfer.Fee.IsZero() {
			fmt.Printf("   💳 Fee: %s\n", formatMoney(transfer.Fee))
			fmt.Printf("   📉 Total deducted: %s\n", formatMoney(transfer.TotalDeducted()))
		}
		if transfer.Note != "" {
			fmt.Printf("   📝 Note: %s\n", transfer.Note)
		}

		return nil
	},
}

func init() {
	transferCmd.Flags().StringP("from", "f", "", "Source wallet ID (required)")
	transferCmd.Flags().StringP("to", "t", "", "Destination wallet ID (required)")
	transferCmd.Flags().StringP("amount", "a", "", "Amount to transfer (required)")
	transferCmd.Flags().StringP("fee", "F", "0", "Transfer fee")
	transferCmd.Flags().StringP("note", "n", "", "Transfer note")

	_ = transferCmd.MarkFlagRequired("from")
	_ = transferCmd.MarkFlagRequired("to")
	_ = transferCmd.MarkFlagRequired("amount")
}
