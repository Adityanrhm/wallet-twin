package repository

import (
	"context"
	"time"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionRepository mendefinisikan operasi data access untuk Transaction.
//
// PENTING: Operasi Create, Update, Delete harus dikoordinasikan dengan
// wallet balance update. Gunakan TransactionManager untuk atomic operations.
type TransactionRepository interface {
	// Create menyimpan transaction baru.
	// TIDAK otomatis update wallet balance - harus dilakukan terpisah.
	Create(ctx context.Context, tx *models.Transaction) error

	// GetByID mengambil transaction berdasarkan ID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)

	// List mengambil transactions dengan filter.
	List(ctx context.Context, filter TransactionFilter, params ListParams) ([]*models.Transaction, error)

	// Update memperbarui transaction.
	Update(ctx context.Context, tx *models.Transaction) error

	// Delete menghapus transaction.
	Delete(ctx context.Context, id uuid.UUID) error

	// GetSummary menghitung total income dan expense untuk periode tertentu.
	// Berguna untuk dashboard dan reports.
	GetSummary(ctx context.Context, filter TransactionFilter) (*TransactionSummary, error)

	// GetByCategory menghitung total per kategori.
	// Berguna untuk pie chart breakdown.
	GetByCategory(ctx context.Context, filter TransactionFilter) ([]*CategorySummary, error)
}

// TransactionFilter adalah filter untuk query transactions.
//
//	// Transaksi bulan ini
//	filter := TransactionFilter{
//	    StartDate: ptr(firstDayOfMonth),
//	    EndDate:   ptr(lastDayOfMonth),
//	}
//
//	// Transaksi expense dari wallet tertentu
//	filter := TransactionFilter{
//	    WalletID: ptr(walletID),
//	    Type:     ptr(models.TransactionTypeExpense),
//	}
type TransactionFilter struct {
	// WalletID filter berdasarkan wallet.
	WalletID *uuid.UUID

	// CategoryID filter berdasarkan category.
	CategoryID *uuid.UUID

	// Type filter berdasarkan tipe (income/expense).
	Type *models.TransactionType

	// StartDate filter transaksi >= tanggal ini.
	StartDate *time.Time

	// EndDate filter transaksi <= tanggal ini.
	EndDate *time.Time

	// Search untuk full-text search di description.
	Search *string

	// Tags filter berdasarkan tags (ANY match).
	Tags []string
}

// TransactionSummary adalah ringkasan transaksi.
type TransactionSummary struct {
	// TotalIncome adalah total pemasukan.
	TotalIncome decimal.Decimal

	// TotalExpense adalah total pengeluaran.
	TotalExpense decimal.Decimal

	// Net adalah selisih (Income - Expense).
	Net decimal.Decimal

	// Count adalah jumlah transaksi.
	Count int
}

// CategorySummary adalah ringkasan per kategori.
type CategorySummary struct {
	// CategoryID adalah ID kategori.
	CategoryID uuid.UUID

	// CategoryName adalah nama kategori.
	CategoryName string

	// Total adalah total amount untuk kategori ini.
	Total decimal.Decimal

	// Count adalah jumlah transaksi.
	Count int

	// Percentage adalah persentase dari total.
	Percentage float64
}
