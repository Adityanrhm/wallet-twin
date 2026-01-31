// Package config mengelola konfigurasi aplikasi menggunakan Viper.
//
// Viper adalah library configuration management yang populer di Go.
// Digunakan oleh banyak project besar seperti Hugo dan Kubernetes CLI.
//
// Fitur Viper yang kita gunakan:
// 1. Membaca dari file YAML/JSON/TOML
// 2. Membaca dari environment variables
// 3. Default values
// 4. Automatic environment binding
//
// Prioritas konfigurasi (tertinggi ke terendah):
// 1. Environment variables (WALLET_DATABASE_HOST, dll)
// 2. Config file (config.yaml)
// 3. Default values (didefinisikan di kode)
//
// Environment Variable Format:
//
//	WALLET_DATABASE_HOST     → database.host
//	WALLET_DATABASE_PORT     → database.port
//	WALLET_APP_CURRENCY      → app.currency
//
// Contoh penggunaan:
//
//	cfg, err := config.Load("./config")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println(cfg.Database.Host)      // "localhost"
//	fmt.Println(cfg.App.Currency)       // "IDR"
//	fmt.Println(cfg.Database.ConnectionString())
//	// Output: postgres://postgres:postgres@localhost:5432/wallet_twin?sslmode=disable
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config adalah struct utama yang menyimpan semua konfigurasi aplikasi.
//
// Struct ini di-populate oleh Viper dari config file dan environment variables.
// Tag `mapstructure` digunakan oleh Viper untuk mapping dari config file ke struct fields.
//
// Contoh config.yaml yang akan di-map ke struct ini:
//
//	database:
//	  host: localhost
//	  port: 5432
//	app:
//	  name: Wallet Twin
type Config struct {
	// Database berisi konfigurasi koneksi PostgreSQL
	Database DatabaseConfig `mapstructure:"database"`

	// App berisi konfigurasi umum aplikasi
	App AppConfig `mapstructure:"app"`

	// TUI berisi konfigurasi Terminal UI
	TUI TUIConfig `mapstructure:"tui"`
}

// DatabaseConfig menyimpan konfigurasi koneksi PostgreSQL.
//
// Semua field diperlukan untuk membuat koneksi database.
// SSLMode akan default ke "disable" jika tidak diisi.
//
// Untuk production, pastikan menggunakan SSL dengan ssl_mode: require
type DatabaseConfig struct {
	// Host adalah alamat server database
	// Contoh: "localhost", "db.example.com", "192.168.1.100"
	Host string `mapstructure:"host"`

	// Port adalah port PostgreSQL (default: 5432)
	Port int `mapstructure:"port"`

	// Name adalah nama database yang akan digunakan
	Name string `mapstructure:"name"`

	// User adalah username untuk autentikasi
	User string `mapstructure:"user"`

	// Password adalah password untuk autentikasi
	// SECURITY: Di production, gunakan environment variable!
	Password string `mapstructure:"password"`

	// SSLMode mengatur mode SSL untuk koneksi
	// Options: disable, require, verify-ca, verify-full
	SSLMode string `mapstructure:"ssl_mode"`
}

// AppConfig menyimpan konfigurasi umum aplikasi.
type AppConfig struct {
	// Name adalah nama aplikasi yang ditampilkan di UI
	Name string `mapstructure:"name"`

	// Currency adalah kode mata uang default (ISO 4217)
	// Contoh: "IDR", "USD", "EUR"
	Currency string `mapstructure:"currency"`

	// Locale untuk formatting tanggal dan angka
	// Contoh: "id-ID", "en-US"
	Locale string `mapstructure:"locale"`
}

// TUIConfig menyimpan konfigurasi untuk Terminal UI.
type TUIConfig struct {
	// Theme adalah nama theme warna
	// Options: "default", "dark", "light"
	Theme string `mapstructure:"theme"`

	// RefreshRate adalah interval refresh dashboard dalam milliseconds
	RefreshRate int `mapstructure:"refresh_rate"`
}

// Load membaca konfigurasi dari environment variables.
//
// Environment Variable Format:
//
//	WT_DATABASE_HOST     → database.host
//	WT_DATABASE_PORT     → database.port
//	WT_APP_CURRENCY      → app.currency
//
// Prefix adalah "WT" (Wallet Twin).
//
// Contoh:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
func Load(_ string) (*Config, error) {
	// 1. Set default values
	setDefaults()

	// 2. Enable automatic environment variable binding
	// Prefix "WT" → WT_DATABASE_HOST, WT_APP_NAME, dll
	viper.SetEnvPrefix("WT")

	// Replace "." dengan "_" untuk nested keys
	// database.host → DATABASE_HOST
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Automatically read matching env vars
	viper.AutomaticEnv()

	// 3. Unmarshal ke Config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// setDefaults mengatur nilai default untuk semua konfigurasi.
//
// Defaults digunakan ketika:
// - Config file tidak ada
// - Key tertentu tidak ada di config file
// - Environment variable tidak di-set
//
// Ini memastikan aplikasi bisa berjalan dengan konfigurasi minimal.
func setDefaults() {
	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "wallet_twin")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.ssl_mode", "disable")

	// App defaults
	viper.SetDefault("app.name", "Wallet Twin")
	viper.SetDefault("app.currency", "IDR")
	viper.SetDefault("app.locale", "id-ID")

	// TUI defaults
	viper.SetDefault("tui.theme", "default")
	viper.SetDefault("tui.refresh_rate", 1000)
}

// ConnectionString membuat PostgreSQL connection string dari DatabaseConfig.
//
// Format yang dihasilkan:
//
//	postgres://user:password@host:port/dbname?sslmode=disable
//
// Format ini compatible dengan pgx dan database/sql.
//
// Contoh output:
//
//	postgres://postgres:secret@localhost:5432/wallet_twin?sslmode=disable
//
// SECURITY NOTE:
// Connection string berisi password! Jangan log atau print ke output.
func (d *DatabaseConfig) ConnectionString() string {
	// Format: postgres://user:password@host:port/dbname?sslmode=X
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
		d.SSLMode,
	)
}

// Validate memeriksa apakah konfigurasi valid.
//
// Validasi yang dilakukan:
// - Database host tidak kosong
// - Database port dalam range valid (1-65535)
// - Database name tidak kosong
// - Currency code valid (3 karakter)
//
// Return error jika ada validasi yang gagal.
func (c *Config) Validate() error {
	// Validate database config
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate app config
	if len(c.App.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter ISO code (e.g., IDR, USD)")
	}

	return nil
}
