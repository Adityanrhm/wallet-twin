// Package models - Category entity
//
// Category digunakan untuk mengkategorikan transaksi.
// Ada 2 tipe: income (pemasukan) dan expense (pengeluaran).
//
// Category mendukung hierarki (parent-child) untuk sub-kategori:
//
//	Food & Dining (parent)
//	├── Groceries (child)
//	├── Restaurant (child)
//	└── Coffee (child)
package models

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// CategoryType adalah tipe kategori.
type CategoryType string

const (
	// CategoryTypeIncome untuk kategori pemasukan
	CategoryTypeIncome CategoryType = "income"

	// CategoryTypeExpense untuk kategori pengeluaran
	CategoryTypeExpense CategoryType = "expense"
)

// IsValid mengecek apakah category type valid.
func (t CategoryType) IsValid() bool {
	switch t {
	case CategoryTypeIncome, CategoryTypeExpense:
		return true
	}
	return false
}

// String returns string representation.
func (t CategoryType) String() string {
	return string(t)
}

// Category merepresentasikan kategori transaksi.
//
// Category bisa memiliki parent (untuk sub-kategori).
// Jika ParentID nil, itu adalah top-level category.
//
// Contoh struktur:
//
//	// Top-level category
//	food := &models.Category{
//	    Name: "Food & Dining",
//	    Type: models.CategoryTypeExpense,
//	    Icon: "🍔",
//	}
//
//	// Sub-category
//	groceries := &models.Category{
//	    Name:     "Groceries",
//	    Type:     models.CategoryTypeExpense,
//	    Icon:     "🥬",
//	    ParentID: &food.ID,
//	}
type Category struct {
	// ID adalah unique identifier.
	// Tidak menggunakan BaseModel karena Category tidak perlu UpdatedAt.
	ID uuid.UUID `json:"id" db:"id"`

	// Name adalah nama kategori.
	// Contoh: "Food & Dining", "Salary", "Transportation"
	Name string `json:"name" db:"name"`

	// Type menentukan apakah ini kategori income atau expense.
	// Ini menentukan di form mana kategori ini muncul.
	Type CategoryType `json:"type" db:"type"`

	// Color adalah warna hex untuk UI visualization.
	// Contoh: "#EF4444" (red), "#10B981" (green)
	Color string `json:"color,omitempty" db:"color"`

	// Icon adalah emoji atau nama icon.
	// Contoh: "🍔", "💰", "🚗"
	Icon string `json:"icon,omitempty" db:"icon"`

	// ParentID untuk hierarki kategori.
	// nil = top-level category
	// non-nil = sub-category
	//
	// Menggunakan pointer agar bisa nil (nullable di DB).
	ParentID *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`

	// SortOrder untuk custom ordering.
	// Lower number = tampil lebih dulu.
	SortOrder int `json:"sort_order" db:"sort_order"`

	// CreatedAt timestamp.
	CreatedAt string `json:"created_at" db:"created_at"`
}

// Validation errors
var (
	ErrCategoryNameRequired = errors.New("category name is required")
	ErrCategoryNameTooLong  = errors.New("category name must be less than 100 characters")
	ErrCategoryInvalidType  = errors.New("invalid category type")
)

// Validate memvalidasi category.
func (c *Category) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return ErrCategoryNameRequired
	}
	if len(c.Name) > 100 {
		return ErrCategoryNameTooLong
	}
	if !c.Type.IsValid() {
		return ErrCategoryInvalidType
	}
	return nil
}

// NewCategory membuat category baru.
//
//	cat := models.NewCategory("Food & Dining", models.CategoryTypeExpense)
//	cat.Icon = "🍔"
//	cat.Color = "#EF4444"
func NewCategory(name string, catType CategoryType) *Category {
	return &Category{
		ID:   NewID(),
		Name: name,
		Type: catType,
	}
}

// IsSubCategory mengecek apakah ini sub-category.
//
//	if cat.IsSubCategory() {
//	    fmt.Println("This is a sub-category")
//	}
func (c *Category) IsSubCategory() bool {
	return c.ParentID != nil
}
