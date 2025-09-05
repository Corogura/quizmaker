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

-- name: CreateQuizQuestions :exec
INSERT INTO quiz_questions (id, quiz_id, question_number, question_text, choice1, choice2, choice3, choice4, answer)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: GetQuiz :one
SELECT * FROM quizzes JOIN quiz_questions ON quizzes.id = quiz_questions.quiz_id
WHERE quizzes.id = ?;

-- name: GetQuizIDFromPath :one
SELECT id, title, user_id, deleted_at FROM quizzes WHERE path = ?;

-- name: GetQuestionCountInQuiz :one
SELECT COUNT(*) AS question_count FROM quiz_questions WHERE quiz_id = ?;

-- name: DeleteQuiz :exec
UPDATE quizzes SET deleted_at = ? WHERE id = ?;

-- name: DeleteQuizQuestion :exec
UPDATE quiz_questions SET deleted_at = ? WHERE id = ?;

-- name: GetQuestionFromQuestionNumber :one
SELECT * FROM quiz_questions WHERE question_number = ? AND quiz_id = ?;

-- name: UpdateQuizTitle :exec
UPDATE quizzes SET title = ?, updated_at = ? WHERE id = ?;

-- name: GetAllQuizzesByUserID :many
SELECT * FROM quizzes WHERE user_id = ? AND deleted_at IS NULL ORDER BY updated_at DESC;

-- name: GetAllQuestionsInQuiz :many
SELECT * FROM quiz_questions WHERE quiz_id = ? AND deleted_at IS NULL ORDER BY question_number ASC;