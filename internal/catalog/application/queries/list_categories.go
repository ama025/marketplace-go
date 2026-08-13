package queries // Слой Application (Use Cases / Queries)

import (
	"context"

	"marketplace/internal/catalog/domain/entities"    // Доменные сущности (Category)
	"marketplace/internal/catalog/domain/repositories" // Абстрактные интерфейсы репозиториев
	"marketplace/internal/catalog/domain/spec"         // Параметры пагинации
)

// CategoriesHandler — обработчик бизнес-сценария "Получить список всех категорий".
// Зависит ТОЛЬКО от абстрактного интерфейса (repositories.CategoryRepository).
type CategoriesHandler struct {
	repo repositories.CategoryRepository
}

// NewCategoriesHandler — конструктор создания сценария получения категорий.
func NewCategoriesHandler(repo repositories.CategoryRepository) *CategoriesHandler {
	return &CategoriesHandler{repo: repo}
}

// Handle — основной метод выполнения бизнес-сценария (без пагинации).
func (q *CategoriesHandler) Handle(ctx context.Context) ([]entities.Category, error) {
	return q.repo.Categories(ctx)
}

// HandlePaged — метод выполнения бизнес-сценария с пагинацией.
// Нормализует параметры пагинации и возвращает страницу категорий с метаданными.
//
// Параметры:
//   - ctx (context.Context): контекст выполнения
//   - args (spec.QueryArgs): параметры пагинации (PageIndex, PageSize)
//
// Возвращает:
//   - spec.Pagination[entities.Category]: страница категорий + метаданные пагинации
//   - error: ошибку, если чтение из репозитория не удалось
func (q *CategoriesHandler) HandlePaged(ctx context.Context, args spec.QueryArgs) (spec.Pagination[entities.Category], error) {
	// Нормализуем параметры пагинации: приводим pageIndex и pageSize к допустимым значениям
	args.NormaLize()

	// Делегируем задачу репозиторию: он вернёт страницу категорий и общее количество
	categories, totalCount, err := q.repo.CategoriesPaged(ctx, args)
	if err != nil {
		return spec.Pagination[entities.Category]{}, err
	}

	// Собираем и возвращаем страничный ответ
	return spec.Pagination[entities.Category]{
		PageIndex:  args.PageIndex,
		PageSize:   args.PageSize,
		TotalCount: totalCount,
		Items:      categories,
	}, nil
}
