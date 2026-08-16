package queries

import (
	"context"

	"marketplace/internal/promotion/domain/entities"
	"marketplace/internal/promotion/domain/repositories"
)

type FindByCatalogItemHandler struct {
	repo repositories.PromotionRepository
}

func NewFindByCatalogItemHandler(repo repositories.PromotionRepository) *FindByCatalogItemHandler {
	return &FindByCatalogItemHandler{repo: repo}
}

func (h *FindByCatalogItemHandler) Handle(ctx context.Context, itemID string) (*entities.Discount, error) {
	return h.repo.FindByCatalogItem(ctx, itemID)
}
