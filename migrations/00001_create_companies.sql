-- +goose Up
CREATE TABLE companies (
    id         INT         GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS companies;