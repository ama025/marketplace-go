package queries

import (
	"context"

	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"
	"marketplace/internal/catalog/domain/spec"
)

type CategoriesHandler struct {
	repo repositories.CategoryRepository
}

func NewCategoriesHandler(repo repositories.CategoryRepository) *CategoriesHandler {
	return &CategoriesHandler{repo: repo}
}

func (q *CategoriesHandler) Handle(ctx context.Context) ([]entities.Category, error) {
	return q.repo.Categories(ctx)
}

func (q *CategoriesHandler) HandlePaged(ctx context.Context, args spec.QueryArgs) (spec.Pagination[entities.Category], error) {

	args.NormaLize()

	categories, totalCount, err := q.repo.CategoriesPaged(ctx, args)
	if err != nil {
		return spec.Pagination[entities.Category]{}, err
	}

	return spec.Pagination[entities.Category]{
		PageIndex:  args.PageIndex,
		PageSize:   args.PageSize,
		TotalCount: totalCount,
		Items:      categories,
	}, nil
}
