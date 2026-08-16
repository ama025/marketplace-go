package api

import (
	"marketplace/internal/catalog/api/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	brands *handlers.BrandsHandler,
	categories *handlers.CategoriesHandler,
	items *handlers.CatalogItemsHandler,
) {

	v1 := r.Group("/api/v1")

	v1.GET("/brands", brands.Brands)
	v1.GET("/categories", categories.Categories)
	v1.GET("/catalog-items", items.ListCatalogItems)
	v1.GET("/catalog-items/:id", items.CatalogItemByID)
	v1.GET("/catalog-items/title/:title", items.CatalogItemByTitle)
	v1.GET("/catalog-items/brand/:brand", items.CatalogItemByBrand)

	v1.POST("/catalog-items", items.CreateCatalogItem)

	v1.PUT("/catalog-items", items.UpdateCatalogItem)

	v1.DELETE("/catalog-items/:id", items.DeleteCatalogItem)
}
