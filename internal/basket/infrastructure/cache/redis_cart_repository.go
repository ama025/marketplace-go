// Package cache — реализация репозитория корзины через Redis (in-memory кэш).
// Используется для быстрого чтения/записи корзины без обращения к PostgreSQL.
package cache

import (
	"context"
	"encoding/json" // Сериализация корзины в JSON для хранения в Redis
	"fmt"
	"time" // TTL — время жизни ключа в Redis

	"marketplace/internal/basket/domain"
	"marketplace/internal/shared"

	"github.com/redis/go-redis/v9"
)

const (
	// cartKeyPrefix — префикс ключа в Redis. Итоговый ключ: "cart:john_doe"
	cartKeyPrefix = "cart:"

	// defaultTTL — корзина живёт в Redis 7 дней без активности.
	// Если пользователь не заходил — корзина автоматически удаляется.
	defaultTTL = 7 * 24 * time.Hour
)

// redisCartRepository — реализация ShoppingCartRepository поверх Redis.
// Хранит корзину как JSON-строку по ключу "cart:{accountName}".
type redisCartRepository struct {
	client *redis.Client // Клиент подключения к Redis
}

// NewRedisCartRepository — конструктор Redis-репозитория корзины.
//
// Параметры:
//   - client: готовый клиент go-redis (*redis.Client)
//
// Пример создания клиента:
//
//	rdb := redis.NewClient(&redis.Options{
//	    Addr:     "localhost:6379",
//	    Password: "12345678",
//	})
//	repo := cache.NewRedisCartRepository(rdb)
func NewRedisCartRepository(client *redis.Client) *redisCartRepository {
	return &redisCartRepository{client: client}
}

// Save сохраняет корзину в Redis как JSON-строку.
// Если корзина с таким accountName уже существует — перезаписывает её (upsert).
// TTL сбрасывается при каждом сохранении (активность продлевает жизнь корзины).
//
// Ключ в Redis: "cart:{accountName}"
// Значение:     JSON-сериализованная структура domain.ShoppingCart
func (r *redisCartRepository) Save(ctx context.Context, cart *domain.ShoppingCart) error {
	// Сериализуем корзину в JSON
	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("redis cart repository: marshal cart: %w", err)
	}

	// Записываем в Redis с TTL. SET key value EX ttl
	key := cartKey(cart.AccountName)
	if err := r.client.Set(ctx, key, data, defaultTTL).Err(); err != nil {
		return fmt.Errorf("redis cart repository: set key %q: %w", key, err)
	}

	return nil
}

// Get возвращает корзину из Redis по имени аккаунта.
//
// Возвращает:
//   - (*domain.ShoppingCart, nil) — корзина найдена
//   - (nil, nil)                  — корзина не найдена (redis.Nil)
//   - (nil, error)                — ошибка соединения или десериализации
func (r *redisCartRepository) Get(ctx context.Context, accountName string) (*domain.ShoppingCart, error) {
	key := cartKey(accountName)

	// Читаем JSON из Redis
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		// redis.Nil — ключ не существует, это не ошибка
		if err == redis.Nil {
			return nil, shared.StatusNotFound{Resource: "ShoppingCart", Key: accountName}
		}
		return nil, fmt.Errorf("redis cart repository: get key %q: %w", key, err)
	}

	// Десериализуем JSON обратно в доменную модель
	var cart domain.ShoppingCart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, fmt.Errorf("redis cart repository: unmarshal cart: %w", err)
	}

	return &cart, nil
}

// Delete удаляет корзину из Redis.
// Используется после оформления заказа — корзина больше не нужна.
//
// Возвращает nil если ключ не существовал (идемпотентная операция).
func (r *redisCartRepository) Delete(ctx context.Context, accountName string) error {
	key := cartKey(accountName)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis cart repository: delete key %q: %w", key, err)
	}

	return nil
}

// cartKey формирует ключ Redis для корзины конкретного пользователя.
// Формат: "cart:{accountName}", например: "cart:john_doe"
func cartKey(accountName string) string {
	return cartKeyPrefix + accountName
}
