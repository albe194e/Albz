-- name: CreateMessage :exec
INSERT INTO messages (
  id,
  conversation_id,
  sender_id,
  recipient_id,
  body,
  created_at,
  direction,
  delivery_state
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: ListMessagesByConversation :many
SELECT
  id,
  conversation_id,
  sender_id,
  recipient_id,
  body,
  created_at,
  direction,
  delivery_state
FROM messages
WHERE conversation_id = ?
ORDER BY created_at ASC;

-- name: UpdateMessageDeliveryState :exec
UPDATE messages
SET delivery_state = ?
WHERE id = ?;

-- name: GetMessage :one
SELECT
  id,
  conversation_id,
  sender_id,
  recipient_id,
  body,
  created_at,
  direction,
  delivery_state
FROM messages
WHERE id = ?;