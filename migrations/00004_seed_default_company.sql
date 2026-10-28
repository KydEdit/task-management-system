-- +goose Up
INSERT INTO companies (name) VALUES ('Default Company');

-- +goose Down
DELETE FROM companies WHERE name = 'Default Company';