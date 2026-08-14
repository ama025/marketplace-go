package domain

// ShoppingCart представляет корзину покупок конкретного пользователя.
type ShoppingCart struct {
	AccountName string             `json:"accountName"` // Уникальное имя (логин) владельца корзины
	Items       []ShoppingCartItem `json:"items"`       // Список позиций (товаров) в корзине
}

// TotalPrice возвращает суммарную стоимость всех товаров в корзине.
// Стоимость каждой позиции = Quantity * UnitPrice.
func (c *ShoppingCart) TotalPrice() float64 {
	var total float64

	for _, item := range c.Items {
		// Прибавляем стоимость текущей позиции к общей сумме
		total += float64(item.Quantity) * item.UnitPrice
	}

	return total
}