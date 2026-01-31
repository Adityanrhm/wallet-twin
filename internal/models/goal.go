// Package models - Goal entity
//
// Goals membantu user menabung untuk tujuan tertentu.
// User set target dan deadline, aplikasi track progress.
//
// Contoh:
// - Emergency Fund: Rp 10.000.000 (deadline: 6 bulan)
// - Holiday Trip: Rp 5.000.000 (deadline: Desember)
// - New Laptop: Rp 15.000.000 (deadline: 1 tahun)
//
// User menambah kontribusi ke goal untuk menambah current_amount.
package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GoalStatus adalah status goal.
type GoalStatus string

const (
	// GoalStatusActive untuk goal yang masih berjalan
	GoalStatusActive GoalStatus = "active"

	// GoalStatusCompleted untuk goal yang sudah tercapai
	GoalStatusCompleted GoalStatus = "completed"

	// GoalStatusCancelled untuk goal yang dibatalkan
	GoalStatusCancelled GoalStatus = "cancelled"
)

// IsValid mengecek apakah goal status valid.
func (s GoalStatus) IsValid() bool {
	switch s {
	case GoalStatusActive, GoalStatusCompleted, GoalStatusCancelled:
		return true
	}
	return false
}

// String returns string representation.
func (s GoalStatus) String() string {
	return string(s)
}

// Goal merepresentasikan target tabungan.
//
// Goal adalah entity untuk tracking progress menuju financial goal.
// User dapat menambah kontribusi untuk meningkatkan current_amount.
//
// Contoh:
//
//	goal := &models.Goal{
//	    BaseModel:     models.BaseModel{ID: models.NewID()},
//	    Name:          "Emergency Fund",
//	    TargetAmount:  decimal.NewFromInt(10000000),
//	    CurrentAmount: decimal.Zero,
//	    Status:        models.GoalStatusActive,
//	}
//
//	// Tambah kontribusi
//	goal.AddContribution(decimal.NewFromInt(500000))
//	// goal.CurrentAmount sekarang 500000
//
//	// Cek progress
//	progress := goal.GetProgress() // 5%
type Goal struct {
	// Embed BaseModel untuk ID dan timestamps
	BaseModel

	// Name adalah nama goal.
	// Contoh: "Emergency Fund", "Holiday Trip"
	Name string `json:"name" db:"name"`

	// Description adalah deskripsi goal (opsional).
	Description string `json:"description,omitempty" db:"description"`

	// TargetAmount adalah jumlah target yang ingin dicapai.
	TargetAmount decimal.Decimal `json:"target_amount" db:"target_amount"`

	// CurrentAmount adalah jumlah yang sudah terkumpul.
	// Di-update setiap ada kontribusi.
	CurrentAmount decimal.Decimal `json:"current_amount" db:"current_amount"`

	// Deadline adalah target tanggal pencapaian (opsional).
	// nil = tidak ada deadline.
	Deadline *time.Time `json:"deadline,omitempty" db:"deadline"`

	// Status adalah status goal.
	Status GoalStatus `json:"status" db:"status"`

	// Color untuk UI.
	Color string `json:"color,omitempty" db:"color"`

	// Icon.
	Icon string `json:"icon,omitempty" db:"icon"`
}

