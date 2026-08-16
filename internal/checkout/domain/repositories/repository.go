package repositories

import (
	"context"

	"marketplace/internal/checkout/domain/entities"

	"github.com/google/uuid"
)

type OrderRepository interface {

	Create(ctx context.Context, order entities.Order) (entities.Order, error)

	GetByID(ctx context.Context, id uuid.UUID) (*entities.Order, error)

	GetByAccount(ctx context.Context, accountName string) ([]entities.Order, error)

	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.OrderStatus) error
}
