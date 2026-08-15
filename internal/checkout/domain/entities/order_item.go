package entities

import "github.com/google/uuid"

// OrderItem — позиция заказа (один товар из корзины).
type OrderItem struct {
	ID        uuid.UUID // UUID позиции
	OrderID   uuid.UUID // Принадлежит заказу
	ItemID    uuid.UUID // UUID товара из каталога
	ItemTitle string    // Название товара на момент заказа
	Quantity  int       // Количество
	UnitPrice float64   // Цена за единицу
	Discount  float64   // % скидки на момент заказа (0 если нет)
}
