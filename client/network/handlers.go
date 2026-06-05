package network

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

func (c *Client) SendFriendRequest(requestID string, payload protocol.FriendRequestSendPayload) error {
	return c.writeJSON(protocol.Envelope[protocol.FriendRequestSendPayload]{
		Type:      protocol.EventFriendRequestSend,
		RequestID: requestID,
		Payload:   payload,
	})
}

func (c *Client) AcceptFriendRequest(requestID string, payload protocol.FriendRequestAcceptPayload) error {
	return c.writeJSON(protocol.Envelope[protocol.FriendRequestAcceptPayload]{
		Type:      protocol.EventFriendRequestAccept,
		RequestID: requestID,
		Payload:   payload,
	})
}

func (c *Client) RejectFriendRequest(requestID string, payload protocol.FriendRequestRejectPayload) error {
	return c.writeJSON(protocol.Envelope[protocol.FriendRequestRejectPayload]{
		Type:      protocol.EventFriendRequestReject,
		RequestID: requestID,
		Payload:   payload,
	})
}
