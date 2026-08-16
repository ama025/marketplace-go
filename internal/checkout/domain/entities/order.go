package entities

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          uuid.UUID
	AccountName string
	Status      OrderStatus
	Items       []OrderItem
	TotalPrice  float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
