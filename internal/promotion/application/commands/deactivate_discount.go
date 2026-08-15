package commands

import (
	"context"

	"marketplace/internal/promotion/domain/repositories"
)

// DeactivateDiscountCommand — входные данные для деактивации скидки.
type DeactivateDiscountCommand struct {
	DiscountID string // UUID скидки которую нужно деактивировать
}

// DeactivateDiscountHandler — use case «Отключить скидку».
// Скидка не удаляется физически — только помечается active=false.
// Это позволяет сохранить историю и откатиться при необходимости.
type DeactivateDiscountHandler struct {
	repo repositories.PromotionRepository
}

func NewDeactivateDiscountHandler(repo repositories.PromotionRepository) *DeactivateDiscountHandler {
	return &DeactivateDiscountHandler{repo: repo}
}

// Handle деактивирует скидку по её UUID.
func (h *DeactivateDiscountHandler) Handle(ctx context.Context, cmd DeactivateDiscountCommand) error {
	return h.repo.Deactivate(ctx, cmd.DiscountID)
}
