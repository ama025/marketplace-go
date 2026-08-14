// Package domain содержит бизнес-модели корзины покупок.
package domain

import "github.com/google/uuid"

// ShoppingCartItem представляет один товар в корзине покупателя.
type ShoppingCartItem struct {
	ItemId    uuid.UUID `json:"itemId"`    // UUID товара из каталога
	Quantity  int       `json:"quantity"`  // Количество единиц товара
	UnitPrice float64   `json:"unitPrice"` // Цена за одну единицу товара
	ItemTitle *string   `json:"itemTitle"` // Название товара (может быть nil, если ещё не загружено)
	ItemNote  *string   `json:"itemNote"`  // Произвольная заметка покупателя к позиции (необязательна)
}
