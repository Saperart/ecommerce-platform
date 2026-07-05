-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS cart;

CREATE TABLE cart.items
(
    user_id BIGINT  NOT NULL,
    sku     INTEGER NOT NULL,
    count   INTEGER NOT NULL CHECK (count > 0),
    PRIMARY KEY (user_id, sku)
);

CREATE INDEX idx_cart_items_sku ON cart.items (sku);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cart.items;
DROP SCHEMA IF EXISTS cart;
-- +goose StatementEnd
