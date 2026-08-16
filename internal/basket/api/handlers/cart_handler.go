

package handlers

import (
	"net/http"
	"time"

	"marketplace/internal/basket/domain"
	"marketplace/internal/basket/domain/repositories"
	"marketplace/internal/shared/messaging"

	"github.com/gin-gonic/gin"
)

type saveCartRequest struct {
	AccountName string                `json:"accountName" binding:"required"`
	Items       []saveCartItemRequest `json:"items" binding:"required"`
}

type saveCartItemRequest struct {
	ItemId    string  `json:"itemId" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitPrice float64 `json:"unitPrice" binding:"required,min=0"`
	ItemTitle *string `json:"itemTitle"`
	ItemNote  *string `json:"itemNote"`
}

type CartHandler struct {
	repo      repositories.ShoppingCartRepository
	publisher *messaging.Publisher
}

func NewCartHandler(repo repositories.ShoppingCartRepository) *CartHandler {
	return &CartHandler{repo: repo}
}

func NewCartHandlerWithPublisher(repo repositories.ShoppingCartRepository, publisher *messaging.Publisher) *CartHandler {
	return &CartHandler{repo: repo, publisher: publisher}
}

func (h *CartHandler) SaveCart(c *gin.Context) {

	var req saveCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart := domain.ShoppingCart{
		AccountName: req.AccountName,
		Items:       make([]domain.ShoppingCartItem, 0, len(req.Items)),
	}

	for _, reqItem := range req.Items {

		itemID, err := parseUUID(reqItem.ItemId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid itemId: " + reqItem.ItemId,
			})
			return
		}

		cart.Items = append(cart.Items, domain.ShoppingCartItem{
			ItemId:    itemID,
			Quantity:  reqItem.Quantity,
			UnitPrice: reqItem.UnitPrice,
			ItemTitle: reqItem.ItemTitle,
			ItemNote:  reqItem.ItemNote,
		})
	}

	if err := h.repo.Save(c.Request.Context(), &cart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CartHandler) GetCart(c *gin.Context) {

	accountName := c.Param("accountName")

	cart, err := h.repo.Get(c.Request.Context(), accountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if cart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart not found for account: " + accountName})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accountName": cart.AccountName,
		"items":       cart.Items,
		"totalPrice":  cart.TotalPrice(),
	})
}

func (h *CartHandler) DeleteCart(c *gin.Context) {
	accountName := c.Param("accountName")

	if err := h.repo.Delete(c.Request.Context(), accountName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CartHandler) CheckoutCart(c *gin.Context) {
	accountName := c.Param("accountName")

	cart, err := h.repo.Get(c.Request.Context(), accountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cart == nil || len(cart.Items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart is empty or not found"})
		return
	}

	if h.publisher != nil {
		event := messaging.OrderConfirmedEvent{
			AccountName: accountName,
			ConfirmedAt: time.Now(),
			Items:       make([]messaging.OrderConfirmedItem, 0, len(cart.Items)),
		}

		for _, item := range cart.Items {
			title := ""
			if item.ItemTitle != nil {
				title = *item.ItemTitle
			}
			event.Items = append(event.Items, messaging.OrderConfirmedItem{
				ItemID:    item.ItemId.String(),
				ItemTitle: title,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Discount:  0,
			})
		}

		if err := h.publisher.PublishOrderConfirmed(c.Request.Context(), event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish order: " + err.Error()})
			return
		}
	}

	if err := h.repo.Delete(c.Request.Context(), accountName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":     "order accepted, processing asynchronously",
		"accountName": accountName,
	})
}
