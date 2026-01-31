package export

import (
	"context"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// PDFExporter creates professional PDF reports.
type PDFExporter struct {
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
}

// NewPDFExporter creates a new PDFExporter.
func NewPDFExporter(
	walletRepo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
) *PDFExporter {
	return &PDFExporter{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
	}
}

// TransactionsToPDF exports transactions to a professional PDF file.
func (e *PDFExporter) TransactionsToPDF(ctx context.Context, filename string, filter repository.TransactionFilter) error {
	// Get data
	params := repository.ListParams{Limit: 1000, Offset: 0}
	transactions, err := e.transactionRepo.List(ctx, filter, params)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header
	pdf.SetFillColor(79, 70, 229) // Purple
	pdf.Rect(0, 0, 210, 35, "F")

	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetY(12)
	pdf.CellFormat(0, 10, "TRANSACTION REPORT", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Generated: %s", time.Now().Format("02 January 2006, 15:04")), "", 1, "C", false, 0, "")

	// Reset colors
	pdf.SetTextColor(0, 0, 0)
	pdf.SetY(45)

	// Summary box
	var totalIncome, totalExpense float64
	for _, tx := range transactions {
		amount, _ := tx.Amount.Float64()
		if tx.Type == models.TransactionTypeIncome {
			totalIncome += amount
		} else {
			totalExpense += amount
		}
	}

	pdf.SetFillColor(248, 250, 252)
	pdf.RoundedRect(15, 45, 180, 30, 3, "1234", "F")

	pdf.SetY(50)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(60, 8, "SUMMARY", "", 0, "C", false, 0, "")
	pdf.CellFormat(60, 8, "", "", 0, "C", false, 0, "")
	pdf.CellFormat(60, 8, "", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	
	// Income
	pdf.SetTextColor(22, 163, 74) // Green
	pdf.CellFormat(60, 6, fmt.Sprintf("Income: Rp %.0f", totalIncome), "", 0, "C", false, 0, "")
	
	// Expense
	pdf.SetTextColor(220, 38, 38) // Red
	pdf.CellFormat(60, 6, fmt.Sprintf("Expense: Rp %.0f", totalExpense), "", 0, "C", false, 0, "")
	
	// Net
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(60, 6, fmt.Sprintf("Net: Rp %.0f", totalIncome-totalExpense), "", 1, "C", false, 0, "")

	// Table header
	pdf.SetY(85)
	pdf.SetFillColor(79, 70, 229)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 10)

	colWidths := []float64{25, 20, 35, 100}
	headers := []string{"Date", "Type", "Amount", "Description"}

	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 9)

	for i, tx := range transactions {
		// Alternate row colors
		if i%2 == 0 {
			pdf.SetFillColor(248, 250, 252)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		pdf.CellFormat(colWidths[0], 7, tx.TransactionDate.Format("02-Jan-06"), "1", 0, "C", true, 0, "")

		// Type with color
		typeStr := string(tx.Type)
		if tx.Type == models.TransactionTypeIncome {
			pdf.SetTextColor(22, 163, 74)
		} else {
			pdf.SetTextColor(220, 38, 38)
		}
		pdf.CellFormat(colWidths[1], 7, typeStr, "1", 0, "C", true, 0, "")
		pdf.SetTextColor(0, 0, 0)

		amount, _ := tx.Amount.Float64()
		pdf.CellFormat(colWidths[2], 7, fmt.Sprintf("Rp %.0f", amount), "1", 0, "R", true, 0, "")

		// Truncate description
		desc := tx.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		pdf.CellFormat(colWidths[3], 7, desc, "1", 0, "L", true, 0, "")

		pdf.Ln(-1)

		// Add new page if needed
		if pdf.GetY() > 270 {
			pdf.AddPage()
			pdf.SetY(20)
		}
	}

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 10, fmt.Sprintf("Wallet Twin - Total: %d transactions", len(transactions)), "", 0, "C", false, 0, "")

	return pdf.OutputFileAndClose(filename)
}

// WalletsToPDF exports wallets to a professional PDF file.
func (e *PDFExporter) WalletsToPDF(ctx context.Context, filename string) error {
	wallets, err := e.walletRepo.List(ctx, repository.WalletFilter{})
	if err != nil {
		return fmt.Errorf("failed to get wallets: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header
	pdf.SetFillColor(79, 70, 229)
	pdf.Rect(0, 0, 210, 35, "F")

	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetY(12)
	pdf.CellFormat(0, 10, "WALLET SUMMARY", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Generated: %s", time.Now().Format("02 January 2006, 15:04")), "", 1, "C", false, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetY(45)

	// Calculate total
	var totalBalance float64
	for _, w := range wallets {
		if w.IsActive {
			bal, _ := w.Balance.Float64()
			totalBalance += bal
		}
	}

	// Total balance box
	pdf.SetFillColor(16, 185, 129) // Green
	pdf.RoundedRect(15, 45, 180, 25, 3, "1234", "F")
	
	pdf.SetY(52)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(0, 10, fmt.Sprintf("Total Balance: Rp %.0f", totalBalance), "", 1, "C", false, 0, "")

	// Table
	pdf.SetY(80)
	pdf.SetFillColor(79, 70, 229)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 10)

	colWidths := []float64{50, 30, 50, 25, 25}
	headers := []string{"Name", "Type", "Balance", "Currency", "Status"}

	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 10)

	for i, w := range wallets {
		if i%2 == 0 {
			pdf.SetFillColor(248, 250, 252)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		name := w.Name
		if w.Icon != "" {
			name = w.Icon + " " + w.Name
		}
		if len(name) > 25 {
			name = name[:22] + "..."
		}

		pdf.CellFormat(colWidths[0], 8, name, "1", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[1], 8, string(w.Type), "1", 0, "C", true, 0, "")

		balance, _ := w.Balance.Float64()
		pdf.CellFormat(colWidths[2], 8, fmt.Sprintf("Rp %.0f", balance), "1", 0, "R", true, 0, "")
		pdf.CellFormat(colWidths[3], 8, w.Currency, "1", 0, "C", true, 0, "")

		status := "Active"
		if !w.IsActive {
			status = "Inactive"
		}
		pdf.CellFormat(colWidths[4], 8, status, "1", 0, "C", true, 0, "")

		pdf.Ln(-1)
	}

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 10, fmt.Sprintf("Wallet Twin - %d wallets", len(wallets)), "", 0, "C", false, 0, "")

	return pdf.OutputFileAndClose(filename)
}
