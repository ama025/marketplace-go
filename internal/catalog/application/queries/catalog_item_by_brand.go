package queries

import (
	"context"

	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"
	"marketplace/internal/catalog/domain/spec"
)

type CatalogItemByBrandHandler struct {
	repo repositories.CatalogItemRepository
}

func NewCatalogItemByBrandHandler(repo repositories.CatalogItemRepository) *CatalogItemByBrandHandler {
	return &CatalogItemByBrandHandler{repo: repo}
}

func (h *CatalogItemByBrandHandler) Handle(ctx context.Context, brandTitle string) ([]entities.CatalogItem, error) {
	return h.repo.ItemsByBrand(ctx, brandTitle)
}

func (h *CatalogItemByBrandHandler) HandlePaged(ctx context.Context, brandTitle string, args spec.QueryArgs) (spec.Pagination[entities.CatalogItem], error) {

	args.NormaLize()

	items, totalCount, err := h.repo.ItemsByBrandPaged(ctx, brandTitle, args)
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
