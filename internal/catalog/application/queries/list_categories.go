package queries // Слой Application (Use Cases / Queries)

import (
	"context"

	"marketplace/internal/catalog/domain/entities"    // Доменные сущности (Category)
	"marketplace/internal/catalog/domain/repositories" // Абстрактные интерфейсы репозиториев
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

// Handle — основной метод выполнения бизнес-сценария.
func (q *CategoriesHandler) Handle(ctx context.Context) ([]entities.Category, error) {
	return q.repo.Categories(ctx)
}
