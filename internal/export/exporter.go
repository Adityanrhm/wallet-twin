// Package export menyediakan fungsi untuk export data ke berbagai format.
//
// Format yang didukung:
// - CSV: Comma-separated values, mudah dibuka di Excel
// - JSON: JavaScript Object Notation, untuk backup atau integrasi
//
// Usage:
//
//	exporter := export.NewExporter(repos)
//
//	// Export ke CSV
//	err := exporter.TransactionsToCSV(ctx, "transactions.csv", filter)
//
//	// Export ke JSON
//	err := exporter.WalletsToJSON(ctx, "wallets.json")
package export

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// Exporter handles data export operations.
type Exporter struct {
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	categoryRepo    repository.CategoryRepository
	goalRepo        repository.GoalRepository
}

// NewExporter creates a new Exporter.
func NewExporter(
	walletRepo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	categoryRepo repository.CategoryRepository,
	goalRepo repository.GoalRepository,
) *Exporter {
	return &Exporter{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		goalRepo:        goalRepo,
	}
}

// ==================== CSV Export ====================

// TransactionsToCSV exports transactions to a CSV file.
func (e *Exporter) TransactionsToCSV(ctx context.Context, filename string, filter repository.TransactionFilter) error {
	// Get transactions
	params := repository.ListParams{Limit: 10000, Offset: 0}
	transactions, err := e.transactionRepo.List(ctx, filter, params)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write CSV
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"ID", "Date", "Type", "Amount", "Description", "Wallet ID", "Category ID", "Tags"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Rows
	for _, tx := range transactions {
		categoryID := ""
		if tx.CategoryID != nil {
			categoryID = tx.CategoryID.String()
		}

		tags := ""
		if len(tx.Tags) > 0 {
			for i, t := range tx.Tags {
				if i > 0 {
					tags += ";"
				}
				tags += t
			}
		}

		row := []string{
			tx.ID.String(),
			tx.TransactionDate.Format("2006-01-02"),
			string(tx.Type),
			tx.Amount.String(),
			tx.Description,
			tx.WalletID.String(),
			categoryID,
			tags,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// WalletsToCSV exports wallets to a CSV file.
func (e *Exporter) WalletsToCSV(ctx context.Context, filename string) error {
	wallets, err := e.walletRepo.List(ctx, repository.WalletFilter{})
	if err != nil {
		return fmt.Errorf("failed to get wallets: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"ID", "Name", "Type", "Balance", "Currency", "Color", "Icon", "Is Active", "Created At"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Rows
	for _, w := range wallets {
		row := []string{
			w.ID.String(),
			w.Name,
			string(w.Type),
			w.Balance.String(),
			w.Currency,
			w.Color,
			w.Icon,
			fmt.Sprintf("%t", w.IsActive),
			w.CreatedAt.Format(time.RFC3339),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// ==================== JSON Export ====================

// ExportData adalah struktur untuk full backup.
type ExportData struct {
	ExportedAt   time.Time            `json:"exported_at"`
	Version      string               `json:"version"`
	Wallets      []*models.Wallet     `json:"wallets"`
	Categories   []*models.Category   `json:"categories"`
	Transactions []*models.Transaction `json:"transactions"`
	Goals        []*models.Goal       `json:"goals"`
}

// ToJSON exports all data to a JSON file (full backup).
func (e *Exporter) ToJSON(ctx context.Context, filename string) error {
	// Get all data
	wallets, err := e.walletRepo.List(ctx, repository.WalletFilter{})
	if err != nil {
		return fmt.Errorf("failed to get wallets: %w", err)
	}

	categories, err := e.categoryRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	params := repository.ListParams{Limit: 100000, Offset: 0}
	transactions, err := e.transactionRepo.List(ctx, repository.TransactionFilter{}, params)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	goals, err := e.goalRepo.List(ctx, repository.GoalFilter{})
	if err != nil {
		return fmt.Errorf("failed to get goals: %w", err)
	}

	// Create export data
	data := ExportData{
		ExportedAt:   time.Now(),
		Version:      "1.0.0",
		Wallets:      wallets,
		Categories:   categories,
		Transactions: transactions,
		Goals:        goals,
	}

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// WalletsToJSON exports wallets to a JSON file.
func (e *Exporter) WalletsToJSON(ctx context.Context, filename string) error {
	wallets, err := e.walletRepo.List(ctx, repository.WalletFilter{})
	if err != nil {
		return fmt.Errorf("failed to get wallets: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(wallets)
}

// TransactionsToJSON exports transactions to a JSON file.
func (e *Exporter) TransactionsToJSON(ctx context.Context, filename string, filter repository.TransactionFilter) error {
	params := repository.ListParams{Limit: 100000, Offset: 0}
	transactions, err := e.transactionRepo.List(ctx, filter, params)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(transactions)
}
