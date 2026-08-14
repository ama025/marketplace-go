// Package persistence — слой Infrastructure (Persistence).
// Содержит конкретную реализацию репозитория корзины покупок через PostgreSQL.
package persistence

import (
	"context"      // Контекст для управления таймаутами и отменой запросов
	"database/sql" // Пакет для работы с реляционными БД через *sql.DB
	"fmt"          // Пакет для форматирования ошибок

	"marketplace/internal/basket/domain"              // Доменные модели корзины
	"marketplace/internal/basket/domain/repositories" // Интерфейс репозитория

	"github.com/google/uuid" // Пакет для генерации и работы с UUID
)

// Убеждаемся на этапе компиляции, что shoppingCartRepository реализует интерфейс ShoppingCartRepository.
// Если метод не реализован — компилятор выдаст ошибку, а не молчаливо проигнорирует.
var _ repositories.ShoppingCartRepository = (*shoppingCartRepository)(nil)

// shoppingCartRepository — конкретная реализация репозитория корзины покупок.
// Хранит ссылку на пул соединений с PostgreSQL.
type shoppingCartRepository struct {
	db *sql.DB // Пул соединений с базой данных PostgreSQL
}

// NewShoppingCartRepository — конструктор репозитория корзины.
// Принимает *sql.DB и возвращает готовую реализацию интерфейса ShoppingCartRepository.
func NewShoppingCartRepository(db *sql.DB) *shoppingCartRepository {
	return &shoppingCartRepository{db: db}
}

