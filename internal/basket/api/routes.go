
package api

import (
	"marketplace/internal/basket/api/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cart *handlers.CartHandler) {

	v1 := r.Group("/api/v1")

	v1.POST("/cart", cart.SaveCart)

	v1.GET("/cart/:accountName", cart.GetCart)

	v1.DELETE("/cart/:accountName", cart.DeleteCart)

	v1.POST("/cart/:accountName/checkout", cart.CheckoutCart)
}
