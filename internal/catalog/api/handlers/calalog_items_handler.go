package handlers // Слой API (Handlers) — принимает HTTP-запросы и возвращает HTTP-ответы

import (
	// Контекст для передачи в нижележащие слои
	"marketplace/internal/catalog/application/commands"
	"marketplace/internal/catalog/application/queries" // Импорт сценариев (Use Cases) из слоя Application
	"marketplace/internal/catalog/domain/spec"         // Параметры фильтрации, сортировки и пагинации
	"net/http"                                         // Стандартный пакет HTTP для кодов состояния и типов ответов

	"github.com/gin-gonic/gin" // HTTP-фреймворк Gin, используемый для маршрутизации
	"github.com/google/uuid"
)


// CatalogItemsHandler — HTTP-обработчик для получения списка товаров каталога.
// В соответствии с принципами Чистой Архитектуры, он зависит ТОЛЬКО от абстрактного
// интерфейса (queries.CatalogItemsHandler), а не от конкретной реализации PostgreSQL или Gin.
// Фактическая бизнес-логика находится в пакете application/queries.
type CatalogItemsHandler struct {
	// scenarios.CatalogItemsHandler - это интерфейс из слоя Application,
	// который определяет, какие бизнес-сценарии доступны для этого хендлера.
	// Здесь используется внедрение зависимости (Dependency Injection).
	catalogItemsHandler *queries.CatalogItemsHandler
	catalogItemByID     *queries.CatalogItemByIDHandler
	catalogItemByTitle  *queries.CatalogItemByTitleHandler
	catalogItemByBrand  *queries.CatalogItemByBrandHandler // Обработчик поиска товаров по бренду

	createCatalogItem *commands.CreateCatalogItemHandler
	updateCatalogItem *commands.UpdateCatalogItemHandler
	deleteCatalogItem *commands.DeleteCatalogItemHandler
}

// NewCatalogItemsHandler — конструктор для создания HTTP-обработчика получения товаров каталога.
// Принимает любую структуру, реализующую интерфейс queries.CatalogItemsHandler.
//
// Параметры:
//   - handler: интерфейс сценария получения товаров каталога (например, *queries.CatalogItemsHandler)
func NewCatalogItemsHandler(catalogItems *queries.CatalogItemsHandler, catalogItemByID *queries.CatalogItemByIDHandler, catalogItemByTitle *queries.CatalogItemByTitleHandler, catalogItemByBrand *queries.CatalogItemByBrandHandler, createItem *commands.CreateCatalogItemHandler, updateItem *commands.UpdateCatalogItemHandler, deleteCatalogItem *commands.DeleteCatalogItemHandler) *CatalogItemsHandler {
	return &CatalogItemsHandler{catalogItemsHandler: catalogItems, catalogItemByID: catalogItemByID, catalogItemByTitle: catalogItemByTitle, catalogItemByBrand: catalogItemByBrand, createCatalogItem: createItem, updateCatalogItem: updateItem, deleteCatalogItem: deleteCatalogItem}
}

// ListCatalogItems — обработчик HTTP GET-запроса для получения списка товаров каталога.
// Реагирует на GET-запросы по адресу:
//   - /api/v1/catalog-items
//
// Поддерживаемые query-параметры:
//   - pageIndex (int)    — номер страницы (начиная с 1)
//   - pageSize  (int)    — размер страницы (максимум ограничен MaxPageSize)
//   - brandId   (string) — UUID бренда для фильтрации
//   - categoryId(string) — UUID категории для фильтрации
//   - search    (string) — текстовый поиск по названию и краткому описанию
//   - sort      (string) — сортировка: price_asc | price_desc | title_asc | title_desc
func (h *CatalogItemsHandler) ListCatalogItems(c *gin.Context) {
	// ShouldBindQuery считывает query string (?brandId=...&sort=price_asc&pageIndex=1…)
	// и заполняет поля структуры QueryArgs согласно тегам `form:"..."`
	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Вызываем бизнес-сценарий — он нормализует пагинацию и делегирует запрос репозиторию
	result, err := h.catalogItemsHandler.Handle(c.Request.Context(), args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем страничный ответ: items + метаданные пагинации (pageIndex, pageSize, totalCount)
	c.JSON(http.StatusOK, result)
}


func (h *CatalogItemsHandler) CatalogItemByID(c *gin.Context) {
	// Получаем ID из параметров запроса
	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.catalogItemByID.Handle(c.Request.Context(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})

}

func (h *CatalogItemsHandler) CatalogItemByTitle(c *gin.Context) {
	title := c.Param("title")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	// ShouldBindQuery считывает query string (?pageIndex=1&pageSize=10)
	// и заполняет поля структуры QueryArgs согласно тегам `form:"..."`
	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Вызываем бизнес-сценарий с пагинацией — он нормализует параметры и делегирует запрос репозиторию
	result, err := h.catalogItemByTitle.HandlePaged(c.Request.Context(), title, args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем страничный ответ: items + метаданные пагинации (pageIndex, pageSize, totalCount)
	c.JSON(http.StatusOK, result)
}

// CatalogItemByBrand — обработчик HTTP GET-запроса для поиска товаров по названию бренда с пагинацией.
// Реагирует на GET-запросы по адресу:
//   - /api/v1/catalog-items/brand/:brand
//
// Поддерживаемые query-параметры:
//   - pageIndex (int) — номер страницы (начиная с 1)
//   - pageSize  (int) — размер страницы (максимум ограничен MaxPageSize)
//
// Параметры:
//   - c: *gin.Context — контекст HTTP-запроса/ответа Gin
func (h *CatalogItemsHandler) CatalogItemByBrand(c *gin.Context) {
	// Извлекаем параметр "brand" из URL-пути (например, /brand/Nike → "Nike")
	brand := c.Param("brand")

	// Валидируем, что параметр не пустой
	if brand == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "brand is required"})
		return
	}

	// ShouldBindQuery считывает query string (?pageIndex=1&pageSize=10)
	// и заполняет поля структуры QueryArgs согласно тегам `form:"..."`
	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Вызываем бизнес-сценарий с пагинацией — он нормализует параметры и делегирует запрос репозиторию
	result, err := h.catalogItemByBrand.HandlePaged(c.Request.Context(), brand, args)
	if err != nil {
		// Если произошла ошибка — возвращаем 500 Internal Server Error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем страничный ответ: items + метаданные пагинации (pageIndex, pageSize, totalCount)
	c.JSON(http.StatusOK, result)
}

func (h *CatalogItemsHandler) CreateCatalogItem(c *gin.Context) {
	var cmd commands.CreateCatalogItemCommand

	if err := c.ShouldBindJSON(&cmd); err != nil { //shouldbindjson это функция которая принимает http запрос и пытается его преобразовать в структуру commands.CreateCatalogItemCommand если приходит
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.createCatalogItem.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})

}

func (h *CatalogItemsHandler) UpdateCatalogItem(c *gin.Context) {
	var cmd commands.UpdateCatalogItemCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isSuccess, err := h.updateCatalogItem.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isSuccess": isSuccess})
}

func (h *CatalogItemsHandler) DeleteCatalogItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}

	isSuccess, err := h.deleteCatalogItem.Handle(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isSuccess": isSuccess})
}
