// Package service berisi business logic layer.
//
// Service Layer dalam Clean Architecture:
// - Berisi use cases / business rules
// - Orchestrate repositories untuk operasi kompleks
// - Handle transactions yang melibatkan multiple repositories
// - Validation business rules
//
// Service TIDAK boleh:
// - Akses database langsung (gunakan repository)
// - Tahu tentang HTTP, CLI, atau UI lainnya
// - Berisi presentation logic
//
// Contoh use case di WalletService:
//
//	func (s *WalletService) Transfer(ctx context.Context, from, to uuid.UUID, amount decimal.Decimal) error {
//	    // 1. Validate: cek saldo cukup
//	    fromWallet, err := s.walletRepo.GetByID(ctx, from)
//	    if err != nil {
//	        return err
//	    }
//	    if fromWallet.Balance.LessThan(amount) {
//	        return ErrInsufficientBalance
//	    }
//
//	    // 2. Begin transaction (atomic operation)
//	    tx, err := s.db.Begin(ctx)
//	    if err != nil {
//	        return err
//	    }
//	    defer tx.Rollback(ctx)
//
//	    // 3. Deduct from source wallet
//	    err = s.walletRepo.UpdateBalance(ctx, tx, from, fromWallet.Balance.Sub(amount))
//	    if err != nil {
//	        return err
//	    }
//
//	    // 4. Add to destination wallet
//	    toWallet, _ := s.walletRepo.GetByID(ctx, to)
//	    err = s.walletRepo.UpdateBalance(ctx, tx, to, toWallet.Balance.Add(amount))
//	    if err != nil {
//	        return err
//	    }
//
//	    // 5. Record transfer
//	    err = s.transferRepo.Create(ctx, tx, &models.Transfer{...})
//	    if err != nil {
//	        return err
//	    }
//
//	    // 6. Commit transaction
//	    return tx.Commit(ctx)
//	}
package service

// TODO: Add service implementations
