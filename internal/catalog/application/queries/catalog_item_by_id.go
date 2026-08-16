package queries

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"

	"github.com/google/uuid"
)

type CatalogItemByIDHandler struct {
	repo repositories.CatalogItemRepository
}

func NewCatalogItemByIDHandler(repo repositories.CatalogItemRepository) *CatalogItemByIDHandler {
	return &CatalogItemByIDHandler{repo: repo}
}

func (h *CatalogItemByIDHandler) Handle(ctx context.Context, id uuid.UUID) (*entities.CatalogItem, error) {
	return h.repo.Item(ctx, id)
}
