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

// walletRepository adalah implementasi PostgreSQL untuk WalletRepository.
type walletRepository struct {
	pool *pgxpool.Pool
}

// NewWalletRepository membuat WalletRepository baru.
//
// Contoh penggunaan:
//
//	pool, _ := pgxpool.New(ctx, connString)
//	walletRepo := postgres.NewWalletRepository(pool)
//
//	wallet := models.NewWallet("Cash", models.WalletTypeCash)
//	err := walletRepo.Create(ctx, wallet)
func NewWalletRepository(pool *pgxpool.Pool) repository.WalletRepository {
	return &walletRepository{pool: pool}
}

// Create menyimpan wallet baru ke database.
//
// SQL yang dieksekusi:
//
//	INSERT INTO wallets (id, name, type, balance, currency, color, icon, is_active, created_at, updated_at)
//	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
func (r *walletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	query := `
		INSERT INTO wallets (id, name, type, balance, currency, color, icon, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		wallet.ID,
		wallet.Name,
		wallet.Type,
		wallet.Balance,
		wallet.Currency,
		wallet.Color,
		wallet.Icon,
		wallet.IsActive,
	)

	return convertError(err)
}

// GetByID mengambil wallet berdasarkan ID.
//
// Return repository.ErrNotFound jika tidak ditemukan.
func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	query := `
		SELECT id, name, type, balance, currency, color, icon, is_active, created_at, updated_at
		FROM wallets
		WHERE id = $1
	`

	wallet := &models.Wallet{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&wallet.ID,
		&wallet.Name,
		&wallet.Type,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.Color,
		&wallet.Icon,
		&wallet.IsActive,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		return nil, convertError(err)
	}

	return wallet, nil
}

// List mengambil wallets dengan filter.
//
// Filter bersifat optional. Jika nil, tidak difilter.
// Hasil diurutkan berdasarkan created_at DESC.
func (r *walletRepository) List(ctx context.Context, filter repository.WalletFilter) ([]*models.Wallet, error) {
	// Build query dinamis dengan WHERE clauses
	query := `
		SELECT id, name, type, balance, currency, color, icon, is_active, created_at, updated_at
		FROM wallets
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE clauses berdasarkan filter
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

	if filter.Currency != nil {
		conditions = append(conditions, fmt.Sprintf("currency = $%d", argIndex))
		args = append(args, *filter.Currency)
		argIndex++
	}

	// Append WHERE clause jika ada conditions
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	// Execute query
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	// Scan results
	var wallets []*models.Wallet
	for rows.Next() {
		wallet := &models.Wallet{}
		err := rows.Scan(
			&wallet.ID,
			&wallet.Name,
			&wallet.Type,
			&wallet.Balance,
			&wallet.Currency,
			&wallet.Color,
			&wallet.Icon,
			&wallet.IsActive,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}

	return wallets, rows.Err()
}

// Update memperbarui wallet.
//
// PENTING: updated_at dihandle oleh trigger di database.
func (r *walletRepository) Update(ctx context.Context, wallet *models.Wallet) error {
	query := `
		UPDATE wallets
		SET name = $2, type = $3, balance = $4, currency = $5, color = $6, icon = $7, is_active = $8
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		wallet.ID,
		wallet.Name,
		wallet.Type,
		wallet.Balance,
		wallet.Currency,
		wallet.Color,
		wallet.Icon,
		wallet.IsActive,
	)

	if err != nil {
		return convertError(err)
	}

	// Check if wallet was found
	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete melakukan soft delete (set is_active = false).
//
// Soft delete digunakan karena:
// 1. Preserve referential integrity (transaksi tetap punya wallet_id valid)
// 2. Data bisa di-recover jika diperlukan
// 3. Untuk reporting historical data
func (r *walletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE wallets SET is_active = false WHERE id = $1 AND is_active = true`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// UpdateBalance mengupdate saldo wallet secara atomic.
//
// Operasi ini menggunakan query langsung tanpa read-modify-write
// untuk menghindari race condition pada concurrent access.
func (r *walletRepository) UpdateBalance(ctx context.Context, id uuid.UUID, newBalance decimal.Decimal) error {
	query := `UPDATE wallets SET balance = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, newBalance)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// GetTotalBalance menghitung total saldo semua wallet aktif.
//
// Query menggunakan COALESCE untuk handle case jika tidak ada wallet.
func (r *walletRepository) GetTotalBalance(ctx context.Context) (decimal.Decimal, error) {
	query := `SELECT COALESCE(SUM(balance), 0) FROM wallets WHERE is_active = true`

	var total decimal.Decimal
	err := r.pool.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return decimal.Zero, convertError(err)
	}

	return total, nil
}
