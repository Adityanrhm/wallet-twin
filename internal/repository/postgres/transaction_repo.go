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

// transactionRepository adalah implementasi PostgreSQL untuk TransactionRepository.
type transactionRepository struct {
	pool *pgxpool.Pool
}

// NewTransactionRepository membuat TransactionRepository baru.
func NewTransactionRepository(pool *pgxpool.Pool) repository.TransactionRepository {
	return &transactionRepository{pool: pool}
}

// Create menyimpan transaction baru.
func (r *transactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions 
			(id, wallet_id, category_id, type, amount, description, tags, transaction_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		tx.ID,
		tx.WalletID,
		tx.CategoryID,
		tx.Type,
		tx.Amount,
		tx.Description,
		tx.Tags,
		tx.TransactionDate,
	)

	return convertError(err)
}

// GetByID mengambil transaction berdasarkan ID.
func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	query := `
		SELECT id, wallet_id, category_id, type, amount, description, tags, 
		       transaction_date, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	tx := &models.Transaction{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tx.ID,
		&tx.WalletID,
		&tx.CategoryID,
		&tx.Type,
		&tx.Amount,
		&tx.Description,
		&tx.Tags,
		&tx.TransactionDate,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err != nil {
		return nil, convertError(err)
	}

	return tx, nil
}

// List mengambil transactions dengan filter.
func (r *transactionRepository) List(
	ctx context.Context,
	filter repository.TransactionFilter,
	params repository.ListParams,
) ([]*models.Transaction, error) {
	params.Validate()

	query := `
		SELECT id, wallet_id, category_id, type, amount, description, tags,
		       transaction_date, created_at, updated_at
		FROM transactions
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE clauses
	if filter.WalletID != nil {
		conditions = append(conditions, fmt.Sprintf("wallet_id = $%d", argIndex))
		args = append(args, *filter.WalletID)
		argIndex++
	}

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, string(*filter.Type))
		argIndex++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_date >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_date <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("description ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY transaction_date DESC, created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		tx := &models.Transaction{}
		err := rows.Scan(
			&tx.ID,
			&tx.WalletID,
			&tx.CategoryID,
			&tx.Type,
			&tx.Amount,
			&tx.Description,
			&tx.Tags,
			&tx.TransactionDate,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

// Update memperbarui transaction.
func (r *transactionRepository) Update(ctx context.Context, tx *models.Transaction) error {
	query := `
		UPDATE transactions
		SET wallet_id = $2, category_id = $3, type = $4, amount = $5, 
		    description = $6, tags = $7, transaction_date = $8
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		tx.ID,
		tx.WalletID,
		tx.CategoryID,
		tx.Type,
		tx.Amount,
		tx.Description,
		tx.Tags,
		tx.TransactionDate,
	)

	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete menghapus transaction.
func (r *transactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM transactions WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// GetSummary menghitung total income dan expense.
func (r *transactionRepository) GetSummary(
	ctx context.Context,
	filter repository.TransactionFilter,
) (*repository.TransactionSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as total_expense,
			COUNT(*) as count
		FROM transactions
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.WalletID != nil {
		conditions = append(conditions, fmt.Sprintf("wallet_id = $%d", argIndex))
		args = append(args, *filter.WalletID)
		argIndex++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_date >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_date <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	summary := &repository.TransactionSummary{}
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&summary.TotalIncome,
		&summary.TotalExpense,
		&summary.Count,
	)

	if err != nil {
		return nil, convertError(err)
	}

	summary.Net = summary.TotalIncome.Sub(summary.TotalExpense)

	return summary, nil
}

// GetByCategory menghitung total per kategori.
func (r *transactionRepository) GetByCategory(
	ctx context.Context,
	filter repository.TransactionFilter,
) ([]*repository.CategorySummary, error) {
	query := `
		SELECT 
			c.id,
			c.name,
			COALESCE(SUM(t.amount), 0) as total,
			COUNT(t.id) as count
		FROM categories c
		LEFT JOIN transactions t ON t.category_id = c.id
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Filter by transaction type
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("c.type = $%d", argIndex))
		args = append(args, string(*filter.Type))
		argIndex++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("(t.transaction_date >= $%d OR t.id IS NULL)", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("(t.transaction_date <= $%d OR t.id IS NULL)", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY c.id, c.name ORDER BY total DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var summaries []*repository.CategorySummary
	var grandTotal decimal.Decimal

	for rows.Next() {
		s := &repository.CategorySummary{}
		err := rows.Scan(&s.CategoryID, &s.CategoryName, &s.Total, &s.Count)
		if err != nil {
			return nil, err
		}
		grandTotal = grandTotal.Add(s.Total)
		summaries = append(summaries, s)
	}

	// Calculate percentages
	if !grandTotal.IsZero() {
		for _, s := range summaries {
			pct, _ := s.Total.Div(grandTotal).Mul(decimal.NewFromInt(100)).Float64()
			s.Percentage = pct
		}
	}

	return summaries, rows.Err()
}
