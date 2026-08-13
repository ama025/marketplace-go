package persistence // Слой Infrastructure (Persistence) — конкретная реализация взаимодействия с базой данных PostgreSQL через SQL

import (
	"context"      // Стандартный контекст Go для управления временем выполнения операций
	"database/sql" // Пакет для отправки запросов в реляционную БД (*sql.DB)
	"fmt"          // Пакет для форматирования ошибок (fmt.Errorf)
	"strings"      // Пакет для работы со строками (построение динамического WHERE-условия)

	"marketplace/internal/catalog/domain/entities" // Импортируем доменную модель CatalogItem
	"marketplace/internal/catalog/domain/spec"     // Параметры фильтрации, сортировки и пагинации

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


func (r *itemRepository) Update(ctx context.Context,item *entities.CatalogItem) (bool, error) {
	var brandID,categoryID *uuid.UUID

	if item.Brand != nil {
		brandID = &item.Brand.Id
	}

	if item.Category != nil {
		categoryID = &item.Category.Id
	}

	sqlQuery := `
		UPDATE
		catalog_items
		SET
			title = $2,
			short_description = $3,
			full_description = $4,
			image_url = $5,
			brand_id = $6,
			category_id = $7,
			price = $8
		WHERE
			id = $1;
	`

	result,err := r.db.ExecContext(ctx,sqlQuery,item.Id,item.Title,item.ShortDescription,item.FullDescription,item.ImageURL,brandID,categoryID,item.Price) //Метод ExecContext в Go служит для выполнения SQL-запросов, которые изменяют данные или структуру базы, но не возвращают строки с данными 
	if err != nil {
		return false, fmt.Errorf("failed to update item: %w", err)
	}

	n,_ := result.RowsAffected()
	return n>0, nil

	
}


// Delete удаляет товар из каталога по его уникальному идентификатору (UUID).
// Возвращает true, если товар был успешно удален, и false, если товара с таким ID не существовало.
func (r *itemRepository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	// Подготавливаем SQL-запрос. Использование плейсхолдера $1 защищает от SQL-инъекций.
	sqlDeleteItem := "DELETE FROM catalog_items WHERE id = $1"
	
	// Выполняем запрос в контексте. ExecContext используется для операций, 
	// которые не возвращают строки данных (INSERT, UPDATE, DELETE).
	result, err := r.db.ExecContext(
		ctx,           // Передаем контекст для контроля таймаутов и отмены запроса
		sqlDeleteItem, // Сам SQL-запрос
		id,            // Значение, которое подставится вместо $1
	)
	// Если во время общения с базой данных произошла системная ошибка (например, пропала связь)
	if err != nil {
		// Возвращаем false и оборачиваем ошибку (%w), добавляя контекст (какой ID не удалось удалить)
		return false, fmt.Errorf("delete item[%v]: %w", id, err)
	}
	
	// Получаем количество строк, которые физически были затронуты (удалены) этим запросом.
	// Ошибку игнорируем (_), так как для DELETE в большинстве драйверов (например, pgx) она всегда nil.
	n, _ := result.RowsAffected() 
	
	// Если n > 0, значит строка существовала и была удалена (функция вернет true).
	// Если n == 0, значит в базе и так не было товара с таким ID (функция вернет false).
	return n > 0, nil
}


// ItemsByBrand — ищет товары каталога по названию бренда (частичное совпадение).
// Использует SQL-оператор LIKE с подстановочными символами '%' для поиска подстроки.
//
// Параметры:
//   - ctx: контекст запроса для управления таймаутами и отменой
//   - brandTitle: строка для поиска в названии бренда (например, "Nike" найдет "Nike", "Nike Air" и т.д.)
//
// Возвращает:
//   - []entities.CatalogItem: список товаров, бренд которых содержит указанную подстроку
//   - error: ошибку SQL-запроса или сканирования строк
func (r *itemRepository) ItemsByBrand(ctx context.Context, brandTitle string) ([]entities.CatalogItem, error) {
	// Формируем SQL-запрос: базовый SELECT с JOIN + фильтрация по названию бренда.
	// Оператор LIKE с '%' по обе стороны ищет подстроку в любом месте названия бренда.
	// $1 — плейсхолдер PostgreSQL, защищающий от SQL-инъекций.
	query := sqlCatalogItemsQuery + `WHERE b.title LIKE '%' || $1 || '%'`

	// Выполняем SQL-запрос с контекстом, передавая название бренда как параметр
	rows, err := r.db.QueryContext(ctx, query, brandTitle)
	if err != nil {
		// Оборачиваем ошибку с контекстом для удобной отладки
		return nil, fmt.Errorf("items by brand query: %w", err)
	}
	// Гарантируем закрытие курсора после завершения чтения
	defer rows.Close()

	// Слайс для накопления найденных товаров
	var items []entities.CatalogItem

	// Итерируем по всем строкам результата
	for rows.Next() {
		// Сканируем текущую строку в структуру CatalogItem
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan catalog item by brand error: %w", err)
		}
		// Добавляем товар в итоговый слайс
		items = append(items, item)
	}

	// Возвращаем результат и возможную ошибку итерации
	return items, rows.Err()
}

