-- name: messages__get_by_id :one
SELECT user_id, deleted_at
FROM messages
WHERE id = $1;

