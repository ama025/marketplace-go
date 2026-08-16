package queries

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"
	"marketplace/internal/catalog/domain/spec"
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

func (h *CatalogItemByTitleHandler) HandlePaged(ctx context.Context, title string, args spec.QueryArgs) (spec.Pagination[entities.CatalogItem], error) {

	args.NormaLize()

	items, totalCount, err := h.repo.ItemsByTitlePaged(ctx, title, args)
	if err != nil {
		return spec.Pagination[entities.CatalogItem]{}, err
	}

	return spec.Pagination[entities.CatalogItem]{
		PageIndex:  args.PageIndex,
		PageSize:   args.PageSize,
		TotalCount: totalCount,
		Items:      items,
	}, nil
}