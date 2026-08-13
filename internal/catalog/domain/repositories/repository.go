package repositories // Слой Domain Repositories — определение контрактов (интерфейсов) для работы с хранилищами данных

import (
	"context" // Стандартный контекст для передачи сигналов отмены и таймаутов

	"marketplace/internal/catalog/domain/entities" // Импортируем доменные сущности
	"marketplace/internal/catalog/domain/spec"     // Параметры фильтрации, сортировки и пагинации

	"github.com/google/uuid"
)

// CatalogItemRepository — абстрактный интерфейс репозитория для работы с товарами каталога.
// Любая конкретная реализация (например, PostgreSQL, MongoDB, In-Memory) должна реализовать эти методы.
type CatalogItemRepository interface {
	Items(ctx context.Context) ([]entities.CatalogItem, error)                                    // Метод получения всех товаров
	ItemsWithFilter(ctx context.Context, args spec.QueryArgs) ([]entities.CatalogItem, int, error) // Фильтрация, сортировка, пагинация
	ItemsByTitlePaged(ctx context.Context, title string, args spec.QueryArgs) ([]entities.CatalogItem, int, error) // Поиск по названию с пагинацией
	ItemsByBrandPaged(ctx context.Context, brandTitle string, args spec.QueryArgs) ([]entities.CatalogItem, int, error) // Поиск по бренду с пагинацией
	Item(ctx context.Context, id uuid.UUID) (*entities.CatalogItem, error)
	ItemsByTitle(ctx context.Context, title string) ([]entities.CatalogItem, error)
	ItemsByBrand(ctx context.Context, brandTitle string) ([]entities.CatalogItem, error)
	Create(ctx context.Context, item *entities.CatalogItem) (*entities.CatalogItem, error)
	Update(ctx context.Context, item *entities.CatalogItem) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
}

// BrandRepository — абстрактный интерфейс репозитория для работы с брендами.
type BrandRepository interface {
	Brands(ctx context.Context) ([]entities.Brand, error)                                       // Метод получения всех брендов
	BrandsPaged(ctx context.Context, args spec.QueryArgs) ([]entities.Brand, int, error)        // Метод получения брендов с пагинацией
}

// CategoryRepository — абстрактный интерфейс репозитория для работы с категориями.
type CategoryRepository interface {
	Categories(ctx context.Context) ([]entities.Category, error)                                         // Метод получения всех категорий
	CategoriesPaged(ctx context.Context, args spec.QueryArgs) ([]entities.Category, int, error)          // Метод получения категорий с пагинацией
}
