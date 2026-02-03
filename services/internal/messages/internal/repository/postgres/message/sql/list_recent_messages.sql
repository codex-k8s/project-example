-- name: messages__list_recent :many
SELECT id, user_id, text, created_at, deleted_at
FROM messages
ORDER BY created_at DESC
LIMIT $1;

