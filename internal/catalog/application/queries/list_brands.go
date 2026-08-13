package queries // Слой Application (Use Cases / Queries) — содержит сценарии использования (бизнес-логику)

import (
	"context" // Пакет для работы с контекстом (таймауты, отмена выполнения)

	"marketplace/internal/catalog/domain/entities"    // Импортируем доменные сущности (Brand)
	"marketplace/internal/catalog/domain/repositories" // Импортируем абстрактные интерфейсы репозиториев
	"marketplace/internal/catalog/domain/spec"         // Параметры пагинации
)

// BrandsHandler — обработчик бизнес-сценария "Получить список всех брендов".
// В соответствии с принципами Чистой Архитектуры, он зависит ТОЛЬКО от абстрактного
// интерфейса (repositories.BrandRepository), а не от конкретной реализации PostgreSQL или MySQL.
type BrandsHandler struct {
	repo repositories.BrandRepository // Абстрактный интерфейс репозитория
}

// NewBrandsHandler — конструктор создания сценария получения брендов.
// Принимает любую структуру, реализующую интерфейс repositories.BrandRepository.
func NewBrandsHandler(repo repositories.BrandRepository) *BrandsHandler {
	return &BrandsHandler{repo: repo}
}

// Handle — основной метод выполнения бизнес-сценария (без пагинации).
//
// Параметры:
//   - ctx (context.Context): контекст выполнения для контроля таймаутов
//
// Возвращает:
//   - []entities.Brand: массив доменных объектов брендов
//   - error: ошибку, если чтение из репозитория не удалось
func (q *BrandsHandler) Handle(ctx context.Context) ([]entities.Brand, error) {
	// Делегируем задачу получения данных репозиторию
	return q.repo.Brands(ctx)
}

// HandlePaged — метод выполнения бизнес-сценария с пагинацией.
// Нормализует параметры пагинации и возвращает страницу брендов с метаданными.
//
// Параметры:
//   - ctx (context.Context): контекст выполнения
//   - args (spec.QueryArgs): параметры пагинации (PageIndex, PageSize)
//
// Возвращает:
//   - spec.Pagination[entities.Brand]: страница брендов + метаданные пагинации
//   - error: ошибку, если чтение из репозитория не удалось
func (q *BrandsHandler) HandlePaged(ctx context.Context, args spec.QueryArgs) (spec.Pagination[entities.Brand], error) {
	// Нормализуем параметры пагинации: приводим pageIndex и pageSize к допустимым значениям
	args.NormaLize()

	// Делегируем задачу репозиторию: он вернёт страницу брендов и общее количество
	brands, totalCount, err := q.repo.BrandsPaged(ctx, args)
	if err != nil {
		return spec.Pagination[entities.Brand]{}, err
	}

	// Собираем и возвращаем страничный ответ
	return spec.Pagination[entities.Brand]{
		PageIndex:  args.PageIndex,
		PageSize:   args.PageSize,
		TotalCount: totalCount,
		Items:      brands,
	}, nil
}
