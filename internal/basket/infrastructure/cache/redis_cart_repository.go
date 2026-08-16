

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"marketplace/internal/basket/domain"
	"marketplace/internal/shared"

	"github.com/redis/go-redis/v9"
)

const (

	cartKeyPrefix = "cart:"

	defaultTTL = 7 * 24 * time.Hour
)

type redisCartRepository struct {
	client *redis.Client
}

func NewRedisCartRepository(client *redis.Client) *redisCartRepository {
	return &redisCartRepository{client: client}
}

func (r *redisCartRepository) Save(ctx context.Context, cart *domain.ShoppingCart) error {

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("redis cart repository: marshal cart: %w", err)
	}

	key := cartKey(cart.AccountName)
	if err := r.client.Set(ctx, key, data, defaultTTL).Err(); err != nil {
		return fmt.Errorf("redis cart repository: set key %q: %w", key, err)
	}

	return nil
}

func (r *redisCartRepository) Get(ctx context.Context, accountName string) (*domain.ShoppingCart, error) {
	key := cartKey(accountName)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {

		if err == redis.Nil {
			return nil, shared.StatusNotFound{Resource: "ShoppingCart", Key: accountName}
		}
		return nil, fmt.Errorf("redis cart repository: get key %q: %w", key, err)
	}

	var cart domain.ShoppingCart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, fmt.Errorf("redis cart repository: unmarshal cart: %w", err)
	}

	return &cart, nil
}

func (r *redisCartRepository) Delete(ctx context.Context, accountName string) error {
	key := cartKey(accountName)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis cart repository: delete key %q: %w", key, err)
	}

	return nil
}

func cartKey(accountName string) string {
	return cartKeyPrefix + accountName
}
