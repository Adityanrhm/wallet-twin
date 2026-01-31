// Package models berisi entity/domain objects untuk aplikasi Wallet Twin.
//
// Dalam Clean Architecture, models (entities) adalah:
// - Pure business objects tanpa dependencies ke layer lain
// - Berisi business rules dan validation
// - Tidak tahu tentang database, UI, atau framework apapun
//
// File ini berisi common types yang digunakan oleh semua entities.
package models

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel adalah struct yang di-embed oleh semua entities.
// Berisi fields yang common untuk semua entities:
// - ID: Primary key UUID
// - CreatedAt: Waktu record dibuat
// - UpdatedAt: Waktu terakhir record diupdate
//
// Cara penggunaan (embedding):
//
//	type Wallet struct {
//	    BaseModel  // ID, CreatedAt, UpdatedAt otomatis tersedia
//	    Name    string
//	    Balance decimal.Decimal
//	}
//
// Keuntungan embedding:
// - DRY: Tidak perlu duplikasi fields di setiap struct
// - Konsisten: Semua entities punya fields yang sama
// - Mudah extend: Tambah field di BaseModel, semua entities dapat
type BaseModel struct {
	// ID adalah unique identifier menggunakan UUID v4.
	//
	// Kenapa UUID bukan auto-increment integer?
	// 1. Security: Tidak bisa di-enumerate (coba ID 1, 2, 3...)
	// 2. Distributed: Bisa generate ID tanpa koordinasi dengan DB
	// 3. Unique globally: Aman untuk merge data dari berbagai sources
	// 4. URL-safe: Bisa dipakai langsung di URL
	//
	// Format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	// Contoh: 550e8400-e29b-41d4-a716-446655440000
	ID uuid.UUID `json:"id" db:"id"`

	// CreatedAt adalah waktu ketika record pertama kali dibuat.
	// Di-set otomatis oleh database (DEFAULT NOW()).
	// Tidak boleh diubah setelah record dibuat.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt adalah waktu terakhir record diupdate.
	// Di-set otomatis oleh trigger di database.
	// Berguna untuk:
	// - Audit trail
	// - Optimistic locking
	// - Cache invalidation
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewID generates a new UUID v4 for entity creation.
//
// Gunakan saat membuat entity baru:
//
//	wallet := &models.Wallet{
//	    BaseModel: models.BaseModel{ID: models.NewID()},
//	    Name:      "Cash",
//	}
func NewID() uuid.UUID {
	return uuid.New()
}

// IsZero checks if the ID is zero/empty.
// Berguna untuk cek apakah entity sudah di-persist ke database.
//
//	if wallet.ID.IsZero() {
//	    // Entity baru, belum ada di database
//	    return repo.Create(wallet)
//	}
//	// Entity sudah ada, update
//	return repo.Update(wallet)
func (b *BaseModel) IsZero() bool {
	return b.ID == uuid.Nil
}
