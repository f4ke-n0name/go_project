-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE categories
(
    id         UUID PRIMARY KEY      DEFAULT uuid_generate_v4(),
    name       VARCHAR(100) NOT NULL,
    parent_id  UUID         REFERENCES categories (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE parts
(
    id          UUID PRIMARY KEY        DEFAULT uuid_generate_v4(),
    sku         VARCHAR(100)   NOT NULL UNIQUE,
    name        VARCHAR(255)   NOT NULL,
    description TEXT,
    brand       VARCHAR(100)   NOT NULL,
    price       NUMERIC(10, 2) NOT NULL,
    category_id UUID           REFERENCES categories (id) ON DELETE SET NULL,
    stock       INT            NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE compatibility
(
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    part_id   UUID         NOT NULL REFERENCES parts (id) ON DELETE CASCADE,
    make      VARCHAR(100) NOT NULL,
    model     VARCHAR(100) NOT NULL,
    year_from INT          NOT NULL,
    year_to   INT          NOT NULL
);

CREATE INDEX idx_parts_sku ON parts (sku);
CREATE INDEX idx_parts_brand ON parts (brand);
CREATE INDEX idx_parts_category ON parts (category_id);
CREATE INDEX idx_compatibility_part ON compatibility (part_id);
CREATE INDEX idx_compatibility_make_model ON compatibility (make, model);

CREATE INDEX idx_parts_fts ON parts USING GIN (
                                               to_tsvector('english', name || ' ' || brand)
    );

-- +goose Down
DROP TABLE IF EXISTS compatibility;
DROP TABLE IF EXISTS parts;
DROP TABLE IF EXISTS categories;