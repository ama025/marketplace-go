package api

import (
	"marketplace/internal/checkout/api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes регистрирует все маршруты checkout-сервиса.
func RegisterRoutes(r *gin.Engine, orderHandler *handlers.OrderHandler) {
	v1 := r.Group("/api/v1")
	{
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)          // POST /api/v1/orders
			orders.GET("/:id", orderHandler.GetOrderByID)     // GET  /api/v1/orders/:id
			orders.GET("", orderHandler.GetOrdersByAccount)   // GET  /api/v1/orders?account=...
		}
	}
}
