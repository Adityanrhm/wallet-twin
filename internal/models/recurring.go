// Package models - Recurring entity
//
// Recurring transaction adalah transaksi yang terjadi secara berkala.
// Sistem akan otomatis generate transaksi saat jatuh tempo.
//
// Contoh:
// - Gaji bulanan (income, monthly, tanggal 25)
// - Langganan Netflix (expense, monthly)
// - Bayar listrik (expense, monthly)
//
// Workflow:
// 1. User setup recurring transaction
// 2. Sistem check setiap hari untuk next_due
// 3. Jika next_due <= today, generate transaksi
// 4. Update next_due ke periode berikutnya
package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RecurringFrequency adalah frekuensi recurring.
type RecurringFrequency string

const (
	// RecurringDaily untuk harian
	RecurringDaily RecurringFrequency = "daily"

	// RecurringWeekly untuk mingguan
	RecurringWeekly RecurringFrequency = "weekly"

	// RecurringMonthly untuk bulanan (paling umum)
	RecurringMonthly RecurringFrequency = "monthly"

	// RecurringYearly untuk tahunan
	RecurringYearly RecurringFrequency = "yearly"
)

// IsValid mengecek apakah frequency valid.
func (f RecurringFrequency) IsValid() bool {
	switch f {
	case RecurringDaily, RecurringWeekly, RecurringMonthly, RecurringYearly:
		return true
	}
	return false
}

// String returns string representation.
func (f RecurringFrequency) String() string {
	return string(f)
}

// RecurringTransaction merepresentasikan transaksi berulang.
//
// RecurringTransaction adalah template untuk generate Transaction.
// Setiap kali jatuh tempo (NextDue), sistem akan:
// 1. Create Transaction baru dari template ini
// 2. Update saldo wallet sesuai
// 3. Advance NextDue ke periode berikutnya
//
// Contoh:
//
//	recurring := &models.RecurringTransaction{
//	    ID:          models.NewID(),
//	    WalletID:    bcaWallet.ID,
//	    CategoryID:  &salaryCategoryID,
//	    Type:        models.TransactionTypeIncome,
//	    Amount:      decimal.NewFromInt(5000000),
//	    Description: "Gaji Bulanan",
//	    Frequency:   models.RecurringMonthly,
//	    NextDue:     time.Date(2026, 2, 25, 0, 0, 0, 0, time.Local),
//	}
type RecurringTransaction struct {
	// ID adalah unique identifier.
	ID uuid.UUID `json:"id" db:"id"`

	// WalletID untuk transaksi yang akan di-generate.
	WalletID uuid.UUID `json:"wallet_id" db:"wallet_id"`

	// CategoryID untuk transaksi.
	CategoryID *uuid.UUID `json:"category_id,omitempty" db:"category_id"`

	// Type adalah tipe transaksi: income atau expense.
	Type TransactionType `json:"type" db:"type"`

	// Amount adalah jumlah per transaksi.
	Amount decimal.Decimal `json:"amount" db:"amount"`

	// Description untuk transaksi yang di-generate.
	Description string `json:"description" db:"description"`

	// Frequency adalah seberapa sering transaksi terjadi.
	Frequency RecurringFrequency `json:"frequency" db:"frequency"`

	// NextDue adalah tanggal jatuh tempo berikutnya.
	// Ini yang di-check oleh scheduler.
	NextDue time.Time `json:"next_due" db:"next_due"`

	// EndDate adalah tanggal akhir recurring (opsional).
	// nil = recurring selamanya.
	EndDate *time.Time `json:"end_date,omitempty" db:"end_date"`

	// IsActive menentukan apakah recurring aktif.
	IsActive bool `json:"is_active" db:"is_active"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Validation errors
var (
	ErrRecurringNoWallet        = errors.New("wallet is required")
	ErrRecurringInvalidType     = errors.New("invalid transaction type")
	ErrRecurringInvalidAmount   = errors.New("amount must be positive")
	ErrRecurringInvalidFreq     = errors.New("invalid frequency")
	ErrRecurringInvalidEndDate  = errors.New("end date must be after next due")
)

// Validate memvalidasi recurring transaction.
func (r *RecurringTransaction) Validate() error {
	if r.WalletID == uuid.Nil {
		return ErrRecurringNoWallet
	}
	if !r.Type.IsValid() {
		return ErrRecurringInvalidType
	}
	if r.Amount.IsNegative() || r.Amount.IsZero() {
		return ErrRecurringInvalidAmount
	}
	if !r.Frequency.IsValid() {
		return ErrRecurringInvalidFreq
	}
	if r.EndDate != nil && r.EndDate.Before(r.NextDue) {
		return ErrRecurringInvalidEndDate
	}
	r.Description = strings.TrimSpace(r.Description)
	return nil
}

// NewRecurringTransaction membuat recurring transaction baru.
func NewRecurringTransaction(
	walletID uuid.UUID,
	txType TransactionType,
	amount decimal.Decimal,
	freq RecurringFrequency,
	nextDue time.Time,
) *RecurringTransaction {
	return &RecurringTransaction{
		ID:        NewID(),
		WalletID:  walletID,
		Type:      txType,
		Amount:    amount,
		Frequency: freq,
		NextDue:   nextDue,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
}

// IsDue mengecek apakah recurring sudah jatuh tempo.
//
//	if recurring.IsDue() {
//	    // Generate transaction
//	}
func (r *RecurringTransaction) IsDue() bool {
	return r.IsActive && !r.NextDue.After(time.Now())
}

// AdvanceNextDue memajukan NextDue ke periode berikutnya.
// Panggil setelah generate transaction.
//
//	recurring.AdvanceNextDue()
func (r *RecurringTransaction) AdvanceNextDue() {
	switch r.Frequency {
	case RecurringDaily:
		r.NextDue = r.NextDue.AddDate(0, 0, 1)
	case RecurringWeekly:
		r.NextDue = r.NextDue.AddDate(0, 0, 7)
	case RecurringMonthly:
		r.NextDue = r.NextDue.AddDate(0, 1, 0)
	case RecurringYearly:
		r.NextDue = r.NextDue.AddDate(1, 0, 0)
	}

	// Deactivate if past end date
	if r.EndDate != nil && r.NextDue.After(*r.EndDate) {
		r.IsActive = false
	}
}

// ToTransaction mengkonversi recurring ke Transaction.
// Panggil ini saat generate transaction dari recurring.
//
//	tx := recurring.ToTransaction()
//	// Simpan tx ke database
func (r *RecurringTransaction) ToTransaction() *Transaction {
	return &Transaction{
		BaseModel:       BaseModel{ID: NewID()},
		WalletID:        r.WalletID,
		CategoryID:      r.CategoryID,
		Type:            r.Type,
		Amount:          r.Amount,
		Description:     r.Description,
		TransactionDate: r.NextDue,
	}
}
