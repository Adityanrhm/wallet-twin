package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// WalletService menangani business logic untuk wallet operations.
//
// Responsibilities:
// - CRUD operations untuk wallet
// - Validasi business rules
// - Menghitung total balance
//
// WalletService TIDAK langsung update balance.
// Balance diupdate melalui TransactionService saat ada transaksi.
type WalletService struct {
	repo repository.WalletRepository
}

// NewWalletService membuat WalletService baru.
//
//	walletRepo := postgres.NewWalletRepository(pool)
//	walletService := service.NewWalletService(walletRepo)
func NewWalletService(repo repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

// Create membuat wallet baru.
//
// Validasi:
// - Name tidak boleh kosong
// - Type harus valid (cash, bank, ewallet)
// - Currency harus 3 karakter
//
// Contoh:
//
//	wallet, err := walletService.Create(ctx, service.CreateWalletInput{
//	    Name:     "BCA Tabungan",
//	    Type:     models.WalletTypeBank,
//	    Currency: "IDR",
//	})
func (s *WalletService) Create(ctx context.Context, input CreateWalletInput) (*models.Wallet, error) {
	wallet := &models.Wallet{
		BaseModel: models.BaseModel{ID: models.NewID()},
		Name:      input.Name,
		Type:      input.Type,
		Balance:   input.InitialBalance,
		Currency:  input.Currency,
		Color:     input.Color,
		Icon:      input.Icon,
		IsActive:  true,
	}

	// Validate wallet
	if err := wallet.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create in database
	if err := s.repo.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

// GetByID mengambil wallet berdasarkan ID.
func (s *WalletService) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return wallet, nil
}

// List mengambil semua wallets dengan filter.
func (s *WalletService) List(ctx context.Context, filter repository.WalletFilter) ([]*models.Wallet, error) {
	wallets, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}
	return wallets, nil
}

// ListActive mengambil semua wallet aktif.
// Shortcut untuk filter IsActive = true.
func (s *WalletService) ListActive(ctx context.Context) ([]*models.Wallet, error) {
	isActive := true
	return s.List(ctx, repository.WalletFilter{IsActive: &isActive})
}

// Update memperbarui wallet.
func (s *WalletService) Update(ctx context.Context, input UpdateWalletInput) (*models.Wallet, error) {
	// Get existing wallet
	wallet, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Update fields
	if input.Name != nil {
		wallet.Name = *input.Name
	}
	if input.Type != nil {
		wallet.Type = *input.Type
	}
	if input.Currency != nil {
		wallet.Currency = *input.Currency
	}
	if input.Color != nil {
		wallet.Color = *input.Color
	}
	if input.Icon != nil {
		wallet.Icon = *input.Icon
	}

	// Validate
	if err := wallet.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update in database
	if err := s.repo.Update(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return wallet, nil
}

// Delete menghapus wallet (soft delete).
func (s *WalletService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}
	return nil
}

// GetTotalBalance menghitung total saldo semua wallet aktif.
func (s *WalletService) GetTotalBalance(ctx context.Context) (decimal.Decimal, error) {
	total, err := s.repo.GetTotalBalance(ctx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get total balance: %w", err)
	}
	return total, nil
}

// CreateWalletInput adalah input untuk membuat wallet baru.
type CreateWalletInput struct {
	Name           string
	Type           models.WalletType
	Currency       string
	InitialBalance decimal.Decimal
	Color          string
	Icon           string
}

// UpdateWalletInput adalah input untuk update wallet.
// Field yang nil tidak akan diupdate.
type UpdateWalletInput struct {
	ID       uuid.UUID
	Name     *string
	Type     *models.WalletType
	Currency *string
	Color    *string
	Icon     *string
}
