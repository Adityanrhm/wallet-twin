// Package service berisi business logic untuk aplikasi Wallet Twin.
//
// Dalam Clean Architecture, service layer adalah tempat untuk:
// - Business rules dan validations
// - Orchestrating repository operations
// - Transactions (atomic operations)
// - Domain events handling
//
// Service BUKAN tempat untuk:
// - Database queries (itu tugas repository)
// - HTTP handling (itu tugas handler/controller)
// - CLI parsing (itu tugas command)
//
// Setiap service menerima dependencies via constructor (Dependency Injection).
// Ini memudahkan testing dengan mock objects.
//
// Contoh:
//
//	walletRepo := postgres.NewWalletRepository(pool)
//	walletService := service.NewWalletService(walletRepo)
//
//	// Dalam test
//	mockRepo := &MockWalletRepository{}
//	walletService := service.NewWalletService(mockRepo)
package service
