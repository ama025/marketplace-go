// Package repositories определяет контракты (интерфейсы) хранилищ данных корзины покупок.
// Любая конкретная реализация (PostgreSQL, In-Memory и т.д.) должна реализовать эти интерфейсы.
package repositories

import (
	"context" // Стандартный контекст для управления таймаутами и отменой запросов

	"marketplace/internal/basket/domain" // Доменные модели корзины
)

// ShoppingCartRepository — абстрактный интерфейс репозитория корзины покупок.
// Описывает операции сохранения и получения корзины из хранилища.
type ShoppingCartRepository interface {
	// Save сохраняет корзину покупок (создаёт или обновляет).
	// Если корзина с данным accountName уже существует — позиции пересоздаются (upsert).
	// Возвращает ошибку, если сохранение не удалось.
	Save(ctx context.Context, cart *domain.ShoppingCart) error

	// Get возвращает корзину покупок по имени аккаунта.
	// Если корзина не найдена — возвращает nil, nil.
	Get(ctx context.Context, accountName string) (*domain.ShoppingCart, error)

	// Delete удаляет корзину покупок по имени аккаунта.
	// Идемпотентная операция — не возвращает ошибку, если корзина не существовала.
	Delete(ctx context.Context, accountName string) error
}
