// Package models - Transaction entity
//
// Transaction adalah inti dari aplikasi - setiap pemasukan dan pengeluaran
// dicatat sebagai Transaction.
//
// Transaction selalu terhubung ke:
// - Wallet: Dari mana/kemana uang mengalir
// - Category: Untuk apa transaksi ini
//
// PENTING: Saat transaction dibuat/dihapus, saldo wallet harus diupdate!
// Ini ditangani oleh TransactionService, bukan di entity level.
package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransactionType adalah tipe transaksi.
type TransactionType string

const (
	// TransactionTypeIncome untuk pemasukan (menambah saldo)
	TransactionTypeIncome TransactionType = "income"

	// TransactionTypeExpense untuk pengeluaran (mengurangi saldo)
	TransactionTypeExpense TransactionType = "expense"
)

// IsValid mengecek apakah transaction type valid.
func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypeIncome, TransactionTypeExpense:
		return true
	}
	return false
}

// String returns string representation.
func (t TransactionType) String() string {
	return string(t)
}

// IsIncome returns true if this is an income transaction.
func (t TransactionType) IsIncome() bool {
	return t == TransactionTypeIncome
}

// IsExpense returns true if this is an expense transaction.
func (t TransactionType) IsExpense() bool {
	return t == TransactionTypeExpense
}

// Transaction merepresentasikan transaksi keuangan.
//
// Setiap transaction mempengaruhi saldo wallet:
// - Income: wallet.Balance += amount
// - Expense: wallet.Balance -= amount
//
// Contoh penggunaan:
//
//	tx := &models.Transaction{
//	    BaseModel:       models.BaseModel{ID: models.NewID()},
//	    WalletID:        wallet.ID,
//	    CategoryID:      &categoryID,
//	    Type:            models.TransactionTypeExpense,
//	    Amount:          decimal.NewFromInt(50000),
//	    Description:     "Makan siang",
//	    TransactionDate: time.Now(),
//	}
type Transaction struct {
	// Embed BaseModel untuk ID dan timestamps
	BaseModel

	// WalletID adalah foreign key ke wallet.
	// Required - setiap transaksi harus punya wallet.
	WalletID uuid.UUID `json:"wallet_id" db:"wallet_id"`

	// CategoryID adalah foreign key ke category.
	// Optional - bisa nil (uncategorized).
	// Menggunakan pointer agar nullable.
	CategoryID *uuid.UUID `json:"category_id,omitempty" db:"category_id"`

	// Type adalah tipe transaksi: income atau expense.
	Type TransactionType `json:"type" db:"type"`

	// Amount adalah jumlah transaksi.
	// Selalu positif! Tipe menentukan apakah add atau subtract.
	// Menggunakan Decimal untuk presisi keuangan.
	Amount decimal.Decimal `json:"amount" db:"amount"`

	// Description adalah catatan transaksi.
	// Optional tapi sangat direkomendasikan untuk tracking.
	// Contoh: "Makan siang di warteg", "Gaji Januari"
	Description string `json:"description,omitempty" db:"description"`

	// Tags adalah label tambahan untuk filtering.
	// Contoh: ["work", "lunch"], ["monthly"]
	Tags []string `json:"tags,omitempty" db:"tags"`

	// TransactionDate adalah tanggal transaksi.
	// Bisa berbeda dengan CreatedAt (backdate transaction).
	// Contoh: User input hari ini untuk transaksi kemarin.
	TransactionDate time.Time `json:"transaction_date" db:"transaction_date"`
}

// Validation errors
var (
	ErrTransactionInvalidType   = errors.New("invalid transaction type")
	ErrTransactionInvalidAmount = errors.New("transaction amount must be positive")
	ErrTransactionNoWallet      = errors.New("wallet is required")
)

// Validate memvalidasi transaction.
func (t *Transaction) Validate() error {
	if t.WalletID == uuid.Nil {
		return ErrTransactionNoWallet
	}
	if !t.Type.IsValid() {
		return ErrTransactionInvalidType
	}
	if t.Amount.IsNegative() || t.Amount.IsZero() {
		return ErrTransactionInvalidAmount
	}
	t.Description = strings.TrimSpace(t.Description)
	return nil
}

// NewTransaction membuat transaction baru dengan defaults.
//
//	tx := models.NewTransaction(walletID, models.TransactionTypeExpense, decimal.NewFromInt(50000))
//	tx.Description = "Makan siang"
//	tx.CategoryID = &foodCategoryID
func NewTransaction(walletID uuid.UUID, txType TransactionType, amount decimal.Decimal) *Transaction {
	return &Transaction{
		BaseModel:       BaseModel{ID: NewID()},
		WalletID:        walletID,
		Type:            txType,
		Amount:          amount,
		TransactionDate: time.Now(),
	}
}

// SetCategory sets the category for this transaction.
// Convenience method untuk set nullable field.
//
//	tx.SetCategory(foodCategoryID)
func (t *Transaction) SetCategory(categoryID uuid.UUID) {
	t.CategoryID = &categoryID
}

// ClearCategory removes the category (set to uncategorized).
func (t *Transaction) ClearCategory() {
	t.CategoryID = nil
}

// AddTag menambah tag ke transaction.
//
//	tx.AddTag("work")
//	tx.AddTag("lunch")
func (t *Transaction) AddTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return
	}
	// Check duplicate
	for _, existing := range t.Tags {
		if existing == tag {
			return
		}
	}
	t.Tags = append(t.Tags, tag)
}

// HasTag mengecek apakah transaction memiliki tag tertentu.
//
//	if tx.HasTag("work") {
//	    // ...
//	}
func (t *Transaction) HasTag(tag string) bool {
	tag = strings.ToLower(tag)
	for _, existing := range t.Tags {
		if existing == tag {
			return true
		}
	}
	return false
}
