-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS watched_directories(
    id INTEGER PRIMARY KEY NOT NULL,
    bucket VARCHAR(50) UNIQUE NOT NULL,
    path TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS watched_directories;
-- +goose StatementEnd
