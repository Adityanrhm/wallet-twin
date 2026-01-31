// Package postgres berisi implementasi repository menggunakan PostgreSQL.
//
// Setiap file dalam package ini mengimplementasikan interface
// yang didefinisikan di package repository parent.
//
// Naming convention:
// - wallet.go mengimplementasi repository.WalletRepository
// - category.go mengimplementasi repository.CategoryRepository
// - dst.
//
// Semua implementasi menggunakan pgx untuk query.
// Pattern yang digunakan:
//
//	func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
//	    // 1. Define query
//	    query := `SELECT id, name, balance FROM wallets WHERE id = $1`
//
//	    // 2. Execute query
//	    row := r.pool.QueryRow(ctx, query, id)
//
//	    // 3. Scan result ke struct
//	    var wallet models.Wallet
//	    err := row.Scan(&wallet.ID, &wallet.Name, &wallet.Balance)
//	    if err != nil {
//	        if errors.Is(err, pgx.ErrNoRows) {
//	            return nil, repository.ErrNotFound
//	        }
//	        return nil, err
//	    }
//
//	    return &wallet, nil
//	}
package postgres

// TODO: Add implementation files
