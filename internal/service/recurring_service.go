package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// RecurringService menangani business logic untuk recurring transactions.
//
// Recurring transaction adalah transaksi yang terjadi secara berkala.
// Service ini menyediakan method untuk:
// - CRUD recurring transactions
// - Process yang jatuh tempo (generate actual transactions)
type RecurringService struct {
	recurringRepo repository.RecurringRepository
	txService     *TransactionService
}

// NewRecurringService membuat RecurringService baru.
func NewRecurringService(
	recurringRepo repository.RecurringRepository,
	txService *TransactionService,
) *RecurringService {
	return &RecurringService{
		recurringRepo: recurringRepo,
		txService:     txService,
	}
}

// Create membuat recurring transaction baru.
func (s *RecurringService) Create(ctx context.Context, input CreateRecurringInput) (*models.RecurringTransaction, error) {
	recurring := &models.RecurringTransaction{
		ID:          models.NewID(),
		WalletID:    input.WalletID,
		CategoryID:  input.CategoryID,
		Type:        input.Type,
		Amount:      input.Amount,
		Description: input.Description,
		Frequency:   input.Frequency,
		NextDue:     input.NextDue,
		EndDate:     input.EndDate,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := recurring.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.recurringRepo.Create(ctx, recurring); err != nil {
		return nil, fmt.Errorf("failed to create recurring: %w", err)
	}

	return recurring, nil
}

// GetByID mengambil recurring berdasarkan ID.
func (s *RecurringService) GetByID(ctx context.Context, id uuid.UUID) (*models.RecurringTransaction, error) {
	recurring, err := s.recurringRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring: %w", err)
	}
	return recurring, nil
}

// List mengambil semua recurring transactions.
func (s *RecurringService) List(ctx context.Context, filter repository.RecurringFilter) ([]*models.RecurringTransaction, error) {
	recurrings, err := s.recurringRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list recurring: %w", err)
	}
	return recurrings, nil
}

// ListActive mengambil recurring aktif.
func (s *RecurringService) ListActive(ctx context.Context) ([]*models.RecurringTransaction, error) {
	isActive := true
	return s.List(ctx, repository.RecurringFilter{IsActive: &isActive})
}

// GetDue mengambil recurring yang jatuh tempo.
func (s *RecurringService) GetDue(ctx context.Context) ([]*models.RecurringTransaction, error) {
	recurrings, err := s.recurringRepo.GetDue(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get due recurring: %w", err)
	}
	return recurrings, nil
}

// ProcessDue memproses semua recurring yang jatuh tempo.
//
// Ini adalah method utama yang dipanggil oleh scheduler.
// Untuk setiap recurring yang due:
// 1. Generate transaction
// 2. Advance next_due ke periode berikutnya
//
// Return jumlah transaksi yang berhasil di-generate.
func (s *RecurringService) ProcessDue(ctx context.Context) (int, error) {
	recurrings, err := s.GetDue(ctx)
	if err != nil {
		return 0, err
	}

	processed := 0
	for _, recurring := range recurrings {
		// Generate transaction
		input := CreateTransactionInput{
			WalletID:    recurring.WalletID,
			CategoryID:  recurring.CategoryID,
			Type:        recurring.Type,
			Amount:      recurring.Amount,
			Description: recurring.Description,
			Date:        recurring.NextDue,
		}

		_, err := s.txService.Create(ctx, input)
		if err != nil {
			// Log error but continue with others
			fmt.Printf("Failed to process recurring %s: %v\n", recurring.ID, err)
			continue
		}

		// Advance next due
		recurring.AdvanceNextDue()
		if err := s.recurringRepo.Update(ctx, recurring); err != nil {
			fmt.Printf("Failed to update recurring %s: %v\n", recurring.ID, err)
			continue
		}

		processed++
	}

	return processed, nil
}

// Update memperbarui recurring.
func (s *RecurringService) Update(ctx context.Context, input UpdateRecurringInput) (*models.RecurringTransaction, error) {
	recurring, err := s.recurringRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring: %w", err)
	}

	if input.Amount != nil {
		recurring.Amount = *input.Amount
	}
	if input.Description != nil {
		recurring.Description = *input.Description
	}
	if input.NextDue != nil {
		recurring.NextDue = *input.NextDue
	}
	if input.EndDate != nil {
		recurring.EndDate = input.EndDate
	}
	if input.IsActive != nil {
		recurring.IsActive = *input.IsActive
	}

	if err := recurring.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.recurringRepo.Update(ctx, recurring); err != nil {
		return nil, fmt.Errorf("failed to update recurring: %w", err)
	}

	return recurring, nil
}

// Delete menghapus recurring.
func (s *RecurringService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.recurringRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete recurring: %w", err)
	}
	return nil
}

// Deactivate menonaktifkan recurring.
func (s *RecurringService) Deactivate(ctx context.Context, id uuid.UUID) error {
	isActive := false
	_, err := s.Update(ctx, UpdateRecurringInput{
		ID:       id,
		IsActive: &isActive,
	})
	return err
}

// CreateRecurringInput adalah input untuk membuat recurring.
type CreateRecurringInput struct {
	WalletID    uuid.UUID
	CategoryID  *uuid.UUID
	Type        models.TransactionType
	Amount      decimal.Decimal
	Description string
	Frequency   models.RecurringFrequency
	NextDue     time.Time
	EndDate     *time.Time
}

// UpdateRecurringInput adalah input untuk update recurring.
type UpdateRecurringInput struct {
	ID          uuid.UUID
	Amount      *decimal.Decimal
	Description *string
	NextDue     *time.Time
	EndDate     *time.Time
	IsActive    *bool
}