// ItemsWithFilter — метод получения товаров каталога с поддержкой:
//   - Фильтрации по brand_id, category_id и поисковой строке (title/short_description)
//   - Сортировки по цене и названию (price_asc, price_desc, title_asc, title_desc)
//   - Пагинации (LIMIT / OFFSET)
//
// Возвращает:
//   - []entities.CatalogItem: страница товаров
//   - int: общее количество товаров, удовлетворяющих фильтру (для расчёта числа страниц)
//   - error: ошибку SQL-запроса или сканирования
func (r *itemRepository) ItemsWithFilter(ctx context.Context, args spec.QueryArgs) ([]entities.CatalogItem, int, error) {
	// --- Динамическое построение WHERE-условия ---
	// conditions — набор SQL-фрагментов вида "ci.brand_id = $N"
	// sqlArgs — список значений-аргументов, соответствующих плейсхолдерам $1, $2, …
	var conditions []string
	var sqlArgs []any

	// Счётчик плейсхолдеров: PostgreSQL использует $1, $2, …, а не ?
	argIdx := 1

	// Фильтр по бренду: если передан brandId — парсим UUID и добавляем условие
	if brandID, err := args.ParseBrandId(); err != nil {
		return nil, 0, fmt.Errorf("invalid brandId: %w", err)
	} else if brandID != nil {
		conditions = append(conditions, fmt.Sprintf("ci.brand_id = $%d", argIdx))
		sqlArgs = append(sqlArgs, *brandID)
		argIdx++
	}

	// Фильтр по категории: аналогично brandId
	if categoryID, err := args.ParseCategoryId(); err != nil {
		return nil, 0, fmt.Errorf("invalid categoryId: %w", err)
	} else if categoryID != nil {
		conditions = append(conditions, fmt.Sprintf("ci.category_id = $%d", argIdx))
		sqlArgs = append(sqlArgs, *categoryID)
		argIdx++
	}

	// Полнотекстовый поиск по названию и краткому описанию товара (ILIKE — без учёта регистра)
	if args.Search != nil && *args.Search != "" {
		pattern := fmt.Sprintf("ci.title ILIKE $%d OR ci.short_description ILIKE $%d", argIdx, argIdx+1)
		conditions = append(conditions, "("+pattern+")")
		searchVal := "%" + *args.Search + "%"
		sqlArgs = append(sqlArgs, searchVal, searchVal)
		argIdx += 2
	}

	// Собираем WHERE-секцию: если есть условия — объединяем через AND
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// --- Определение ORDER BY ---
	// Допустимые значения sort: price_asc, price_desc, title_asc, title_desc
	// По умолчанию сортируем по ci.title ASC
	orderClause := "ORDER BY ci.title ASC"
	if args.Sort != nil {
		switch *args.Sort {
		case "price_asc":
			orderClause = "ORDER BY ci.price ASC"
		case "price_desc":
			orderClause = "ORDER BY ci.price DESC"
		case "title_asc":
			orderClause = "ORDER BY ci.title ASC"
		case "title_desc":
			orderClause = "ORDER BY ci.title DESC"
		}
	}

	// --- Запрос общего количества (COUNT) для пагинации ---
	// Выполняем отдельный COUNT-запрос с теми же WHERE-условиями, но без LIMIT/OFFSET,
	// чтобы UI мог рассчитать общее число страниц.
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM catalog_items ci
		LEFT JOIN brands b ON b.id = ci.brand_id
		LEFT JOIN categories c ON c.id = ci.category_id
		%s
	`, whereClause)

	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, sqlArgs...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count catalog items: %w", err)
	}

	// --- Пагинация: LIMIT и OFFSET ---
	// LIMIT — количество строк на странице
	// OFFSET — сколько строк пропустить (вычисляется из номера страницы)
	limit := args.PageSize
	offset := (args.PageIndex - 1) * args.PageSize

	// Добавляем LIMIT и OFFSET как следующие плейсхолдеры
	limitPlaceholder := fmt.Sprintf("$%d", argIdx)
	offsetPlaceholder := fmt.Sprintf("$%d", argIdx+1)
	sqlArgs = append(sqlArgs, limit, offset)

	// --- Финальный SQL-запрос с фильтром, сортировкой и пагинацией ---
	dataQuery := fmt.Sprintf(`
		%s
		%s
		%s
		LIMIT %s OFFSET %s
	`, sqlCatalogItemsQuery, whereClause, orderClause, limitPlaceholder, offsetPlaceholder)

	rows, err := r.db.QueryContext(ctx, dataQuery, sqlArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("items with filter query: %w", err)
	}
	defer rows.Close()

	// Накапливаем результаты
	var items []entities.CatalogItem
	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan catalog item (filter): %w", err)
		}
		items = append(items, item)
	}

	return items, totalCount, rows.Err()
}

// ItemsByTitlePaged — поиск товаров по названию с поддержкой пагинации.
// Выполняет два запроса: COUNT для общего числа совпадений и SELECT с LIMIT/OFFSET для страницы.
//
// Параметры:
//   - ctx: контекст запроса
//   - title: строка поиска (частичное совпадение через LIKE)
//   - args: параметры пагинации (PageIndex, PageSize)
//
// Возвращает:
//   - []entities.CatalogItem: список товаров на текущей странице
//   - int: общее число найденных товаров
//   - error: ошибку SQL-запроса или сканирования
func (r *itemRepository) ItemsByTitlePaged(ctx context.Context, title string, args spec.QueryArgs) ([]entities.CatalogItem, int, error) {
	// Запрос общего числа совпадений (без LIMIT/OFFSET)
	countQuery := `
		SELECT COUNT(*)
		FROM catalog_items ci
		LEFT JOIN brands b ON b.id = ci.brand_id
		LEFT JOIN categories c ON c.id = ci.category_id
		WHERE ci.title LIKE '%' || $1 || '%'
	`
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, title).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("items by title count: %w", err)
	}

	// Вычисляем смещение (OFFSET) из номера страницы
	offset := (args.PageIndex - 1) * args.PageSize

	// Запрос страницы товаров с LIMIT/OFFSET
	dataQuery := sqlCatalogItemsQuery + `WHERE ci.title LIKE '%' || $1 || '%' ORDER BY ci.title ASC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, dataQuery, title, args.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("items by title paged query: %w", err)
	}
	defer rows.Close()

	var items []entities.CatalogItem
	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan catalog item by title: %w", err)
		}
		items = append(items, item)
	}

	return items, totalCount, rows.Err()
}

