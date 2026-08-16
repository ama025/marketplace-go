package queries

import (
	"context"

	"marketplace/internal/promotion/domain/entities"
	"marketplace/internal/promotion/domain/repositories"
)

type FindManyByCatalogItemsHandler struct {
	repo repositories.PromotionRepository
}

func NewFindManyByCatalogItemsHandler(repo repositories.PromotionRepository) *FindManyByCatalogItemsHandler {
	return &FindManyByCatalogItemsHandler{repo: repo}
}

func (h *FindManyByCatalogItemsHandler) Handle(ctx context.Context, itemIDs []string) ([]entities.Discount, error) {
	return h.repo.FindManyByCatalogItems(ctx, itemIDs)
}
