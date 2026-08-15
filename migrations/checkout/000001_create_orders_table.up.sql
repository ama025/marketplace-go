-- 000001_create_orders_table.up.sql
-- Таблицы для сервиса оформления заказов (checkout).

-- Заказ — результат оформления корзины.
CREATE TABLE IF NOT EXISTS orders (
    id           UUID           NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    account_name VARCHAR(255)   NOT NULL,                        -- Покупатель
    status       VARCHAR(50)    NOT NULL DEFAULT 'pending',      -- pending | confirmed | cancelled
    total_price  NUMERIC(12, 2) NOT NULL DEFAULT 0,              -- Итоговая сумма
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Позиции заказа (товары из корзины)
CREATE TABLE IF NOT EXISTS order_items (
    id          UUID           NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID           NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id     UUID           NOT NULL,                         -- UUID товара из каталога
    item_title  VARCHAR(255)   NOT NULL,
    quantity    INT            NOT NULL DEFAULT 1,
    unit_price  NUMERIC(10, 2) NOT NULL DEFAULT 0,
    discount    NUMERIC(5, 2)  NOT NULL DEFAULT 0                -- % скидки на момент заказа
);

CREATE INDEX idx_orders_account    ON orders(account_name);
CREATE INDEX idx_orders_status     ON orders(status);
CREATE INDEX idx_order_items_order ON order_items(order_id);
