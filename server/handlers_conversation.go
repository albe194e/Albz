package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/albe194e/albz/shared/protocol"
)

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
			FromDeviceID:       client.deviceID,
			FromContactCode:    client.contactCode,
		},
	}

	for _, recipient := range recipients {
		if err := recipient.writeJSON(created); err != nil {
			return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "recipient is not currently online")
		}
	}

	return nil
}
