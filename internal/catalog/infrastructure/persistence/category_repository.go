package persistence // Слой Infrastructure (Persistence) — работа с PostgreSQL через SQL

import (
	"context"      // Контекст Go для управления отменой и таймаутами
	"database/sql" // Пакет для отправки запросов в реляционную БД (*sql.DB)
	"fmt"          // Пакет для форматирования ошибок (fmt.Errorf)

	"marketplace/internal/catalog/domain/entities" // Доменная модель Category
)

// CategoryRepository — конкретная структура репозитория категорий.
type CategoryRepository struct {
	db *sql.DB // Ссылка на пул соединений с PostgreSQL
}

// NewCategoryRepository — конструктор репозитория категорий.
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Categories — метод выполнения SQL-запроса для получения всех категорий из таблицы "categories".
// Реализует доменный интерфейс repositories.CategoryRepository.
func (r *CategoryRepository) Categories(ctx context.Context) ([]entities.Category, error) {
	// Отправляем SQL-запрос SELECT к таблице categories с помощью QueryContext
	rows, err := r.db.QueryContext(ctx, `
	SELECT
		id, title
	FROM categories
	ORDER BY title
	`)

	if err != nil {
		return nil, fmt.Errorf("categories query: %w", err)
	}

	defer rows.Close()

	var categories []entities.Category

	// Читаем строки из результата БД
	for rows.Next() {
		var c entities.Category

		if err := rows.Scan(&c.Id, &c.Title); err != nil {
			return nil, fmt.Errorf("scan category error: %w", err)
		}

		categories = append(categories, c)
	}

	return categories, rows.Err()
}
