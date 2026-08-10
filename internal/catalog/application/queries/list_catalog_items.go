package queries // Слой Application (Use Cases / Queries) — содержит сценарии использования (бизнес-логику)

import (
	"context" // Пакет для работы с контекстом (таймауты, отмена выполнения)

	"marketplace/internal/catalog/domain/entities"    // Импортируем доменные сущности (CatalogItem)
	"marketplace/internal/catalog/domain/repositories" // Импортируем абстрактные интерфейсы репозиториев
)

// CatalogItemsHandler — обработчик бизнес-сценария "Получить список всех товаров каталога".
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
// Параметры:
//   - ctx (context.Context): контекст выполнения для контроля таймаутов
//
// Возвращает:
//   - []entities.CatalogItem: массив доменных объектов товаров каталога (с брендами и категориями)
//   - error: ошибку, если чтение из репозитория не удалось
func (q *CatalogItemsHandler) Handle(ctx context.Context) ([]entities.CatalogItem, error) {
	// Делегируем задачу получения данных репозиторию.
	// Этот слой не знает, откуда берутся данные (PostgreSQL, MongoDB, мок) —
	// он просто вызывает метод интерфейса.
	return q.repo.Items(ctx)
}
