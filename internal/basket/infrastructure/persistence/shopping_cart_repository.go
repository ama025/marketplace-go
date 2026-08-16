

package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"marketplace/internal/basket/domain"
	"marketplace/internal/basket/domain/repositories"

	"github.com/google/uuid"
)

var _ repositories.ShoppingCartRepository = (*shoppingCartRepository)(nil)

type shoppingCartRepository struct {
	db *sql.DB
}

func NewShoppingCartRepository(db *sql.DB) *shoppingCartRepository {
	return &shoppingCartRepository{db: db}
}

func (r *shoppingCartRepository) Save(ctx context.Context, cart *domain.ShoppingCart) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	const upsertCart = `
		INSERT INTO shopping_carts (account_name)
		VALUES ($1)
		ON CONFLICT (account_name) DO NOTHING
	`
	if _, err = tx.ExecContext(ctx, upsertCart, cart.AccountName); err != nil {
		return fmt.Errorf("upsert shopping cart [%s]: %w", cart.AccountName, err)
	}

	const deleteItems = `DELETE FROM shopping_cart_items WHERE account_name = $1`
	if _, err = tx.ExecContext(ctx, deleteItems, cart.AccountName); err != nil {
		return fmt.Errorf("delete cart items [%s]: %w", cart.AccountName, err)
	}

	const insertItem = `
		INSERT INTO shopping_cart_items (id, account_name, item_id, quantity, unit_price, item_title, item_note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, item := range cart.Items {

		itemRowID := uuid.New()

		if _, err = tx.ExecContext(ctx, insertItem,
			itemRowID,
			cart.AccountName,
			item.ItemId,
			item.Quantity,
			item.UnitPrice,
			item.ItemTitle,
			item.ItemNote,
		); err != nil {
			return fmt.Errorf("insert cart item [itemId=%s]: %w", item.ItemId, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit save cart [%s]: %w", cart.AccountName, err)
	}

	return nil
}

func (r *shoppingCartRepository) Get(ctx context.Context, accountName string) (*domain.ShoppingCart, error) {

	const cartQuery = `SELECT account_name FROM shopping_carts WHERE account_name = $1`

	var foundName string
	err := r.db.QueryRowContext(ctx, cartQuery, accountName).Scan(&foundName)
	if err == sql.ErrNoRows {

		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get shopping cart [%s]: %w", accountName, err)
	}

	const itemsQuery = `
		SELECT item_id, quantity, unit_price, item_title, item_note
		FROM shopping_cart_items
		WHERE account_name = $1
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, accountName)
	if err != nil {
		return nil, fmt.Errorf("get cart items [%s]: %w", accountName, err)
	}
	defer rows.Close()

	var items []domain.ShoppingCartItem
	for rows.Next() {
		var item domain.ShoppingCartItem

		if err := rows.Scan(
			&item.ItemId,
			&item.Quantity,
			&item.UnitPrice,
			&item.ItemTitle,
			&item.ItemNote,
		); err != nil {
			return nil, fmt.Errorf("scan cart item [%s]: %w", accountName, err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error cart items [%s]: %w", accountName, err)
	}

	return &domain.ShoppingCart{
		AccountName: accountName,
		Items:       items,
	}, nil
}

func (r *shoppingCartRepository) Delete(ctx context.Context, accountName string) error {

	_, err := r.db.ExecContext(ctx,
		`DELETE FROM cart_items
		 WHERE cart_id = (SELECT id FROM shopping_carts WHERE account_name = $1)`,
		accountName,
	)
	if err != nil {
		return fmt.Errorf("delete cart items [%s]: %w", accountName, err)
	}

	_, err = r.db.ExecContext(ctx,
		`DELETE FROM shopping_carts WHERE account_name = $1`,
		accountName,
	)
	if err != nil {
		return fmt.Errorf("delete shopping cart [%s]: %w", accountName, err)
	}

	return nil
}

