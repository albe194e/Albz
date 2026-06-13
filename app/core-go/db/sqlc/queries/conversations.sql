-- name: GetConversationsByUserID :many
SELECT conversations.*
FROM conversations
JOIN conversation_participants ON conversation_participants.conversation_id = conversations.id
WHERE conversation_participants.user_id = ?
ORDER BY COALESCE(conversations.updated_at, conversations.created_at) DESC, conversations.id DESC;

-- name: GetConversationByID :one
SELECT *
FROM conversations
WHERE id = ?;

-- name: CreateConversation :one
INSERT INTO conversations (id, name, type, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: AddParticipant :exec
INSERT INTO conversation_participants (conversation_id, user_id, created_at)
VALUES (?, ?, ?);

-- name: ListConversationParticipantIDs :many
SELECT user_id
FROM conversation_participants
WHERE conversation_id = ?
ORDER BY created_at ASC, id ASC;
