package messages

import dbsql "github.com/albe194e/albz/app/core-go/db/sqlc/generated"

const (
	DirectionIncoming    = "incoming"
	DirectionOutgoing    = "outgoing"
	DeliveryStateSending = "sending"
	DeliveryStateFailed  = "failed"
)

type Message struct {
	dbsql.Message
}

type CreateParams = dbsql.CreateMessageParams
type UpdateDeliveryStateParams = dbsql.UpdateMessageDeliveryStateByClientMessageIDParams
