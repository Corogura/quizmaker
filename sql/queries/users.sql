-- name: CreateUser :exec
INSERT INTO users (id, created_at, updated_at, email, hashed_pw)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
);
--

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;
--

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;
--

-- name: UpdatePassword :exec
UPDATE users
SET hashed_pw = ?, updated_at = ?
WHERE id = ?;
--