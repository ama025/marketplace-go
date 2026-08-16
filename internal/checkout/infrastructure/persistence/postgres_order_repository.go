package persistence

import (
	"context"
	"database/sql"

	"marketplace/internal/checkout/domain/entities"
	"marketplace/internal/checkout/domain/repositories"

	"github.com/google/uuid"
)

type postgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) repositories.OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) Create(ctx context.Context, order entities.Order) (entities.Order, error) {
	order.ID = uuid.New()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO orders (id, account_name, status, total_price)
		 VALUES ($1, $2, $3, $4)`,
		order.ID, order.AccountName, order.Status, order.TotalPrice,
	)
	if err != nil {
		return entities.Order{}, err
	}

	for i := range order.Items {
		order.Items[i].ID = uuid.New()
		order.Items[i].OrderID = order.ID

		_, err := r.db.ExecContext(ctx,
			`INSERT INTO order_items (id, order_id, item_id, item_title, quantity, unit_price, discount)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			order.Items[i].ID,
			order.Items[i].OrderID,
			order.Items[i].ItemID,
			order.Items[i].ItemTitle,
			order.Items[i].Quantity,
			order.Items[i].UnitPrice,
			order.Items[i].Discount,
		)
		if err != nil {
			return entities.Order{}, err
		}
	}

	return order, nil
}

func (r *postgresOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Order, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, account_name, status, total_price, created_at, updated_at
		 FROM orders WHERE id = $1`,
		id,
	)

	var o entities.Order
	err := row.Scan(&o.ID, &o.AccountName, &o.Status, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *postgresOrderRepository) GetByAccount(ctx context.Context, accountName string) ([]entities.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, account_name, status, total_price, created_at, updated_at
		 FROM orders WHERE account_name = $1
		 ORDER BY created_at DESC`,
		accountName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []entities.Order
	for rows.Next() {
		var o entities.Order
		if err := rows.Scan(&o.ID, &o.AccountName, &o.Status, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *postgresOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.OrderStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	return err
}
