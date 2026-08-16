package queries

import (
	"context"

	"marketplace/internal/checkout/domain/entities"
	"marketplace/internal/checkout/domain/repositories"
)

type GetOrdersByAccountHandler struct {
	repo repositories.OrderRepository
}

func NewGetOrdersByAccountHandler(repo repositories.OrderRepository) *GetOrdersByAccountHandler {
	return &GetOrdersByAccountHandler{repo: repo}
}

func (h *GetOrdersByAccountHandler) Handle(ctx context.Context, accountName string) ([]entities.Order, error) {
	return h.repo.GetByAccount(ctx, accountName)
}
