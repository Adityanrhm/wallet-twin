// Package app berisi bootstrap dan dependency injection untuk aplikasi.
//
// Dependency Injection (DI) adalah design pattern dimana dependencies
// diberikan ke object dari luar, bukan dibuat di dalam object.
//
// Kenapa DI penting?
//
//  1. Testability: Bisa inject mock dependencies saat testing
//  2. Flexibility: Mudah swap implementation (misal: ganti database)
//  3. Decoupling: Components tidak tightly-coupled
//
// Contoh TANPA DI (bad):
//
//	type WalletService struct {}
//	func (s *WalletService) GetWallet(id string) {
//	    db := database.NewPostgres(...)  // <-- Hardcoded dependency!
//	    db.Query(...)
//	}
//
// Contoh DENGAN DI (good):
//
//	type WalletService struct {
//	    repo repository.WalletRepository  // <-- Injected dari luar
//	}
//	func (s *WalletService) GetWallet(id string) {
//	    s.repo.GetByID(...)  // <-- Menggunakan injected dependency
//	}
//
// Dalam package ini, App struct adalah "composition root" yang
// menghubungkan semua dependencies bersama.
package app

import (
	"fmt"

	"github.com/Adityanrhm/wallet-twin/internal/config"
	"github.com/Adityanrhm/wallet-twin/internal/database"
)

// App adalah struct utama yang menyimpan semua dependencies aplikasi.
//
// App bertindak sebagai:
// - Dependency Injection Container
// - Application Lifecycle Manager
// - Central access point untuk semua services
//
// Pattern ini sering disebut "Composition Root" dalam DI terminology.
// Semua wiring dependencies dilakukan di satu tempat (New function).
//
// Contoh penggunaan:
//
//	app, err := app.New("./config")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer app.Close()
//
//	// Akses services melalui app
//	wallets, err := app.WalletService.List()
type App struct {
	// Config menyimpan konfigurasi aplikasi
	// Diload dari config.yaml dan environment variables
	Config *config.Config

	// DB adalah koneksi ke PostgreSQL
	// Gunakan untuk operasi database
	DB *database.PostgresDB

	// Services akan ditambahkan di sini setelah dibuat:
	// WalletService  *service.WalletService
	// CategoryService *service.CategoryService
	// TransactionService *service.TransactionService
	// ... dst
}

// New membuat instance baru dari App dengan semua dependencies.
//
// Flow initialization:
//  1. Load configuration dari file dan env vars
//  2. Validate configuration
//  3. Connect ke database
//  4. Initialize repositories (akan ditambahkan nanti)
//  5. Initialize services (akan ditambahkan nanti)
//  6. Return App yang siap digunakan
//
// Parameter:
//   - configPath: path ke config file tanpa extension
//     Contoh: "./config" akan mencari config.yaml
//
// Return error jika ada langkah initialization yang gagal.
// Caller harus memanggil Close() saat selesai menggunakan App.
//
// Contoh:
//
//	app, err := app.New("./config")
//	if err != nil {
//	    log.Fatal("Failed to initialize app:", err)
//	}
//	defer app.Close()
func New(configPath string) (*App, error) {
	// 1. Load configuration
	// Config diload pertama karena diperlukan oleh semua komponen lain
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Validate configuration
	// Pastikan semua required values terisi dengan benar
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 3. Connect ke database
	// Database connection adalah fundamental, jadi connect early
	// Ini juga memvalidasi bahwa database accessible
	db, err := database.NewPostgres(cfg.Database.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 4. Initialize repositories (akan ditambahkan nanti)
	// Repositories menggunakan db untuk akses data
	//
	// Contoh saat repositories sudah dibuat:
	// walletRepo := postgres.NewWalletRepository(db.Pool)
	// categoryRepo := postgres.NewCategoryRepository(db.Pool)

	// 5. Initialize services (akan ditambahkan nanti)
	// Services menggunakan repositories untuk business logic
	//
	// Contoh saat services sudah dibuat:
	// walletService := service.NewWalletService(walletRepo)
	// categoryService := service.NewCategoryService(categoryRepo)

	// 6. Return App dengan semua dependencies
	return &App{
		Config: cfg,
		DB:     db,
		// Services akan ditambahkan di sini
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
