-- +goose Up
ALTER TABLE quiz_questions
ADD COLUMN deleted_at TEXT;

-- +goose Down
ALTER TABLE quiz_questions
DROP COLUMN deleted_at;