// Save сохраняет корзину покупок в базе данных.
//
// Логика работы:
//  1. Upsert корзины (строки в таблице shopping_carts): создаёт, если нет, иначе ничего не делает.
//  2. Удаляет все старые позиции корзины (чтобы избежать дублирования при обновлении).
//  3. Вставляет текущие позиции корзины заново.
//
// Все три операции выполняются в одной транзакции — атомарно.
// Если любой шаг упадёт — вся транзакция откатится (Rollback).
func (r *shoppingCartRepository) Save(ctx context.Context, cart *domain.ShoppingCart) error {
	// Начинаем транзакцию: все изменения будут применены только при успешном Commit
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// defer с проверкой: если к моменту выхода из функции err != nil — откатываем транзакцию.
	// Если всё прошло успешно и был вызван Commit — Rollback вернёт безвредную ошибку sql.ErrTxDone.
	defer func() {
		if err != nil {
			tx.Rollback() //nolint:errcheck // ошибка Rollback нас не интересует — основная ошибка уже есть
		}
	}()

	// Шаг 1: Upsert корзины.
	// ON CONFLICT (account_name) DO NOTHING — если корзина уже существует, просто пропускаем INSERT.
	const upsertCart = `
		INSERT INTO shopping_carts (account_name)
		VALUES ($1)
		ON CONFLICT (account_name) DO NOTHING
	`
	if _, err = tx.ExecContext(ctx, upsertCart, cart.AccountName); err != nil {
		return fmt.Errorf("upsert shopping cart [%s]: %w", cart.AccountName, err)
	}

	// Шаг 2: Удаляем все текущие позиции корзины.
	// Это проще и надёжнее, чем пытаться сравнивать и обновлять каждую позицию отдельно.
	const deleteItems = `DELETE FROM shopping_cart_items WHERE account_name = $1`
	if _, err = tx.ExecContext(ctx, deleteItems, cart.AccountName); err != nil {
		return fmt.Errorf("delete cart items [%s]: %w", cart.AccountName, err)
	}

	// Шаг 3: Вставляем актуальные позиции корзины.
	const insertItem = `
		INSERT INTO shopping_cart_items (id, account_name, item_id, quantity, unit_price, item_title, item_note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, item := range cart.Items {
		// Генерируем новый UUID для каждой позиции корзины
		itemRowID := uuid.New()

		if _, err = tx.ExecContext(ctx, insertItem,
			itemRowID,        // $1 — первичный ключ строки позиции
			cart.AccountName, // $2 — внешний ключ на shopping_carts
			item.ItemId,      // $3 — UUID товара из каталога
			item.Quantity,    // $4 — количество единиц
			item.UnitPrice,   // $5 — цена за единицу
			item.ItemTitle,   // $6 — название товара (nullable)
			item.ItemNote,    // $7 — заметка покупателя (nullable)
		); err != nil {
			return fmt.Errorf("insert cart item [itemId=%s]: %w", item.ItemId, err)
		}
	}

	// Фиксируем транзакцию: все изменения записываются в БД атомарно
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit save cart [%s]: %w", cart.AccountName, err)
	}

	return nil
}

// Get возвращает корзину покупок по имени аккаунта вместе со всеми позициями.
//
// Логика работы:
//  1. Проверяет наличие корзины в таблице shopping_carts.
//  2. Если найдена — загружает все связанные позиции из shopping_cart_items.
//  3. Если корзина не найдена — возвращает nil, nil (не ошибка).
func (r *shoppingCartRepository) Get(ctx context.Context, accountName string) (*domain.ShoppingCart, error) {
	// Шаг 1: Проверяем существование корзины
	const cartQuery = `SELECT account_name FROM shopping_carts WHERE account_name = $1`

	var foundName string
	err := r.db.QueryRowContext(ctx, cartQuery, accountName).Scan(&foundName)
	if err == sql.ErrNoRows {
		// Корзина не найдена — это нормальная ситуация (новый пользователь)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get shopping cart [%s]: %w", accountName, err)
	}

	// Шаг 2: Загружаем все позиции корзины
	const itemsQuery = `
		SELECT item_id, quantity, unit_price, item_title, item_note
		FROM shopping_cart_items
		WHERE account_name = $1
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, accountName)
	if err != nil {
		return nil, fmt.Errorf("get cart items [%s]: %w", accountName, err)
	}
	defer rows.Close() // Освобождаем ресурсы соединения после чтения

	// Собираем позиции корзины
	var items []domain.ShoppingCartItem
	for rows.Next() {
		var item domain.ShoppingCartItem
		// Сканируем строку в поля структуры; порядок должен совпадать с SELECT
		if err := rows.Scan(
			&item.ItemId,    // item_id
			&item.Quantity,  // quantity
			&item.UnitPrice, // unit_price
			&item.ItemTitle, // item_title (nullable → *string)
			&item.ItemNote,  // item_note  (nullable → *string)
		); err != nil {
			return nil, fmt.Errorf("scan cart item [%s]: %w", accountName, err)
		}
		items = append(items, item)
	}

	// Проверяем, не возникло ли ошибок во время итерации по курсору
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error cart items [%s]: %w", accountName, err)
	}

	// Возвращаем собранную корзину
	return &domain.ShoppingCart{
		AccountName: accountName,
		Items:       items,
	}, nil
}

// Delete удаляет корзину и все её позиции из PostgreSQL по имени аккаунта.
// Сначала удаляет позиции (cart_items), затем саму корзину (shopping_carts).
// Идемпотентная операция — если корзина не существовала, возвращает nil.
func (r *shoppingCartRepository) Delete(ctx context.Context, accountName string) error {
	// Удаляем позиции корзины (дочерние записи — сначала)
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM cart_items
		 WHERE cart_id = (SELECT id FROM shopping_carts WHERE account_name = $1)`,
		accountName,
	)
	if err != nil {
		return fmt.Errorf("delete cart items [%s]: %w", accountName, err)
	}

	// Удаляем саму корзину
	_, err = r.db.ExecContext(ctx,
		`DELETE FROM shopping_carts WHERE account_name = $1`,
		accountName,
	)
	if err != nil {
		return fmt.Errorf("delete shopping cart [%s]: %w", accountName, err)
	}

	return nil
}

