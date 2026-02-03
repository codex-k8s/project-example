-- name: messages__create :one
INSERT INTO messages (user_id, text)
VALUES ($1, $2)
RETURNING id, user_id, text, created_at, deleted_at;

