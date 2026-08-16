package commands

import (
	"context"

	"marketplace/internal/checkout/domain/entities"
	"marketplace/internal/checkout/domain/repositories"

	"github.com/google/uuid"
)

type CreateOrderCommand struct {
	AccountName string
	Items       []OrderItemInput
}

type OrderItemInput struct {
	ItemID    uuid.UUID
	ItemTitle string
	Quantity  int
	UnitPrice float64
	Discount  float64
}

type CreateOrderHandler struct {
	repo repositories.OrderRepository
}

func NewCreateOrderHandler(repo repositories.OrderRepository) *CreateOrderHandler {
	return &CreateOrderHandler{repo: repo}
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) (entities.Order, error) {
	items := make([]entities.OrderItem, 0, len(cmd.Items))
	var total float64

	for _, i := range cmd.Items {
		price := i.UnitPrice * float64(i.Quantity) * (1 - i.Discount/100)
		total += price

		items = append(items, entities.OrderItem{
			ItemID:    i.ItemID,
			ItemTitle: i.ItemTitle,
			Quantity:  i.Quantity,
			UnitPrice: i.UnitPrice,
			Discount:  i.Discount,
		})
	}

	order := entities.Order{
		AccountName: cmd.AccountName,
		Status:      entities.OrderStatusPending,
		Items:       items,
		TotalPrice:  total,
	}

	return h.repo.Create(ctx, order)
}
