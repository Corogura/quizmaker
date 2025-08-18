-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT 
    users.id,
    users.created_at,
    users.updated_at,
    users.email,
    refresh_tokens.expires_at AS token_expires_at,
    refresh_tokens.revoked_at AS token_revoked_at
FROM users
JOIN refresh_tokens
    ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = ?;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = ?, revoked_at = ?
WHERE token = ?;