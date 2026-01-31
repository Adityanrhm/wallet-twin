package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// GoalService menangani business logic untuk savings goals.
//
// Goal adalah target tabungan yang ingin dicapai user.
// Service ini menyediakan:
// - CRUD goals
// - Add contributions
// - Track progress
type GoalService struct {
	goalRepo repository.GoalRepository
}

// NewGoalService membuat GoalService baru.
func NewGoalService(goalRepo repository.GoalRepository) *GoalService {
	return &GoalService{goalRepo: goalRepo}
}

// Create membuat goal baru.
func (s *GoalService) Create(ctx context.Context, input CreateGoalInput) (*models.Goal, error) {
	goal := models.NewGoal(input.Name, input.TargetAmount)
	goal.Description = input.Description
	goal.Deadline = input.Deadline
	goal.Color = input.Color
	goal.Icon = input.Icon

	if err := goal.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.goalRepo.Create(ctx, goal); err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	return goal, nil
}

// GetByID mengambil goal berdasarkan ID.
func (s *GoalService) GetByID(ctx context.Context, id uuid.UUID) (*models.Goal, error) {
	goal, err := s.goalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}
	return goal, nil
}

// List mengambil semua goals.
func (s *GoalService) List(ctx context.Context, filter repository.GoalFilter) ([]*models.Goal, error) {
	goals, err := s.goalRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list goals: %w", err)
	}
	return goals, nil
}

// ListActive mengambil goal aktif.
func (s *GoalService) ListActive(ctx context.Context) ([]*models.Goal, error) {
	status := models.GoalStatusActive
	return s.List(ctx, repository.GoalFilter{Status: &status})
}

// AddContribution menambahkan kontribusi ke goal.
//
// Contoh:
//
//	err := goalService.AddContribution(ctx, goalID, service.AddContributionInput{
//	    Amount: decimal.NewFromInt(500000),
//	    Note:   "Bonus freelance",
//	})
func (s *GoalService) AddContribution(ctx context.Context, goalID uuid.UUID, input AddContributionInput) error {
	contribution := models.NewContribution(goalID, input.Amount)
	contribution.Note = input.Note

	if err := contribution.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// AddContribution in repo also updates goal.current_amount
	if err := s.goalRepo.AddContribution(ctx, contribution); err != nil {
		return fmt.Errorf("failed to add contribution: %w", err)
	}

	// Check if goal is now completed
	goal, err := s.goalRepo.GetByID(ctx, goalID)
	if err != nil {
		return nil // Contribution added, but couldn't check completion
	}

	if goal.IsCompleted() && goal.Status == models.GoalStatusActive {
		goal.Status = models.GoalStatusCompleted
		_ = s.goalRepo.Update(ctx, goal)
	}

	return nil
}

// GetContributions mengambil history kontribusi.
func (s *GoalService) GetContributions(
	ctx context.Context,
	goalID uuid.UUID,
	params repository.ListParams,
) ([]*models.GoalContribution, error) {
	contributions, err := s.goalRepo.GetContributions(ctx, goalID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get contributions: %w", err)
	}
	return contributions, nil
}

// GetProgress menghitung progress goal.
func (s *GoalService) GetProgress(ctx context.Context, id uuid.UUID) (*GoalProgress, error) {
	goal, err := s.goalRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	return &GoalProgress{
		Goal:              goal,
		Progress:          goal.GetProgress(),
		Remaining:         goal.GetRemaining(),
		IsCompleted:       goal.IsCompleted(),
		DaysUntilDeadline: goal.DaysUntilDeadline(),
	}, nil
}

// Update memperbarui goal.
func (s *GoalService) Update(ctx context.Context, input UpdateGoalInput) (*models.Goal, error) {
	goal, err := s.goalRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	if input.Name != nil {
		goal.Name = *input.Name
	}
	if input.Description != nil {
		goal.Description = *input.Description
	}
	if input.TargetAmount != nil {
		goal.TargetAmount = *input.TargetAmount
	}
	if input.Deadline != nil {
		goal.Deadline = input.Deadline
	}
	if input.Status != nil {
		goal.Status = *input.Status
	}
	if input.Color != nil {
		goal.Color = *input.Color
	}
	if input.Icon != nil {
		goal.Icon = *input.Icon
	}

	if err := goal.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.goalRepo.Update(ctx, goal); err != nil {
		return nil, fmt.Errorf("failed to update goal: %w", err)
	}

	return goal, nil
}

// Delete menghapus goal.
func (s *GoalService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.goalRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete goal: %w", err)
	}
	return nil
}

// MarkCompleted menandai goal sebagai completed.
func (s *GoalService) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	status := models.GoalStatusCompleted
	_, err := s.Update(ctx, UpdateGoalInput{
		ID:     id,
		Status: &status,
	})
	return err
}

// Cancel membatalkan goal.
func (s *GoalService) Cancel(ctx context.Context, id uuid.UUID) error {
	status := models.GoalStatusCancelled
	_, err := s.Update(ctx, UpdateGoalInput{
		ID:     id,
		Status: &status,
	})
	return err
}

// CreateGoalInput adalah input untuk membuat goal.
type CreateGoalInput struct {
	Name         string
	Description  string
	TargetAmount decimal.Decimal
	Deadline     *time.Time
	Color        string
	Icon         string
}

// UpdateGoalInput adalah input untuk update goal.
type UpdateGoalInput struct {
	ID           uuid.UUID
	Name         *string
	Description  *string
	TargetAmount *decimal.Decimal
	Deadline     *time.Time
	Status       *models.GoalStatus
	Color        *string
	Icon         *string
}

// AddContributionInput adalah input untuk menambah kontribusi.
type AddContributionInput struct {
	Amount decimal.Decimal
	Note   string
}

// GoalProgress adalah ringkasan progress goal.
type GoalProgress struct {
	Goal              *models.Goal
	Progress          float64         // Percentage (0-100)
	Remaining         decimal.Decimal // Amount remaining
	IsCompleted       bool
	DaysUntilDeadline int // -1 if no deadline or past
}
