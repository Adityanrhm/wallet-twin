// Package app berisi bootstrap dan dependency injection untuk aplikasi.
//
// Dependency Injection (DI) adalah design pattern dimana dependencies
// diberikan ke object dari luar, bukan dibuat di dalam object.
//
// Dalam package ini, App struct adalah "composition root" yang
// menghubungkan semua dependencies bersama.
package app

import (
	"fmt"

	"github.com/Adityanrhm/wallet-twin/internal/config"
	"github.com/Adityanrhm/wallet-twin/internal/database"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
	"github.com/Adityanrhm/wallet-twin/internal/repository/postgres"
)

// Repos menyimpan semua repository instances.
type Repos struct {
	Wallet      repository.WalletRepository
	Category    repository.CategoryRepository
	Transaction repository.TransactionRepository
	Transfer    repository.TransferRepository
	Budget      repository.BudgetRepository
	Recurring   repository.RecurringRepository
	Goal        repository.GoalRepository
}

// App adalah struct utama yang menyimpan semua dependencies aplikasi.
type App struct {
	// Config menyimpan konfigurasi aplikasi
	Config *config.Config

	// DB adalah koneksi ke PostgreSQL
	DB *database.PostgresDB

	// Repos menyimpan semua repository instances
	Repos *Repos
}

// New membuat instance baru dari App dengan semua dependencies.
func New(configPath string) (*App, error) {
	// 1. Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 3. Connect ke database
	db, err := database.NewPostgres(cfg.Database.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 4. Initialize repositories
	repos := &Repos{
		Wallet:      postgres.NewWalletRepository(db.Pool),
		Category:    postgres.NewCategoryRepository(db.Pool),
		Transaction: postgres.NewTransactionRepository(db.Pool),
		Transfer:    postgres.NewTransferRepository(db.Pool),
		Budget:      postgres.NewBudgetRepository(db.Pool),
		Recurring:   postgres.NewRecurringRepository(db.Pool),
		Goal:        postgres.NewGoalRepository(db.Pool),
	}

	// 5. Return App dengan semua dependencies
	return &App{
		Config: cfg,
		DB:     db,
		Repos:  repos,
	}, nil
}

// Close membersihkan semua resources yang digunakan oleh App.
//
// PENTING: Selalu panggil Close() saat aplikasi selesai!
// Best practice adalah menggunakan defer:
//
//	app, err := app.New("./config")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer app.Close()  // <-- Cleanup otomatis saat function selesai
//
// Close akan:
//   - Menutup connection pool database
//   - Cleanup resources lainnya (jika ada)
//
// Close aman dipanggil multiple times.
func (a *App) Close() error {
	// Close database connection
	if a.DB != nil {
		a.DB.Close()
	}

	// Cleanup resources lainnya akan ditambahkan di sini
	// Contoh: close file handles, stop background workers, dll

	return nil
}
