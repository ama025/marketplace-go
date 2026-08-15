package persistence

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"

	"marketplace/internal/promotion/domain/entities"
	"marketplace/internal/promotion/domain/repositories"
)

// mysqlPromotionRepository — MySQL реализация PromotionRepository.
// Живёт в infrastructure-слое и знает о деталях БД.
// Application-слой видит только интерфейс из domain/repositories.
type mysqlPromotionRepository struct {
	db *sql.DB
}

// NewMySQLPromotionRepository создаёт репозиторий поверх переданного *sql.DB.
func NewMySQLPromotionRepository(db *sql.DB) repositories.PromotionRepository {
	return &mysqlPromotionRepository{db: db}
}

// Add вставляет новую скидку в таблицу discounts и возвращает её UUID.
func (r *mysqlPromotionRepository) Add(ctx context.Context, d entities.Discount) (string, error) {
	d.ID = uuid.NewString()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO discounts (id, item_id, percent, active, starts_at, ends_at)
		 VALUES (?, ?, ?, TRUE, ?, ?)`,
		d.ID, d.ItemID, d.Percent, nullTime(d.StartsAt), nullTime(d.EndsAt),
	)
	if err != nil {
		return "", err
	}
	return d.ID, nil
}

// FindByCatalogItem возвращает активную скидку для товара или nil если её нет.
func (r *mysqlPromotionRepository) FindByCatalogItem(ctx context.Context, itemID string) (*entities.Discount, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, item_id, percent, active, created_at, updated_at
		 FROM discounts
		 WHERE item_id = ? AND active = TRUE
		 ORDER BY created_at DESC
		 LIMIT 1`,
		itemID,
	)

	d, err := scanDiscount(row)
	if err == sql.ErrNoRows {
		return nil, nil // скидки нет — это норма
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

// FindManyByCatalogItems возвращает активные скидки для списка товаров (batch).
// Товары без скидки просто отсутствуют в результате.
func (r *mysqlPromotionRepository) FindManyByCatalogItems(ctx context.Context, itemIDs []string) ([]entities.Discount, error) {
	if len(itemIDs) == 0 {
		return nil, nil
	}

	// Строим placeholders: ?,?,? для IN-запроса
	placeholders := strings.Repeat("?,", len(itemIDs))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, len(itemIDs))
	for i, id := range itemIDs {
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, item_id, percent, active, created_at, updated_at
		 FROM discounts
		 WHERE item_id IN (`+placeholders+`) AND active = TRUE`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var discounts []entities.Discount
	for rows.Next() {
		var d entities.Discount
		if err := rows.Scan(
			&d.ID, &d.ItemID, &d.Percent, &d.Active,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		discounts = append(discounts, d)
	}
	return discounts, rows.Err()
}

// Deactivate помечает скидку как неактивную (active = FALSE).
// Физически запись не удаляется — история сохраняется.
func (r *mysqlPromotionRepository) Deactivate(ctx context.Context, discountID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE discounts SET active = FALSE WHERE id = ?`,
		discountID,
	)
	return err
}

// ─── helpers ──────────────────────────────────────────────────────────────────

type discountScanner interface {
	Scan(dest ...any) error
}

func scanDiscount(row discountScanner) (*entities.Discount, error) {
	var d entities.Discount
	err := row.Scan(&d.ID, &d.ItemID, &d.Percent, &d.Active, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// nullTime конвертирует *time.Time в sql.NullTime для корректной записи NULL в MySQL.
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
