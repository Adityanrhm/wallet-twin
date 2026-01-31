package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestWallet_Validate(t *testing.T) {
	tests := []struct {
		name    string
		wallet  *Wallet
		wantErr bool
	}{
		{
			name: "valid wallet",
			wallet: &Wallet{
				BaseModel: BaseModel{ID: uuid.New()},
				Name:      "BCA",
				Type:      WalletTypeBank,
				Currency:  "IDR",
				Balance:   decimal.NewFromInt(100000),
			},
			wantErr: false,
		},
		{
			name: "empty name",
			wallet: &Wallet{
				BaseModel: BaseModel{ID: uuid.New()},
				Name:      "",
				Type:      WalletTypeBank,
				Currency:  "IDR",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			wallet: &Wallet{
				BaseModel: BaseModel{ID: uuid.New()},
				Name:      "BCA",
				Type:      WalletType("invalid"),
				Currency:  "IDR",
			},
			wantErr: true,
		},
		{
			name: "empty currency",
			wallet: &Wallet{
				BaseModel: BaseModel{ID: uuid.New()},
				Name:      "BCA",
				Type:      WalletTypeBank,
				Currency:  "",
			},
			wantErr: true,
		},
		{
			name: "negative balance",
			wallet: &Wallet{
				BaseModel: BaseModel{ID: uuid.New()},
				Name:      "BCA",
				Type:      WalletTypeBank,
				Currency:  "IDR",
				Balance:   decimal.NewFromInt(-1000),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.wallet.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Wallet.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransaction_Validate(t *testing.T) {
	walletID := uuid.New()

	tests := []struct {
		name    string
		tx      *Transaction
		wantErr bool
	}{
		{
			name: "valid income",
			tx: &Transaction{
				BaseModel:       BaseModel{ID: uuid.New()},
				WalletID:        walletID,
				Type:            TransactionTypeIncome,
				Amount:          decimal.NewFromInt(50000),
				TransactionDate: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid expense",
			tx: &Transaction{
				BaseModel:       BaseModel{ID: uuid.New()},
				WalletID:        walletID,
				Type:            TransactionTypeExpense,
				Amount:          decimal.NewFromInt(25000),
				Description:     "Makan siang",
				TransactionDate: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "zero amount",
			tx: &Transaction{
				BaseModel:       BaseModel{ID: uuid.New()},
				WalletID:        walletID,
				Type:            TransactionTypeExpense,
				Amount:          decimal.Zero,
				TransactionDate: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			tx: &Transaction{
				BaseModel:       BaseModel{ID: uuid.New()},
				WalletID:        walletID,
				Type:            TransactionTypeExpense,
				Amount:          decimal.NewFromInt(-1000),
				TransactionDate: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			tx: &Transaction{
				BaseModel:       BaseModel{ID: uuid.New()},
				WalletID:        walletID,
				Type:            TransactionType("invalid"),
				Amount:          decimal.NewFromInt(50000),
				TransactionDate: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Transaction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGoal_GetProgress(t *testing.T) {
	tests := []struct {
		name    string
		current int64
		target  int64
		want    float64
	}{
		{"0%", 0, 1000000, 0.0},
		{"50%", 500000, 1000000, 50.0},
		{"100%", 1000000, 1000000, 100.0},
		{"over 100%", 1500000, 1000000, 150.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goal{
				CurrentAmount: decimal.NewFromInt(tt.current),
				TargetAmount:  decimal.NewFromInt(tt.target),
			}
			got := g.GetProgress()
			if got != tt.want {
				t.Errorf("Goal.GetProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoal_IsCompleted(t *testing.T) {
	tests := []struct {
		name    string
		current int64
		target  int64
		want    bool
	}{
		{"not completed", 500000, 1000000, false},
		{"exactly completed", 1000000, 1000000, true},
		{"over completed", 1200000, 1000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goal{
				CurrentAmount: decimal.NewFromInt(tt.current),
				TargetAmount:  decimal.NewFromInt(tt.target),
			}
			if got := g.IsCompleted(); got != tt.want {
				t.Errorf("Goal.IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransfer_TotalDeducted(t *testing.T) {
	transfer := &Transfer{
		Amount: decimal.NewFromInt(500000),
		Fee:    decimal.NewFromInt(6500),
	}

	expected := decimal.NewFromInt(506500)
	got := transfer.TotalDeducted()

	if !got.Equal(expected) {
		t.Errorf("Transfer.TotalDeducted() = %v, want %v", got, expected)
	}
}