// ItemsByBrandPaged — поиск товаров по названию бренда с поддержкой пагинации.
// Выполняет два запроса: COUNT для общего числа совпадений и SELECT с LIMIT/OFFSET для страницы.
//
// Параметры:
//   - ctx: контекст запроса
//   - brandTitle: строка поиска по названию бренда (частичное совпадение через LIKE)
//   - args: параметры пагинации (PageIndex, PageSize)
//
// Возвращает:
//   - []entities.CatalogItem: список товаров на текущей странице
//   - int: общее число найденных товаров
//   - error: ошибку SQL-запроса или сканирования
func (r *itemRepository) ItemsByBrandPaged(ctx context.Context, brandTitle string, args spec.QueryArgs) ([]entities.CatalogItem, int, error) {
	// Запрос общего числа совпадений (без LIMIT/OFFSET)
	countQuery := `
		SELECT COUNT(*)
		FROM catalog_items ci
		LEFT JOIN brands b ON b.id = ci.brand_id
		LEFT JOIN categories c ON c.id = ci.category_id
		WHERE b.title LIKE '%' || $1 || '%'
	`
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, brandTitle).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("items by brand count: %w", err)
	}

	// Вычисляем смещение (OFFSET) из номера страницы
	offset := (args.PageIndex - 1) * args.PageSize

	// Запрос страницы товаров с LIMIT/OFFSET
	dataQuery := sqlCatalogItemsQuery + `WHERE b.title LIKE '%' || $1 || '%' ORDER BY ci.title ASC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, dataQuery, brandTitle, args.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("items by brand paged query: %w", err)
	}
	defer rows.Close()

	var items []entities.CatalogItem
	for rows.Next() {
		item, err := scanCatalogItem(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan catalog item by brand: %w", err)
		}
		items = append(items, item)
	}

	return items, totalCount, rows.Err()
}
