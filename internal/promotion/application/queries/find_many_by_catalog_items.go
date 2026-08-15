package queries

import (
	"context"

	"marketplace/internal/promotion/domain/entities"
	"marketplace/internal/promotion/domain/repositories"
)

// FindManyByCatalogItemsHandler — use case «Получить скидки для списка товаров».
//
// Используется basket-ом при загрузке корзины:
// вместо N отдельных запросов — один batch-запрос для всех товаров.
type FindManyByCatalogItemsHandler struct {
	repo repositories.PromotionRepository
}

// NewFindManyByCatalogItemsHandler создаёт обработчик batch-запроса.
func NewFindManyByCatalogItemsHandler(repo repositories.PromotionRepository) *FindManyByCatalogItemsHandler {
	return &FindManyByCatalogItemsHandler{repo: repo}
}

// Handle возвращает активные скидки для переданных UUID товаров.
// Товары без скидки отсутствуют в результирующем слайсе (не возвращается ошибка).
func (h *FindManyByCatalogItemsHandler) Handle(ctx context.Context, itemIDs []string) ([]entities.Discount, error) {
	return h.repo.FindManyByCatalogItems(ctx, itemIDs)
}
