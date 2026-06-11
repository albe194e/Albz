package protocol

type EventType string

const (
	EventMessageSend            EventType = "message.send"
	EventMessageCreated         EventType = "message.created"
	EventMessageDelivery        EventType = "message.delivery"
	EventConversationCreate     EventType = "conversation.create"
	EventConversationCreated    EventType = "conversation.created"
	EventContactRequestSend     EventType = "contact_request.send"
	EventContactRequestReceived EventType = "contact_request.received"
	EventContactRequestAccept   EventType = "contact_request.accept"
	EventContactRequestAccepted EventType = "contact_request.accepted"
	EventContactRequestReject   EventType = "contact_request.reject"
	EventContactRequestRejected EventType = "contact_request.rejected"
	EventError                  EventType = "error"
)

const (
	ErrorCodeInvalidMessage      = "invalid_message"
	ErrorCodeUnsupported         = "unsupported_event"
	ErrorCodeInternal            = "internal_error"
	ErrorCodeRecipientOffline    = "recipient_offline"
	ErrorCodeContactCodeNotFound = "contact_code_not_found"
	ErrorCodeInvalidRecipient    = "invalid_recipient"
)

type Envelope[T any] struct {
	Type      EventType `json:"type"`
	EventID   string    `json:"event_id,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp int64     `json:"timestamp,omitempty"`
	Payload   T         `json:"payload"`
}

type MessageSendPayload struct {
	ClientMessageID string   `json:"client_message_id"`
	ConversationID  string   `json:"conversation_id"`
	ToUserIDs       []string `json:"to_user_ids"`
	Body            string   `json:"body"`
	SentAt          int64    `json:"sent_at"`
}

type MessageCreatedPayload struct {
	MessageID          string   `json:"message_id"`
	ConversationID     string   `json:"conversation_id"`
	FromUserID         string   `json:"from_user_id"`
	ParticipantUserIDs []string `json:"participant_user_ids"`
	Body               string   `json:"body"`
	SentAt             int64    `json:"sent_at"`
}

type DeliveryStatus string

const (
	DeliveryStatusDelivered        DeliveryStatus = "delivered"
	DeliveryStatusRecipientOffline DeliveryStatus = "recipient_offline"
	DeliveryStatusRejected         DeliveryStatus = "rejected"
)

type MessageDeliveryPayload struct {
	ClientMessageID string         `json:"client_message_id"`
	Status          DeliveryStatus `json:"status"`
}

type ConversationCreatePayload struct {
	ConversationID   string   `json:"conversation_id"`
	ConversationName string   `json:"conversation_name,omitempty"`
	ToUserIDs        []string `json:"to_user_ids"`
}

type ConversationCreatedPayload struct {
	ConversationID     string   `json:"conversation_id"`
	ConversationName   string   `json:"conversation_name,omitempty"`
	ParticipantUserIDs []string `json:"participant_user_ids"`
	FromUserID         string   `json:"from_user_id"`
	FromContactCode    string   `json:"from_contact_code"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PublicContactProfile struct {
	UserID      string `json:"user_id,omitempty"`
	ContactCode string `json:"contact_code"`
	Name        string `json:"name"`
	Username    string `json:"username"`
}

type ContactRequestSendPayload struct {
	ContactCode string               `json:"contact_code"`
	FromProfile PublicContactProfile `json:"from_profile"`
}

type ContactRequestReceivedPayload struct {
	FromProfile PublicContactProfile `json:"from_profile"`
}

type ContactRequestAcceptPayload struct {
	FromUserID    string               `json:"from_user_id"`
	AcceptProfile PublicContactProfile `json:"accept_profile"`
}

type ContactRequestAcceptedPayload struct {
	Profile PublicContactProfile `json:"profile"`
}

type ContactRequestRejectPayload struct {
	FromUserID string `json:"from_user_id"`
}

type ContactRequestRejectedPayload struct {
	UserID string `json:"user_id"`
}
