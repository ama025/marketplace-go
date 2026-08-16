package handlers

import (
	"net/http"

	"marketplace/internal/checkout/application/commands"
	"marketplace/internal/checkout/application/queries"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderHandler struct {
	createOrder     *commands.CreateOrderHandler
	getByID         *queries.GetOrderByIDHandler
	getByAccount    *queries.GetOrdersByAccountHandler
}

func NewOrderHandler(
	createOrder *commands.CreateOrderHandler,
	getByID *queries.GetOrderByIDHandler,
	getByAccount *queries.GetOrdersByAccountHandler,
) *OrderHandler {
	return &OrderHandler{
		createOrder:  createOrder,
		getByID:      getByID,
		getByAccount: getByAccount,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var cmd commands.CreateOrderCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.createOrder.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := h.getByID.Handle(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrdersByAccount(c *gin.Context) {
	account := c.Query("account")
	if account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account query param is required"})
		return
	}

	orders, err := h.getByAccount.Handle(c.Request.Context(), account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}
