package network

import (
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/albe194e/albz/shared/protocol"
	"github.com/gorilla/websocket"
)

const authProofContext = "haddle-relay-auth-v1"

type Handlers struct {
	OnConversationCreated    func(protocol.Envelope[protocol.ConversationCreatedPayload])
	OnMessageCreated         func(protocol.Envelope[protocol.MessageCreatedPayload])
	OnMessageDelivery        func(protocol.Envelope[protocol.MessageDeliveryPayload])
	OnContactRequestReceived func(protocol.Envelope[protocol.ContactRequestReceivedPayload])
	OnContactRequestAccepted func(protocol.Envelope[protocol.ContactRequestAcceptedPayload])
	OnContactRequestRejected func(protocol.Envelope[protocol.ContactRequestRejectedPayload])
	OnError                  func(protocol.Envelope[protocol.ErrorPayload])
	OnDisconnect             func(error)
}

type Client struct {
	serverURL string
	handlers  Handlers

	mu          sync.RWMutex
	conn        *websocket.Conn
	userID      string
	deviceID    string
	contactCode string
	writeMu     sync.Mutex
}

type rawEnvelope struct {
	Type      protocol.EventType `json:"type"`
	EventID   string             `json:"event_id,omitempty"`
	RequestID string             `json:"request_id,omitempty"`
	Timestamp int64              `json:"timestamp,omitempty"`
	Payload   json.RawMessage    `json:"payload"`
}

func NewClient(serverURL string, handlers Handlers) *Client {
	return &Client{
		serverURL: serverURL,
		handlers:  handlers,
	}
}

func (c *Client) Connect(userID, deviceID, contactCode string, devicePublicKey, devicePrivateKey []byte) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user ID is required")
	}
	if strings.TrimSpace(deviceID) == "" {
		return fmt.Errorf("device ID is required")
	}
	if strings.TrimSpace(contactCode) == "" {
		return fmt.Errorf("contact code is required")
	}
	if len(devicePublicKey) == 0 {
		return fmt.Errorf("device public key is required")
	}
	if len(devicePrivateKey) == 0 {
		return fmt.Errorf("device private key is required")
	}

	c.mu.RLock()
	if c.conn != nil && c.userID == userID && c.deviceID == deviceID && c.contactCode == contactCode {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	wsURL, err := url.Parse(c.serverURL)
	if err != nil {
		return fmt.Errorf("parse server URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}

	if err := c.authenticateConnection(conn, userID, deviceID, contactCode, devicePublicKey, devicePrivateKey); err != nil {
		_ = conn.Close()
		return err
	}

	var previous *websocket.Conn
	c.mu.Lock()
	previous = c.conn
	c.conn = conn
	c.userID = userID
	c.deviceID = deviceID
	c.contactCode = contactCode
	c.mu.Unlock()

	if previous != nil {
		_ = previous.Close()
	}

	go c.readLoop(conn)
	return nil
}

func (c *Client) authenticateConnection(conn *websocket.Conn, userID, deviceID, contactCode string, devicePublicKey, devicePrivateKey []byte) error {
	if err := conn.WriteJSON(protocol.Envelope[protocol.DeviceRegisterPayload]{
		Type: protocol.EventDeviceRegister,
		Payload: protocol.DeviceRegisterPayload{
			UserID:          userID,
			DeviceID:        deviceID,
			DevicePublicKey: append([]byte(nil), devicePublicKey...),
			ContactCode:     contactCode,
		},
	}); err != nil {
		return fmt.Errorf("send device registration: %w", err)
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read auth challenge: %w", err)
	}

	envelope, err := decodeRawEnvelope(data)
	if err != nil {
		return err
	}

	switch envelope.Type {
	case protocol.EventAuthChallenge:
		var payload protocol.AuthChallengePayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode auth challenge payload: %w", err)
		}

		proof, err := computeAuthProof(devicePrivateKey, payload.ServerPublicKey, payload.Nonce, userID, deviceID, contactCode)
		if err != nil {
			return err
		}

		if err := conn.WriteJSON(protocol.Envelope[protocol.AuthRespondPayload]{
			Type: protocol.EventAuthRespond,
			Payload: protocol.AuthRespondPayload{
				UserID:   userID,
				DeviceID: deviceID,
				Proof:    proof,
			},
		}); err != nil {
			return fmt.Errorf("send auth response: %w", err)
		}
	case protocol.EventAuthFailure:
		return decodeAuthFailure(envelope)
	case protocol.EventError:
		return decodeErrorPayload(envelope)
	default:
		return fmt.Errorf("unexpected handshake event %q", envelope.Type)
	}

	_, data, err = conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read auth success: %w", err)
	}

	envelope, err = decodeRawEnvelope(data)
	if err != nil {
		return err
	}

	switch envelope.Type {
	case protocol.EventAuthSuccess:
		var payload protocol.AuthSuccessPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode auth success payload: %w", err)
		}
		if payload.UserID != userID || payload.DeviceID != deviceID {
			return fmt.Errorf("auth success did not match expected device")
		}
		return nil
	case protocol.EventAuthFailure:
		return decodeAuthFailure(envelope)
	case protocol.EventError:
		return decodeErrorPayload(envelope)
	default:
		return fmt.Errorf("unexpected post-auth event %q", envelope.Type)
	}
}

func decodeRawEnvelope(data []byte) (rawEnvelope, error) {
	var envelope rawEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return rawEnvelope{}, fmt.Errorf("decode envelope: %w", err)
	}

	return envelope, nil
}

func decodeErrorPayload(envelope rawEnvelope) error {
	var payload protocol.ErrorPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("decode error payload: %w", err)
	}

	return errors.New(payload.Message)
}

