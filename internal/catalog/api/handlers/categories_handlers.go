package handlers

import (
	"net/http"

	"marketplace/internal/catalog/application/queries"
	"marketplace/internal/catalog/domain/spec"

	"github.com/gin-gonic/gin"
)

type CategoriesHandler struct {
	categories *queries.CategoriesHandler
}

func NewCategoriesHandler(categories *queries.CategoriesHandler) *CategoriesHandler {
	return &CategoriesHandler{categories: categories}
}

func (h *CategoriesHandler) Categories(c *gin.Context) {

	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.categories.HandlePaged(c.Request.Context(), args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
