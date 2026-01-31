// Package models - Wallet entity
//
// Wallet merepresentasikan dompet/akun keuangan pengguna.
// Ini adalah entity utama untuk tracking saldo.
//
// Contoh wallet:
// - Cash (uang tunai)
// - BCA, Mandiri, BNI (rekening bank)
// - GoPay, OVO, DANA (e-wallet)
package models

import (
	"errors"
	"strings"

	"github.com/shopspring/decimal"
)

// WalletType adalah tipe wallet.
// Menggunakan string constant (bukan iota) agar:
// - Readable di database
// - Mudah di-serialize ke JSON
// - Tidak berubah jika urutan berubah
type WalletType string

const (
	// WalletTypeCash untuk uang tunai
	WalletTypeCash WalletType = "cash"

	// WalletTypeBank untuk rekening bank
	WalletTypeBank WalletType = "bank"

	// WalletTypeEWallet untuk dompet digital
	WalletTypeEWallet WalletType = "ewallet"
)

// IsValid mengecek apakah wallet type valid.
//
//	if !walletType.IsValid() {
//	    return errors.New("invalid wallet type")
//	}
func (t WalletType) IsValid() bool {
	switch t {
	case WalletTypeCash, WalletTypeBank, WalletTypeEWallet:
		return true
	}
	return false
}

// String returns string representation of wallet type.
func (t WalletType) String() string {
	return string(t)
}

// Wallet merepresentasikan dompet/akun keuangan.
//
// Wallet adalah entity "aggregate root" - entity utama yang
// memiliki identity sendiri dan lifecycle independen.
//
// Balance di-track secara real-time dan diupdate setiap ada transaksi.
// PENTING: Jangan update Balance secara manual, gunakan service methods.
//
// Contoh penggunaan:
//
//	wallet := &models.Wallet{
//	    BaseModel: models.BaseModel{ID: models.NewID()},
//	    Name:      "BCA Tabungan",
//	    Type:      models.WalletTypeBank,
//	    Balance:   decimal.NewFromInt(1000000),
//	    Currency:  "IDR",
//	}
//
//	if err := wallet.Validate(); err != nil {
//	    log.Fatal(err)
//	}
type Wallet struct {
	// Embed BaseModel untuk ID dan timestamps
	BaseModel

	// Name adalah nama wallet yang ditampilkan ke user.
	// Required, minimal 1 karakter.
	// Contoh: "Cash", "BCA", "GoPay"
	Name string `json:"name" db:"name"`

	// Type adalah tipe wallet.
	// Default: WalletTypeCash
	Type WalletType `json:"type" db:"type"`

	// Balance adalah saldo wallet saat ini.
	// Menggunakan decimal.Decimal untuk presisi keuangan.
	// JANGAN pakai float64 untuk uang! (precision issues)
	//
	// Kenapa Decimal?
	// - Float: 0.1 + 0.2 = 0.30000000000000004 (WRONG!)
	// - Decimal: 0.1 + 0.2 = 0.3 (CORRECT!)
	Balance decimal.Decimal `json:"balance" db:"balance"`

	// Currency adalah kode mata uang ISO 4217.
	// 3 karakter uppercase.
	// Contoh: "IDR", "USD", "EUR"
	Currency string `json:"currency" db:"currency"`

	// Color adalah warna hex untuk UI.
	// Optional, format: "#RRGGBB"
	// Contoh: "#7C3AED"
	Color string `json:"color,omitempty" db:"color"`

	// Icon adalah emoji atau nama icon.
	// Optional.
	// Contoh: "💰", "🏦", "wallet"
	Icon string `json:"icon,omitempty" db:"icon"`

	// IsActive menentukan apakah wallet ditampilkan.
	// FALSE = soft deleted (tersembunyi tapi data tetap ada)
	IsActive bool `json:"is_active" db:"is_active"`
}

// Validation errors
var (
	ErrWalletNameRequired    = errors.New("wallet name is required")
	ErrWalletNameTooLong     = errors.New("wallet name must be less than 100 characters")
	ErrWalletInvalidType     = errors.New("invalid wallet type")
	ErrWalletInvalidCurrency = errors.New("currency must be a 3-letter ISO code")
	ErrWalletNegativeBalance = errors.New("wallet balance cannot be negative")
)

// Validate memvalidasi wallet sebelum disimpan.
//
// Validasi yang dilakukan:
// - Name tidak kosong dan tidak terlalu panjang
// - Type valid (cash, bank, ewallet)
// - Currency 3 karakter
// - Balance tidak negatif
//
// Contoh:
//
//	if err := wallet.Validate(); err != nil {
//	    return fmt.Errorf("invalid wallet: %w", err)
//	}
func (w *Wallet) Validate() error {
	// Validate name
	w.Name = strings.TrimSpace(w.Name)
	if w.Name == "" {
		return ErrWalletNameRequired
	}
	if len(w.Name) > 100 {
		return ErrWalletNameTooLong
	}

	// Validate type
	if !w.Type.IsValid() {
		return ErrWalletInvalidType
	}

	// Validate currency (3 letters)
	w.Currency = strings.ToUpper(strings.TrimSpace(w.Currency))
	if len(w.Currency) != 3 {
		return ErrWalletInvalidCurrency
	}

	// Validate balance (tidak boleh negatif untuk wallet biasa)
	if w.Balance.IsNegative() {
		return ErrWalletNegativeBalance
	}

	return nil
}

// NewWallet membuat wallet baru dengan default values.
// Convenience function untuk membuat wallet dengan sensible defaults.
//
// Contoh:
//
//	wallet := models.NewWallet("BCA", models.WalletTypeBank)
//	wallet.Balance = decimal.NewFromInt(1000000)
func NewWallet(name string, walletType WalletType) *Wallet {
	return &Wallet{
		BaseModel: BaseModel{ID: NewID()},
		Name:      name,
		Type:      walletType,
		Balance:   decimal.Zero,
		Currency:  "IDR",
		IsActive:  true,
	}
}

// AddBalance menambah saldo wallet (untuk income).
// Positive amount only.
//
//	wallet.AddBalance(decimal.NewFromInt(500000))
func (w *Wallet) AddBalance(amount decimal.Decimal) {
	w.Balance = w.Balance.Add(amount)
}

// SubtractBalance mengurangi saldo wallet (untuk expense).
// Positive amount only. Akan return error jika saldo tidak cukup.
//
//	err := wallet.SubtractBalance(decimal.NewFromInt(50000))
//	if err != nil {
//	    log.Println("Insufficient balance")
//	}
func (w *Wallet) SubtractBalance(amount decimal.Decimal) error {
	if w.Balance.LessThan(amount) {
		return errors.New("insufficient balance")
	}
	w.Balance = w.Balance.Sub(amount)
	return nil
}
