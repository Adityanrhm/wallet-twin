package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// TransferService menangani business logic untuk transfer operations.
//
// Transfer adalah operasi ATOMIC yang melibatkan 2 wallets:
// - Wallet sumber: balance -= (amount + fee)
// - Wallet tujuan: balance += amount
//
// Fee adalah biaya transfer yang "hilang" (tidak masuk ke manapun).
type TransferService struct {
	transferRepo repository.TransferRepository
	walletRepo   repository.WalletRepository
	txManager    repository.TransactionManager
}

// NewTransferService membuat TransferService baru.
func NewTransferService(
	transferRepo repository.TransferRepository,
	walletRepo repository.WalletRepository,
	txManager repository.TransactionManager,
) *TransferService {
	return &TransferService{
		transferRepo: transferRepo,
		walletRepo:   walletRepo,
		txManager:    txManager,
	}
}

// Create membuat transfer baru dan update kedua wallet balances.
//
// Contoh:
//
//	transfer, err := transferService.Create(ctx, service.CreateTransferInput{
//	    FromWalletID: bcaID,
//	    ToWalletID:   gopayID,
//	    Amount:       decimal.NewFromInt(500000),
//	    Fee:          decimal.NewFromInt(6500),
//	    Note:         "Top up GoPay",
//	})
func (s *TransferService) Create(ctx context.Context, input CreateTransferInput) (*models.Transfer, error) {
	// Validate same wallet
	if input.FromWalletID == input.ToWalletID {
		return nil, errors.New("cannot transfer to the same wallet")
	}

	// Get source wallet
	fromWallet, err := s.walletRepo.GetByID(ctx, input.FromWalletID)
	if err != nil {
		return nil, fmt.Errorf("source wallet not found: %w", err)
	}

	if !fromWallet.IsActive {
		return nil, errors.New("source wallet is inactive")
	}

	// Get destination wallet
	toWallet, err := s.walletRepo.GetByID(ctx, input.ToWalletID)
	if err != nil {
		return nil, fmt.Errorf("destination wallet not found: %w", err)
	}

	if !toWallet.IsActive {
		return nil, errors.New("destination wallet is inactive")
	}

	// Calculate total deducted from source
	totalDeducted := input.Amount.Add(input.Fee)

	// Check balance
	if fromWallet.Balance.LessThan(totalDeducted) {
		return nil, ErrInsufficientBalance
	}

	// Create transfer model
	transfer := models.NewTransfer(input.FromWalletID, input.ToWalletID, input.Amount)
	transfer.Fee = input.Fee
	transfer.Note = input.Note

	if err := transfer.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Calculate new balances
	fromNewBalance := fromWallet.Balance.Sub(totalDeducted)
	toNewBalance := toWallet.Balance.Add(input.Amount)

	// Execute in transaction (ATOMIC)
	err = s.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		// Create transfer record
		if err := s.transferRepo.Create(ctx, transfer); err != nil {
			return fmt.Errorf("failed to create transfer: %w", err)
		}

		// Update source wallet
		if err := s.walletRepo.UpdateBalance(ctx, fromWallet.ID, fromNewBalance); err != nil {
			return fmt.Errorf("failed to update source balance: %w", err)
		}

		// Update destination wallet
		if err := s.walletRepo.UpdateBalance(ctx, toWallet.ID, toNewBalance); err != nil {
			return fmt.Errorf("failed to update destination balance: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return transfer, nil
}

// GetByID mengambil transfer berdasarkan ID.
func (s *TransferService) GetByID(ctx context.Context, id uuid.UUID) (*models.Transfer, error) {
	transfer, err := s.transferRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}
	return transfer, nil
}

// List mengambil transfers dengan filter.
func (s *TransferService) List(
	ctx context.Context,
	filter repository.TransferFilter,
	params repository.ListParams,
) ([]*models.Transfer, error) {
	transfers, err := s.transferRepo.List(ctx, filter, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfers: %w", err)
	}
	return transfers, nil
}

// GetByWallet mengambil semua transfers yang melibatkan wallet tertentu.
func (s *TransferService) GetByWallet(
	ctx context.Context,
	walletID uuid.UUID,
	params repository.ListParams,
) ([]*models.Transfer, error) {
	filter := repository.TransferFilter{WalletID: &walletID}
	return s.List(ctx, filter, params)
}

// CreateTransferInput adalah input untuk membuat transfer.
type CreateTransferInput struct {
	FromWalletID uuid.UUID
	ToWalletID   uuid.UUID
	Amount       decimal.Decimal
	Fee          decimal.Decimal
	Note         string
}
