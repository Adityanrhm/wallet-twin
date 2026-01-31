// Package database mengelola koneksi ke PostgreSQL menggunakan pgx.
//
// pgx adalah PostgreSQL driver native untuk Go yang menawarkan:
// - Performa lebih baik daripada database/sql
// - Native PostgreSQL types support
// - Connection pooling dengan pgxpool
// - COPY protocol untuk bulk insert
// - LISTEN/NOTIFY support
//
// Kenapa pgx bukan database/sql?
//
//  1. Performance: pgx tidak perlu convert ke interface{} untuk setiap scan
//  2. Features: Support PostgreSQL-specific features seperti ARRAY, JSONB, dll
//  3. Pooling: Built-in connection pool yang lebih efisien
//
// Connection Pooling:
//
// Pool memungkinkan reuse koneksi database. Tanpa pool, setiap query
// akan membuat koneksi baru → lambat dan boros resource.
//
// Dengan pool:
//
//	Query 1 → Ambil koneksi dari pool → Execute → Kembalikan ke pool
//	Query 2 → Ambil koneksi dari pool → Execute → Kembalikan ke pool
//	                          ↓
//	              Koneksi yang sama bisa dipakai ulang!
//
// Contoh penggunaan:
//
//	db, err := database.NewPostgres(cfg.Database.ConnectionString())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Gunakan db.Pool untuk query
//	rows, err := db.Pool.Query(ctx, "SELECT * FROM wallets WHERE user_id = $1", userID)
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB adalah wrapper untuk pgxpool.Pool.
//
// Wrapper ini menyediakan:
// - Encapsulation: Hide internal implementation details
// - Additional methods: Health check, migrations, utilities
// - Easier testing: Bisa di-mock untuk unit tests
//
// Pool field adalah thread-safe, bisa digunakan dari multiple goroutines
// secara concurrent tanpa synchronization tambahan.
type PostgresDB struct {
	// Pool adalah connection pool ke PostgreSQL.
	// Thread-safe untuk concurrent access.
	//
	// Gunakan Pool untuk semua operasi database:
	//   db.Pool.Query(ctx, sql, args...)
	//   db.Pool.QueryRow(ctx, sql, args...)
	//   db.Pool.Exec(ctx, sql, args...)
	Pool *pgxpool.Pool

	// connString disimpan untuk keperluan reconnection atau logging
	// SECURITY: Jangan log connString karena berisi password!
	connString string
}

// NewPostgres membuat koneksi baru ke PostgreSQL dengan connection pooling.
//
// Fungsi ini melakukan:
//  1. Parse connection string
//  2. Configure pool settings (max connections, timeouts, dll)
//  3. Create connection pool
//  4. Test koneksi dengan Ping
//
// Connection String Format:
//
//	postgres://user:password@host:port/dbname?sslmode=disable
//
// Pool Configuration:
//   - MaxConns: 10 (maximum concurrent connections)
//   - MinConns: 2 (minimum idle connections to keep)
//   - MaxConnLifetime: 1 hour (recreate connection after this time)
//   - MaxConnIdleTime: 30 minutes (close idle connection after this)
//
// Error dikembalikan jika:
//   - Connection string invalid
//   - Tidak bisa connect ke database
//   - Ping gagal
//
// Contoh:
//
//	db, err := database.NewPostgres("postgres://postgres:pass@localhost:5432/wallet_twin?sslmode=disable")
//	if err != nil {
//	    log.Fatal("Cannot connect to database:", err)
//	}
//	defer db.Close()
func NewPostgres(connString string) (*PostgresDB, error) {
	// Parse connection string ke config object
	// Ini memungkinkan kita untuk modify config sebelum create pool
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool settings
	// Nilai-nilai ini bisa di-tune berdasarkan kebutuhan aplikasi
	//
	// MaxConns: Jangan terlalu tinggi karena PostgreSQL punya limit koneksi
	// Default PostgreSQL max_connections = 100, sisakan untuk admin, tools, dll
	config.MaxConns = 10

	// MinConns: Minimal koneksi yang selalu siap
	// Mengurangi latency saat ada burst request karena tidak perlu buka koneksi baru
	config.MinConns = 2

	// MaxConnLifetime: Recreate koneksi setelah 1 jam
	// Penting untuk:
	// - Handle database restart/failover
	// - Clean up leaked connections
	// - Prevent stale connections di network
	config.MaxConnLifetime = time.Hour

	// MaxConnIdleTime: Close koneksi idle setelah 30 menit
	// Mengurangi resource usage saat aplikasi idle
	config.MaxConnIdleTime = 30 * time.Minute

	// HealthCheckPeriod: Check koneksi setiap 1 menit
	// Pool akan otomatis remove koneksi yang tidak healthy
	config.HealthCheckPeriod = time.Minute

	// Create context dengan timeout untuk initial connection
	// Jika tidak bisa connect dalam 10 detik, gagalkan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection dengan Ping
	// Ini memastikan database accessible dan credentials benar
	if err := pool.Ping(ctx); err != nil {
		pool.Close() // Cleanup jika gagal
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{
		Pool:       pool,
		connString: connString,
	}, nil
}

// Close menutup semua koneksi dalam pool.
//
// PENTING: Selalu panggil Close() saat aplikasi shutdown!
// Gunakan defer untuk memastikan cleanup:
//
//	db, err := database.NewPostgres(connString)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()  // <-- Ini memastikan pool ditutup saat app exit
//
// Close akan:
// - Menunggu queries yang sedang berjalan selesai
// - Menutup semua koneksi
// - Release semua resources
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Ping melakukan health check ke database.
//
// Use cases:
// - Kubernetes liveness/readiness probe
// - Health check endpoint
// - Verify koneksi masih valid
// - Monitoring dan alerting
//
// Return nil jika database healthy dan dapat diakses.
// Return error jika koneksi gagal atau timeout.
//
// Contoh untuk HTTP health endpoint:
//
//	func healthHandler(db *database.PostgresDB) http.HandlerFunc {
//	    return func(w http.ResponseWriter, r *http.Request) {
//	        if err := db.Ping(r.Context()); err != nil {
//	            w.WriteHeader(http.StatusServiceUnavailable)
//	            w.Write([]byte("Database unhealthy"))
//	            return
//	        }
//	        w.Write([]byte("OK"))
//	    }
//	}
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats mengembalikan statistik connection pool.
//
// Berguna untuk monitoring dan debugging:
// - AcquireCount: Total koneksi yang pernah di-acquire
// - AcquiredConns: Koneksi yang sedang digunakan
// - TotalConns: Total koneksi dalam pool
// - IdleConns: Koneksi yang idle/available
//
// Contoh:
//
//	stats := db.Stats()
//	fmt.Printf("Active: %d, Idle: %d, Total: %d\n",
//	    stats.AcquiredConns(),
//	    stats.IdleConns(),
//	    stats.TotalConns(),
//	)
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}
