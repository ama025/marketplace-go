package persistence // Слой Infrastructure (Persistence) — конкретная реализация взаимодействия с базой данных PostgreSQL через SQL

import (
	"context"      // Стандартный контекст Go для управления временем выполнения операций
	"database/sql" // Пакет для отправки запросов в реляционную БД (*sql.DB)
	"fmt"          // Пакет для форматирования ошибок (fmt.Errorf)

	"marketplace/internal/catalog/domain/entities" // Импортируем доменную модель Brand
	"marketplace/internal/catalog/domain/spec"     // Параметры пагинации
)

// BrandRepository — конкретная структура репозитория брендов.
// Работает напрямую с базой данных PostgreSQL через драйвер *sql.DB.
type BrandRepository struct {
	db *sql.DB // Ссылка на пул соединений с базой данных
}

// NewBrandRepository — конструктор репозитория брендов.
func NewBrandRepository(db *sql.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

// Brands — метод выполнения SQL-запроса для получения всех брендов из таблицы "brands".
// Реализует доменный интерфейс repositories.BrandRepository.
//
// Параметры:
//   - ctx (context.Context): контекст запроса (отменяет SQL-запрос, если клиент отсоединился)
//
// Возвращает:
//   - []entities.Brand: полученный список брендов
//   - error: ошибку SQL-запроса или сканирования строк
func (r *BrandRepository) Brands(ctx context.Context) ([]entities.Brand, error) {
	// Отправляем SQL-запрос SELECT к таблице brands с помощью QueryContext (для поддержки отмены по ctx)
	rows, err := r.db.QueryContext(ctx, ` 
	SELECT
		id, title
	FROM brands
	ORDER BY title
	`) // Запрос выбирает поля id и title всех брендов, отсортированных по имени

	// Если SQL-запрос упал с ошибкой — возвращаем ее с оберткой fmt.Errorf
	if err != nil {
		return nil, fmt.Errorf("brands query: %w", err)
	}

	// defer rows.Close() гарантирует освобождение ресурсов соединений с БД после окончания чтения
	defer rows.Close()

	// Слайс для накопления результатов
	var brands []entities.Brand

	// Проходимся в цикле rows.Next() по всем возвращенным строкам ответа БД
	for rows.Next() {
		var b entities.Brand

		// Scan считывает колонки текущей строки (id, title) в поля структуры b.Id и b.Title
		if err := rows.Scan(&b.Id, &b.Title); err != nil {
			return nil, fmt.Errorf("scan brand error: %w", err)
		}

		// Добавляем прочитанную сущность бренда в итоговый массив
		brands = append(brands, b)
	}

	// Проверяем, не возникло ли ошибок во время итерирования по курсору rows
	return brands, rows.Err()
}

// BrandsPaged — метод получения брендов с пагинацией.
// Выполняет два запроса: COUNT для подсчёта общего числа брендов и SELECT с LIMIT/OFFSET для текущей страницы.
//
// Параметры:
//   - ctx (context.Context): контекст запроса
//   - args (spec.QueryArgs): параметры пагинации (PageIndex, PageSize)
//
// Возвращает:
//   - []entities.Brand: список брендов на текущей странице
//   - int: общее количество брендов (для расчёта числа страниц)
//   - error: ошибку SQL-запроса или сканирования строк
func (r *BrandRepository) BrandsPaged(ctx context.Context, args spec.QueryArgs) ([]entities.Brand, int, error) {
	// Запрос общего количества брендов (без LIMIT/OFFSET)
	var totalCount int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM brands`).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("brands count query: %w", err)
	}

	// Вычисляем смещение (OFFSET) из номера страницы
	offset := (args.PageIndex - 1) * args.PageSize

	// Запрос страницы брендов с LIMIT и OFFSET
	rows, err := r.db.QueryContext(ctx, `
	SELECT
		id, title
	FROM brands
	ORDER BY title
	LIMIT $1 OFFSET $2
	`, args.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("brands paged query: %w", err)
	}
	defer rows.Close()

	var brands []entities.Brand

	for rows.Next() {
		var b entities.Brand
		if err := rows.Scan(&b.Id, &b.Title); err != nil {
			return nil, 0, fmt.Errorf("scan brand error: %w", err)
		}
		brands = append(brands, b)
	}

	return brands, totalCount, rows.Err()
}
