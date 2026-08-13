package commands

import (
	"context"
	"marketplace/internal/catalog/domain/repositories"

	"github.com/google/uuid"
)

type DeleteCatalogItemCommand struct {
	ID uuid.UUID `json: "id"`
}

type DeleteCatalogItemHandler struct {
	repo repositories.CatalogItemRepository
}

func NewDeleteCatalogItemHandler(repo repositories.CatalogItemRepository) *DeleteCatalogItemHandler {
	return &DeleteCatalogItemHandler{repo: repo}
}

func (h *DeleteCatalogItemHandler) Handle(ctx context.Context, id uuid.UUID) (bool, error) {
	existing, err := h.repo.Item(ctx, id)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, nil
	}

	return h.repo.Delete(ctx, id)
}
