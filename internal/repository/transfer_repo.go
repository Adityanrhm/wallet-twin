package repository

import (
	"context"
	"time"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/google/uuid"
)

// TransferRepository mendefinisikan operasi data access untuk Transfer.
type TransferRepository interface {
	// Create menyimpan transfer baru.
	// TIDAK otomatis update wallet balances - harus dalam transaction.
	Create(ctx context.Context, transfer *models.Transfer) error

	// GetByID mengambil transfer berdasarkan ID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error)

	// List mengambil transfers dengan filter.
	List(ctx context.Context, filter TransferFilter, params ListParams) ([]*models.Transfer, error)
}

// TransferFilter adalah filter untuk query transfers.
type TransferFilter struct {
	// WalletID filter transfer yang melibatkan wallet ini (from OR to).
	WalletID *uuid.UUID

	// FromWalletID filter berdasarkan wallet sumber.
	FromWalletID *uuid.UUID

	// ToWalletID filter berdasarkan wallet tujuan.
	ToWalletID *uuid.UUID

	// StartDate filter transfer >= tanggal ini.
	StartDate *time.Time

	// EndDate filter transfer <= tanggal ini.
	EndDate *time.Time
}
