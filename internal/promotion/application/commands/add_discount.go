package commands

import (
	"context"

	"marketplace/internal/promotion/domain/entities"
	"marketplace/internal/promotion/domain/repositories"
)

type AddDiscountCommand struct {
	ItemID   string
	Percent  float64
	StartsAt *string
	EndsAt   *string
}

type AddDiscountHandler struct {
	repo repositories.PromotionRepository
}

func NewAddDiscountHandler(repo repositories.PromotionRepository) *AddDiscountHandler {
	return &AddDiscountHandler{repo: repo}
}

func (h *AddDiscountHandler) Handle(ctx context.Context, cmd AddDiscountCommand) (string, error) {
	d := entities.Discount{
		ItemID:  cmd.ItemID,
		Percent: cmd.Percent,
		Active:  true,
	}
	return h.repo.Add(ctx, d)
}
