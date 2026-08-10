package api // Пакет api отвечает за маршрутизацию (routing) HTTP-запросов к соответствующим хендлерам

import (
	"marketplace/internal/catalog/api/handlers" // Импортируем HTTP-хендлеры

	"github.com/gin-gonic/gin" // Фреймворк Gin для регистрации маршрутов
)

// RegisterRoutes связывает HTTP-пути (URLs) с функциями-обработчиками (хендлерами).
//
// Параметры:
//   - r (*gin.Engine): экземпляр роутера Gin, созданный в main.go
//   - brands (*handlers.BrandsHandler): хендлер для работы с брендами
func RegisterRoutes(
	r *gin.Engine,
	brands *handlers.BrandsHandler,
	categories *handlers.CategoriesHandler,
	items *handlers.CatalogItemsHandler,
) {
	// Создаем группу маршрутов с общим префиксом "/api/v1".
	v1 := r.Group("/api/v1")

	// Регистрируем эндпоинты
	v1.GET("/brands", brands.Brands)
	v1.GET("/categories", categories.Categories)
	v1.GET("/catalog-items", items.ListCatalogItems)
	v1.GET("/catalog-items/:id", items.CatalogItemByID)
	v1.GET("/catalog-items/title/:title", items.CatalogItemByTitle)
	
	v1.POST("/catalog-items", items.CreateCatalogItem)
}

