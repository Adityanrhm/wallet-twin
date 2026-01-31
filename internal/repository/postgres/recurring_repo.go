package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// recurringRepository adalah implementasi PostgreSQL untuk RecurringRepository.
type recurringRepository struct {
	pool *pgxpool.Pool
}

// NewRecurringRepository membuat RecurringRepository baru.
func NewRecurringRepository(pool *pgxpool.Pool) repository.RecurringRepository {
	return &recurringRepository{pool: pool}
}

// Create menyimpan recurring transaction baru.
func (r *recurringRepository) Create(ctx context.Context, recurring *models.RecurringTransaction) error {
	query := `
		INSERT INTO recurring_transactions 
			(id, wallet_id, category_id, type, amount, description, frequency, next_due, end_date, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		recurring.ID,
		recurring.WalletID,
		recurring.CategoryID,
		recurring.Type,
		recurring.Amount,
		recurring.Description,
		recurring.Frequency,
		recurring.NextDue,
		recurring.EndDate,
		recurring.IsActive,
	)

	return convertError(err)
}

// GetByID mengambil recurring berdasarkan ID.
func (r *recurringRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.RecurringTransaction, error) {
	query := `
		SELECT id, wallet_id, category_id, type, amount, description, frequency, 
		       next_due, end_date, is_active, created_at
		FROM recurring_transactions
		WHERE id = $1
	`

	rec := &models.RecurringTransaction{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rec.ID,
		&rec.WalletID,
		&rec.CategoryID,
		&rec.Type,
		&rec.Amount,
		&rec.Description,
		&rec.Frequency,
		&rec.NextDue,
		&rec.EndDate,
		&rec.IsActive,
		&rec.CreatedAt,
	)

	if err != nil {
		return nil, convertError(err)
	}

	return rec, nil
}

// List mengambil recurring transactions dengan filter.
func (r *recurringRepository) List(
	ctx context.Context,
	filter repository.RecurringFilter,
) ([]*models.RecurringTransaction, error) {
	query := `
		SELECT id, wallet_id, category_id, type, amount, description, frequency,
		       next_due, end_date, is_active, created_at
		FROM recurring_transactions
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.WalletID != nil {
		conditions = append(conditions, fmt.Sprintf("wallet_id = $%d", argIndex))
		args = append(args, *filter.WalletID)
		argIndex++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, string(*filter.Type))
		argIndex++
	}

	if filter.Frequency != nil {
		conditions = append(conditions, fmt.Sprintf("frequency = $%d", argIndex))
		args = append(args, string(*filter.Frequency))
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY next_due ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var recurrings []*models.RecurringTransaction
	for rows.Next() {
		rec := &models.RecurringTransaction{}
		err := rows.Scan(
			&rec.ID,
			&rec.WalletID,
			&rec.CategoryID,
			&rec.Type,
			&rec.Amount,
			&rec.Description,
			&rec.Frequency,
			&rec.NextDue,
			&rec.EndDate,
			&rec.IsActive,
			&rec.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		recurrings = append(recurrings, rec)
	}

	return recurrings, rows.Err()
}

// GetDue mengambil recurring yang jatuh tempo (next_due <= today).
func (r *recurringRepository) GetDue(ctx context.Context) ([]*models.RecurringTransaction, error) {
	query := `
		SELECT id, wallet_id, category_id, type, amount, description, frequency,
		       next_due, end_date, is_active, created_at
		FROM recurring_transactions
		WHERE is_active = true AND next_due <= CURRENT_DATE
		ORDER BY next_due ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var recurrings []*models.RecurringTransaction
	for rows.Next() {
		rec := &models.RecurringTransaction{}
		err := rows.Scan(
			&rec.ID,
			&rec.WalletID,
			&rec.CategoryID,
			&rec.Type,
			&rec.Amount,
			&rec.Description,
			&rec.Frequency,
			&rec.NextDue,
			&rec.EndDate,
			&rec.IsActive,
			&rec.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		recurrings = append(recurrings, rec)
	}

	return recurrings, rows.Err()
}

// Update memperbarui recurring.
func (r *recurringRepository) Update(ctx context.Context, recurring *models.RecurringTransaction) error {
	query := `
		UPDATE recurring_transactions
		SET wallet_id = $2, category_id = $3, type = $4, amount = $5, description = $6,
		    frequency = $7, next_due = $8, end_date = $9, is_active = $10
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		recurring.ID,
		recurring.WalletID,
		recurring.CategoryID,
		recurring.Type,
		recurring.Amount,
		recurring.Description,
		recurring.Frequency,
		recurring.NextDue,
		recurring.EndDate,
		recurring.IsActive,
	)

	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete menghapus recurring.
func (r *recurringRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM recurring_transactions WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// UpdateNextDue mengupdate next_due date.
func (r *recurringRepository) UpdateNextDue(ctx context.Context, id uuid.UUID, nextDue time.Time) error {
	query := `UPDATE recurring_transactions SET next_due = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, nextDue)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
