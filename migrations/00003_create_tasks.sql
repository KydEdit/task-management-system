-- +goose Up
CREATE TABLE tasks (
    id          INT     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title       TEXT    NOT NULL,
    description TEXT    NOT NULL DEFAULT '',
    user_id     INT     NOT NULL,
    completed   BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_tasks_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS tasks;