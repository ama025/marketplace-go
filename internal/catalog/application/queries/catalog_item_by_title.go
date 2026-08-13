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

// Handle — поиск товаров по названию (без пагинации).
func (h *CatalogItemByTitleHandler) Handle(ctx context.Context, title string) ([]entities.CatalogItem, error) {
	return h.repo.ItemsByTitle(ctx, title)
}

// HandlePaged — выполняет поиск товаров по названию с пагинацией.
// Нормализует параметры пагинации и возвращает страницу результатов.
//
// Параметры:
//   - ctx: контекст запроса (таймауты, отмена)
//   - title: строка поиска (частичное совпадение по названию)
//   - args: параметры пагинации (PageIndex, PageSize)
//
// Возвращает:
//   - spec.Pagination[entities.CatalogItem]: страница товаров + метаданные пагинации
//   - error: ошибку, если запрос к БД не удался
func (h *CatalogItemByTitleHandler) HandlePaged(ctx context.Context, title string, args spec.QueryArgs) (spec.Pagination[entities.CatalogItem], error) {
	// Нормализуем параметры пагинации: приводим pageIndex и pageSize к допустимым значениям
	args.NormaLize()

	// Делегируем запрос репозиторию: он вернёт страницу товаров и общее число совпадений
	items, totalCount, err := h.repo.ItemsByTitlePaged(ctx, title, args)
	if err != nil {
		return spec.Pagination[entities.CatalogItem]{}, err
	}

	// Собираем и возвращаем страничный ответ
	return spec.Pagination[entities.CatalogItem]{
		PageIndex:  args.PageIndex,
		PageSize:   args.PageSize,
		TotalCount: totalCount,
		Items:      items,
	}, nil
}