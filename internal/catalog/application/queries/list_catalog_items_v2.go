package queries

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"
	"marketplace/internal/catalog/domain/spec"
)

type CatalogItemsV2Handler struct {
	repo repositories.CatalogItemRepository
}

func NewCatalogItemsV2Handler(repo repositories.CatalogItemRepository) *CatalogItemsV2Handler {
	return &CatalogItemsV2Handler{repo: repo}
}

func (h *CatalogItemsV2Handler) Handle(ctx context.Context, args spec.QueryArgs) (spec.Pagination[entities.CatalogItem], error) {

	args.NormaLize()

	items, totalCount, err := h.repo.ItemsWithFilter(ctx, args)
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
