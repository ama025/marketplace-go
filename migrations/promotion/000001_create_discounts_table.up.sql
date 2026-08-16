

CREATE TABLE IF NOT EXISTS discounts (
    id         CHAR(36)       NOT NULL PRIMARY KEY,
    item_id    CHAR(36)       NOT NULL,
    percent    DECIMAL(5, 2)  NOT NULL CHECK (percent BETWEEN 0 AND 100),
    active     BOOLEAN        NOT NULL DEFAULT TRUE,
    starts_at  DATETIME       NULL,
    ends_at    DATETIME       NULL,
    created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_discounts_item_id (item_id),
    INDEX idx_discounts_active  (active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
