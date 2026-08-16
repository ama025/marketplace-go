package domain

type ShoppingCart struct {
	AccountName string             `json:"accountName"`
	Items       []ShoppingCartItem `json:"items"`
}

func (c *ShoppingCart) TotalPrice() float64 {
	var total float64

	for _, item := range c.Items {

		total += float64(item.Quantity) * item.UnitPrice
	}

	return total
}