package handlers

import (
	"net/http"

	"marketplace/internal/catalog/application/queries"
	"marketplace/internal/catalog/domain/spec"

	"github.com/gin-gonic/gin"
)

type BrandsHandler struct {
	brands *queries.BrandsHandler
}

func NewBrandsHandler(brands *queries.BrandsHandler) *BrandsHandler {
	return &BrandsHandler{brands: brands}
}

func (h *BrandsHandler) Brands(c *gin.Context) {

	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.brands.HandlePaged(c.Request.Context(), args)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