// GoalContribution merepresentasikan kontribusi ke goal.
//
// Setiap kali user menabung untuk goal, buat GoalContribution.
// Ini untuk tracking history kontribusi.
//
//	contribution := &models.GoalContribution{
//	    ID:        models.NewID(),
//	    GoalID:    goal.ID,
//	    Amount:    decimal.NewFromInt(500000),
//	    Note:      "Bonus dari freelance",
//	}
type GoalContribution struct {
	// ID adalah unique identifier.
	ID uuid.UUID `json:"id" db:"id"`

	// GoalID adalah goal yang dikontribusi.
	GoalID uuid.UUID `json:"goal_id" db:"goal_id"`

	// Amount adalah jumlah kontribusi.
	Amount decimal.Decimal `json:"amount" db:"amount"`

	// Note adalah catatan (opsional).
	Note string `json:"note,omitempty" db:"note"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Validation errors
var (
	ErrGoalNameRequired     = errors.New("goal name is required")
	ErrGoalNameTooLong      = errors.New("goal name must be less than 100 characters")
	ErrGoalInvalidTarget    = errors.New("target amount must be positive")
	ErrGoalInvalidStatus    = errors.New("invalid goal status")
	ErrContributionInvalid  = errors.New("contribution amount must be positive")
	ErrContributionNoGoal   = errors.New("goal is required for contribution")
)

// Validate memvalidasi goal.
func (g *Goal) Validate() error {
	g.Name = strings.TrimSpace(g.Name)
	if g.Name == "" {
		return ErrGoalNameRequired
	}
	if len(g.Name) > 100 {
		return ErrGoalNameTooLong
	}
	if g.TargetAmount.IsNegative() || g.TargetAmount.IsZero() {
		return ErrGoalInvalidTarget
	}
	if !g.Status.IsValid() {
		return ErrGoalInvalidStatus
	}
	return nil
}

// Validate memvalidasi contribution.
func (c *GoalContribution) Validate() error {
	if c.GoalID == uuid.Nil {
		return ErrContributionNoGoal
	}
	if c.Amount.IsNegative() || c.Amount.IsZero() {
		return ErrContributionInvalid
	}
	return nil
}

// NewGoal membuat goal baru.
//
//	goal := models.NewGoal("Emergency Fund", decimal.NewFromInt(10000000))
func NewGoal(name string, target decimal.Decimal) *Goal {
	return &Goal{
		BaseModel:     BaseModel{ID: NewID()},
		Name:          name,
		TargetAmount:  target,
		CurrentAmount: decimal.Zero,
		Status:        GoalStatusActive,
	}
}

// NewContribution membuat kontribusi baru.
//
//	contribution := models.NewContribution(goal.ID, decimal.NewFromInt(500000))
//	contribution.Note = "Bonus freelance"
func NewContribution(goalID uuid.UUID, amount decimal.Decimal) *GoalContribution {
	return &GoalContribution{
		ID:        NewID(),
		GoalID:    goalID,
		Amount:    amount,
		CreatedAt: time.Now(),
	}
}

// GetProgress menghitung persentase progress goal (0-100).
//
//	progress := goal.GetProgress() // 75.5
func (g *Goal) GetProgress() float64 {
	if g.TargetAmount.IsZero() {
		return 0
	}
	progress, _ := g.CurrentAmount.Div(g.TargetAmount).Mul(decimal.NewFromInt(100)).Float64()
	return progress
}

// GetRemaining menghitung sisa yang perlu dikumpulkan.
//
//	remaining := goal.GetRemaining()
func (g *Goal) GetRemaining() decimal.Decimal {
	remaining := g.TargetAmount.Sub(g.CurrentAmount)
	if remaining.IsNegative() {
		return decimal.Zero
	}
	return remaining
}

// IsCompleted mengecek apakah goal sudah tercapai.
//
//	if goal.IsCompleted() {
//	    goal.Status = models.GoalStatusCompleted
//	}
func (g *Goal) IsCompleted() bool {
	return g.CurrentAmount.GreaterThanOrEqual(g.TargetAmount)
}

// AddContribution menambah current amount.
// Untuk actual contribution, buat GoalContribution record juga.
//
//	goal.AddContribution(decimal.NewFromInt(500000))
//	if goal.IsCompleted() {
//	    goal.Status = models.GoalStatusCompleted
//	}
func (g *Goal) AddContribution(amount decimal.Decimal) {
	g.CurrentAmount = g.CurrentAmount.Add(amount)
}

// DaysUntilDeadline menghitung hari tersisa sampai deadline.
// Return -1 jika tidak ada deadline atau sudah lewat.
//
//	days := goal.DaysUntilDeadline()
//	if days > 0 {
//	    fmt.Printf("%d days remaining\n", days)
//	}
func (g *Goal) DaysUntilDeadline() int {
	if g.Deadline == nil {
		return -1
	}
	duration := time.Until(*g.Deadline)
	if duration < 0 {
		return -1
	}
	return int(duration.Hours() / 24)
}
