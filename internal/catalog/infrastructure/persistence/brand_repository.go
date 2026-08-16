package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/spec"
)

type BrandRepository struct {
	db *sql.DB
}

func NewBrandRepository(db *sql.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

func (r *BrandRepository) Brands(ctx context.Context) ([]entities.Brand, error) {

	rows, err := r.db.QueryContext(ctx, `
	SELECT
		id, title
	FROM brands
	ORDER BY title
	`)

	if err != nil {
		return nil, fmt.Errorf("brands query: %w", err)
	}

	defer rows.Close()

	var brands []entities.Brand

	for rows.Next() {
		var b entities.Brand

		if err := rows.Scan(&b.Id, &b.Title); err != nil {
			return nil, fmt.Errorf("scan brand error: %w", err)
		}

		brands = append(brands, b)
	}

	return brands, rows.Err()
}

func (r *BrandRepository) BrandsPaged(ctx context.Context, args spec.QueryArgs) ([]entities.Brand, int, error) {

	var totalCount int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM brands`).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("brands count query: %w", err)
	}

	offset := (args.PageIndex - 1) * args.PageSize

	rows, err := r.db.QueryContext(ctx, `
	SELECT
		id, title
	FROM brands
	ORDER BY title
	LIMIT $1 OFFSET $2
	`, args.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("brands paged query: %w", err)
	}
	defer rows.Close()

	var brands []entities.Brand

	for rows.Next() {
		var b entities.Brand
		if err := rows.Scan(&b.Id, &b.Title); err != nil {
			return nil, 0, fmt.Errorf("scan brand error: %w", err)
		}
		brands = append(brands, b)
	}

	return brands, totalCount, rows.Err()
}
