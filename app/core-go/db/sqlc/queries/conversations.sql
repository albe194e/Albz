-- name: GetConversationsByUserID :many
SELECT conversations.id, conversations.name
FROM conversations
JOIN conversation_participants ON conversation_participants.conversation_id = conversations.id
WHERE conversation_participants.participant_id = ?
ORDER BY conversations.id DESC;

-- name: GetConversationByID :one
SELECT *
FROM conversations
WHERE id = ?;

-- name: CreateConversation :one
INSERT INTO conversations (id, name)
VALUES (?, ?)
RETURNING *;

-- name: AddParticipant :exec
INSERT INTO conversation_participants (conversation_id, participant_id)
VALUES (?, ?);

-- name: ListConversationParticipantIDs :many
SELECT participant_id
FROM conversation_participants
WHERE conversation_id = ?;

