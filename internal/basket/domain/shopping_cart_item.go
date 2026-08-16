
package domain

import "github.com/google/uuid"

type ShoppingCartItem struct {
	ItemId    uuid.UUID `json:"itemId"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unitPrice"`
	ItemTitle *string   `json:"itemTitle"`
	ItemNote  *string   `json:"itemNote"`
}
