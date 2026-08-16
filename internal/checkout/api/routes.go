package api

import (
	"marketplace/internal/checkout/api/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, orderHandler *handlers.OrderHandler) {
	v1 := r.Group("/api/v1")
	{
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrderByID)
			orders.GET("", orderHandler.GetOrdersByAccount)
		}
	}
}
