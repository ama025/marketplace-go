package queries // Слой Application (Use Cases / Queries) — содержит сценарии использования (бизнес-логику)

import (
	"context" // Пакет для работы с контекстом (таймауты, отмена выполнения)

	"marketplace/internal/catalog/domain/entities"    // Импортируем доменные сущности (CatalogItem)
	"marketplace/internal/catalog/domain/repositories" // Импортируем абстрактные интерфейсы репозиториев
	"marketplace/internal/catalog/domain/spec"         // Параметры фильтрации, сортировки и пагинации
)

// CatalogItemsHandler — обработчик бизнес-сценария «Получить список товаров каталога с фильтрацией».
// В соответствии с принципами Чистой Архитектуры, он зависит ТОЛЬКО от абстрактного
// интерфейса (repositories.CatalogItemRepository), а не от конкретной реализации PostgreSQL или MySQL.
type CatalogItemsHandler struct {
	repo repositories.CatalogItemRepository // Абстрактный интерфейс репозитория товаров
}

// NewCatalogItemsHandler — конструктор создания сценария получения товаров каталога.
// Принимает любую структуру, реализующую интерфейс repositories.CatalogItemRepository.
func NewCatalogItemsHandler(repo repositories.CatalogItemRepository) *CatalogItemsHandler {
	return &CatalogItemsHandler{repo: repo}
}

// Handle — основной метод выполнения бизнес-сценария.
//
// Принимает QueryArgs с параметрами фильтрации, сортировки и пагинации.
// Нормализует параметры (минимальный/максимальный pageSize) перед передачей в репозиторий.
//
// Параметры:
//   - ctx (context.Context): контекст выполнения для контроля таймаутов
//   - args (spec.QueryArgs): параметры запроса (brandId, categoryId, search, sort, pageIndex, pageSize)
//
// Возвращает:
//   - spec.Pagination[entities.CatalogItem]: страница товаров + метаданные пагинации
//   - error: ошибку, если чтение из репозитория не удалось
func (q *CatalogItemsHandler) Handle(ctx context.Context, args spec.QueryArgs) (spec.Pagination[entities.CatalogItem], error) {
	// Нормализуем параметры пагинации: приводим pageIndex и pageSize к допустимым значениям
	args.NormaLize()

	// Делегируем задачу репозиторию: он построит SQL с WHERE, ORDER BY, LIMIT/OFFSET
	items, totalCount, err := q.repo.ItemsWithFilter(ctx, args)
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
