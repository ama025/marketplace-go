package entities

import "github.com/google/uuid"

type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ItemID    uuid.UUID
	ItemTitle string
	Quantity  int
	UnitPrice float64
	Discount  float64
}
