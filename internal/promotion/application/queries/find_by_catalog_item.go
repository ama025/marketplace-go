package queries

import (
	"context"

	"marketplace/internal/promotion/domain/entities"
	"marketplace/internal/promotion/domain/repositories"
)

// FindByCatalogItemHandler — use case «Найти скидку для товара каталога».
//
// Basket вызывает этот сценарий перед отображением цены:
//  1. Передаёт UUID товара
//  2. Получает скидку (или nil если скидки нет)
//  3. Рассчитывает итоговую цену: price * (1 - discount.Percent/100)
type FindByCatalogItemHandler struct {
	repo repositories.PromotionRepository
}

// NewFindByCatalogItemHandler создаёт обработчик запроса.
// Принимает любую реализацию PromotionRepository (MySQL, Mock и т.д.).
func NewFindByCatalogItemHandler(repo repositories.PromotionRepository) *FindByCatalogItemHandler {
	return &FindByCatalogItemHandler{repo: repo}
}

// Handle выполняет поиск активной скидки для товара.
//
// Возвращает:
//   - *entities.Discount — скидка, если есть
//   - nil, nil           — если скидки нет (не ошибка)
//   - nil, err           — если произошла ошибка БД
func (h *FindByCatalogItemHandler) Handle(ctx context.Context, itemID string) (*entities.Discount, error) {
	return h.repo.FindByCatalogItem(ctx, itemID)
}
