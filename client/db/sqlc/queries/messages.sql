-- name: CreateMessage :exec
INSERT INTO messages (
  conversation_id,
  sender_id,
  client_message_id,
  body,
  created_at,
  delivery_state
) VALUES (
  ?, ?, ?, ?, ?, ?
);

-- name: GetMessage :one
SELECT
  id,
  conversation_id,
  sender_id,
  client_message_id,
  body,
  created_at,
  delivery_state
FROM messages
WHERE id = ?;

-- name: GetMessages :many
SELECT
  id,
  conversation_id,
  sender_id,
  client_message_id,
  body,
  created_at,
  delivery_state
FROM messages;

-- name: ListMessagesByConversation :many
SELECT
  id,
  conversation_id,
  sender_id,
  client_message_id,
  body,
  created_at,
  delivery_state
FROM messages
WHERE conversation_id = ?
ORDER BY created_at ASC;

-- name: UpdateMessageDeliveryStateByClientMessageID :exec
UPDATE messages
SET delivery_state = ?
WHERE client_message_id = ?;
