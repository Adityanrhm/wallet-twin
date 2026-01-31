package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// transferRepository adalah implementasi PostgreSQL untuk TransferRepository.
type transferRepository struct {
	pool *pgxpool.Pool
}

// NewTransferRepository membuat TransferRepository baru.
func NewTransferRepository(pool *pgxpool.Pool) repository.TransferRepository {
	return &transferRepository{pool: pool}
}

// Create menyimpan transfer baru.
func (r *transferRepository) Create(ctx context.Context, transfer *models.Transfer) error {
	query := `
		INSERT INTO transfers (id, from_wallet_id, to_wallet_id, amount, fee, note)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		transfer.ID,
		transfer.FromWalletID,
		transfer.ToWalletID,
		transfer.Amount,
		transfer.Fee,
		transfer.Note,
	)

	return convertError(err)
}

// GetByID mengambil transfer berdasarkan ID.
func (r *transferRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error) {
	query := `
		SELECT id, from_wallet_id, to_wallet_id, amount, fee, note, created_at
		FROM transfers
		WHERE id = $1
	`

	t := &models.Transfer{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.FromWalletID,
		&t.ToWalletID,
		&t.Amount,
		&t.Fee,
		&t.Note,
		&t.CreatedAt,
	)

	if err != nil {
		return nil, convertError(err)
	}

	return t, nil
}

// List mengambil transfers dengan filter.
func (r *transferRepository) List(
	ctx context.Context,
	filter repository.TransferFilter,
	params repository.ListParams,
) ([]*models.Transfer, error) {
	params.Validate()

	query := `
		SELECT id, from_wallet_id, to_wallet_id, amount, fee, note, created_at
		FROM transfers
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// WalletID = from OR to
	if filter.WalletID != nil {
		conditions = append(conditions, fmt.Sprintf("(from_wallet_id = $%d OR to_wallet_id = $%d)", argIndex, argIndex))
		args = append(args, *filter.WalletID)
		argIndex++
	}

	if filter.FromWalletID != nil {
		conditions = append(conditions, fmt.Sprintf("from_wallet_id = $%d", argIndex))
		args = append(args, *filter.FromWalletID)
		argIndex++
	}

	if filter.ToWalletID != nil {
		conditions = append(conditions, fmt.Sprintf("to_wallet_id = $%d", argIndex))
		args = append(args, *filter.ToWalletID)
		argIndex++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var transfers []*models.Transfer
	for rows.Next() {
		t := &models.Transfer{}
		err := rows.Scan(
			&t.ID,
			&t.FromWalletID,
			&t.ToWalletID,
			&t.Amount,
			&t.Fee,
			&t.Note,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}

	return transfers, rows.Err()
}
