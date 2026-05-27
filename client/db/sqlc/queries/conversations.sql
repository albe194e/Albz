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

-- name: CreateConversation :exec
INSERT INTO conversations (id, name)
VALUES (?, ?);

-- name: AddParticipant :exec
INSERT INTO conversation_participants (conversation_id, participant_id)
VALUES (?, ?);

