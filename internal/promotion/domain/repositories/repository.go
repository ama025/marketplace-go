package repositories

import (
	"context"

	"marketplace/internal/promotion/domain/entities"
)

type PromotionRepository interface {

	Add(ctx context.Context, d entities.Discount) (string, error)

	FindByCatalogItem(ctx context.Context, itemID string) (*entities.Discount, error)

	FindManyByCatalogItems(ctx context.Context, itemIDs []string) ([]entities.Discount, error)

	Deactivate(ctx context.Context, discountID string) error
}
