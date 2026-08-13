package queries

import (
	"context"
	"marketplace/internal/catalog/domain/entities"
	"marketplace/internal/catalog/domain/repositories"
	"marketplace/internal/catalog/domain/spec"
)

// CatalogItemsV2Handler — альтернативный обработчик получения товаров каталога.
// Делегирует запрос напрямую в ItemsWithFilter репозитория, минуя нормализацию в Handle.
// Используется для эндпоинтов v2, где нормализация делается на уровне выше.
type CatalogItemsV2Handler struct {
	repo repositories.CatalogItemRepository
}

func NewCatalogItemsV2Handler(repo repositories.CatalogItemRepository) *CatalogItemsV2Handler {
	return &CatalogItemsV2Handler{repo: repo}
}

// Handle — получает страницу товаров с фильтрацией, сортировкой и пагинацией.
// Нормализует параметры пагинации перед передачей в репозиторий.
//
// Параметры:
//   - ctx: контекст запроса (таймауты, отмена)
//   - args: параметры запроса (brandId, categoryId, search, sort, pageIndex, pageSize)
//
// Возвращает:
//   - spec.Pagination[entities.CatalogItem]: страница товаров + метаданные пагинации
//   - error: ошибку, если запрос к БД не удался
func (h *CatalogItemsV2Handler) Handle(ctx context.Context, args spec.QueryArgs) (spec.Pagination[entities.CatalogItem], error) {
	// Нормализуем параметры пагинации: приводим pageIndex и pageSize к допустимым значениям
	args.NormaLize()

	// Делегируем запрос репозиторию: он построит SQL с WHERE, ORDER BY, LIMIT/OFFSET
	items, totalCount, err := h.repo.ItemsWithFilter(ctx, args)
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
