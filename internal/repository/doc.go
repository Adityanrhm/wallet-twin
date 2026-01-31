// Package repository mendefinisikan interface untuk data access.
//
// Repository Pattern:
// - Memisahkan business logic dari data access logic
// - Membuat services testable dengan mock repositories
// - Memungkinkan swap database tanpa mengubah business logic
//
// Dalam package ini, kita hanya mendefinisikan INTERFACE.
// Implementation (PostgreSQL) ada di sub-package repository/postgres.
//
// Contoh:
//
//	// Interface di repository/wallet_repo.go
//	type WalletRepository interface {
//	    Create(ctx context.Context, wallet *models.Wallet) error
//	    GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
//	}
//
//	// Implementation di repository/postgres/wallet.go
//	type walletRepository struct {
//	    db *pgxpool.Pool
//	}
//
//	func (r *walletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
//	    // PostgreSQL specific implementation
//	}
//
// Ini memungkinkan kita untuk:
// 1. Test services dengan mock repository
// 2. Ganti PostgreSQL ke MySQL tanpa ubah services
// 3. Unit test tanpa database
package repository

// TODO: Add base repository interface dan common types
