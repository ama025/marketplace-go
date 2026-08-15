-- 000001_create_discounts_table.up.sql
-- Таблица скидок для товаров каталога.
-- Скидка задаётся в процентах (0-100) и может быть ограничена по времени.

CREATE TABLE IF NOT EXISTS discounts (
    id         CHAR(36)       NOT NULL PRIMARY KEY,             -- UUID скидки
    item_id    CHAR(36)       NOT NULL,                         -- UUID товара из каталога
    percent    DECIMAL(5, 2)  NOT NULL CHECK (percent BETWEEN 0 AND 100), -- % скидки
    active     BOOLEAN        NOT NULL DEFAULT TRUE,            -- включена ли скидка
    starts_at  DATETIME       NULL,                             -- начало действия (NULL = сразу)
    ends_at    DATETIME       NULL,                             -- конец действия (NULL = бессрочно)
    created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_discounts_item_id (item_id),
    INDEX idx_discounts_active  (active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
