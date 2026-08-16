

package repositories

import (
	"context"

	"marketplace/internal/basket/domain"
)

type ShoppingCartRepository interface {

	Save(ctx context.Context, cart *domain.ShoppingCart) error

	Get(ctx context.Context, accountName string) (*domain.ShoppingCart, error)

	Delete(ctx context.Context, accountName string) error
}
