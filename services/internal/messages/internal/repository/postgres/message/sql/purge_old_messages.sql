-- name: messages__purge_old :many
UPDATE messages
SET deleted_at = now()
WHERE created_at < $1 AND deleted_at IS NULL
RETURNING id, deleted_at;

