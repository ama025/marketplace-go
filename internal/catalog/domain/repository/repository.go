package repository

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
)

type CatalogItemRepository interface {
	Items(ctx context.Context) ([]entities.CatalogItem, error)
}
