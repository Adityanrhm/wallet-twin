// Package postgres berisi implementasi PostgreSQL untuk repository interfaces.
//
// Package ini adalah "adapter" dalam terminology Clean Architecture.
// Mengimplementasikan interface dari package repository menggunakan PostgreSQL.
//
// Semua implementasi menggunakan pgxpool untuk connection pooling.
// pgx adalah PostgreSQL driver yang lebih performant dari database/sql.
//
// Pattern yang digunakan:
//
// 1. Struct dengan pool: Setiap repository struct menyimpan reference ke pool.
//
//	type walletRepository struct {
//	    pool *pgxpool.Pool
//	}
//
// 2. Constructor dengan pool injection:
//
//	func NewWalletRepository(pool *pgxpool.Pool) repository.WalletRepository {
//	    return &walletRepository{pool: pool}
//	}
//
// 3. Query methods menggunakan pool:
//
//	func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
//	    row := r.pool.QueryRow(ctx, "SELECT ... FROM wallets WHERE id = $1", id)
//	    // scan result...
//	}
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// Error codes PostgreSQL yang umum digunakan.
// Ref: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	// PGErrUniqueViolation adalah error code untuk duplicate key.
	PGErrUniqueViolation = "23505"

	// PGErrForeignKeyViolation adalah error code untuk FK violation.
	PGErrForeignKeyViolation = "23503"

	// PGErrNotNullViolation adalah error code untuk not null violation.
	PGErrNotNullViolation = "23502"
)

// TransactionManager adalah implementasi PostgreSQL untuk repository.TransactionManager.
//
// Digunakan untuk operasi atomic yang melibatkan multiple repositories.
// Contoh: Transfer antar wallet harus update 2 wallet + create transfer record.
//
//	err := txManager.WithTransaction(ctx, func(ctx context.Context) error {
//	    // Semua operasi di sini dalam satu transaction
//	    return nil
//	})
type TransactionManager struct {
	pool *pgxpool.Pool
}

// NewTransactionManager membuat TransactionManager baru.
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// WithTransaction menjalankan fn dalam database transaction.
//
// Flow:
// 1. Begin transaction
// 2. Execute fn dengan context yang menyimpan tx
// 3. Jika fn return error -> Rollback
// 4. Jika fn return nil -> Commit
//
// PENTING: Repository implementations harus check context untuk transaction.
// Jika ada transaction di context, gunakan tx tersebut bukan pool.
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn repository.TxFunc) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}

	// Defer rollback - no-op if already committed
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Store tx in context
	ctx = context.WithValue(ctx, txKey{}, tx)

	// Execute function
	if err = fn(ctx); err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit(ctx)
}

// txKey adalah key untuk menyimpan transaction di context.
type txKey struct{}

// GetTx mengambil transaction dari context.
// Return nil jika tidak ada transaction.
func GetTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

// convertError mengkonversi PostgreSQL error ke repository error.
// Ini membantu abstraksi sehingga caller tidak perlu depend pada pgx errors.
func convertError(err error) error {
	if err == nil {
		return nil
	}

	// Check for "no rows"
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrNotFound
	}

	// Check PostgreSQL specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PGErrUniqueViolation:
			return repository.ErrDuplicateKey
		case PGErrForeignKeyViolation:
			return repository.ErrForeignKeyViolation
		}
	}

	return err
}
