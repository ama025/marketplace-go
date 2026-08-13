package queries // Слой Application (Queries) — сценарий поиска товаров по названию бренда

import (
	"context" // Стандартный контекст для передачи таймаутов и отмены

	"marketplace/internal/catalog/domain/entities"    // Доменная сущность CatalogItem
	"marketplace/internal/catalog/domain/repositories" // Интерфейс репозитория товаров
	"marketplace/internal/catalog/domain/spec"         // Параметры пагинации
)

// CatalogItemByBrandHandler — обработчик запроса на поиск товаров по названию бренда.
// Содержит ссылку на интерфейс репозитория (Dependency Injection).
type CatalogItemByBrandHandler struct {
	repo repositories.CatalogItemRepository // Абстрактный репозиторий — не зависим от конкретной БД
}

// NewCatalogItemByBrandHandler — конструктор обработчика поиска товаров по бренду.
// Принимает любую реализацию интерфейса CatalogItemRepository.
func NewCatalogItemByBrandHandler(repo repositories.CatalogItemRepository) *CatalogItemByBrandHandler {
	return &CatalogItemByBrandHandler{repo: repo}
}

// Handle — выполняет бизнес-сценарий: ищет все товары, принадлежащие указанному бренду (без пагинации).
// Делегирует работу репозиторию, передавая контекст и название бренда.
//
// Параметры:
//   - ctx: контекст запроса (таймауты, отмена)
//   - brandTitle: название бренда для поиска (частичное совпадение)
//
// Возвращает:
//   - []entities.CatalogItem: список найденных товаров
//   - error: ошибку, если запрос к БД не удался
func (h *CatalogItemByBrandHandler) Handle(ctx context.Context, brandTitle string) ([]entities.CatalogItem, error) {
	return h.repo.ItemsByBrand(ctx, brandTitle) // Передаём запрос в репозиторий
}

// HandlePaged — выполняет поиск товаров по бренду с пагинацией.
// Нормализует параметры пагинации и возвращает страницу результатов.
//
// Параметры:
//   - ctx: контекст запроса (таймауты, отмена)
//   - brandTitle: название бренда для поиска (частичное совпадение)
//   - args: параметры пагинации (PageIndex, PageSize)
//
// Возвращает:
//   - spec.Pagination[entities.CatalogItem]: страница товаров + метаданные пагинации
//   - error: ошибку, если запрос к БД не удался
func (h *CatalogItemByBrandHandler) HandlePaged(ctx context.Context, brandTitle string, args spec.QueryArgs) (spec.Pagination[entities.CatalogItem], error) {
	// Нормализуем параметры пагинации: приводим pageIndex и pageSize к допустимым значениям
	args.NormaLize()

	// Делегируем запрос репозиторию: он вернёт страницу товаров и общее число совпадений
	items, totalCount, err := h.repo.ItemsByBrandPaged(ctx, brandTitle, args)
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
