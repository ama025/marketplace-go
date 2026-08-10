package handlers // Слой API (Handlers) — принимает HTTP-запросы и возвращает HTTP-ответы

import (
	// Контекст для передачи в нижележащие слои
	"marketplace/internal/catalog/application/commands"
	"marketplace/internal/catalog/application/queries" // Импорт сценариев (Use Cases) из слоя Application
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

	createCatalogItem *commands.CreateCatalogItemHandler
}

// NewCatalogItemsHandler — конструктор для создания HTTP-обработчика получения товаров каталога.
// Принимает любую структуру, реализующую интерфейс queries.CatalogItemsHandler.
//
// Параметры:
//   - handler: интерфейс сценария получения товаров каталога (например, *queries.CatalogItemsHandler)
func NewCatalogItemsHandler(catalogItems *queries.CatalogItemsHandler, catalogItemByID *queries.CatalogItemByIDHandler, catalogItemByTitle *queries.CatalogItemByTitleHandler, createItem *commands.CreateCatalogItemHandler) *CatalogItemsHandler {
	return &CatalogItemsHandler{catalogItemsHandler: catalogItems, catalogItemByID: catalogItemByID, catalogItemByTitle: catalogItemByTitle, createCatalogItem: createItem}
}

// ListCatalogItems — обработчик HTTP GET-запроса для получения списка товаров каталога.
// Реагирует на GET-запросы по адресу:
//   - /v1/catalog/items
//
// Метод использует внедренный сценарий catalogItemsHandler для выполнения бизнес-логики.
// Он не содержит SQL-запросов или прямого доступа к базе данных — всё делегируется в слой Application.
//
// Параметры:
//   - c: *gin.Context — контекст HTTP-запроса/ответа Gin
func (h *CatalogItemsHandler) ListCatalogItems(c *gin.Context) {
	// Вызываем бизнес-сценарий (Use Case) для получения товаров каталога.
	// Используется контекст HTTP-запроса (c.Request.Context()), который содержит
	// информацию о таймауте и отмене запроса.
	items, err := h.catalogItemsHandler.Handle(c.Request.Context())

	// Обработка возможной ошибки выполнения бизнес-сценария
	if err != nil {
		// Отправляем HTTP 500 (Internal Server Error) с JSON-сообщением об ошибке.
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Если ошибок нет, отправляем HTTP 200 (OK) с JSON-ответом, содержащим массив товаров.
	c.JSON(http.StatusOK, gin.H{"data": items})
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

	items, err := h.catalogItemByTitle.Handle(c.Request.Context(), title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if items == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
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
