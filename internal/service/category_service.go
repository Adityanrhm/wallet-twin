package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// CategoryService menangani business logic untuk category operations.
type CategoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService membuat CategoryService baru.
func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// Create membuat category baru.
func (s *CategoryService) Create(ctx context.Context, input CreateCategoryInput) (*models.Category, error) {
	category := &models.Category{
		ID:        models.NewID(),
		Name:      input.Name,
		Type:      input.Type,
		Color:     input.Color,
		Icon:      input.Icon,
		ParentID:  input.ParentID,
		SortOrder: input.SortOrder,
	}

	if err := category.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate parent exists if specified
	if input.ParentID != nil {
		parent, err := s.repo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
		// Sub-category must have same type as parent
		if parent.Type != input.Type {
			return nil, fmt.Errorf("sub-category type must match parent type")
		}
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// GetByID mengambil category berdasarkan ID.
func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return category, nil
}

// GetByType mengambil kategori berdasarkan tipe.
func (s *CategoryService) GetByType(ctx context.Context, catType models.CategoryType) ([]*models.Category, error) {
	categories, err := s.repo.GetByType(ctx, catType)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	return categories, nil
}

// GetIncomeCategories mengambil semua kategori income.
func (s *CategoryService) GetIncomeCategories(ctx context.Context) ([]*models.Category, error) {
	return s.GetByType(ctx, models.CategoryTypeIncome)
}

// GetExpenseCategories mengambil semua kategori expense.
func (s *CategoryService) GetExpenseCategories(ctx context.Context) ([]*models.Category, error) {
	return s.GetByType(ctx, models.CategoryTypeExpense)
}

// GetWithChildren mengambil category dengan sub-categories.
func (s *CategoryService) GetWithChildren(ctx context.Context, id uuid.UUID) (*CategoryWithChildren, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	return &CategoryWithChildren{
		Category: category,
		Children: children,
	}, nil
}

// List mengambil semua kategori.
func (s *CategoryService) List(ctx context.Context) ([]*models.Category, error) {
	return s.repo.List(ctx)
}

// Update memperbarui category.
func (s *CategoryService) Update(ctx context.Context, input UpdateCategoryInput) (*models.Category, error) {
	category, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if input.Name != nil {
		category.Name = *input.Name
	}
	if input.Color != nil {
		category.Color = *input.Color
	}
	if input.Icon != nil {
		category.Icon = *input.Icon
	}
	if input.SortOrder != nil {
		category.SortOrder = *input.SortOrder
	}

	if err := category.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// Delete menghapus category.
func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}

// CreateCategoryInput adalah input untuk membuat category.
type CreateCategoryInput struct {
	Name      string
	Type      models.CategoryType
	Color     string
	Icon      string
	ParentID  *uuid.UUID
	SortOrder int
}

// UpdateCategoryInput adalah input untuk update category.
type UpdateCategoryInput struct {
	ID        uuid.UUID
	Name      *string
	Color     *string
	Icon      *string
	SortOrder *int
}

// CategoryWithChildren adalah category dengan sub-categories.
type CategoryWithChildren struct {
	Category *models.Category
	Children []*models.Category
}
