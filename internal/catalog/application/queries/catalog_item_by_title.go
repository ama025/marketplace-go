package queries

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"
)

type CatalogItemByTitleHandler struct {
	repo repositories.CatalogItemRepository
}

func NewCatalogItemByTitleHandler(repo repositories.CatalogItemRepository) *CatalogItemByTitleHandler {
	return &CatalogItemByTitleHandler{repo: repo}
}

func (h *CatalogItemByTitleHandler) Handle(ctx context.Context, title string) ([]entities.CatalogItem, error) {
	return h.repo.ItemsByTitle(ctx, title)
}