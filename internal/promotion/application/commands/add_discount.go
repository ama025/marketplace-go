package commands

import (
	"context"

	"marketplace/internal/promotion/domain/entities"
	"marketplace/internal/promotion/domain/repositories"
)

// AddDiscountCommand — входные данные для создания скидки.
type AddDiscountCommand struct {
	ItemID   string     // UUID товара из каталога
	Percent  float64    // Процент скидки (0–100)
	StartsAt *string    // RFC3339 строка или nil (сразу)
	EndsAt   *string    // RFC3339 строка или nil (бессрочно)
}

// AddDiscountHandler — use case «Создать скидку для товара».
type AddDiscountHandler struct {
	repo repositories.PromotionRepository
}

func NewAddDiscountHandler(repo repositories.PromotionRepository) *AddDiscountHandler {
	return &AddDiscountHandler{repo: repo}
}

// Handle создаёт скидку и возвращает её UUID.
func (h *AddDiscountHandler) Handle(ctx context.Context, cmd AddDiscountCommand) (string, error) {
	d := entities.Discount{
		ItemID:  cmd.ItemID,
		Percent: cmd.Percent,
		Active:  true,
	}
	return h.repo.Add(ctx, d)
}
