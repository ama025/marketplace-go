package repositories

import (
	"context"

	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/spec"

	"github.com/google/uuid"
)

type CatalogItemRepository interface {
	Items(ctx context.Context) ([]entities.CatalogItem, error)
	ItemsWithFilter(ctx context.Context, args spec.QueryArgs) ([]entities.CatalogItem, int, error)
	ItemsByTitlePaged(ctx context.Context, title string, args spec.QueryArgs) ([]entities.CatalogItem, int, error)
	ItemsByBrandPaged(ctx context.Context, brandTitle string, args spec.QueryArgs) ([]entities.CatalogItem, int, error)
	Item(ctx context.Context, id uuid.UUID) (*entities.CatalogItem, error)
	ItemsByTitle(ctx context.Context, title string) ([]entities.CatalogItem, error)
	ItemsByBrand(ctx context.Context, brandTitle string) ([]entities.CatalogItem, error)
	Create(ctx context.Context, item *entities.CatalogItem) (*entities.CatalogItem, error)
	Update(ctx context.Context, item *entities.CatalogItem) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
}

type BrandRepository interface {
	Brands(ctx context.Context) ([]entities.Brand, error)
	BrandsPaged(ctx context.Context, args spec.QueryArgs) ([]entities.Brand, int, error)
}

type CategoryRepository interface {
	Categories(ctx context.Context) ([]entities.Category, error)
	CategoriesPaged(ctx context.Context, args spec.QueryArgs) ([]entities.Category, int, error)
}
