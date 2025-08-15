-- +goose Up
CREATE TABLE users(
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    hashed_pw TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;