func decodeAuthFailure(envelope rawEnvelope) error {
	var payload protocol.AuthFailurePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("decode auth failure payload: %w", err)
	}

	if strings.TrimSpace(payload.Message) == "" {
		return fmt.Errorf("relay authentication failed")
	}

	return errors.New(payload.Message)
}

func computeAuthProof(devicePrivateKeyBytes, serverPublicKeyBytes, nonce []byte, userID, deviceID, contactCode string) ([]byte, error) {
	privateKey, err := ecdh.X25519().NewPrivateKey(devicePrivateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("load device private key: %w", err)
	}

	serverPublicKey, err := ecdh.X25519().NewPublicKey(serverPublicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("load server public key: %w", err)
	}

	sharedSecret, err := privateKey.ECDH(serverPublicKey)
	if err != nil {
		return nil, fmt.Errorf("derive relay auth secret: %w", err)
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

func (c *Client) Close() error {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.userID = ""
	c.deviceID = ""
	c.contactCode = ""
	c.mu.Unlock()

	if conn == nil {
		return nil
	}

	return conn.Close()
}

func (c *Client) readLoop(conn *websocket.Conn) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			c.clearConnection(conn)
			if c.handlers.OnDisconnect != nil {
				c.handlers.OnDisconnect(err)
			}
			return
		}

		if err := c.handleIncoming(data); err != nil {
			if c.handlers.OnError != nil {
				c.handlers.OnError(protocol.Envelope[protocol.ErrorPayload]{
					Type: protocol.EventError,
					Payload: protocol.ErrorPayload{
						Code:    protocol.ErrorCodeInvalidMessage,
						Message: err.Error(),
					},
				})
			}
		}
	}
}

func (c *Client) handleIncoming(data []byte) error {
	var envelope rawEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode envelope: %w", err)
	}

	switch envelope.Type {
	case protocol.EventConversationCreated:
		var payload protocol.ConversationCreatedPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode conversation.created payload: %w", err)
		}
		if c.handlers.OnConversationCreated != nil {
			c.handlers.OnConversationCreated(protocol.Envelope[protocol.ConversationCreatedPayload]{
				Type:      envelope.Type,
				EventID:   envelope.EventID,
				RequestID: envelope.RequestID,
				Timestamp: envelope.Timestamp,
				Payload:   payload,
			})
		}
	case protocol.EventMessageCreated:
		var payload protocol.MessageCreatedPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode message.created payload: %w", err)
		}
		if c.handlers.OnMessageCreated != nil {
			c.handlers.OnMessageCreated(protocol.Envelope[protocol.MessageCreatedPayload]{
				Type:      envelope.Type,
				EventID:   envelope.EventID,
				RequestID: envelope.RequestID,
				Timestamp: envelope.Timestamp,
				Payload:   payload,
			})
		}
	case protocol.EventMessageDelivery:
		var payload protocol.MessageDeliveryPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode message.delivery payload: %w", err)
		}
		if c.handlers.OnMessageDelivery != nil {
			c.handlers.OnMessageDelivery(protocol.Envelope[protocol.MessageDeliveryPayload]{
				Type:      envelope.Type,
				EventID:   envelope.EventID,
				RequestID: envelope.RequestID,
				Timestamp: envelope.Timestamp,
				Payload:   payload,
			})
		}
	case protocol.EventContactRequestReceived:
		var payload protocol.ContactRequestReceivedPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode contact_request.received payload: %w", err)
		}
		if c.handlers.OnContactRequestReceived != nil {
			c.handlers.OnContactRequestReceived(protocol.Envelope[protocol.ContactRequestReceivedPayload]{
				Type:      envelope.Type,
				EventID:   envelope.EventID,
				RequestID: envelope.RequestID,
				Timestamp: envelope.Timestamp,
				Payload:   payload,
			})
		}
	case protocol.EventContactRequestAccepted:
		var payload protocol.ContactRequestAcceptedPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode contact_request.accepted payload: %w", err)
		}
		if c.handlers.OnContactRequestAccepted != nil {
			c.handlers.OnContactRequestAccepted(protocol.Envelope[protocol.ContactRequestAcceptedPayload]{
				Type:      envelope.Type,
				EventID:   envelope.EventID,
				RequestID: envelope.RequestID,
				Timestamp: envelope.Timestamp,
				Payload:   payload,
			})
		}
	case protocol.EventContactRequestRejected:
		var payload protocol.ContactRequestRejectedPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode contact_request.rejected payload: %w", err)
		}
		if c.handlers.OnContactRequestRejected != nil {
			c.handlers.OnContactRequestRejected(protocol.Envelope[protocol.ContactRequestRejectedPayload]{
				Type:      envelope.Type,
				EventID:   envelope.EventID,
				RequestID: envelope.RequestID,
				Timestamp: envelope.Timestamp,
				Payload:   payload,
			})
		}
	case protocol.EventError:
		var payload protocol.ErrorPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return fmt.Errorf("decode error payload: %w", err)
		}
		if c.handlers.OnError != nil {
			c.handlers.OnError(protocol.Envelope[protocol.ErrorPayload]{
				Type:      envelope.Type,
				EventID:   envelope.EventID,
				RequestID: envelope.RequestID,
				Timestamp: envelope.Timestamp,
				Payload:   payload,
			})
		}
	default:
		return fmt.Errorf("unsupported event type %q", envelope.Type)
	}

	return nil
}

func (c *Client) writeJSON(payload any) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("websocket is not connected")
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := conn.WriteJSON(payload); err != nil {
		return fmt.Errorf("write websocket message: %w", err)
	}

	return nil
}

func (c *Client) clearConnection(conn *websocket.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == conn {
		c.conn = nil
		c.userID = ""
		c.deviceID = ""
		c.contactCode = ""
	}
}
