package relay

import "github.com/albe194e/albz/shared/protocol"

func (c *Client) SendMessage(requestID string, payload protocol.MessageSendPayload) error {
	envelope := protocol.Envelope[protocol.MessageSendPayload]{
		Type:      protocol.EventMessageSend,
		RequestID: requestID,
		Payload:   payload,
	}

	return c.writeJSON(envelope)
}

func (c *Client) CreateConversation(requestID string, payload protocol.ConversationCreatePayload) error {
	return c.writeJSON(protocol.Envelope[protocol.ConversationCreatePayload]{
		Type:      protocol.EventConversationCreate,
		RequestID: requestID,
		Payload:   payload,
	})
}

func (c *Client) SendContactRequest(requestID string, payload protocol.ContactRequestSendPayload) error {
	return c.writeJSON(protocol.Envelope[protocol.ContactRequestSendPayload]{
		Type:      protocol.EventContactRequestSend,
		RequestID: requestID,
		Payload:   payload,
	})
}

func (c *Client) AcceptContactRequest(requestID string, payload protocol.ContactRequestAcceptPayload) error {
	return c.writeJSON(protocol.Envelope[protocol.ContactRequestAcceptPayload]{
		Type:      protocol.EventContactRequestAccept,
		RequestID: requestID,
		Payload:   payload,
	})
}

func (c *Client) RejectContactRequest(requestID string, payload protocol.ContactRequestRejectPayload) error {
	return c.writeJSON(protocol.Envelope[protocol.ContactRequestRejectPayload]{
		Type:      protocol.EventContactRequestReject,
		RequestID: requestID,
		Payload:   payload,
	})
}
