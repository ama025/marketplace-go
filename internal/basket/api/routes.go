// Package api отвечает за регистрацию HTTP-маршрутов корзины покупок.
package api

import (
	"marketplace/internal/basket/api/handlers" // HTTP-хендлеры корзины

	"github.com/gin-gonic/gin" // HTTP-фреймворк Gin
)

// RegisterRoutes регистрирует все HTTP-маршруты basket-сервиса на переданном роутере.
//
// Параметры:
//   - r (*gin.Engine): экземпляр роутера Gin, созданный в main.go
//   - cart (*handlers.CartHandler): обработчик корзины покупок
func RegisterRoutes(r *gin.Engine, cart *handlers.CartHandler) {
	// Группа маршрутов с общим префиксом /api/v1
	v1 := r.Group("/api/v1")

	// POST /api/v1/cart — сохранение (создание или обновление) корзины покупок
	v1.POST("/cart", cart.SaveCart)

	// GET /api/v1/cart/:accountName — получение корзины покупок по имени аккаунта
	v1.GET("/cart/:accountName", cart.GetCart)

	// DELETE /api/v1/cart/:accountName — удаление корзины (после оформления заказа)
	v1.DELETE("/cart/:accountName", cart.DeleteCart)

	// POST /api/v1/cart/:accountName/checkout — оформить заказ:
	// публикует OrderConfirmed в RabbitMQ → Checkout создаёт Order → корзина очищается
	v1.POST("/cart/:accountName/checkout", cart.CheckoutCart)
}
