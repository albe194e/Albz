package conversations

import dbsql "github.com/albe194e/albz/app/core-go/db/sqlc/generated"

const (
	TypeDirect = "direct"
	TypeRoom   = "room"
)

type Conversation struct {
	dbsql.Conversation
}

type Participant struct {
	dbsql.ConversationParticipant
}

type CreateParams = dbsql.CreateConversationParams
type AddParticipantParams = dbsql.AddParticipantParams
