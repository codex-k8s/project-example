-- name: users__get_by_id :one
SELECT id, username, password_hash, created_at
FROM users
WHERE id = $1;

