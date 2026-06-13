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
			FromDeviceID:       client.deviceID,
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
