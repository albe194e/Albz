package protocol

type EventType string

const (
	EventMessageSend           EventType = "message.send"
	EventMessageCreated        EventType = "message.created"
	EventMessageDelivery       EventType = "message.delivery"
	EventConversationCreate    EventType = "conversation.create"
	EventConversationCreated   EventType = "conversation.created"
	EventFriendRequestSend     EventType = "friend_request.send"
	EventFriendRequestReceived EventType = "friend_request.received"
	EventFriendRequestAccept   EventType = "friend_request.accept"
	EventFriendRequestAccepted EventType = "friend_request.accepted"
	EventFriendRequestReject   EventType = "friend_request.reject"
	EventFriendRequestRejected EventType = "friend_request.rejected"
	EventError                 EventType = "error"
)

const (
	ErrorCodeInvalidMessage     = "invalid_message"
	ErrorCodeUnsupported        = "unsupported_event"
	ErrorCodeInternal           = "internal_error"
	ErrorCodeRecipientOffline   = "recipient_offline"
	ErrorCodeFriendCodeNotFound = "friend_code_not_found"
	ErrorCodeInvalidRecipient   = "invalid_recipient"
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
	FromFriendCode     string   `json:"from_friend_code"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PublicFriendProfile struct {
	UserID     string `json:"user_id,omitempty"`
	FriendCode string `json:"friend_code"`
	Name       string `json:"name"`
	Username   string `json:"username"`
}

type FriendRequestSendPayload struct {
	FriendCode  string              `json:"friend_code"`
	FromProfile PublicFriendProfile `json:"from_profile"`
}

type FriendRequestReceivedPayload struct {
	FromProfile PublicFriendProfile `json:"from_profile"`
}

type FriendRequestAcceptPayload struct {
	FromUserID    string              `json:"from_user_id"`
	AcceptProfile PublicFriendProfile `json:"accept_profile"`
}

type FriendRequestAcceptedPayload struct {
	Profile PublicFriendProfile `json:"profile"`
}

type FriendRequestRejectPayload struct {
	FromUserID string `json:"from_user_id"`
}

type FriendRequestRejectedPayload struct {
	UserID string `json:"user_id"`
}
