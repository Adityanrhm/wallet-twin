// Package repository mendefinisikan interface untuk data access.
//
// Repository Pattern adalah design pattern yang memisahkan business logic
// dari data access logic. Ini memberikan beberapa keuntungan:
//
// 1. TESTABILITY: Bisa inject mock repository saat unit testing
// 2. FLEXIBILITY: Mudah swap database (PostgreSQL → MySQL)
// 3. SEPARATION OF CONCERNS: Business logic terpisah dari SQL
//
// Dalam package ini hanya ada INTERFACE.
// Implementation ada di sub-package (misalnya: repository/postgres).
//
// Contoh penggunaan:
//
//	// Service hanya depend pada interface, bukan implementation
//	type WalletService struct {
//	    repo repository.WalletRepository  // Interface
//	}
//
//	// Di production, inject PostgreSQL implementation
//	walletRepo := postgres.NewWalletRepository(db.Pool)
//	service := &WalletService{repo: walletRepo}
//
//	// Di test, inject mock
//	mockRepo := &MockWalletRepository{}
//	service := &WalletService{repo: mockRepo}
package repository

import (
	"context"
	"errors"
)

// Common errors yang bisa terjadi di semua repositories.
// Gunakan errors.Is() untuk compare.
//
//	if errors.Is(err, repository.ErrNotFound) {
//	    // Handle not found
//	}
var (
	// ErrNotFound dikembalikan ketika record tidak ditemukan.
	ErrNotFound = errors.New("record not found")

	// ErrDuplicateKey dikembalikan ketika insert/update violate unique constraint.
	ErrDuplicateKey = errors.New("duplicate key violation")

	// ErrForeignKeyViolation dikembalikan ketika foreign key tidak valid.
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Querier adalah interface untuk database operations.
// Ini memungkinkan repository methods bekerja dengan:
// - *pgxpool.Pool (untuk operasi normal)
// - pgx.Tx (untuk operasi dalam transaction)
//
// Pattern ini penting untuk atomic operations:
//
//	tx, err := db.Pool.Begin(ctx)
//	if err != nil {
//	    return err
//	}
//	defer tx.Rollback(ctx)
//
//	// Semua operasi dalam transaction yang sama
//	err = walletRepo.UpdateBalance(ctx, tx, walletID, newBalance)
//	err = transferRepo.Create(ctx, tx, transfer)
//
//	return tx.Commit(ctx)
type Querier interface {
	// Exec untuk INSERT, UPDATE, DELETE
	// Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)

	// Query untuk SELECT multiple rows
	// Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// QueryRow untuk SELECT single row
	// QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// ListParams adalah parameter umum untuk list/pagination.
//
//	params := repository.ListParams{
//	    Limit:  20,
//	    Offset: 0,
//	}
type ListParams struct {
	// Limit adalah jumlah maksimal records yang dikembalikan.
	// Default: 20, Max: 100
	Limit int

	// Offset untuk pagination.
	// Skip N records pertama.
	Offset int
}

// DefaultListParams mengembalikan default pagination params.
func DefaultListParams() ListParams {
	return ListParams{
		Limit:  20,
		Offset: 0,
	}
}

// Validate memvalidasi dan sanitize list params.
func (p *ListParams) Validate() {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}

// TxFunc adalah function yang akan dijalankan dalam transaction.
// Digunakan oleh TransactionManager.
type TxFunc func(ctx context.Context) error

// TransactionManager adalah interface untuk mengelola database transactions.
// Ini abstraction untuk atomic operations yang melibatkan multiple repositories.
//
//	err := txManager.WithTransaction(ctx, func(ctx context.Context) error {
//	    // Semua operasi di sini dalam satu transaction
//	    if err := walletRepo.UpdateBalance(ctx, ...); err != nil {
//	        return err  // Akan rollback
//	    }
//	    if err := transferRepo.Create(ctx, ...); err != nil {
//	        return err  // Akan rollback
//	    }
//	    return nil  // Akan commit
//	})
type TransactionManager interface {
	// WithTransaction menjalankan fn dalam database transaction.
	// Jika fn return error, transaction di-rollback.
	// Jika fn return nil, transaction di-commit.
	WithTransaction(ctx context.Context, fn TxFunc) error
}
