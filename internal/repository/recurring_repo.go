package repository

import (
	"context"
	"time"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/google/uuid"
)

// RecurringRepository mendefinisikan operasi data access untuk RecurringTransaction.
type RecurringRepository interface {
	// Create menyimpan recurring transaction baru.
	Create(ctx context.Context, recurring *models.RecurringTransaction) error

	// GetByID mengambil recurring berdasarkan ID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.RecurringTransaction, error)

	// List mengambil semua recurring transactions dengan filter.
	List(ctx context.Context, filter RecurringFilter) ([]*models.RecurringTransaction, error)

	// GetDue mengambil recurring yang sudah jatuh tempo (next_due <= today).
	// Digunakan oleh scheduler untuk generate transactions.
	GetDue(ctx context.Context) ([]*models.RecurringTransaction, error)

	// Update memperbarui recurring.
	Update(ctx context.Context, recurring *models.RecurringTransaction) error

	// Delete menghapus recurring.
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateNextDue mengupdate next_due date setelah generate transaction.
	UpdateNextDue(ctx context.Context, id uuid.UUID, nextDue time.Time) error
}

// RecurringFilter adalah filter untuk query recurring transactions.
type RecurringFilter struct {
	// WalletID filter berdasarkan wallet.
	WalletID *uuid.UUID

	// IsActive filter berdasarkan status aktif.
	IsActive *bool

	// Type filter berdasarkan tipe (income/expense).
	Type *models.TransactionType

	// Frequency filter berdasarkan frekuensi.
	Frequency *models.RecurringFrequency
}
