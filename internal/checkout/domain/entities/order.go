package entities

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus — статус заказа.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // ожидает подтверждения
	OrderStatusConfirmed OrderStatus = "confirmed" // подтверждён
	OrderStatusCancelled OrderStatus = "cancelled" // отменён
)

// Order — доменная модель заказа.
type Order struct {
	ID          uuid.UUID   // UUID заказа
	AccountName string      // Покупатель
	Status      OrderStatus // Текущий статус
	Items       []OrderItem // Позиции заказа
	TotalPrice  float64     // Итоговая сумма
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
