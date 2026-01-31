// Package database menyediakan utilities untuk koneksi dan migrasi database.
//
// File ini berisi migration runner menggunakan golang-migrate.
//
// golang-migrate adalah library standard untuk database migrations di Go.
// Digunakan oleh banyak project besar untuk manage database schema.
//
// Konsep Migration:
//
// Migration adalah cara untuk mengelola perubahan database schema secara
// incremental dan version-controlled. Setiap migration memiliki:
// - UP: Script untuk apply perubahan (CREATE TABLE, ADD COLUMN, dll)
// - DOWN: Script untuk rollback perubahan (DROP TABLE, DROP COLUMN, dll)
//
// Contoh migration files:
//
//	migrations/
//	├── 000001_create_wallets.up.sql     # CREATE TABLE wallets(...)
//	├── 000001_create_wallets.down.sql   # DROP TABLE wallets
//	├── 000002_create_categories.up.sql  # CREATE TABLE categories(...)
//	└── 000002_create_categories.down.sql
//
// Workflow:
//  1. Buat migration files dengan naming convention: {version}_{name}.{up|down}.sql
//  2. Jalankan migrasi: wallet init (atau migrate up)
//  3. Jika perlu rollback: migrate down
//
// golang-migrate menyimpan versi migration yang sudah dijalankan di tabel
// schema_migrations, sehingga tahu migration mana yang sudah applied.
package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	// Blank import untuk driver PostgreSQL
	// Driver ini perlu di-import agar golang-migrate tau cara connect ke PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Blank import untuk source file
	// Ini memungkinkan membaca migration files dari filesystem
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrator adalah wrapper untuk golang-migrate.
//
// Menyediakan interface yang lebih sederhana untuk:
// - Up: Apply semua pending migrations
// - Down: Rollback semua migrations
// - Steps: Apply/rollback N migrations
// - Version: Cek versi migration saat ini
type Migrator struct {
	// migrate adalah instance golang-migrate
	migrate *migrate.Migrate
}

// NewMigrator membuat instance Migrator baru.
//
// Parameters:
//   - databaseURL: PostgreSQL connection string
//     Format: postgres://user:password@host:port/dbname?sslmode=disable
//   - migrationsPath: Path ke folder migration files
//     Contoh: "file://./migrations"
//
// Contoh:
//
//	migrator, err := database.NewMigrator(
//	    "postgres://postgres:pass@localhost:5432/wallet_twin?sslmode=disable",
//	    "file://./migrations",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer migrator.Close()
func NewMigrator(databaseURL, migrationsPath string) (*Migrator, error) {
	// Buat instance migrate dengan source dan database
	// Source: "file://./migrations" - folder berisi SQL files
	// Database: connection string PostgreSQL
	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

// Up menjalankan semua pending migrations.
//
// Migration dijalankan secara berurutan berdasarkan version number.
// Jika salah satu migration gagal, proses berhenti dan return error.
//
// Idempotent: Aman dijalankan berulang kali. Migration yang sudah
// applied akan di-skip.
//
// Contoh:
//
//	err := migrator.Up()
//	if err != nil {
//	    log.Fatal("Migration failed:", err)
//	}
//	fmt.Println("All migrations applied successfully!")
func (m *Migrator) Up() error {
	err := m.migrate.Up()
	if err != nil {
		// ErrNoChange means all migrations already applied
		// Ini bukan error, tapi kondisi normal
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Down rollback semua migrations.
//
// WARNING: Ini akan menghapus semua tabel dan data!
// Gunakan dengan hati-hati, terutama di production.
//
// Biasanya digunakan untuk:
// - Testing: Reset database sebelum test
// - Development: Reset saat mengubah schema
//
// JANGAN gunakan di production kecuali benar-benar perlu!
func (m *Migrator) Down() error {
	err := m.migrate.Down()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	return nil
}

// Steps menjalankan N migrations ke atas (n > 0) atau ke bawah (n < 0).
//
// Parameters:
//   - n: Jumlah steps
//     Positif: migrate up N steps
//     Negatif: migrate down N steps
//
// Contoh:
//
//	// Apply 1 migration
//	err := migrator.Steps(1)
//
//	// Rollback 1 migration
//	err := migrator.Steps(-1)
//
// Berguna untuk:
// - Rollback satu migration yang bermasalah
// - Apply migration satu per satu untuk debugging
func (m *Migrator) Steps(n int) error {
	err := m.migrate.Steps(n)
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("failed to run %d migration steps: %w", n, err)
	}
	return nil
}

// Version mengembalikan versi migration saat ini.
//
// Return:
//   - version: Nomor versi migration terakhir yang applied
//   - dirty: True jika migration terakhir gagal di tengah jalan
//   - error: Error jika gagal membaca versi
//
// Dirty state terjadi jika migration gagal sebelum selesai.
// Kamu perlu fix secara manual dan force version.
//
// Contoh:
//
//	version, dirty, err := migrator.Version()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)
func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Force sets migration version tanpa menjalankan migration.
//
// WARNING: Gunakan dengan sangat hati-hati!
//
// Use case:
// - Fix dirty state setelah migration gagal
// - Skip migration yang bermasalah (setelah fix manual)
//
// Contoh:
//
//	// Setelah fix manual migration 5
//	err := migrator.Force(5)
func (m *Migrator) Force(version int) error {
	return m.migrate.Force(version)
}

// Close menutup koneksi database yang digunakan oleh migrator.
//
// Selalu panggil Close setelah selesai menggunakan Migrator.
// Gunakan defer untuk memastikan cleanup:
//
//	migrator, err := database.NewMigrator(...)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer migrator.Close()
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	return dbErr
}
