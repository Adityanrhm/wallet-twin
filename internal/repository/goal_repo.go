package repository

import (
	"context"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GoalRepository mendefinisikan operasi data access untuk Goal.
type GoalRepository interface {
	// Create menyimpan goal baru.
	Create(ctx context.Context, goal *models.Goal) error

	// GetByID mengambil goal berdasarkan ID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Goal, error)

	// List mengambil semua goals dengan filter.
	List(ctx context.Context, filter GoalFilter) ([]*models.Goal, error)

	// Update memperbarui goal.
	Update(ctx context.Context, goal *models.Goal) error

	// Delete menghapus goal.
	Delete(ctx context.Context, id uuid.UUID) error

	// AddContribution menambahkan kontribusi ke goal.
	// Ini atomic operation yang juga update current_amount.
	AddContribution(ctx context.Context, contribution *models.GoalContribution) error

	// GetContributions mengambil history kontribusi untuk goal.
	GetContributions(ctx context.Context, goalID uuid.UUID, params ListParams) ([]*models.GoalContribution, error)

	// UpdateCurrentAmount mengupdate current_amount goal.
	UpdateCurrentAmount(ctx context.Context, id uuid.UUID, amount decimal.Decimal) error
}

// GoalFilter adalah filter untuk query goals.
type GoalFilter struct {
	// Status filter berdasarkan status.
	Status *models.GoalStatus
}
