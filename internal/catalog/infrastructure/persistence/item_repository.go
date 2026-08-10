package persistence // Слой Infrastructure (Persistence) — конкретная реализация взаимодействия с базой данных PostgreSQL через SQL

import (
	"context"      // Стандартный контекст Go для управления временем выполнения операций
	"database/sql" // Пакет для отправки запросов в реляционную БД (*sql.DB)
	"fmt"          // Пакет для форматирования ошибок (fmt.Errorf)

	"marketplace/internal/catalog/domain/entities" // Импортируем доменную модель CatalogItem

	"github.com/google/uuid" // Пакет для работы с UUID-идентификаторами
)

const sqlCatalogItemsQuery = `
SELECT ci.id, ci.short_description, ci.full_description, ci.price,
	b.id as brand_id,b.title as brand_title,
	c.id as category_id,c.title as category_title
FROM catalog_items ci
LEFT JOIN brands b ON b.id = ci.brand_id
LEFT JOIN categories c ON c.id = ci.category_id
`

// itemRepository — конкретная структура репозитория товаров каталога.
// Работает напрямую с базой данных PostgreSQL через драйвер *sql.DB.
type itemRepository struct {
	db *sql.DB // Ссылка на пул соединений с базой данных
}

// NewItemRepository — конструктор репозитория товаров каталога.
// Принимает *sql.DB (пул соединений) и возвращает готовый к использованию репозиторий.
func NewItemRepository(db *sql.DB) *itemRepository {
	return &itemRepository{db: db}
}

// Items — метод выполнения SQL-запроса для получения всех товаров каталога.
// Реализует доменный интерфейс repositories.CatalogItemRepository.
//
// Запрос использует LEFT JOIN для подтягивания связанных данных бренда и категории,
// чтобы вернуть полную информацию о товаре за один SQL-запрос.
//
// Параметры:
//   - ctx (context.Context): контекст запроса (отменяет SQL-запрос, если клиент отсоединился)
//
// Возвращает:
//   - []entities.CatalogItem: полученный список товаров каталога с брендами и категориями
//   - error: ошибку SQL-запроса или сканирования строк
func (r *itemRepository) Items(ctx context.Context) ([]entities.CatalogItem, error) {
	// SQL-запрос с LEFT JOIN: подтягиваем бренд и категорию к каждому товару.
	// LEFT JOIN используется вместо INNER JOIN, потому что товар может не иметь бренда или категории (NULL).
	query := sqlCatalogItemsQuery

	// Отправляем SQL-запрос с поддержкой контекста (отмена по ctx)
	rows, err := r.db.QueryContext(ctx, query)

	// Если SQL-запрос упал с ошибкой — возвращаем её с оберткой fmt.Errorf
	if err != nil {
		return nil, fmt.Errorf("catalog items query: %w", err)
	}

	// defer rows.Close() гарантирует освобождение ресурсов соединений с БД после окончания чтения
	defer rows.Close()

	// Слайс для накопления результатов
	var items []entities.CatalogItem

	// Проходимся в цикле rows.Next() по всем возвращённым строкам ответа БД
	for rows.Next() {
		// Используем вспомогательную функцию scanCatalogItem для чтения одной строки
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan catalog item error: %w", err)
		}

		// Добавляем прочитанную сущность товара в итоговый массив
		items = append(items, item)
	}

	// Проверяем, не возникло ли ошибок во время итерирования по курсору rows
	return items, rows.Err()
}

// scanCatalogItem — вспомогательная функция для чтения одной строки результата SQL-запроса.
// Извлекает данные товара, бренда и категории из текущей строки курсора *sql.Rows.
//
// Бренд и категория могут быть NULL в БД (LEFT JOIN), поэтому используются указатели
// (*uuid.UUID, *string) — если значение NULL, указатель будет nil.
//
// Параметры:
//   - rows (*sql.Rows): курсор SQL-результата, установленный на текущую строку
//
// Возвращает:
//   - entities.CatalogItem: заполненную сущность товара с брендом и категорией (если есть)
//   - error: ошибку сканирования, если типы колонок не совпали или данные повреждены
func scanCatalogItem(rows *sql.Rows) (entities.CatalogItem, error) {
	var item entities.CatalogItem

	// Указатели для nullable-полей: бренд и категория могут отсутствовать (LEFT JOIN → NULL)
	var brandID *uuid.UUID
	var brandTitle *string
	var categoryID *uuid.UUID
	var categoryTitle *string

	// Scan считывает колонки текущей строки в поля структуры и переменные-указатели.
	// Порядок аргументов ДОЛЖЕН совпадать с порядком колонок в SELECT-запросе.
	if err := rows.Scan(&item.Id, &item.ShortDescription, &item.FullDescription, &item.Price, &brandID, &brandTitle, &categoryID, &categoryTitle); err != nil {
		return entities.CatalogItem{}, fmt.Errorf("scan catalog item error: %w", err)
	}

	// Если бренд существует (brandID != nil), создаём объект Brand и прикрепляем к товару
	if brandID != nil {
		item.Brand = &entities.Brand{
			BaseEntity: entities.BaseEntity{
				Id:    *brandID,    // Разыменовываем указатель — получаем uuid.UUID
				Title: *brandTitle, // Разыменовываем указатель — получаем string
			},
		}
	}

	// Если категория существует (categoryID != nil), создаём объект Category и прикрепляем к товару
	if categoryID != nil {
		item.Category = &entities.Category{
			BaseEntity: entities.BaseEntity{
				Id:    *categoryID,
				Title: *categoryTitle,
			},
		}
	}

	// Возвращаем полностью собранный товар каталога
	return item, nil
}

func (r *itemRepository) Item(ctx context.Context, id uuid.UUID) (*entities.CatalogItem, error) {
	query := sqlCatalogItemsQuery + "WHERE ci.id = $1"

	rows, err := r.db.QueryContext(ctx, query, id) // rows вместо row
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	item, err := scanCatalogItem(rows)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &item, nil

}

func (r *itemRepository) ItemsByTitle(ctx context.Context, title string) ([]entities.CatalogItem, error) {
	query := sqlCatalogItemsQuery + `WHERE ci.title LIKE '%' || $1 || '%'`

	rows, err := r.db.QueryContext(ctx, query, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entities.CatalogItem

	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *itemRepository) Create(ctx context.Context, item *entities.CatalogItem) (*entities.CatalogItem, error) {
	if item.Id == uuid.Nil {
		item.Id = uuid.New()
	}

	var brandID, categoryID *uuid.UUID

	if item.Brand != nil {
		brandID = &item.Brand.Id

	}

	if item.Category != nil {
		categoryID = &item.Category.Id
	}

	sqlQuery := `
			INSERT INTO
		catalog_items (
			id,
			title,
			short_description,
			full_description,
			image_url,
			brand_id,
			category_id,
			price
		)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8);
	`

	_, err := r.db.ExecContext(ctx, sqlQuery, item.Id, item.Title, item.ShortDescription, item.FullDescription, item.ImageURL, brandID, categoryID, item.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to insert item: %w", err)
	}

	return item, nil
}
