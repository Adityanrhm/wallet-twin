package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Adityanrhm/wallet-twin/internal/models"
	"github.com/Adityanrhm/wallet-twin/internal/repository"
)

// categoryRepository adalah implementasi PostgreSQL untuk CategoryRepository.
type categoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository membuat CategoryRepository baru.
func NewCategoryRepository(pool *pgxpool.Pool) repository.CategoryRepository {
	return &categoryRepository{pool: pool}
}

// Create menyimpan category baru.
func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (id, name, type, color, icon, parent_id, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Type,
		category.Color,
		category.Icon,
		category.ParentID,
		category.SortOrder,
	)

	return convertError(err)
}

// GetByID mengambil category berdasarkan ID.
func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	query := `
		SELECT id, name, type, color, icon, parent_id, sort_order, created_at
		FROM categories
		WHERE id = $1
	`

	cat := &models.Category{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&cat.ID,
		&cat.Name,
		&cat.Type,
		&cat.Color,
		&cat.Icon,
		&cat.ParentID,
		&cat.SortOrder,
		&cat.CreatedAt,
	)

	if err != nil {
		return nil, convertError(err)
	}

	return cat, nil
}

// GetByType mengambil kategori berdasarkan tipe.
// Hanya top-level categories (parent_id IS NULL).
func (r *categoryRepository) GetByType(ctx context.Context, catType models.CategoryType) ([]*models.Category, error) {
	query := `
		SELECT id, name, type, color, icon, parent_id, sort_order, created_at
		FROM categories
		WHERE type = $1 AND parent_id IS NULL
		ORDER BY sort_order, name
	`

	rows, err := r.pool.Query(ctx, query, catType)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		cat := &models.Category{}
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Type,
			&cat.Color,
			&cat.Icon,
			&cat.ParentID,
			&cat.SortOrder,
			&cat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

// GetChildren mengambil sub-kategori.
func (r *categoryRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*models.Category, error) {
	query := `
		SELECT id, name, type, color, icon, parent_id, sort_order, created_at
		FROM categories
		WHERE parent_id = $1
		ORDER BY sort_order, name
	`

	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		cat := &models.Category{}
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Type,
			&cat.Color,
			&cat.Icon,
			&cat.ParentID,
			&cat.SortOrder,
			&cat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

// List mengambil semua kategori.
func (r *categoryRepository) List(ctx context.Context) ([]*models.Category, error) {
	query := `
		SELECT id, name, type, color, icon, parent_id, sort_order, created_at
		FROM categories
		ORDER BY type, sort_order, name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, convertError(err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		cat := &models.Category{}
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Type,
			&cat.Color,
			&cat.Icon,
			&cat.ParentID,
			&cat.SortOrder,
			&cat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

// Update memperbarui category.
func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $2, type = $3, color = $4, icon = $5, parent_id = $6, sort_order = $7
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Type,
		category.Color,
		category.Icon,
		category.ParentID,
		category.SortOrder,
	)

	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete menghapus category.
func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return convertError(err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
