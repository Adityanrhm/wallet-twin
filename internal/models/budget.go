// Package models - Budget entity
//
// Budget membantu user mengontrol pengeluaran per kategori.
// User set budget bulanan, dan aplikasi track progress.
//
// Contoh:
// - Budget Food & Dining: Rp 2.000.000 per bulan
// - Budget Transportation: Rp 500.000 per bulan
//
// Aplikasi akan alert jika pengeluaran mendekati/melebihi budget.
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BudgetPeriod adalah periode budget.
type BudgetPeriod string

const (
	// BudgetPeriodWeekly untuk budget mingguan
	BudgetPeriodWeekly BudgetPeriod = "weekly"

	// BudgetPeriodMonthly untuk budget bulanan (paling umum)
	BudgetPeriodMonthly BudgetPeriod = "monthly"

	// BudgetPeriodYearly untuk budget tahunan
	BudgetPeriodYearly BudgetPeriod = "yearly"
)

// IsValid mengecek apakah budget period valid.
func (p BudgetPeriod) IsValid() bool {
	switch p {
	case BudgetPeriodWeekly, BudgetPeriodMonthly, BudgetPeriodYearly:
		return true
	}
	return false
}

// String returns string representation.
func (p BudgetPeriod) String() string {
	return string(p)
}

// Budget merepresentasikan anggaran per kategori per periode.
//
// Budget digunakan untuk:
// 1. Set limit pengeluaran per kategori
// 2. Track spending vs budget
// 3. Alert saat mendekati/melebihi budget
//
// Contoh penggunaan:
//
//	budget := &models.Budget{
//	    ID:         models.NewID(),
//	    CategoryID: foodCategoryID,
//	    Amount:     decimal.NewFromInt(2000000),
//	    Period:     models.BudgetPeriodMonthly,
//	    StartDate:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local),
//	}
//
//	// Cek progress
//	spent := decimal.NewFromInt(1500000)
//	progress := budget.CalculateProgress(spent)
//	// progress = 75%
type Budget struct {
	// ID adalah unique identifier.
	ID uuid.UUID `json:"id" db:"id"`

	// CategoryID adalah kategori yang di-budget.
	// Required - budget harus untuk kategori tertentu.
	CategoryID uuid.UUID `json:"category_id" db:"category_id"`

	// Amount adalah jumlah budget.
	// Ini adalah limit maksimal pengeluaran untuk kategori ini.
	Amount decimal.Decimal `json:"amount" db:"amount"`

	// Period adalah periode budget.
	// Default: monthly
	Period BudgetPeriod `json:"period" db:"period"`

	// StartDate adalah tanggal mulai budget.
	// Untuk monthly, biasanya tanggal 1.
	StartDate time.Time `json:"start_date" db:"start_date"`

	// EndDate adalah tanggal akhir budget (opsional).
	// nil = budget berlaku selamanya (recurring).
	EndDate *time.Time `json:"end_date,omitempty" db:"end_date"`

	// IsActive menentukan apakah budget aktif.
	IsActive bool `json:"is_active" db:"is_active"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Validation errors
var (
	ErrBudgetNoCategory    = errors.New("category is required for budget")
	ErrBudgetInvalidAmount = errors.New("budget amount must be positive")
	ErrBudgetInvalidPeriod = errors.New("invalid budget period")
	ErrBudgetInvalidDates  = errors.New("end date must be after start date")
)

// Validate memvalidasi budget.
func (b *Budget) Validate() error {
	if b.CategoryID == uuid.Nil {
		return ErrBudgetNoCategory
	}
	if b.Amount.IsNegative() || b.Amount.IsZero() {
		return ErrBudgetInvalidAmount
	}
	if !b.Period.IsValid() {
		return ErrBudgetInvalidPeriod
	}
	if b.EndDate != nil && b.EndDate.Before(b.StartDate) {
		return ErrBudgetInvalidDates
	}
	return nil
}

// NewBudget membuat budget baru.
//
//	budget := models.NewBudget(foodCategoryID, decimal.NewFromInt(2000000))
func NewBudget(categoryID uuid.UUID, amount decimal.Decimal) *Budget {
	return &Budget{
		ID:         NewID(),
		CategoryID: categoryID,
		Amount:     amount,
		Period:     BudgetPeriodMonthly,
		StartDate:  time.Now(),
		IsActive:   true,
		CreatedAt:  time.Now(),
	}
}

// CalculateProgress menghitung persentase budget yang sudah terpakai.
// Return value 0-100 (bisa > 100 jika over budget).
//
//	spent := decimal.NewFromInt(1500000)
//	progress := budget.CalculateProgress(spent) // 75
func (b *Budget) CalculateProgress(spent decimal.Decimal) float64 {
	if b.Amount.IsZero() {
		return 0
	}
	progress, _ := spent.Div(b.Amount).Mul(decimal.NewFromInt(100)).Float64()
	return progress
}

// IsOverBudget mengecek apakah pengeluaran melebihi budget.
//
//	if budget.IsOverBudget(spent) {
//	    fmt.Println("WARNING: Over budget!")
//	}
func (b *Budget) IsOverBudget(spent decimal.Decimal) bool {
	return spent.GreaterThan(b.Amount)
}

// GetRemaining menghitung sisa budget.
// Return 0 jika sudah over budget.
//
//	remaining := budget.GetRemaining(spent)
func (b *Budget) GetRemaining(spent decimal.Decimal) decimal.Decimal {
	remaining := b.Amount.Sub(spent)
	if remaining.IsNegative() {
		return decimal.Zero
	}
	return remaining
}
