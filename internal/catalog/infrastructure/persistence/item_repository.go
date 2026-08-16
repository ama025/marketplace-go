package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/spec"

	"github.com/google/uuid"
)

const sqlCatalogItemsQuery = `
SELECT ci.id, ci.short_description, ci.full_description, ci.price,
	b.id as brand_id,b.title as brand_title,
	c.id as category_id,c.title as category_title
FROM catalog_items ci
LEFT JOIN brands b ON b.id = ci.brand_id
LEFT JOIN categories c ON c.id = ci.category_id
`

type itemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *itemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) Items(ctx context.Context) ([]entities.CatalogItem, error) {

	query := sqlCatalogItemsQuery

	rows, err := r.db.QueryContext(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("catalog items query: %w", err)
	}

	defer rows.Close()

	var items []entities.CatalogItem

	for rows.Next() {

		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan catalog item error: %w", err)
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func scanCatalogItem(rows *sql.Rows) (entities.CatalogItem, error) {
	var item entities.CatalogItem

	var brandID *uuid.UUID
	var brandTitle *string
	var categoryID *uuid.UUID
	var categoryTitle *string

	if err := rows.Scan(&item.Id, &item.ShortDescription, &item.FullDescription, &item.Price, &brandID, &brandTitle, &categoryID, &categoryTitle); err != nil {
		return entities.CatalogItem{}, fmt.Errorf("scan catalog item error: %w", err)
	}

	if brandID != nil {
		item.Brand = &entities.Brand{
			BaseEntity: entities.BaseEntity{
				Id:    *brandID,
				Title: *brandTitle,
			},
		}
	}

	if categoryID != nil {
		item.Category = &entities.Category{
			BaseEntity: entities.BaseEntity{
				Id:    *categoryID,
				Title: *categoryTitle,
			},
		}
	}

	return item, nil
}

func (r *itemRepository) Item(ctx context.Context, id uuid.UUID) (*entities.CatalogItem, error) {
	query := sqlCatalogItemsQuery + "WHERE ci.id = $1"

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	item, err := scanCatalogItem(rows)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &item, nil

}

func (r *itemRepository) ItemsByTitle(ctx context.Context, title string) ([]entities.CatalogItem, error) {
	query := sqlCatalogItemsQuery + `WHERE ci.title LIKE '%' || $1 || '%'`

	rows, err := r.db.QueryContext(ctx, query, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entities.CatalogItem

	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *itemRepository) Create(ctx context.Context, item *entities.CatalogItem) (*entities.CatalogItem, error) {
	if item.Id == uuid.Nil {
		item.Id = uuid.New()
	}

	var brandID, categoryID *uuid.UUID

	if item.Brand != nil {
		brandID = &item.Brand.Id

	}

	if item.Category != nil {
		categoryID = &item.Category.Id
	}

	sqlQuery := `
			INSERT INTO
		catalog_items (
			id,
			title,
			short_description,
			full_description,
			image_url,
			brand_id,
			category_id,
			price
		)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8);
	`

	_, err := r.db.ExecContext(ctx, sqlQuery, item.Id, item.Title, item.ShortDescription, item.FullDescription, item.ImageURL, brandID, categoryID, item.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to insert item: %w", err)
	}

	return item, nil
}

func (r *itemRepository) Update(ctx context.Context,item *entities.CatalogItem) (bool, error) {
	var brandID,categoryID *uuid.UUID

	if item.Brand != nil {
		brandID = &item.Brand.Id
	}

	if item.Category != nil {
		categoryID = &item.Category.Id
	}

	sqlQuery := `
		UPDATE
		catalog_items
		SET
			title = $2,
			short_description = $3,
			full_description = $4,
			image_url = $5,
			brand_id = $6,
			category_id = $7,
			price = $8
		WHERE
			id = $1;
	`

	result,err := r.db.ExecContext(ctx,sqlQuery,item.Id,item.Title,item.ShortDescription,item.FullDescription,item.ImageURL,brandID,categoryID,item.Price)
	if err != nil {
		return false, fmt.Errorf("failed to update item: %w", err)
	}

	n,_ := result.RowsAffected()
	return n>0, nil

}

func (r *itemRepository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {

	sqlDeleteItem := "DELETE FROM catalog_items WHERE id = $1"

	result, err := r.db.ExecContext(
		ctx,
		sqlDeleteItem,
		id,
	)

	if err != nil {

		return false, fmt.Errorf("delete item[%v]: %w", id, err)
	}

	n, _ := result.RowsAffected()

	return n > 0, nil
}

func (r *itemRepository) ItemsByBrand(ctx context.Context, brandTitle string) ([]entities.CatalogItem, error) {

	query := sqlCatalogItemsQuery + `WHERE b.title LIKE '%' || $1 || '%'`

	rows, err := r.db.QueryContext(ctx, query, brandTitle)
	if err != nil {

		return nil, fmt.Errorf("items by brand query: %w", err)
	}

	defer rows.Close()

	var items []entities.CatalogItem

	for rows.Next() {

		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan catalog item by brand error: %w", err)
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *itemRepository) ItemsWithFilter(ctx context.Context, args spec.QueryArgs) ([]entities.CatalogItem, int, error) {

	var conditions []string
	var sqlArgs []any

	argIdx := 1

	if brandID, err := args.ParseBrandId(); err != nil {
		return nil, 0, fmt.Errorf("invalid brandId: %w", err)
	} else if brandID != nil {
		conditions = append(conditions, fmt.Sprintf("ci.brand_id = $%d", argIdx))
		sqlArgs = append(sqlArgs, *brandID)
		argIdx++
	}

	if categoryID, err := args.ParseCategoryId(); err != nil {
		return nil, 0, fmt.Errorf("invalid categoryId: %w", err)
	} else if categoryID != nil {
		conditions = append(conditions, fmt.Sprintf("ci.category_id = $%d", argIdx))
		sqlArgs = append(sqlArgs, *categoryID)
		argIdx++
	}

	if args.Search != nil && *args.Search != "" {
		pattern := fmt.Sprintf("ci.title ILIKE $%d OR ci.short_description ILIKE $%d", argIdx, argIdx+1)
		conditions = append(conditions, "("+pattern+")")
		searchVal := "%" + *args.Search + "%"
		sqlArgs = append(sqlArgs, searchVal, searchVal)
		argIdx += 2
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderClause := "ORDER BY ci.title ASC"
	if args.Sort != nil {
		switch *args.Sort {
		case "price_asc":
			orderClause = "ORDER BY ci.price ASC"
		case "price_desc":
			orderClause = "ORDER BY ci.price DESC"
		case "title_asc":
			orderClause = "ORDER BY ci.title ASC"
		case "title_desc":
			orderClause = "ORDER BY ci.title DESC"
		}
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM catalog_items ci
		LEFT JOIN brands b ON b.id = ci.brand_id
		LEFT JOIN categories c ON c.id = ci.category_id
		%s
	`, whereClause)

	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, sqlArgs...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count catalog items: %w", err)
	}

	limit := args.PageSize
	offset := (args.PageIndex - 1) * args.PageSize

	limitPlaceholder := fmt.Sprintf("$%d", argIdx)
	offsetPlaceholder := fmt.Sprintf("$%d", argIdx+1)
	sqlArgs = append(sqlArgs, limit, offset)

	dataQuery := fmt.Sprintf(`
		%s
		%s
		%s
		LIMIT %s OFFSET %s
	`, sqlCatalogItemsQuery, whereClause, orderClause, limitPlaceholder, offsetPlaceholder)

	rows, err := r.db.QueryContext(ctx, dataQuery, sqlArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("items with filter query: %w", err)
	}
	defer rows.Close()

	var items []entities.CatalogItem
	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan catalog item (filter): %w", err)
		}
		items = append(items, item)
	}

	return items, totalCount, rows.Err()
}

func (r *itemRepository) ItemsByTitlePaged(ctx context.Context, title string, args spec.QueryArgs) ([]entities.CatalogItem, int, error) {

	countQuery := `
		SELECT COUNT(*)
		FROM catalog_items ci
		LEFT JOIN brands b ON b.id = ci.brand_id
		LEFT JOIN categories c ON c.id = ci.category_id
		WHERE ci.title LIKE '%' || $1 || '%'
	`
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, title).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("items by title count: %w", err)
	}

	offset := (args.PageIndex - 1) * args.PageSize

	dataQuery := sqlCatalogItemsQuery + `WHERE ci.title LIKE '%' || $1 || '%' ORDER BY ci.title ASC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, dataQuery, title, args.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("items by title paged query: %w", err)
	}
	defer rows.Close()

	var items []entities.CatalogItem
	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan catalog item by title: %w", err)
		}
		items = append(items, item)
	}

	return items, totalCount, rows.Err()
}

func (r *itemRepository) ItemsByBrandPaged(ctx context.Context, brandTitle string, args spec.QueryArgs) ([]entities.CatalogItem, int, error) {

	countQuery := `
		SELECT COUNT(*)
		FROM catalog_items ci
		LEFT JOIN brands b ON b.id = ci.brand_id
		LEFT JOIN categories c ON c.id = ci.category_id
		WHERE b.title LIKE '%' || $1 || '%'
	`
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, brandTitle).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("items by brand count: %w", err)
	}

	offset := (args.PageIndex - 1) * args.PageSize

	dataQuery := sqlCatalogItemsQuery + `WHERE b.title LIKE '%' || $1 || '%' ORDER BY ci.title ASC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, dataQuery, brandTitle, args.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("items by brand paged query: %w", err)
	}
	defer rows.Close()

	var items []entities.CatalogItem
	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan catalog item by brand: %w", err)
		}
		items = append(items, item)
	}

	return items, totalCount, rows.Err()
}
