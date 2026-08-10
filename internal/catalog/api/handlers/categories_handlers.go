package handlers // Слой API Handlers (Контроллеры)

import (
	"net/http"

	"marketplace/internal/catalog/application/queries"

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

// Categories — метод-обработчик (HTTP Endpoint) для получения списка всех категорий.
// Вызывается Gin при запросе GET /api/v1/categories.
func (h *CategoriesHandler) Categories(c *gin.Context) {
	categories, err := h.categories.Handle(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}
