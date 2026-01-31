// Package components berisi reusable UI components untuk TUI.
//
// Components adalah building blocks untuk views.
// Setiap component adalah Bubble Tea model yang bisa di-compose.
//
// Components yang tersedia:
// - Table: Tabel data dengan scrolling
// - Menu: Navigation menu
// - Progress: Progress bar untuk budgets dan goals
// - Chart: ASCII charts untuk visualisasi
//
// Composing components:
//
//	type dashboardModel struct {
//	    menu     menu.Model
//	    table    table.Model
//	    progress progress.Model
//	}
//
//	func (m dashboardModel) View() string {
//	    return lipgloss.JoinHorizontal(
//	        lipgloss.Left,
//	        m.menu.View(),
//	        lipgloss.JoinVertical(
//	            lipgloss.Top,
//	            m.table.View(),
//	            m.progress.View(),
//	        ),
//	    )
//	}
package components

// TODO: Add component implementations
