package handlers // Слой API Handlers (Контроллеры)

import (
	"net/http"

	"marketplace/internal/catalog/application/queries"
	"marketplace/internal/catalog/domain/spec" // Параметры пагинации

	"github.com/gin-gonic/gin"
)

// CategoriesHandler — структура HTTP-контроллера для работы с категориями.
type CategoriesHandler struct {
	categories *queries.CategoriesHandler
}

// NewCategoriesHandler — конструктор для создания экземпляра CategoriesHandler.
func NewCategoriesHandler(categories *queries.CategoriesHandler) *CategoriesHandler {
	return &CategoriesHandler{categories: categories}
}

// Categories — метод-обработчик (HTTP Endpoint) для получения списка категорий с пагинацией.
// Вызывается Gin при запросе GET /api/v1/categories.
//
// Поддерживаемые query-параметры:
//   - pageIndex (int) — номер страницы (начиная с 1)
//   - pageSize  (int) — размер страницы (максимум ограничен MaxPageSize)
func (h *CategoriesHandler) Categories(c *gin.Context) {
	// ShouldBindQuery считывает query string (?pageIndex=1&pageSize=10)
	// и заполняет поля структуры QueryArgs согласно тегам `form:"..."`
	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Вызываем бизнес-сценарий с пагинацией — он нормализует параметры и делегирует запрос репозиторию
	result, err := h.categories.HandlePaged(c.Request.Context(), args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем страничный ответ: items + метаданные пагинации (pageIndex, pageSize, totalCount)
	c.JSON(http.StatusOK, result)
}
