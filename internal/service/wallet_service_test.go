package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// Mock repositories for testing

type mockWalletRepo struct {
	wallets map[uuid.UUID]*models.Wallet
}

func newMockWalletRepo() *mockWalletRepo {
	return &mockWalletRepo{
		wallets: make(map[uuid.UUID]*models.Wallet),
	}
}

func (m *mockWalletRepo) Create(ctx context.Context, w *models.Wallet) error {
	m.wallets[w.ID] = w
	return nil
}

func (m *mockWalletRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	if w, ok := m.wallets[id]; ok {
		return w, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockWalletRepo) List(ctx context.Context, filter repository.WalletFilter) ([]*models.Wallet, error) {
	var result []*models.Wallet
	for _, w := range m.wallets {
		if filter.IsActive != nil && w.IsActive != *filter.IsActive {
			continue
		}
		result = append(result, w)
	}
	return result, nil
}

func (m *mockWalletRepo) Update(ctx context.Context, w *models.Wallet) error {
	if _, ok := m.wallets[w.ID]; !ok {
		return repository.ErrNotFound
	}
	m.wallets[w.ID] = w
	return nil
}

func (m *mockWalletRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if w, ok := m.wallets[id]; ok {
		w.IsActive = false
		return nil
	}
	return repository.ErrNotFound
}

func (m *mockWalletRepo) UpdateBalance(ctx context.Context, id uuid.UUID, balance decimal.Decimal) error {
	if w, ok := m.wallets[id]; ok {
		w.Balance = balance
		return nil
	}
	return repository.ErrNotFound
}

func (m *mockWalletRepo) GetTotalBalance(ctx context.Context) (decimal.Decimal, error) {
	total := decimal.Zero
	for _, w := range m.wallets {
		if w.IsActive {
			total = total.Add(w.Balance)
		}
	}
	return total, nil
}

// Tests

func TestWalletService_Create(t *testing.T) {
	repo := newMockWalletRepo()
	svc := NewWalletService(repo)

	tests := []struct {
		name    string
		input   CreateWalletInput
		wantErr bool
	}{
		{
			name: "valid wallet",
			input: CreateWalletInput{
				Name:           "BCA Tabungan",
				Type:           models.WalletTypeBank,
				Currency:       "IDR",
				InitialBalance: decimal.NewFromInt(1000000),
				Icon:           "🏦",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			input: CreateWalletInput{
				Name:     "",
				Type:     models.WalletTypeBank,
				Currency: "IDR",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			input: CreateWalletInput{
				Name:     "Test",
				Type:     models.WalletType("invalid"),
				Currency: "IDR",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := svc.Create(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("WalletService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && wallet == nil {
				t.Error("Expected wallet to be created")
			}
			if !tt.wantErr {
				if wallet.Name != tt.input.Name {
					t.Errorf("Name = %v, want %v", wallet.Name, tt.input.Name)
				}
				if !wallet.Balance.Equal(tt.input.InitialBalance) {
					t.Errorf("Balance = %v, want %v", wallet.Balance, tt.input.InitialBalance)
				}
				if !wallet.IsActive {
					t.Error("Expected wallet to be active")
				}
			}
		})
	}
}

func TestWalletService_GetTotalBalance(t *testing.T) {
	repo := newMockWalletRepo()
	svc := NewWalletService(repo)
	ctx := context.Background()

	// Create wallets
	_, _ = svc.Create(ctx, CreateWalletInput{
		Name:           "BCA",
		Type:           models.WalletTypeBank,
		Currency:       "IDR",
		InitialBalance: decimal.NewFromInt(1000000),
	})

	_, _ = svc.Create(ctx, CreateWalletInput{
		Name:           "Cash",
		Type:           models.WalletTypeCash,
		Currency:       "IDR",
		InitialBalance: decimal.NewFromInt(500000),
	})

	total, err := svc.GetTotalBalance(ctx)
	if err != nil {
		t.Fatalf("GetTotalBalance() error = %v", err)
	}

	expected := decimal.NewFromInt(1500000)
	if !total.Equal(expected) {
		t.Errorf("GetTotalBalance() = %v, want %v", total, expected)
	}
}

func TestWalletService_Delete(t *testing.T) {
	repo := newMockWalletRepo()
	svc := NewWalletService(repo)
	ctx := context.Background()

	// Create wallet
	wallet, _ := svc.Create(ctx, CreateWalletInput{
		Name:     "Test",
		Type:     models.WalletTypeCash,
		Currency: "IDR",
	})

	// Delete
	err := svc.Delete(ctx, wallet.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify soft deleted
	deleted, _ := svc.GetByID(ctx, wallet.ID)
	if deleted.IsActive {
		t.Error("Expected wallet to be inactive after delete")
	}
}
