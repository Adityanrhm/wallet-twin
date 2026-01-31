// Package models - Transfer entity
//
// Transfer adalah operasi khusus yang memindahkan uang dari satu wallet
// ke wallet lain. Berbeda dengan transaksi biasa, transfer melibatkan
// 2 wallet sekaligus secara atomic.
//
// Contoh:
// - Transfer dari BCA ke GoPay Rp 500.000
// - Tarik tunai dari ATM (Bank → Cash)
// - Top up e-wallet (Bank → E-wallet)
//
// Transfer bisa memiliki fee (biaya transfer) yang dibebankan ke wallet sumber.
package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Transfer merepresentasikan transfer antar wallet.
//
// Transfer adalah operasi ATOMIC - harus berhasil atau gagal sepenuhnya.
// Jangan implementasikan ini dengan 2 transaksi terpisah!
//
// Flow transfer:
// 1. Validate (saldo cukup, wallet berbeda, dll)
// 2. Begin database transaction
// 3. Subtract from source wallet (amount + fee)
// 4. Add to destination wallet (amount only, fee tidak ditransfer)
// 5. Create Transfer record
// 6. Commit transaction
//
// Contoh penggunaan:
//
//	transfer := &models.Transfer{
//	    ID:           models.NewID(),
//	    FromWalletID: bcaWallet.ID,
//	    ToWalletID:   gopayWallet.ID,
//	    Amount:       decimal.NewFromInt(500000),
//	    Fee:          decimal.NewFromInt(6500),
//	    Note:         "Top up GoPay",
//	}
type Transfer struct {
	// ID adalah unique identifier.
	ID uuid.UUID `json:"id" db:"id"`

	// FromWalletID adalah wallet sumber (uang keluar).
	FromWalletID uuid.UUID `json:"from_wallet_id" db:"from_wallet_id"`

	// ToWalletID adalah wallet tujuan (uang masuk).
	ToWalletID uuid.UUID `json:"to_wallet_id" db:"to_wallet_id"`

	// Amount adalah jumlah yang ditransfer.
	// Jumlah ini yang masuk ke wallet tujuan.
	Amount decimal.Decimal `json:"amount" db:"amount"`

	// Fee adalah biaya transfer (opsional).
	// Dibebankan ke wallet sumber.
	// Total yang dikurangi dari sumber = Amount + Fee
	//
	// Contoh: Transfer 500.000 dengan fee 6.500
	// - Wallet sumber: -506.500
	// - Wallet tujuan: +500.000
	// - Fee: 6.500 (hilang/biaya)
	Fee decimal.Decimal `json:"fee" db:"fee"`

	// Note adalah catatan transfer.
	Note string `json:"note,omitempty" db:"note"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Validation errors
var (
	ErrTransferSameWallet    = errors.New("cannot transfer to the same wallet")
	ErrTransferInvalidAmount = errors.New("transfer amount must be positive")
	ErrTransferNegativeFee   = errors.New("transfer fee cannot be negative")
	ErrTransferNoFromWallet  = errors.New("source wallet is required")
	ErrTransferNoToWallet    = errors.New("destination wallet is required")
)

// Validate memvalidasi transfer.
func (t *Transfer) Validate() error {
	if t.FromWalletID == uuid.Nil {
		return ErrTransferNoFromWallet
	}
	if t.ToWalletID == uuid.Nil {
		return ErrTransferNoToWallet
	}
	if t.FromWalletID == t.ToWalletID {
		return ErrTransferSameWallet
	}
	if t.Amount.IsNegative() || t.Amount.IsZero() {
		return ErrTransferInvalidAmount
	}
	if t.Fee.IsNegative() {
		return ErrTransferNegativeFee
	}
	t.Note = strings.TrimSpace(t.Note)
	return nil
}

// NewTransfer membuat transfer baru.
//
//	transfer := models.NewTransfer(fromWallet.ID, toWallet.ID, decimal.NewFromInt(500000))
//	transfer.Fee = decimal.NewFromInt(6500)
//	transfer.Note = "Top up GoPay"
func NewTransfer(fromWalletID, toWalletID uuid.UUID, amount decimal.Decimal) *Transfer {
	return &Transfer{
		ID:           NewID(),
		FromWalletID: fromWalletID,
		ToWalletID:   toWalletID,
		Amount:       amount,
		Fee:          decimal.Zero,
		CreatedAt:    time.Now(),
	}
}

// TotalDeducted menghitung total yang dikurangi dari wallet sumber.
// Total = Amount + Fee
//
//	deducted := transfer.TotalDeducted()
//	// deducted = 506500 (jika amount 500000, fee 6500)
func (t *Transfer) TotalDeducted() decimal.Decimal {
	return t.Amount.Add(t.Fee)
}

// SetFee sets the transfer fee.
// Convenience method dengan validation.
//
//	if err := transfer.SetFee(decimal.NewFromInt(6500)); err != nil {
//	    // fee negatif
//	}
func (t *Transfer) SetFee(fee decimal.Decimal) error {
	if fee.IsNegative() {
		return ErrTransferNegativeFee
	}
	t.Fee = fee
	return nil
}
