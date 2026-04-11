-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE order_status AS ENUM (
    'PENDING',
    'CONFIRMED',
    'PROCESSING',
    'SHIPPED',
    'DELIVERED',
    'CANCELLED'
    );

CREATE TABLE order_items
(
    id       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID           NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    part_id  UUID           NOT NULL,
    quantity INT            NOT NULL CHECK (quantity > 0),
    price    NUMERIC(10, 2) NOT NULL
);

CREATE TABLE orders
(
    id          UUID PRIMARY KEY        DEFAULT uuid_generate_v4(),
    user_id     UUID           NOT NULL,
    status      order_status   NOT NULL DEFAULT 'PENDING',
    total_price NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE outbox
(
    id         UUID PRIMARY KEY      DEFAULT uuid_generate_v4(),
    topic      VARCHAR(255) NOT NULL,
    payload    JSONB        NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    sent_at    TIMESTAMPTZ
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_order_items_order ON order_items (order_id);
CREATE INDEX idx_outbox_unsent ON outbox (created_at) WHERE sent_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;