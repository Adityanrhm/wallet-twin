// Package main adalah entry point untuk aplikasi Wallet Twin.
package main

import (
	"fmt"
	"os"

	"github.com/Adityanrhm/wallet-twin/internal/app"
	"github.com/Adityanrhm/wallet-twin/internal/cli"
)

// main adalah entry point aplikasi.
//
// Flow:
//  1. Initialize App dengan semua dependencies
//  2. Run CLI commands (Cobra)
//  3. Cleanup saat exit
func main() {
	// Initialize application
	application, err := app.New("./config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Pastikan cleanup dilakukan saat exit
	defer func() {
		if err := application.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error during cleanup: %v\n", err)
		}
	}()

	// Run CLI commands
	if err := cli.Execute(application); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

