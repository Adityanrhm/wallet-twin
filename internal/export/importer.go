// Package export menyediakan fungsi untuk import data dari berbagai format.
package export

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// Importer handles data import operations.
type Importer struct {
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	categoryRepo    repository.CategoryRepository
	goalRepo        repository.GoalRepository
	txManager       repository.TransactionManager
}

// NewImporter creates a new Importer.
func NewImporter(
	walletRepo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	categoryRepo repository.CategoryRepository,
	goalRepo repository.GoalRepository,
	txManager repository.TransactionManager,
) *Importer {
	return &Importer{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
		goalRepo:        goalRepo,
		txManager:       txManager,
	}
}

// ImportResult contains the result of an import operation.
type ImportResult struct {
	TotalRows     int
	SuccessCount  int
	SkippedCount  int
	Errors        []string
}

// ==================== CSV Import ====================

// TransactionsFromCSV imports transactions from a CSV file.
func (i *Importer) TransactionsFromCSV(ctx context.Context, filename string) (*ImportResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Create column index map
	colIndex := make(map[string]int)
	for idx, col := range header {
		colIndex[strings.ToLower(strings.TrimSpace(col))] = idx
	}

	// Required columns
	requiredCols := []string{"date", "type", "amount", "wallet id"}
	for _, col := range requiredCols {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	result := &ImportResult{}

	// Read rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row read error: %v", err))
			continue
		}

		result.TotalRows++

		// Parse row
		tx, err := i.parseTransactionRow(row, colIndex)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", result.TotalRows, err))
			result.SkippedCount++
			continue
		}

		// Create transaction (without balance update for import)
		if err := i.transactionRepo.Create(ctx, tx); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", result.TotalRows, err))
			result.SkippedCount++
			continue
		}

		result.SuccessCount++
	}

	return result, nil
}

func (i *Importer) parseTransactionRow(row []string, colIndex map[string]int) (*models.Transaction, error) {
	getValue := func(col string) string {
		if idx, ok := colIndex[col]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	// Parse date
	dateStr := getValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %s", dateStr)
	}

	// Parse type
	txType := models.TransactionType(strings.ToLower(getValue("type")))
	if txType != models.TransactionTypeIncome && txType != models.TransactionTypeExpense {
		return nil, fmt.Errorf("invalid type: %s", txType)
	}

	// Parse amount
	amountStr := getValue("amount")
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %s", amountStr)
	}

	// Parse wallet ID
	walletIDStr := getValue("wallet id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet id: %s", walletIDStr)
	}

	// Optional: category ID
	var categoryID *uuid.UUID
	if catIDStr := getValue("category id"); catIDStr != "" {
		catID, err := uuid.Parse(catIDStr)
		if err == nil {
			categoryID = &catID
		}
	}

	// Optional: description
	description := getValue("description")

	// Optional: tags
	var tags []string
	if tagsStr := getValue("tags"); tagsStr != "" {
		tags = strings.Split(tagsStr, ";")
	}

	return &models.Transaction{
		BaseModel:       models.BaseModel{ID: models.NewID()},
		WalletID:        walletID,
		CategoryID:      categoryID,
		Type:            txType,
		Amount:          amount,
		Description:     description,
		Tags:            tags,
		TransactionDate: date,
	}, nil
}

// ==================== JSON Import ====================

// FromJSON imports all data from a JSON backup file.
func (i *Importer) FromJSON(ctx context.Context, filename string) (*ImportResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var data ExportData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	result := &ImportResult{}

	// Import in transaction for atomicity
	err = i.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		// Import wallets
		for _, w := range data.Wallets {
			result.TotalRows++
			if err := i.walletRepo.Create(ctx, w); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("wallet %s: %v", w.Name, err))
				result.SkippedCount++
			} else {
				result.SuccessCount++
			}
		}

		// Import categories
		for _, c := range data.Categories {
			result.TotalRows++
			if err := i.categoryRepo.Create(ctx, c); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("category %s: %v", c.Name, err))
				result.SkippedCount++
			} else {
				result.SuccessCount++
			}
		}

		// Import transactions
		for _, tx := range data.Transactions {
			result.TotalRows++
			if err := i.transactionRepo.Create(ctx, tx); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("transaction %s: %v", tx.ID, err))
				result.SkippedCount++
			} else {
				result.SuccessCount++
			}
		}

		// Import goals
		for _, g := range data.Goals {
			result.TotalRows++
			if err := i.goalRepo.Create(ctx, g); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("goal %s: %v", g.Name, err))
				result.SkippedCount++
			} else {
				result.SuccessCount++
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
