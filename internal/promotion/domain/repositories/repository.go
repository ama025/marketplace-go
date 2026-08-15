package repositories

import (
	"context"

	"marketplace/internal/promotion/domain/entities"
)

// PromotionRepository — контракт для работы со скидками.
// Application-слой зависит только от этого интерфейса,
// конкретная реализация (MySQL) живёт в infrastructure/persistence.
type PromotionRepository interface {
	// Add создаёт новую скидку и возвращает её UUID.
	Add(ctx context.Context, d entities.Discount) (string, error)

	// FindByCatalogItem возвращает активную скидку для товара.
	// Если скидки нет — возвращает nil, nil.
	FindByCatalogItem(ctx context.Context, itemID string) (*entities.Discount, error)

	// FindManyByCatalogItems возвращает активные скидки для списка товаров (batch).
	// Товары без скидки просто отсутствуют в результате.
	FindManyByCatalogItems(ctx context.Context, itemIDs []string) ([]entities.Discount, error)

	// Deactivate помечает скидку как неактивную (active=false).
	// Физически запись не удаляется — история сохраняется.
	Deactivate(ctx context.Context, discountID string) error
}
