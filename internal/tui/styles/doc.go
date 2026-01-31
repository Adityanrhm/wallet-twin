// Package styles berisi Lipgloss styles untuk TUI.
//
// Lipgloss adalah library styling untuk terminal dari Charm.
// Mirip seperti CSS tapi untuk terminal.
//
// Definisi style:
//
//	var (
//	    // Colors
//	    primaryColor   = lipgloss.Color("#7C3AED")  // Purple
//	    secondaryColor = lipgloss.Color("#10B981")  // Green
//	    dangerColor    = lipgloss.Color("#EF4444")  // Red
//	    mutedColor     = lipgloss.Color("#6B7280")  // Gray
//
//	    // Base styles
//	    BaseStyle = lipgloss.NewStyle().
//	        Padding(1, 2)
//
//	    // Title style
//	    TitleStyle = lipgloss.NewStyle().
//	        Foreground(primaryColor).
//	        Bold(true).
//	        Padding(0, 1)
//
//	    // Box styles
//	    BoxStyle = lipgloss.NewStyle().
//	        Border(lipgloss.RoundedBorder()).
//	        BorderForeground(primaryColor).
//	        Padding(1, 2)
//
//	    // Income style (green)
//	    IncomeStyle = lipgloss.NewStyle().
//	        Foreground(secondaryColor)
//
//	    // Expense style (red)
//	    ExpenseStyle = lipgloss.NewStyle().
//	        Foreground(dangerColor)
//	)
//
// Menggunakan styles:
//
//	fmt.Println(TitleStyle.Render("Wallet Twin"))
//	fmt.Println(IncomeStyle.Render("+Rp 500,000"))
//	fmt.Println(ExpenseStyle.Render("-Rp 50,000"))
package styles

// TODO: Add style definitions
