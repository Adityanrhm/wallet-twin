package repository

import (
	"context"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/google/uuid"
)

// CategoryRepository mendefinisikan operasi data access untuk Category.
//
// Category bisa memiliki parent (untuk hierarki).
// Method GetByType paling sering digunakan untuk mendapatkan
// kategori income atau expense.
type CategoryRepository interface {
	// Create menyimpan category baru.
	Create(ctx context.Context, category *models.Category) error

	// GetByID mengambil category berdasarkan ID.
	GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error)

	// GetByType mengambil semua kategori berdasarkan tipe (income/expense).
	// Ini yang paling sering digunakan untuk populate dropdown.
	GetByType(ctx context.Context, catType models.CategoryType) ([]*models.Category, error)

	// GetChildren mengambil sub-kategori dari parent category.
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*models.Category, error)

	// List mengambil semua kategori.
	// Diurutkan berdasarkan type, sort_order.
	List(ctx context.Context) ([]*models.Category, error)

	// Update memperbarui category.
	Update(ctx context.Context, category *models.Category) error

	// Delete menghapus category.
	// Akan error jika masih ada transaksi yang menggunakan category ini.
	Delete(ctx context.Context, id uuid.UUID) error
}
