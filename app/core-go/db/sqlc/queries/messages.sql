-- name: CreateMessage :exec
INSERT INTO messages (
	id,
	conversation_id,
	sender_user_id,
	sender_device_id,
	client_message_id,
	body,
	created_at,
	received_at,
	direction,
	delivery_state
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetMessage :one
SELECT *
FROM messages
WHERE id = ?;

-- name: GetMessages :many
SELECT *
FROM messages
ORDER BY created_at ASC, id ASC;

-- name: ListMessagesByConversation :many
SELECT *
FROM messages
WHERE conversation_id = ?
ORDER BY created_at ASC, id ASC;

-- name: UpdateMessageDeliveryStateByClientMessageID :exec
UPDATE messages
SET delivery_state = ?
WHERE client_message_id = ?;
