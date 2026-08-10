package repositories // Слой Domain Repositories — определение контрактов (интерфейсов) для работы с хранилищами данных

import (
	"context" // Стандартный контекст для передачи сигналов отмены и таймаутов

	"marketplace/internal/catalog/domain/entities" // Импортируем доменные сущности

	"github.com/google/uuid"
)

// CatalogItemRepository — абстрактный интерфейс репозитория для работы с товарами каталога.
// Любая конкретная реализация (например, PostgreSQL, MongoDB, In-Memory) должна реализовать эти методы.
type CatalogItemRepository interface {
	Items(ctx context.Context) ([]entities.CatalogItem, error) // Метод получения всех товаров
	Item(ctx context.Context, id uuid.UUID) (*entities.CatalogItem, error)
	ItemsByTitle(ctx context.Context, title string) ([]entities.CatalogItem, error)
	Create(ctx context.Context, item *entities.CatalogItem) (*entities.CatalogItem, error)
}

// BrandRepository — абстрактный интерфейс репозитория для работы с брендами.
type BrandRepository interface {
	Brands(ctx context.Context) ([]entities.Brand, error) // Метод получения всех брендов
}

// CategoryRepository — абстрактный интерфейс репозитория для работы с категориями.
type CategoryRepository interface {
	Categories(ctx context.Context) ([]entities.Category, error) // Метод получения всех категорий
}
