package queries

import (
	"context"

	"marketplace/internal/checkout/domain/entities"
	"marketplace/internal/checkout/domain/repositories"

	"github.com/google/uuid"
)

// GetOrderByIDHandler — use case «Получить заказ по ID».
type GetOrderByIDHandler struct {
	repo repositories.OrderRepository
}

func NewGetOrderByIDHandler(repo repositories.OrderRepository) *GetOrderByIDHandler {
	return &GetOrderByIDHandler{repo: repo}
}

func (h *GetOrderByIDHandler) Handle(ctx context.Context, id uuid.UUID) (*entities.Order, error) {
	return h.repo.GetByID(ctx, id)
}
