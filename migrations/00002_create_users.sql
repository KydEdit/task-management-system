-- +goose Up
CREATE TABLE users (
    id         INT          GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    company_id INT          NOT NULL,
    CONSTRAINT fk_users_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON DELETE RESTRICT
);

-- +goose Down
DROP TABLE IF EXISTS users;