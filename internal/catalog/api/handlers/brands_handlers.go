package handlers // Слой API Handlers (Контроллеры): отвечает за прием HTTP-запросов и отправку ответов клиенту

import (
	"net/http" // Стандартный пакет Go с константами статусов HTTP (http.StatusOK, http.StatusInternalServerError)

	"marketplace/internal/catalog/application/queries" // Импортируем слой бизнес-логики (Application Queries)
	"marketplace/internal/catalog/domain/spec"         // Параметры пагинации

	"github.com/gin-gonic/gin" // Фреймворк Gin
)

// BrandsHandler — структура HTTP-контроллера для работы с брендами.
// Хранит в себе ссылку на объект бизнес-логики (queries.BrandsHandler).
type BrandsHandler struct {
	brands *queries.BrandsHandler
}

// NewBrandsHandler — конструктор для создания экземпляра BrandsHandler.
// Внедряет зависимость (Dependency Injection) слоя приложений.
func NewBrandsHandler(brands *queries.BrandsHandler) *BrandsHandler {
	return &BrandsHandler{brands: brands}
}

// Brands — метод-обработчик (HTTP Endpoint) для получения списка брендов с пагинацией.
// Вызывается автоматически веб-фреймворком Gin при запросе GET /api/v1/brands.
//
// Поддерживаемые query-параметры:
//   - pageIndex (int) — номер страницы (начиная с 1)
//   - pageSize  (int) — размер страницы (максимум ограничен MaxPageSize)
//
// Параметры:
//   - c (*gin.Context): контекст запроса Gin, содержит заголовки, параметры и методы для отправки ответа (c.JSON).
func (h *BrandsHandler) Brands(c *gin.Context) {
	// ShouldBindQuery считывает query string (?pageIndex=1&pageSize=10)
	// и заполняет поля структуры QueryArgs согласно тегам `form:"..."`
	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Вызываем бизнес-сценарий с пагинацией — он нормализует параметры и делегирует запрос репозиторию
	result, err := h.brands.HandlePaged(c.Request.Context(), args)
	if err != nil {
		// Ошибка 500 (Internal Server Error): возвращаем клиенту JSON с текстом ошибки {"error": "..."}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // Прерываем дальше выполнение метода
	}

	// Возвращаем страничный ответ: items + метаданные пагинации (pageIndex, pageSize, totalCount)
	c.JSON(http.StatusOK, result)
}
