package repository

import (
	"context"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WalletRepository mendefinisikan operasi data access untuk Wallet.
//
// Interface ini mendefinisikan CONTRACT yang harus dipenuhi oleh
// implementation apapun (PostgreSQL, MySQL, in-memory, mock, dll).
//
// Semua method menerima context.Context sebagai parameter pertama.
// Ini penting untuk:
// - Request cancellation
// - Timeout handling
// - Passing request-scoped values
//
// Contoh penggunaan:
//
//	// Get wallet by ID
//	wallet, err := repo.GetByID(ctx, walletID)
//	if errors.Is(err, repository.ErrNotFound) {
//	    return fmt.Errorf("wallet not found")
//	}
//
//	// List all active wallets
//	wallets, err := repo.List(ctx, repository.WalletFilter{IsActive: ptr(true)})
type WalletRepository interface {
	// Create menyimpan wallet baru ke database.
	// Wallet.ID harus sudah di-set sebelum memanggil Create.
	// Return error jika wallet dengan ID yang sama sudah ada.
	Create(ctx context.Context, wallet *models.Wallet) error

	// GetByID mengambil wallet berdasarkan ID.
	// Return ErrNotFound jika wallet tidak ditemukan.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)

	// List mengambil semua wallets dengan filter opsional.
	// Wallets diurutkan berdasarkan created_at DESC.
	List(ctx context.Context, filter WalletFilter) ([]*models.Wallet, error)

	// Update memperbarui wallet yang sudah ada.
	// Hanya field yang berubah yang di-update.
	// Return ErrNotFound jika wallet tidak ditemukan.
	Update(ctx context.Context, wallet *models.Wallet) error

	// Delete menghapus wallet (soft delete - set is_active = false).
	// Return ErrNotFound jika wallet tidak ditemukan.
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateBalance mengupdate saldo wallet.
	// Ini adalah atomic operation - aman untuk concurrent access.
	// Digunakan saat ada transaksi income/expense.
	UpdateBalance(ctx context.Context, id uuid.UUID, newBalance decimal.Decimal) error

	// GetTotalBalance menghitung total saldo semua wallet aktif.
	// Berguna untuk dashboard summary.
	GetTotalBalance(ctx context.Context) (decimal.Decimal, error)
}

// WalletFilter adalah filter untuk query wallets.
// Semua field adalah optional (pointer).
// nil berarti tidak di-filter.
//
//	// Hanya wallet aktif
//	filter := WalletFilter{IsActive: ptr(true)}
//
//	// Hanya wallet tipe bank
//	filter := WalletFilter{Type: ptr(models.WalletTypeBank)}
type WalletFilter struct {
	// IsActive filter berdasarkan status aktif.
	IsActive *bool

	// Type filter berdasarkan tipe wallet.
	Type *models.WalletType

	// Currency filter berdasarkan mata uang.
	Currency *string
}
