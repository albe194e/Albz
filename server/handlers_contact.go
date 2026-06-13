package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/albe194e/albz/shared/protocol"
)

func (s *relayServer) handleContactRequestSend(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.ContactRequestSendPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid contact_request.send payload")
	}

	if strings.TrimSpace(payload.ContactCode) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "contact code is required")
	}
	if strings.TrimSpace(payload.FromProfile.Name) == "" || strings.TrimSpace(payload.FromProfile.Username) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "sender profile is required")
	}
	if payload.ContactCode == client.contactCode {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidRecipient, "cannot send a contact request to yourself")
	}

	recipient := s.getClientByContactCode(payload.ContactCode)
	if recipient == nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeContactCodeNotFound, "contact code is not currently online")
	}

	if err := recipient.writeJSON(protocol.Envelope[protocol.ContactRequestReceivedPayload]{
		Type:      protocol.EventContactRequestReceived,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.ContactRequestReceivedPayload{
			FromProfile: protocol.PublicContactProfile{
				UserID:          client.userID,
				DeviceID:        client.deviceID,
				DevicePublicKey: append([]byte(nil), client.devicePublicKey...),
				ContactCode:     client.contactCode,
				Name:            payload.FromProfile.Name,
				Username:        payload.FromProfile.Username,
			},
		},
	}); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeContactCodeNotFound, "contact code is not currently online")
	}

	return nil
}

func (s *relayServer) handleContactRequestAccept(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.ContactRequestAcceptPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid contact_request.accept payload")
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

	requesterEvent := protocol.Envelope[protocol.ContactRequestAcceptedPayload]{
		Type:      protocol.EventContactRequestAccepted,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.ContactRequestAcceptedPayload{
			Profile: protocol.PublicContactProfile{
				UserID:          client.userID,
				DeviceID:        client.deviceID,
				DevicePublicKey: append([]byte(nil), client.devicePublicKey...),
				ContactCode:     client.contactCode,
				Name:            payload.AcceptProfile.Name,
				Username:        payload.AcceptProfile.Username,
			},
		},
	}
	if err := requester.writeJSON(requesterEvent); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "requester is not currently online")
	}

	return client.writeJSON(protocol.Envelope[protocol.ContactRequestAcceptedPayload]{
		Type:      protocol.EventContactRequestAccepted,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.ContactRequestAcceptedPayload{
			Profile: protocol.PublicContactProfile{
				UserID:          requester.userID,
				DeviceID:        requester.deviceID,
				DevicePublicKey: append([]byte(nil), requester.devicePublicKey...),
				ContactCode:     requester.contactCode,
			},
		},
	})
}

func (s *relayServer) handleContactRequestReject(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.ContactRequestRejectPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid contact_request.reject payload")
	}
	if strings.TrimSpace(payload.FromUserID) == "" {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "from_user_id is required")
	}

	requester := s.getClientByUserID(payload.FromUserID)
	if requester == nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "requester is not currently online")
	}

	requesterEvent := protocol.Envelope[protocol.ContactRequestRejectedPayload]{
		Type:      protocol.EventContactRequestRejected,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.ContactRequestRejectedPayload{
			UserID: client.userID,
		},
	}
	if err := requester.writeJSON(requesterEvent); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeRecipientOffline, "requester is not currently online")
	}

	return client.writeJSON(protocol.Envelope[protocol.ContactRequestRejectedPayload]{
		Type:      protocol.EventContactRequestRejected,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.ContactRequestRejectedPayload{
			UserID: requester.userID,
		},
	})
}
