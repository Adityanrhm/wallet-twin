// Package main adalah entry point untuk aplikasi Wallet Twin.
//
// File main.go biasanya berisi:
// 1. Parse command line arguments
// 2. Initialize application
// 3. Run CLI atau TUI
// 4. Handle graceful shutdown
//
// Dalam Go, package main adalah special package yang menghasilkan
// executable binary. Setiap executable Go harus memiliki:
// - package main
// - func main()
//
// Untuk menjalankan aplikasi:
//
//	# Development
//	go run cmd/wallet/main.go
//
//	# Build dan run
//	go build -o wallet.exe cmd/wallet/main.go
//	./wallet.exe
//
//	# Atau menggunakan go install
//	go install ./cmd/wallet
//	wallet
package main

import (
	"fmt"
	"os"

	"github.com/Adityanrhm/wallet-twin/internal/app"
)

// main adalah entry point aplikasi.
//
// Flow:
//  1. Initialize App dengan semua dependencies
//  2. Setup graceful shutdown handler
//  3. Run CLI commands (Cobra)
//  4. Cleanup saat exit
//
// Exit codes:
//   - 0: Success
//   - 1: Error
func main() {
	// Initialize application
	// "./config" akan mencari config.yaml di current directory
	application, err := app.New("./config")
	if err != nil {
		// Print error dan exit dengan code 1
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Pastikan cleanup dilakukan saat exit
	// defer akan execute saat main() function selesai
	defer func() {
		if err := application.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error during cleanup: %v\n", err)
		}
	}()

	// Temporary: Print success message
	// Ini akan diganti dengan CLI commands (Cobra) di commit selanjutnya
	fmt.Println("🎉 Wallet Twin initialized successfully!")
	fmt.Printf("📊 Database: %s\n", application.Config.Database.Host)
	fmt.Printf("💰 Currency: %s\n", application.Config.App.Currency)

	// TODO: Run CLI commands
	// Akan ditambahkan setelah cli package selesai:
	//
	// if err := cli.Execute(application); err != nil {
	//     fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	//     os.Exit(1)
	// }
}
