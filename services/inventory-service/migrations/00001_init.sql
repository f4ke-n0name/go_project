-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE stock
(
    id         UUID PRIMARY KEY     DEFAULT uuid_generate_v4(),
    part_id    UUID        NOT NULL UNIQUE,
    quantity   INT         NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved   INT         NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE movement_type AS ENUM (
    'reserve',
    'release',
    'restock',
    'writeoff'
    );

CREATE TABLE stock_movements
(
    id         UUID PRIMARY KEY       DEFAULT uuid_generate_v4(),
    part_id    UUID          NOT NULL,
    order_id   UUID,
    type       movement_type NOT NULL,
    quantity   INT           NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stock_part_id ON stock (part_id);
CREATE INDEX idx_stock_movements_part ON stock_movements (part_id);
CREATE INDEX idx_stock_movements_order ON stock_movements (order_id);

-- +goose Down
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS stock;