-- name: users__get_by_username :one
SELECT id, username, password_hash, created_at
FROM users
WHERE username = $1;

