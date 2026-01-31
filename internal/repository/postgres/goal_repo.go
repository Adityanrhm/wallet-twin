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

// goalRepository adalah implementasi PostgreSQL untuk GoalRepository.
type goalRepository struct {
	pool *pgxpool.Pool
}

// NewGoalRepository membuat GoalRepository baru.
func NewGoalRepository(pool *pgxpool.Pool) repository.GoalRepository {
	return &goalRepository{pool: pool}
}

// Create menyimpan goal baru.
func (r *goalRepository) Create(ctx context.Context, goal *models.Goal) error {
	query := `
		INSERT INTO goals (id, name, description, target_amount, current_amount, deadline, status, color, icon)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		goal.ID,
		goal.Name,
		goal.Description,
		goal.TargetAmount,
		goal.CurrentAmount,
		goal.Deadline,
		goal.Status,
		goal.Color,
		goal.Icon,
	)

	return convertError(err)
}

// GetByID mengambil goal berdasarkan ID.
func (r *goalRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Goal, error) {
	query := `
		SELECT id, name, description, target_amount, current_amount, deadline, status, color, icon, created_at, updated_at
		FROM goals
		WHERE id = $1
	`

	g := &models.Goal{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&g.ID,
		&g.Name,
		&g.Description,
		&g.TargetAmount,
		&g.CurrentAmount,
		&g.Deadline,
		&g.Status,
		&g.Color,
		&g.Icon,
		&g.CreatedAt,
		&g.UpdatedAt,
	)

	if err != nil {
		return nil, convertError(err)
	}

	return g, nil
}

// List mengambil goals dengan filter.
func (r *goalRepository) List(ctx context.Context, filter repository.GoalFilter) ([]*models.Goal, error) {
	query := `
		SELECT id, name, description, target_amount, current_amount, deadline, status, color, icon, created_at, updated_at
		FROM goals
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*filter.Status))
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

	var goals []*models.Goal
	for rows.Next() {
		g := &models.Goal{}
		err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.Description,
			&g.TargetAmount,
			&g.CurrentAmount,
			&g.Deadline,
			&g.Status,
			&g.Color,
			&g.Icon,
			&g.CreatedAt,
			&g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}

	return goals, rows.Err()
}

// Update memperbarui goal.
func (r *goalRepository) Update(ctx context.Context, goal *models.Goal) error {
	query := `
		UPDATE goals
		SET name = $2, description = $3, target_amount = $4, current_amount = $5, 
		    deadline = $6, status = $7, color = $8, icon = $9
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		goal.ID,
		goal.Name,
		goal.Description,
		goal.TargetAmount,
		goal.CurrentAmount,
		goal.Deadline,
		goal.Status,
		goal.Color,
		goal.Icon,
	)

	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete menghapus goal.
func (r *goalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM goals WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// AddContribution menambahkan kontribusi ke goal.
// Ini atomic operation yang juga update current_amount.
func (r *goalRepository) AddContribution(ctx context.Context, contribution *models.GoalContribution) error {
	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert contribution
	insertQuery := `
		INSERT INTO goal_contributions (id, goal_id, amount, note)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(ctx, insertQuery,
		contribution.ID,
		contribution.GoalID,
		contribution.Amount,
		contribution.Note,
	)
	if err != nil {
		return convertError(err)
	}

	// Update goal current_amount
	updateQuery := `
		UPDATE goals 
		SET current_amount = current_amount + $2
		WHERE id = $1
	`
	result, err := tx.Exec(ctx, updateQuery, contribution.GoalID, contribution.Amount)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return tx.Commit(ctx)
}

// GetContributions mengambil history kontribusi.
func (r *goalRepository) GetContributions(
	ctx context.Context,
	goalID uuid.UUID,
	params repository.ListParams,
) ([]*models.GoalContribution, error) {
	params.Validate()

	query := `
		SELECT id, goal_id, amount, note, created_at
		FROM goal_contributions
		WHERE goal_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, goalID, params.Limit, params.Offset)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var contributions []*models.GoalContribution
	for rows.Next() {
		c := &models.GoalContribution{}
		err := rows.Scan(
			&c.ID,
			&c.GoalID,
			&c.Amount,
			&c.Note,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		contributions = append(contributions, c)
	}

	return contributions, rows.Err()
}

// UpdateCurrentAmount mengupdate current_amount goal.
func (r *goalRepository) UpdateCurrentAmount(ctx context.Context, id uuid.UUID, amount decimal.Decimal) error {
	query := `UPDATE goals SET current_amount = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, amount)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
