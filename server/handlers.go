package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/albe194e/albz/shared/protocol"
)

func (s *relayServer) handleMessageSend(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.MessageSendPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid message.send payload")
	}

	if payload.ClientMessageID == "" || payload.ConversationID == "" || payload.ToUserID == "" || strings.TrimSpace(payload.Body) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "missing required message fields")
	}

	recipient := s.getClientByUserID(payload.ToUserID)
	if recipient == nil {
		return client.writeJSON(protocol.Envelope[protocol.MessageDeliveryPayload]{
			Type:      protocol.EventMessageDelivery,
			EventID:   newID(),
			RequestID: envelope.RequestID,
			Timestamp: time.Now().Unix(),
			Payload: protocol.MessageDeliveryPayload{
				ClientMessageID: payload.ClientMessageID,
				Status:          protocol.DeliveryStatusRecipientOffline,
			},
		})
	}

	created := protocol.Envelope[protocol.MessageCreatedPayload]{
		Type:      protocol.EventMessageCreated,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.MessageCreatedPayload{
			MessageID:      payload.ClientMessageID,
			ConversationID: payload.ConversationID,
			FromUserID:     client.userID,
			ToUserID:       payload.ToUserID,
			Body:           payload.Body,
			SentAt:         payload.SentAt,
		},
	}

	if err := recipient.writeJSON(created); err != nil {
		return client.writeJSON(protocol.Envelope[protocol.MessageDeliveryPayload]{
			Type:      protocol.EventMessageDelivery,
			EventID:   newID(),
			RequestID: envelope.RequestID,
			Timestamp: time.Now().Unix(),
			Payload: protocol.MessageDeliveryPayload{
				ClientMessageID: payload.ClientMessageID,
				Status:          protocol.DeliveryStatusRejected,
			},
		})
	}

	return client.writeJSON(protocol.Envelope[protocol.MessageDeliveryPayload]{
		Type:      protocol.EventMessageDelivery,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.MessageDeliveryPayload{
			ClientMessageID: payload.ClientMessageID,
			Status:          protocol.DeliveryStatusDelivered,
		},
	})
}

func (s *relayServer) handleConversationCreate(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.ConversationCreatePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid conversation.create payload")
	}

	if strings.TrimSpace(payload.ConversationID) == "" || strings.TrimSpace(payload.ToUserID) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "conversation_id and to_user_id are required")
	}
	if payload.ToUserID == client.userID {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidRecipient, "cannot create a conversation with yourself")
	}

	recipient := s.getClientByUserID(payload.ToUserID)
	if recipient == nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "recipient is not currently online")
	}

	if err := recipient.writeJSON(protocol.Envelope[protocol.ConversationCreatedPayload]{
		Type:      protocol.EventConversationCreated,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.ConversationCreatedPayload{
			ConversationID: payload.ConversationID,
			FromUserID:     client.userID,
			FromFriendCode: client.friendCode,
		},
	}); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "recipient is not currently online")
	}

	return nil
}

func (s *relayServer) handleFriendRequestSend(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.FriendRequestSendPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid friend_request.send payload")
	}

	if strings.TrimSpace(payload.FriendCode) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "friend code is required")
	}
	if strings.TrimSpace(payload.FromProfile.Name) == "" || strings.TrimSpace(payload.FromProfile.Username) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "sender profile is required")
	}
	if payload.FriendCode == client.friendCode {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidRecipient, "cannot send a friend request to yourself")
	}

	recipient := s.getClientByFriendCode(payload.FriendCode)
	if recipient == nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeFriendCodeNotFound, "friend code is not currently online")
	}

	if err := recipient.writeJSON(protocol.Envelope[protocol.FriendRequestReceivedPayload]{
		Type:      protocol.EventFriendRequestReceived,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.FriendRequestReceivedPayload{
			FromProfile: protocol.PublicFriendProfile{
				UserID:     client.userID,
				FriendCode: client.friendCode,
				Name:       payload.FromProfile.Name,
				Username:   payload.FromProfile.Username,
			},
		},
	}); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeFriendCodeNotFound, "friend code is not currently online")
	}

	return nil
}

func (s *relayServer) handleFriendRequestAccept(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.FriendRequestAcceptPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid friend_request.accept payload")
	}
	if strings.TrimSpace(payload.FromUserID) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "from_user_id is required")
	}
	if strings.TrimSpace(payload.AcceptProfile.Name) == "" || strings.TrimSpace(payload.AcceptProfile.Username) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "accept profile is required")
	}

	requester := s.getClientByUserID(payload.FromUserID)
	if requester == nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "requester is not currently online")
	}

	requesterEvent := protocol.Envelope[protocol.FriendRequestAcceptedPayload]{
		Type:      protocol.EventFriendRequestAccepted,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.FriendRequestAcceptedPayload{
			Profile: protocol.PublicFriendProfile{
				UserID:     client.userID,
				FriendCode: client.friendCode,
				Name:       payload.AcceptProfile.Name,
				Username:   payload.AcceptProfile.Username,
			},
		},
	}
	if err := requester.writeJSON(requesterEvent); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "requester is not currently online")
	}

	return client.writeJSON(protocol.Envelope[protocol.FriendRequestAcceptedPayload]{
		Type:      protocol.EventFriendRequestAccepted,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.FriendRequestAcceptedPayload{
			Profile: protocol.PublicFriendProfile{
				UserID:     requester.userID,
				FriendCode: requester.friendCode,
			},
		},
	})
}

func (s *relayServer) handleFriendRequestReject(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.FriendRequestRejectPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid friend_request.reject payload")
	}
	if strings.TrimSpace(payload.FromUserID) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "from_user_id is required")
	}

	requester := s.getClientByUserID(payload.FromUserID)
	if requester == nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "requester is not currently online")
	}

	requesterEvent := protocol.Envelope[protocol.FriendRequestRejectedPayload]{
		Type:      protocol.EventFriendRequestRejected,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.FriendRequestRejectedPayload{
			UserID: client.userID,
		},
	}
	if err := requester.writeJSON(requesterEvent); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "requester is not currently online")
	}

	return client.writeJSON(protocol.Envelope[protocol.FriendRequestRejectedPayload]{
		Type:      protocol.EventFriendRequestRejected,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.FriendRequestRejectedPayload{
			UserID: requester.userID,
		},
	})
}
