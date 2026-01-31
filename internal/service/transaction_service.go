package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// TransactionService menangani business logic untuk transaction operations.
//
// PENTING: Setiap create/delete transaction HARUS update wallet balance.
// Ini adalah ATOMIC operation - harus dalam satu database transaction.
//
// Flow Create Transaction:
// 1. Validate input
// 2. Begin DB transaction
// 3. Create transaction record
// 4. Update wallet balance
// 5. Commit
//
// Jika langkah manapun gagal, semua di-rollback.
type TransactionService struct {
	txRepo     repository.TransactionRepository
	walletRepo repository.WalletRepository
	txManager  repository.TransactionManager
}

// NewTransactionService membuat TransactionService baru.
func NewTransactionService(
	txRepo repository.TransactionRepository,
	walletRepo repository.WalletRepository,
	txManager repository.TransactionManager,
) *TransactionService {
	return &TransactionService{
		txRepo:     txRepo,
		walletRepo: walletRepo,
		txManager:  txManager,
	}
}

// Common errors
var (
	ErrInsufficientBalance = errors.New("insufficient wallet balance")
)

// Create membuat transaksi baru dan update wallet balance.
//
// Income: wallet.balance += amount
// Expense: wallet.balance -= amount (error jika tidak cukup)
//
// Contoh:
//
//	tx, err := txService.Create(ctx, service.CreateTransactionInput{
//	    WalletID:    walletID,
//	    CategoryID:  &categoryID,
//	    Type:        models.TransactionTypeExpense,
//	    Amount:      decimal.NewFromInt(50000),
//	    Description: "Makan siang",
//	})
func (s *TransactionService) Create(ctx context.Context, input CreateTransactionInput) (*models.Transaction, error) {
	// Get wallet and validate
	wallet, err := s.walletRepo.GetByID(ctx, input.WalletID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	if !wallet.IsActive {
		return nil, errors.New("cannot create transaction on inactive wallet")
	}

	// Check balance for expense
	if input.Type == models.TransactionTypeExpense {
		if wallet.Balance.LessThan(input.Amount) {
			return nil, ErrInsufficientBalance
		}
	}

	// Create transaction model
	transaction := &models.Transaction{
		BaseModel:       models.BaseModel{ID: models.NewID()},
		WalletID:        input.WalletID,
		CategoryID:      input.CategoryID,
		Type:            input.Type,
		Amount:          input.Amount,
		Description:     input.Description,
		Tags:            input.Tags,
		TransactionDate: input.Date,
	}

	if transaction.TransactionDate.IsZero() {
		transaction.TransactionDate = time.Now()
	}

	if err := transaction.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Calculate new balance
	newBalance := wallet.Balance
	if input.Type == models.TransactionTypeIncome {
		newBalance = newBalance.Add(input.Amount)
	} else {
		newBalance = newBalance.Sub(input.Amount)
	}

	// Execute in transaction
	err = s.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.txRepo.Create(ctx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newBalance); err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetByID mengambil transaction berdasarkan ID.
func (s *TransactionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	tx, err := s.txRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return tx, nil
}

// List mengambil transactions dengan filter.
func (s *TransactionService) List(
	ctx context.Context,
	filter repository.TransactionFilter,
	params repository.ListParams,
) ([]*models.Transaction, error) {
	transactions, err := s.txRepo.List(ctx, filter, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	return transactions, nil
}

// GetByWallet mengambil transactions untuk wallet tertentu.
func (s *TransactionService) GetByWallet(
	ctx context.Context,
	walletID uuid.UUID,
	params repository.ListParams,
) ([]*models.Transaction, error) {
	filter := repository.TransactionFilter{WalletID: &walletID}
	return s.List(ctx, filter, params)
}

// GetRecent mengambil transaksi terbaru.
func (s *TransactionService) GetRecent(ctx context.Context, limit int) ([]*models.Transaction, error) {
	params := repository.ListParams{Limit: limit, Offset: 0}
	return s.List(ctx, repository.TransactionFilter{}, params)
}

// Delete menghapus transaction dan rollback wallet balance.
func (s *TransactionService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get transaction
	tx, err := s.txRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Get wallet
	wallet, err := s.walletRepo.GetByID(ctx, tx.WalletID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	// Calculate rollback balance
	newBalance := wallet.Balance
	if tx.Type == models.TransactionTypeIncome {
		// Income was added, now subtract
		newBalance = newBalance.Sub(tx.Amount)
	} else {
		// Expense was subtracted, now add back
		newBalance = newBalance.Add(tx.Amount)
	}

	// Execute in transaction
	return s.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.txRepo.Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete transaction: %w", err)
		}

		if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, newBalance); err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		return nil
	})
}

// GetSummary menghitung ringkasan transaksi.
func (s *TransactionService) GetSummary(
	ctx context.Context,
	filter repository.TransactionFilter,
) (*repository.TransactionSummary, error) {
	summary, err := s.txRepo.GetSummary(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}
	return summary, nil
}

// GetMonthlySummary menghitung ringkasan untuk bulan tertentu.
func (s *TransactionService) GetMonthlySummary(
	ctx context.Context,
	year int,
	month time.Month,
) (*repository.TransactionSummary, error) {
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, -1) // Last day of month

	filter := repository.TransactionFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	return s.GetSummary(ctx, filter)
}

// GetCategorySummary menghitung ringkasan per kategori.
func (s *TransactionService) GetCategorySummary(
	ctx context.Context,
	filter repository.TransactionFilter,
) ([]*repository.CategorySummary, error) {
	summaries, err := s.txRepo.GetByCategory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get category summary: %w", err)
	}
	return summaries, nil
}

// CreateTransactionInput adalah input untuk membuat transaction.
type CreateTransactionInput struct {
	WalletID    uuid.UUID
	CategoryID  *uuid.UUID
	Type        models.TransactionType
	Amount      decimal.Decimal
	Description string
	Tags        []string
	Date        time.Time
}
