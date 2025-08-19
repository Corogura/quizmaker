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
INSERT INTO quiz_questions (id, quiz_id, question_text, choice1, choice2, choice3, choice4, answer)
VALUES (
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