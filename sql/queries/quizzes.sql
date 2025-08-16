-- name: CreateQuiz :exec
INSERT INTO quizzes (id, created_at, updated_at, title, user_id, path)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

INSERT INTO quiz_questions (id, quiz_id, question_text, choices)
VALUES (
    ?,
    ?,
    ?,
    ?
);

-- name: GetQuiz :one
SELECT * FROM quizzes JOIN quiz_questions ON quizzes.id = quiz_questions.quiz_id
WHERE quizzes.id = ?;