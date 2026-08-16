package commands

import (
	"context"

	"marketplace/internal/promotion/domain/repositories"
)

type DeactivateDiscountCommand struct {
	DiscountID string
}

type DeactivateDiscountHandler struct {
	repo repositories.PromotionRepository
}

func NewDeactivateDiscountHandler(repo repositories.PromotionRepository) *DeactivateDiscountHandler {
	return &DeactivateDiscountHandler{repo: repo}
}

func (h *DeactivateDiscountHandler) Handle(ctx context.Context, cmd DeactivateDiscountCommand) error {
	return h.repo.Deactivate(ctx, cmd.DiscountID)
}
