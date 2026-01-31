package repository

import (
	"context"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BudgetRepository mendefinisikan operasi data access untuk Budget.
type BudgetRepository interface {
	// Create menyimpan budget baru.
	Create(ctx context.Context, budget *models.Budget) error

	// GetByID mengambil budget berdasarkan ID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Budget, error)

	// GetByCategory mengambil budget aktif untuk kategori tertentu.
	GetByCategory(ctx context.Context, categoryID uuid.UUID) (*models.Budget, error)

	// List mengambil semua budgets dengan filter.
	List(ctx context.Context, filter BudgetFilter) ([]*models.Budget, error)

	// Update memperbarui budget.
	Update(ctx context.Context, budget *models.Budget) error

	// Delete menghapus budget.
	Delete(ctx context.Context, id uuid.UUID) error

	// GetBudgetStatus menghitung status semua budget aktif.
	// Membandingkan budget amount dengan actual spending.
	GetBudgetStatus(ctx context.Context) ([]*BudgetStatus, error)
}

// BudgetFilter adalah filter untuk query budgets.
type BudgetFilter struct {
	// IsActive filter berdasarkan status aktif.
	IsActive *bool

	// CategoryID filter berdasarkan kategori.
	CategoryID *uuid.UUID

	// Period filter berdasarkan periode.
	Period *models.BudgetPeriod
}

// BudgetStatus adalah status budget dengan actual spending.
type BudgetStatus struct {
	// Budget adalah data budget.
	Budget *models.Budget

	// CategoryName adalah nama kategori.
	CategoryName string

	// CategoryIcon adalah icon kategori.
	CategoryIcon string

	// Spent adalah jumlah yang sudah dikeluarkan.
	Spent decimal.Decimal

	// Remaining adalah sisa budget (Amount - Spent).
	Remaining decimal.Decimal

	// Progress adalah persentase (0-100+).
	Progress float64

	// IsOverBudget true jika Spent > Amount.
	IsOverBudget bool
}
