-- name: CreateMessage :exec
INSERT INTO messages (
  conversation_id,
  sender_id,
  body,
  created_at
) VALUES (
  ?, ?, ?, ?
);

-- name: GetMessage :one
SELECT
  id,
  conversation_id,
  sender_id,
  body,
  created_at
FROM messages
WHERE id = ?;

-- name: GetMessages :many
SELECT
  id,
  conversation_id,
  sender_id,
  body,
  created_at
FROM messages;

-- name: ListMessagesByConversation :many
SELECT
  id,
  conversation_id,
  sender_id,
  body,
  created_at
FROM messages
WHERE conversation_id = ?
ORDER BY created_at ASC;
