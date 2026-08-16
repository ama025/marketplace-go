

CREATE TABLE IF NOT EXISTS orders (
    id           UUID           NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    account_name VARCHAR(255)   NOT NULL,
    status       VARCHAR(50)    NOT NULL DEFAULT 'pending',
    total_price  NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id          UUID           NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID           NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id     UUID           NOT NULL,
    item_title  VARCHAR(255)   NOT NULL,
    quantity    INT            NOT NULL DEFAULT 1,
    unit_price  NUMERIC(10, 2) NOT NULL DEFAULT 0,
    discount    NUMERIC(5, 2)  NOT NULL DEFAULT 0
);

CREATE INDEX idx_orders_account    ON orders(account_name);
CREATE INDEX idx_orders_status     ON orders(status);
CREATE INDEX idx_order_items_order ON order_items(order_id);
