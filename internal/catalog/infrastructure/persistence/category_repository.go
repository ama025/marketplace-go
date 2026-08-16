package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/spec"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Categories(ctx context.Context) ([]entities.Category, error) {

	rows, err := r.db.QueryContext(ctx, `
	SELECT
		id, title
	FROM categories
	ORDER BY title
	`)

	if err != nil {
		return nil, fmt.Errorf("categories query: %w", err)
	}

	defer rows.Close()

	var categories []entities.Category

	for rows.Next() {
		var c entities.Category

		if err := rows.Scan(&c.Id, &c.Title); err != nil {
			return nil, fmt.Errorf("scan category error: %w", err)
		}

		categories = append(categories, c)
	}

	return categories, rows.Err()
}

func (r *CategoryRepository) CategoriesPaged(ctx context.Context, args spec.QueryArgs) ([]entities.Category, int, error) {

	var totalCount int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories`).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("categories count query: %w", err)
	}

	offset := (args.PageIndex - 1) * args.PageSize

	rows, err := r.db.QueryContext(ctx, `
	SELECT
		id, title
	FROM categories
	ORDER BY title
	LIMIT $1 OFFSET $2
	`, args.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("categories paged query: %w", err)
	}
	defer rows.Close()

	var categories []entities.Category

	for rows.Next() {
		var c entities.Category
		if err := rows.Scan(&c.Id, &c.Title); err != nil {
			return nil, 0, fmt.Errorf("scan category error: %w", err)
		}
		categories = append(categories, c)
	}

	return categories, totalCount, rows.Err()
}

