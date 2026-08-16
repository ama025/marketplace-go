package handlers

import (

	"marketplace/internal/catalog/application/commands"
	"marketplace/internal/catalog/application/queries"
	"marketplace/internal/catalog/domain/spec"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CatalogItemsHandler struct {

	catalogItemsHandler *queries.CatalogItemsHandler
	catalogItemByID     *queries.CatalogItemByIDHandler
	catalogItemByTitle  *queries.CatalogItemByTitleHandler
	catalogItemByBrand  *queries.CatalogItemByBrandHandler

	createCatalogItem *commands.CreateCatalogItemHandler
	updateCatalogItem *commands.UpdateCatalogItemHandler
	deleteCatalogItem *commands.DeleteCatalogItemHandler
}

func NewCatalogItemsHandler(catalogItems *queries.CatalogItemsHandler, catalogItemByID *queries.CatalogItemByIDHandler, catalogItemByTitle *queries.CatalogItemByTitleHandler, catalogItemByBrand *queries.CatalogItemByBrandHandler, createItem *commands.CreateCatalogItemHandler, updateItem *commands.UpdateCatalogItemHandler, deleteCatalogItem *commands.DeleteCatalogItemHandler) *CatalogItemsHandler {
	return &CatalogItemsHandler{catalogItemsHandler: catalogItems, catalogItemByID: catalogItemByID, catalogItemByTitle: catalogItemByTitle, catalogItemByBrand: catalogItemByBrand, createCatalogItem: createItem, updateCatalogItem: updateItem, deleteCatalogItem: deleteCatalogItem}
}

func (h *CatalogItemsHandler) ListCatalogItems(c *gin.Context) {

	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.catalogItemsHandler.Handle(c.Request.Context(), args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CatalogItemsHandler) CatalogItemByID(c *gin.Context) {

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

	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.catalogItemByTitle.HandlePaged(c.Request.Context(), title, args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CatalogItemsHandler) CatalogItemByBrand(c *gin.Context) {

	brand := c.Param("brand")

	if brand == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "brand is required"})
		return
	}

	var args spec.QueryArgs
	if err := c.ShouldBindQuery(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.catalogItemByBrand.HandlePaged(c.Request.Context(), brand, args)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CatalogItemsHandler) CreateCatalogItem(c *gin.Context) {
	var cmd commands.CreateCatalogItemCommand

	if err := c.ShouldBindJSON(&cmd); err != nil {
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
