package queries

import (
	"context"

	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"
	"marketplace/internal/catalog/domain/spec"
)

type CatalogItemsHandler struct {
	repo repositories.CatalogItemRepository
}

func NewCatalogItemsHandler(repo repositories.CatalogItemRepository) *CatalogItemsHandler {
	return &CatalogItemsHandler{repo: repo}
}

func (q *CatalogItemsHandler) Handle(ctx context.Context, args spec.QueryArgs) (spec.Pagination[entities.CatalogItem], error) {

	args.NormaLize()

	items, totalCount, err := q.repo.ItemsWithFilter(ctx, args)
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
