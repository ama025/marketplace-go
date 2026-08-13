package spec

type Pagination[T any] struct {
	PageIndex  int `json:"pageIndex"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
	Items      []T `json:"items"`
}
