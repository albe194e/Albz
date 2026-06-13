package main

import (
	"context"
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	db "github.com/albe194e/albz/server/db/generated"
	"github.com/albe194e/albz/shared/protocol"
)

func (s *relayServer) handleDeviceRegister(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.DeviceRegisterPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid device.register payload")
	}

	device, err := s.registerOrValidateDevice(payload)
	if err != nil {
		logDebugf(
			"device registration rejected for user=%s device=%s: %v",
			redactID(payload.UserID),
			redactID(payload.DeviceID),
			err,
		)
		return s.sendAuthFailure(client, protocol.ErrorCodeDeviceRegistrationConflict, err.Error())
	}

	serverPrivateKey, nonce, err := createAuthChallenge()
	if err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInternal, "failed to generate auth challenge")
	}

	client.userID = device.UserID
	client.deviceID = device.DeviceID
	client.contactCode = device.ContactCode
	client.devicePublicKey = append([]byte(nil), device.DevicePublicKey...)
	client.pendingAuth = &pendingAuth{
		ServerPrivateKey: serverPrivateKey,
		Nonce:            nonce,
	}
	logDebugf(
		"device registration accepted for user=%s device=%s",
		redactID(device.UserID),
		redactID(device.DeviceID),
	)

	return client.writeJSON(protocol.Envelope[protocol.AuthChallengePayload]{
		Type:      protocol.EventAuthChallenge,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.AuthChallengePayload{
			UserID:          device.UserID,
			DeviceID:        device.DeviceID,
			ServerPublicKey: serverPrivateKey.PublicKey().Bytes(),
			Nonce:           nonce,
		},
	})
}

func (s *relayServer) handleAuthRespond(client *clientConn, envelope rawEnvelope) error {
	var payload protocol.AuthRespondPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInvalidMessage, "invalid auth.respond payload")
	}
	if client.pendingAuth == nil {
		return s.sendAuthFailure(client, protocol.ErrorCodeAuthFailed, "no pending auth challenge for this connection")
	}
	if payload.UserID != client.userID || payload.DeviceID != client.deviceID {
		return s.sendAuthFailure(client, protocol.ErrorCodeAuthFailed, "auth response did not match the pending device")
	}

	device, err := s.getRegisteredDeviceByUserID(client.userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.sendAuthFailure(client, protocol.ErrorCodeDeviceNotRegistered, "device is not registered on the relay")
		}
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInternal, "failed to load registered device")
	}
	if device.DeviceID != client.deviceID {
		return s.sendAuthFailure(client, protocol.ErrorCodeAuthFailed, "device registration mismatch")
	}

	expectedProof, err := computeExpectedAuthProof(
		device.DevicePublicKey,
		client.pendingAuth.ServerPrivateKey,
		client.pendingAuth.Nonce,
		device.UserID,
		device.DeviceID,
		device.ContactCode,
	)
	if err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInternal, "failed to verify relay auth proof")
	}
	if !hmac.Equal(expectedProof, payload.Proof) {
		_ = s.sendAuthFailure(client, protocol.ErrorCodeAuthFailed, "device proof did not verify")
		logDebugf(
			"relay auth proof failed for user=%s device=%s",
			redactID(client.userID),
			redactID(client.deviceID),
		)
		return fmt.Errorf("relay auth proof failed for %s/%s", client.userID, client.deviceID)
	}

	client.authenticated = true
	client.pendingAuth = nil
	if err := s.markDeviceSeen(device.UserID); err != nil {
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeInternal, "failed to update relay device last_seen")
	}

	for _, previous := range s.registerAuthenticatedClient(client) {
		_ = previous.close()
	}
	logInfof(
		"authenticated relay client for user=%s device=%s",
		redactID(client.userID),
		redactID(client.deviceID),
	)

	return client.writeJSON(protocol.Envelope[protocol.AuthSuccessPayload]{
		Type:      protocol.EventAuthSuccess,
		EventID:   newID(),
		RequestID: envelope.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.AuthSuccessPayload{
			UserID:      client.userID,
			DeviceID:    client.deviceID,
			ContactCode: client.contactCode,
		},
	})
}

func (s *relayServer) registerOrValidateDevice(payload protocol.DeviceRegisterPayload) (*db.RegisteredDevice, error) {
	userID := strings.TrimSpace(payload.UserID)
	deviceID := strings.TrimSpace(payload.DeviceID)
	contactCode := strings.TrimSpace(payload.ContactCode)

	if userID == "" || deviceID == "" || contactCode == "" || len(payload.DevicePublicKey) == 0 {
		return nil, fmt.Errorf("user_id, device_id, contact_code, and device_public_key are required")
	}

	if _, err := ecdh.X25519().NewPublicKey(payload.DevicePublicKey); err != nil {
		return nil, fmt.Errorf("device_public_key is invalid")
	}

	existing, err := s.getRegisteredDeviceByUserID(userID)
	if err == nil {
		if existing.DeviceID != deviceID {
			return nil, fmt.Errorf("the relay currently supports only one registered device per user")
		}
		if existing.ContactCode != contactCode {
			return nil, fmt.Errorf("contact code does not match the registered device")
		}
		if !hmac.Equal(existing.DevicePublicKey, payload.DevicePublicKey) {
			return nil, fmt.Errorf("device public key does not match the registered device")
		}
		return &existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("load existing registered device: %w", err)
	}

	now := time.Now().Unix()
	if err := s.store.Q.CreateRegisteredDevice(context.Background(), db.CreateRegisteredDeviceParams{
		UserID:          userID,
		DeviceID:        deviceID,
		DevicePublicKey: append([]byte(nil), payload.DevicePublicKey...),
		ContactCode:     contactCode,
		CreatedAt:       now,
		LastSeen:        now,
		RevokedAt:       sql.NullInt64{},
	}); err != nil {
		return nil, fmt.Errorf("create registered device: %w", err)
	}

	created, err := s.getRegisteredDeviceByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("load registered device after create: %w", err)
	}

	return &created, nil
}

func (s *relayServer) getRegisteredDeviceByUserID(userID string) (db.RegisteredDevice, error) {
	return s.store.Q.GetRegisteredDeviceByUserID(context.Background(), userID)
}

func (s *relayServer) markDeviceSeen(userID string) error {
	return s.store.Q.UpdateRegisteredDeviceLastSeen(context.Background(), db.UpdateRegisteredDeviceLastSeenParams{
		LastSeen: time.Now().Unix(),
		UserID:   userID,
	})
}

func createAuthChallenge() (*ecdh.PrivateKey, []byte, error) {
	serverPrivateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, authChallengeSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	return serverPrivateKey, nonce, nil
}

func computeExpectedAuthProof(devicePublicKeyBytes []byte, serverPrivateKey *ecdh.PrivateKey, nonce []byte, userID, deviceID, contactCode string) ([]byte, error) {
	devicePublicKey, err := ecdh.X25519().NewPublicKey(devicePublicKeyBytes)
	if err != nil {
		return nil, err
	}

	sharedSecret, err := serverPrivateKey.ECDH(devicePublicKey)
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha256.New, sharedSecret)
	_, _ = mac.Write([]byte(authProofContext))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(userID))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(deviceID))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(contactCode))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write(nonce)

	return mac.Sum(nil), nil
}
