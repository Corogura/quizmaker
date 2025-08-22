-- +goose Up
ALTER TABLE quizzes
ADD COLUMN deleted_at TEXT;

-- +goose Down
ALTER TABLE quizzes
DROP COLUMN deleted_at;