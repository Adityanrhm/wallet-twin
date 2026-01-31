# 💰 Wallet Twin

A modern CLI personal finance application built with Go.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

## ✨ Features

- 💼 **Multi-Wallet Support** - Track cash, bank accounts, and e-wallets
- 📝 **Transaction Tracking** - Record income and expenses with categories
- 🔄 **Inter-Wallet Transfers** - Transfer money between accounts with fees
- 📊 **Budget Management** - Set spending limits and track progress
- 🎯 **Savings Goals** - Track progress toward financial goals
- 🖥️ **Interactive TUI** - Beautiful terminal dashboard with Bubble Tea
- 📤 **Export/Import** - Backup and restore data in CSV/JSON format

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- PostgreSQL 14 or later

### Installation

```bash
# Clone the repository
git clone https://github.com/Adityanrhm/wallet-twin.git
cd wallet-twin

# Install dependencies
go mod download

# Setup database
createdb wallet_twin
go run cmd/migrate/main.go up

# Copy and configure
cp config.yaml.example config.yaml
# Edit config.yaml with your database credentials

# Build
go build -o wallet ./cmd/wallet
```

### Usage

```bash
# Show help
./wallet --help

# Launch interactive dashboard
./wallet dashboard

# Wallet commands
./wallet wallet add -n "BCA Savings" -t bank -c IDR -b 1000000
./wallet wallet list
./wallet wallet balance

# Transaction commands
./wallet tx add -w <wallet-id> -t expense -a 50000 -d "Lunch"
./wallet tx list
./wallet tx summary

# Transfer between wallets
./wallet transfer -f <from-id> -t <to-id> -a 500000

# Budget commands
./wallet budget add -c <category-id> -a 2000000 -p monthly
./wallet budget list

# Goal commands
./wallet goal add -n "Emergency Fund" -t 10000000
./wallet goal contribute -g <goal-id> -a 500000
./wallet goal list

# Export/Import
./wallet export all -o backup.json
./wallet import backup backup.json
```

## 📁 Project Structure

```
wallet-twin/
├── cmd/
│   ├── wallet/          # Main CLI application
│   └── migrate/         # Database migration tool
├── internal/
│   ├── app/             # Application bootstrap & DI
│   ├── cli/             # CLI commands (Cobra)
│   ├── config/          # Configuration management
│   ├── database/        # Database connection
│   ├── export/          # Export/Import functionality
│   ├── models/          # Domain models
│   ├── repository/      # Data access layer
│   │   └── postgres/    # PostgreSQL implementation
│   ├── service/         # Business logic layer
│   └── tui/             # Terminal UI (Bubble Tea)
├── migrations/          # SQL migrations
├── config.yaml          # Configuration file
└── go.mod
```

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| CLI Framework | [Cobra](https://github.com/spf13/cobra) |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Database | PostgreSQL with [pgx](https://github.com/jackc/pgx) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Config | [Viper](https://github.com/spf13/viper) |

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/models/...
go test ./internal/service/...
```

## 📝 Configuration

Create `config.yaml` in the project root:

```yaml
app:
  name: "Wallet Twin"
  currency: "IDR"
  debug: false

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your-password"
  name: "wallet_twin"
  sslmode: "disable"
```

Or use environment variables:

```bash
export WT_DATABASE_HOST=localhost
export WT_DATABASE_USER=postgres
export WT_DATABASE_PASSWORD=secret
```

## 🎨 TUI Dashboard

Launch the interactive dashboard:

```bash
./wallet dashboard
```

**Keyboard Shortcuts:**
- `← →` - Navigate between tabs
- `1-5` - Jump to tab
- `r` - Refresh data
- `q` - Quit

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Aditya** - [@Adityanrhm](https://github.com/Adityanrhm)

---

Made with ❤️ and Go
