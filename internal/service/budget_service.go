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

// BudgetService menangani business logic untuk budget operations.
//
// Budget membantu user track pengeluaran per kategori.
// Service ini menghitung status budget (spent, remaining, progress).
type BudgetService struct {
	budgetRepo repository.BudgetRepository
	txRepo     repository.TransactionRepository
}

// NewBudgetService membuat BudgetService baru.
func NewBudgetService(
	budgetRepo repository.BudgetRepository,
	txRepo repository.TransactionRepository,
) *BudgetService {
	return &BudgetService{
		budgetRepo: budgetRepo,
		txRepo:     txRepo,
	}
}

// Create membuat budget baru.
func (s *BudgetService) Create(ctx context.Context, input CreateBudgetInput) (*models.Budget, error) {
	budget := &models.Budget{
		ID:         models.NewID(),
		CategoryID: input.CategoryID,
		Amount:     input.Amount,
		Period:     input.Period,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := budget.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.budgetRepo.Create(ctx, budget); err != nil {
		return nil, fmt.Errorf("failed to create budget: %w", err)
	}

	return budget, nil
}

// GetByID mengambil budget berdasarkan ID.
func (s *BudgetService) GetByID(ctx context.Context, id uuid.UUID) (*models.Budget, error) {
	budget, err := s.budgetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}
	return budget, nil
}

// GetByCategory mengambil budget aktif untuk kategori.
func (s *BudgetService) GetByCategory(ctx context.Context, categoryID uuid.UUID) (*models.Budget, error) {
	budget, err := s.budgetRepo.GetByCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}
	return budget, nil
}

// List mengambil semua budgets.
func (s *BudgetService) List(ctx context.Context, filter repository.BudgetFilter) ([]*models.Budget, error) {
	budgets, err := s.budgetRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}
	return budgets, nil
}

// ListActive mengambil semua budget aktif.
func (s *BudgetService) ListActive(ctx context.Context) ([]*models.Budget, error) {
	isActive := true
	return s.List(ctx, repository.BudgetFilter{IsActive: &isActive})
}

// GetAllStatus menghitung status semua budget aktif.
// Ini yang ditampilkan di dashboard.
func (s *BudgetService) GetAllStatus(ctx context.Context) ([]*repository.BudgetStatus, error) {
	statuses, err := s.budgetRepo.GetBudgetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget status: %w", err)
	}
	return statuses, nil
}

// GetStatus menghitung status budget tertentu.
func (s *BudgetService) GetStatus(ctx context.Context, id uuid.UUID) (*repository.BudgetStatus, error) {
	budget, err := s.budgetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	// Calculate spent amount
	filter := repository.TransactionFilter{
		CategoryID: &budget.CategoryID,
		StartDate:  &budget.StartDate,
	}
	if budget.EndDate != nil {
		filter.EndDate = budget.EndDate
	}

	expenseType := models.TransactionTypeExpense
	filter.Type = &expenseType

	summary, err := s.txRepo.GetSummary(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get spending: %w", err)
	}

	spent := summary.TotalExpense
	remaining := budget.Amount.Sub(spent)
	if remaining.IsNegative() {
		remaining = decimal.Zero
	}

	var progress float64
	if !budget.Amount.IsZero() {
		pct, _ := spent.Div(budget.Amount).Mul(decimal.NewFromInt(100)).Float64()
		progress = pct
	}

	return &repository.BudgetStatus{
		Budget:       budget,
		Spent:        spent,
		Remaining:    remaining,
		Progress:     progress,
		IsOverBudget: spent.GreaterThan(budget.Amount),
	}, nil
}

// Update memperbarui budget.
func (s *BudgetService) Update(ctx context.Context, input UpdateBudgetInput) (*models.Budget, error) {
	budget, err := s.budgetRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	if input.Amount != nil {
		budget.Amount = *input.Amount
	}
	if input.EndDate != nil {
		budget.EndDate = input.EndDate
	}
	if input.IsActive != nil {
		budget.IsActive = *input.IsActive
	}

	if err := budget.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, fmt.Errorf("failed to update budget: %w", err)
	}

	return budget, nil
}

// Delete menghapus budget.
func (s *BudgetService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.budgetRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}
	return nil
}

// CreateBudgetInput adalah input untuk membuat budget.
type CreateBudgetInput struct {
	CategoryID uuid.UUID
	Amount     decimal.Decimal
	Period     models.BudgetPeriod
	StartDate  time.Time
	EndDate    *time.Time
}

// UpdateBudgetInput adalah input untuk update budget.
type UpdateBudgetInput struct {
	ID       uuid.UUID
	Amount   *decimal.Decimal
	EndDate  *time.Time
	IsActive *bool
}
