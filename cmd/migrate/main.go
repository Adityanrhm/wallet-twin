package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Get database URL from env
	dbURL := getDBURL()

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	// Create migrator
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer m.Close()

	switch cmd {
	case "up":
		fmt.Println("⬆️  Running migrations...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✅ Migrations completed!")

	case "down":
		fmt.Println("⬇️  Rolling back last migration...")
		if err := m.Steps(-1); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("✅ Rollback completed!")

	case "reset":
		fmt.Println("🔄 Resetting database...")
		if err := m.Drop(); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}
		fmt.Println("✅ Database reset!")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("📌 Current version: %d (dirty: %v)\n", version, dirty)

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate force <version>")
		}
		var version int
		fmt.Sscanf(os.Args[2], "%d", &version)
		if err := m.Force(version); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Printf("✅ Forced to version %d\n", version)

	default:
		printUsage()
		os.Exit(1)
	}
}

func getDBURL() string {
	host := getEnv("WT_DATABASE_HOST", "localhost")
	port := getEnv("WT_DATABASE_PORT", "5432")
	user := getEnv("WT_DATABASE_USER", "postgres")
	password := getEnv("WT_DATABASE_PASSWORD", "postgres")
	name := getEnv("WT_DATABASE_NAME", "wallet_twin")
	sslmode := getEnv("WT_DATABASE_SSL_MODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, name, sslmode)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func printUsage() {
	fmt.Println(`
Usage: go run cmd/migrate/main.go <command>

Commands:
  up       Run all pending migrations
  down     Rollback last migration
  reset    Drop all tables
  version  Show current migration version
  force N  Force set migration version to N

Example:
  go run cmd/migrate/main.go up
`)
	flag.PrintDefaults()
}
