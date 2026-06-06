package main

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/albe194e/albz/shared/protocol"
)

func (s *relayServer) handleMessageSend(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.MessageSendPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid message.send payload")
	}

	recipientUserIDs, err := normalizeRecipientUserIDs(client.userID, payload.ToUserIDs)
	if err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, err.Error())
	}

	if payload.ClientMessageID == "" || payload.ConversationID == "" || strings.TrimSpace(payload.Body) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "missing required message fields")
	}

	recipients := s.getClientsByUserIDs(recipientUserIDs)
	if len(recipients) != len(recipientUserIDs) {
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

	participantUserIDs := append([]string{client.userID}, recipientUserIDs...)
	created := protocol.Envelope[protocol.MessageCreatedPayload]{
		Type:      protocol.EventMessageCreated,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.MessageCreatedPayload{
			MessageID:          payload.ClientMessageID,
			ConversationID:     payload.ConversationID,
			FromUserID:         client.userID,
			ParticipantUserIDs: participantUserIDs,
			Body:               payload.Body,
			SentAt:             payload.SentAt,
		},
	}

	for _, recipient := range recipients {
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

	recipientUserIDs, err := normalizeRecipientUserIDs(client.userID, payload.ToUserIDs)
	if err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, err.Error())
	}
	if strings.TrimSpace(payload.ConversationID) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "conversation_id is required")
	}

	recipients := s.getClientsByUserIDs(recipientUserIDs)
	if len(recipients) != len(recipientUserIDs) {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "recipient is not currently online")
	}

	participantUserIDs := append([]string{client.userID}, recipientUserIDs...)
	created := protocol.Envelope[protocol.ConversationCreatedPayload]{
		Type:      protocol.EventConversationCreated,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.ConversationCreatedPayload{
			ConversationID:     payload.ConversationID,
			ConversationName:   payload.ConversationName,
			ParticipantUserIDs: participantUserIDs,
			FromUserID:         client.userID,
			FromFriendCode:     client.friendCode,
		},
	}

	for _, recipient := range recipients {
		if err := recipient.writeJSON(created); err != nil {
			return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "recipient is not currently online")
		}
	}

	return nil
}

func normalizeRecipientUserIDs(senderUserID string, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("at least one recipient is required")
	}

	normalized := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			return nil, errors.New("recipient user IDs must not be empty")
		}
		if trimmed == senderUserID {
			return nil, errors.New("cannot create a conversation with yourself")
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}

		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	if len(normalized) == 0 {
		return nil, errors.New("at least one recipient is required")
	}

	return normalized, nil
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
