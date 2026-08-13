package spec

import "github.com/google/uuid"

const MaxPageSize = 4

type QueryArgs struct {
	PageIndex  int     `form:"pageIndex"`
	PageSize   int     `form:"pageSize"`
	BrandId    *string `form:"brandId"`
	CategoryId *string `form:"categoryId"`
	Search     *string `form:"search"`
	Sort       *string `form:"sort"`
}

func (q *QueryArgs) NormaLize() {
	if q.PageIndex < 1 {
		q.PageIndex = 1
	}

	if q.PageSize <= 0 {
		q.PageSize = 2
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}

}

func (q *QueryArgs) ParseBrandId() (*uuid.UUID, error) {
	if q.BrandId == nil || *q.BrandId == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*q.BrandId)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (q *QueryArgs) ParseCategoryId() (*uuid.UUID, error) {
	if q.CategoryId == nil || *q.CategoryId == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*q.CategoryId)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
