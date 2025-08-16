-- +goose Up
CREATE TABLE quiz_questions(
    id TEXT PRIMARY KEY,
    quiz_id TEXT NOT NULL,
    question_text TEXT NOT NULL,
    choices TEXT NOT NULL, -- JSON array of choices
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE quiz_questions;