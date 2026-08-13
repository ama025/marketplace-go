package persistence // Слой Infrastructure (Persistence) — работа с PostgreSQL через SQL

import (
	"context"      // Контекст Go для управления отменой и таймаутами
	"database/sql" // Пакет для отправки запросов в реляционную БД (*sql.DB)
	"fmt"          // Пакет для форматирования ошибок (fmt.Errorf)

	"marketplace/internal/catalog/domain/entities" // Доменная модель Category
	"marketplace/internal/catalog/domain/spec"     // Параметры пагинации
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

// CategoriesPaged — метод получения категорий с пагинацией.
// Выполняет два запроса: COUNT для подсчёта общего числа категорий и SELECT с LIMIT/OFFSET для текущей страницы.
//
// Параметры:
//   - ctx (context.Context): контекст запроса
//   - args (spec.QueryArgs): параметры пагинации (PageIndex, PageSize)
//
// Возвращает:
//   - []entities.Category: список категорий на текущей странице
//   - int: общее количество категорий (для расчёта числа страниц)
//   - error: ошибку SQL-запроса или сканирования строк
func (r *CategoryRepository) CategoriesPaged(ctx context.Context, args spec.QueryArgs) ([]entities.Category, int, error) {
	// Запрос общего количества категорий (без LIMIT/OFFSET)
	var totalCount int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories`).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("categories count query: %w", err)
	}

	// Вычисляем смещение (OFFSET) из номера страницы
	offset := (args.PageIndex - 1) * args.PageSize

	// Запрос страницы категорий с LIMIT и OFFSET
	rows, err := r.db.QueryContext(ctx, `
	SELECT
		id, title
	FROM categories
	ORDER BY title
	LIMIT $1 OFFSET $2
	`, args.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("categories paged query: %w", err)
	}
	defer rows.Close()

	var categories []entities.Category

	for rows.Next() {
		var c entities.Category
		if err := rows.Scan(&c.Id, &c.Title); err != nil {
			return nil, 0, fmt.Errorf("scan category error: %w", err)
		}
		categories = append(categories, c)
	}

	return categories, totalCount, rows.Err()
}

