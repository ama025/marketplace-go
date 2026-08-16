package queries

import (
	"context"

	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"
	"marketplace/internal/catalog/domain/spec"
)

type BrandsHandler struct {
	repo repositories.BrandRepository
}

func NewBrandsHandler(repo repositories.BrandRepository) *BrandsHandler {
	return &BrandsHandler{repo: repo}
}

func (q *BrandsHandler) Handle(ctx context.Context) ([]entities.Brand, error) {

	return q.repo.Brands(ctx)
}

func (q *BrandsHandler) HandlePaged(ctx context.Context, args spec.QueryArgs) (spec.Pagination[entities.Brand], error) {

	args.NormaLize()

	brands, totalCount, err := q.repo.BrandsPaged(ctx, args)
	if err != nil {
		return spec.Pagination[entities.Brand]{}, err
	}

	return spec.Pagination[entities.Brand]{
		PageIndex:  args.PageIndex,
		PageSize:   args.PageSize,
		TotalCount: totalCount,
		Items:      brands,
	}, nil
}
