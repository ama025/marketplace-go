CREATE TABLE IF NOT EXISTS shopping_carts (
    account_name VARCHAR(255) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS shopping_cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_name VARCHAR(255) NOT NULL REFERENCES shopping_carts(account_name) ON DELETE CASCADE,
    item_id UUID NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    unit_price NUMERIC(10,2) NOT NULL DEFAULT 0,
    item_title VARCHAR(255),
    item_note TEXT
);

CREATE INDEX idx_shopping_cart_items_account ON shopping_cart_items(account_name);