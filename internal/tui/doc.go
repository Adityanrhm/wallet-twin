// Package tui berisi Terminal User Interface menggunakan Bubble Tea.
//
// Bubble Tea adalah framework TUI dari Charm yang menggunakan
// The Elm Architecture (TEA) - sebuah pattern untuk building UIs.
//
// The Elm Architecture terdiri dari 3 bagian:
//
// 1. MODEL: State aplikasi
//
//	type model struct {
//	    wallets  []models.Wallet
//	    selected int
//	    loading  bool
//	}
//
// 2. UPDATE: Mengupdate state berdasarkan message
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case tea.KeyMsg:
//	        switch msg.String() {
//	        case "up", "k":
//	            m.selected--
//	        case "down", "j":
//	            m.selected++
//	        case "q":
//	            return m, tea.Quit
//	        }
//	    case walletsLoadedMsg:
//	        m.wallets = msg.wallets
//	        m.loading = false
//	    }
//	    return m, nil
//	}
//
// 3. VIEW: Render state ke string
//
//	func (m model) View() string {
//	    if m.loading {
//	        return "Loading..."
//	    }
//
//	    var s strings.Builder
//	    for i, w := range m.wallets {
//	        cursor := " "
//	        if i == m.selected {
//	            cursor = ">"
//	        }
//	        s.WriteString(fmt.Sprintf("%s %s\n", cursor, w.Name))
//	    }
//	    return s.String()
//	}
//
// Keuntungan TEA:
// - Predictable state management
// - Easy to test (pure functions)
// - Time-travel debugging
// - Immutable updates
package tui

// TODO: Add TUI implementation
