package repositories

import (
	"context"

	"marketplace/internal/checkout/domain/entities"

	"github.com/google/uuid"
)

// OrderRepository — контракт для работы с заказами.
// Application-слой зависит только от этого интерфейса.
type OrderRepository interface {
	// Create создаёт новый заказ и возвращает его с заполненным ID.
	Create(ctx context.Context, order entities.Order) (entities.Order, error)

	// GetByID возвращает заказ по UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Order, error)

	// GetByAccount возвращает все заказы покупателя.
	GetByAccount(ctx context.Context, accountName string) ([]entities.Order, error)

	// UpdateStatus меняет статус заказа.
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.OrderStatus) error
}
