package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// budgetRepository adalah implementasi PostgreSQL untuk BudgetRepository.
type budgetRepository struct {
	pool *pgxpool.Pool
}

// NewBudgetRepository membuat BudgetRepository baru.
func NewBudgetRepository(pool *pgxpool.Pool) repository.BudgetRepository {
	return &budgetRepository{pool: pool}
}

// Create menyimpan budget baru.
func (r *budgetRepository) Create(ctx context.Context, budget *models.Budget) error {
	query := `
		INSERT INTO budgets (id, category_id, amount, period, start_date, end_date, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		budget.ID,
		budget.CategoryID,
		budget.Amount,
		budget.Period,
		budget.StartDate,
		budget.EndDate,
		budget.IsActive,
	)

	return convertError(err)
}

// GetByID mengambil budget berdasarkan ID.
func (r *budgetRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Budget, error) {
	query := `
		SELECT id, category_id, amount, period, start_date, end_date, is_active, created_at
		FROM budgets
		WHERE id = $1
	`

	b := &models.Budget{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.CategoryID,
		&b.Amount,
		&b.Period,
		&b.StartDate,
		&b.EndDate,
		&b.IsActive,
		&b.CreatedAt,
	)

	if err != nil {
		return nil, convertError(err)
	}

	return b, nil
}

// GetByCategory mengambil budget aktif untuk kategori.
func (r *budgetRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID) (*models.Budget, error) {
	query := `
		SELECT id, category_id, amount, period, start_date, end_date, is_active, created_at
		FROM budgets
		WHERE category_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1
	`

	b := &models.Budget{}
	err := r.pool.QueryRow(ctx, query, categoryID).Scan(
		&b.ID,
		&b.CategoryID,
		&b.Amount,
		&b.Period,
		&b.StartDate,
		&b.EndDate,
		&b.IsActive,
		&b.CreatedAt,
	)

	if err != nil {
		return nil, convertError(err)
	}

	return b, nil
}

// List mengambil budgets dengan filter.
func (r *budgetRepository) List(ctx context.Context, filter repository.BudgetFilter) ([]*models.Budget, error) {
	query := `
		SELECT id, category_id, amount, period, start_date, end_date, is_active, created_at
		FROM budgets
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	if filter.Period != nil {
		conditions = append(conditions, fmt.Sprintf("period = $%d", argIndex))
		args = append(args, string(*filter.Period))
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var budgets []*models.Budget
	for rows.Next() {
		b := &models.Budget{}
		err := rows.Scan(
			&b.ID,
			&b.CategoryID,
			&b.Amount,
			&b.Period,
			&b.StartDate,
			&b.EndDate,
			&b.IsActive,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}

	return budgets, rows.Err()
}

// Update memperbarui budget.
func (r *budgetRepository) Update(ctx context.Context, budget *models.Budget) error {
	query := `
		UPDATE budgets
		SET category_id = $2, amount = $3, period = $4, start_date = $5, end_date = $6, is_active = $7
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		budget.ID,
		budget.CategoryID,
		budget.Amount,
		budget.Period,
		budget.StartDate,
		budget.EndDate,
		budget.IsActive,
	)

	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete menghapus budget.
func (r *budgetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM budgets WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// GetBudgetStatus menghitung status semua budget aktif.
func (r *budgetRepository) GetBudgetStatus(ctx context.Context) ([]*repository.BudgetStatus, error) {
	query := `
		SELECT 
			b.id, b.category_id, b.amount, b.period, b.start_date, b.end_date, b.is_active, b.created_at,
			c.name as category_name,
			COALESCE(c.icon, '') as category_icon,
			COALESCE(
				(SELECT SUM(t.amount) 
				 FROM transactions t 
				 WHERE t.category_id = b.category_id 
				   AND t.type = 'expense'
				   AND t.transaction_date >= b.start_date
				   AND (b.end_date IS NULL OR t.transaction_date <= b.end_date)
				), 0
			) as spent
		FROM budgets b
		JOIN categories c ON c.id = b.category_id
		WHERE b.is_active = true
		ORDER BY b.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var statuses []*repository.BudgetStatus
	for rows.Next() {
		b := &models.Budget{}
		s := &repository.BudgetStatus{Budget: b}

		err := rows.Scan(
			&b.ID,
			&b.CategoryID,
			&b.Amount,
			&b.Period,
			&b.StartDate,
			&b.EndDate,
			&b.IsActive,
			&b.CreatedAt,
			&s.CategoryName,
			&s.CategoryIcon,
			&s.Spent,
		)
		if err != nil {
			return nil, err
		}

		// Calculate remaining and progress
		s.Remaining = b.Amount.Sub(s.Spent)
		if s.Remaining.IsNegative() {
			s.Remaining = decimal.Zero
		}

		if !b.Amount.IsZero() {
			pct, _ := s.Spent.Div(b.Amount).Mul(decimal.NewFromInt(100)).Float64()
			s.Progress = pct
		}

		s.IsOverBudget = s.Spent.GreaterThan(b.Amount)

		statuses = append(statuses, s)
	}

	return statuses, rows.Err()
}
