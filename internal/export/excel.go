package export

import (
	"context"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// ExcelExporter creates professional Excel reports.
type ExcelExporter struct {
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	categoryRepo    repository.CategoryRepository
}

// NewExcelExporter creates a new ExcelExporter.
func NewExcelExporter(
	walletRepo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	categoryRepo repository.CategoryRepository,
) *ExcelExporter {
	return &ExcelExporter{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		categoryRepo:    categoryRepo,
	}
}

// Excel styles
var (
	headerStyle = &excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4F46E5"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	}

	titleStyle = &excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  16,
			Color: "1E293B",
		},
	}

	incomeStyle = &excelize.Style{
		Font: &excelize.Font{Color: "16A34A"},
		NumFmt: 4,
	}

	expenseStyle = &excelize.Style{
		Font: &excelize.Font{Color: "DC2626"},
		NumFmt: 4,
	}

	moneyStyle = &excelize.Style{
		NumFmt: 4,
		Alignment: &excelize.Alignment{Horizontal: "right"},
	}

	alternateRowStyle = &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"F8FAFC"},
			Pattern: 1,
		},
	}
)

// TransactionsToExcel exports transactions to a professional Excel file.
func (e *ExcelExporter) TransactionsToExcel(ctx context.Context, filename string, filter repository.TransactionFilter) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Transactions"
	f.SetSheetName("Sheet1", sheetName)

	// Get data
	params := repository.ListParams{Limit: 10000, Offset: 0}
	transactions, err := e.transactionRepo.List(ctx, filter, params)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	// Create styles
	headerStyleID, _ := f.NewStyle(headerStyle)
	titleStyleID, _ := f.NewStyle(titleStyle)
	incomeStyleID, _ := f.NewStyle(incomeStyle)
	expenseStyleID, _ := f.NewStyle(expenseStyle)
	moneyStyleID, _ := f.NewStyle(moneyStyle)

	// Title
	f.SetCellValue(sheetName, "A1", "📊 Transaction Report")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyleID)
	f.MergeCell(sheetName, "A1", "F1")

	// Subtitle
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Generated: %s", time.Now().Format("02 January 2006, 15:04")))

	// Headers
	headers := []string{"Date", "Type", "Amount", "Description", "Wallet ID", "Category"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c4", 'A'+i)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyleID)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 12)
	f.SetColWidth(sheetName, "C", "C", 18)
	f.SetColWidth(sheetName, "D", "D", 40)
	f.SetColWidth(sheetName, "E", "E", 38)
	f.SetColWidth(sheetName, "F", "F", 20)

	// Data rows
	var totalIncome, totalExpense float64
	for i, tx := range transactions {
		row := i + 5
		
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), tx.TransactionDate.Format("02-Jan-2006"))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), string(tx.Type))
		
		amount, _ := tx.Amount.Float64()
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), amount)
		
		if tx.Type == models.TransactionTypeIncome {
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), incomeStyleID)
			totalIncome += amount
		} else {
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), expenseStyleID)
			totalExpense += amount
		}

		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), tx.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), tx.WalletID.String())
		
		categoryName := "-"
		if tx.CategoryID != nil {
			categoryName = tx.CategoryID.String()[:8] + "..."
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), categoryName)
	}

	// Summary section
	summaryRow := len(transactions) + 7
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "📈 SUMMARY")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), titleStyleID)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow+1), "Total Income:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow+1), totalIncome)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", summaryRow+1), fmt.Sprintf("B%d", summaryRow+1), incomeStyleID)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow+2), "Total Expense:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow+2), totalExpense)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", summaryRow+2), fmt.Sprintf("B%d", summaryRow+2), expenseStyleID)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow+3), "Net:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow+3), totalIncome-totalExpense)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", summaryRow+3), fmt.Sprintf("B%d", summaryRow+3), moneyStyleID)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow+4), "Total Transactions:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow+4), len(transactions))

	return f.SaveAs(filename)
}

// WalletsToExcel exports wallets to a professional Excel file.
func (e *ExcelExporter) WalletsToExcel(ctx context.Context, filename string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Wallets"
	f.SetSheetName("Sheet1", sheetName)

	wallets, err := e.walletRepo.List(ctx, repository.WalletFilter{})
	if err != nil {
		return fmt.Errorf("failed to get wallets: %w", err)
	}

	// Create styles
	headerStyleID, _ := f.NewStyle(headerStyle)
	titleStyleID, _ := f.NewStyle(titleStyle)
	moneyStyleID, _ := f.NewStyle(moneyStyle)

	// Title
	f.SetCellValue(sheetName, "A1", "💼 Wallet Summary")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyleID)
	f.MergeCell(sheetName, "A1", "E1")

	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Generated: %s", time.Now().Format("02 January 2006, 15:04")))

	// Headers
	headers := []string{"Name", "Type", "Balance", "Currency", "Status"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c4", 'A'+i)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyleID)
	}

	// Column widths
	f.SetColWidth(sheetName, "A", "A", 25)
	f.SetColWidth(sheetName, "B", "B", 12)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 10)
	f.SetColWidth(sheetName, "E", "E", 12)

	// Data
	var totalBalance float64
	for i, w := range wallets {
		row := i + 5
		
		name := w.Name
		if w.Icon != "" {
			name = w.Icon + " " + w.Name
		}
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), name)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), string(w.Type))
		
		balance, _ := w.Balance.Float64()
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), balance)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), moneyStyleID)
		
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), w.Currency)
		
		status := "Active"
		if !w.IsActive {
			status = "Inactive"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), status)

		if w.IsActive {
			totalBalance += balance
		}
	}

	// Total
	totalRow := len(wallets) + 6
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", totalRow), "TOTAL BALANCE:")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("A%d", totalRow), titleStyleID)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", totalRow), totalBalance)
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", totalRow), fmt.Sprintf("C%d", totalRow), moneyStyleID)

	return f.SaveAs(filename)
}
