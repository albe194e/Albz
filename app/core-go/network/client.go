package network

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/albe194e/albz/shared/protocol"
	"github.com/gorilla/websocket"
)

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

func (c *Client) Connect(userID string, contactCode string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user ID is required")
	}
	if strings.TrimSpace(contactCode) == "" {
		return fmt.Errorf("contact code is required")
	}

	c.mu.RLock()
	if c.conn != nil && c.userID == userID && c.contactCode == contactCode {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	wsURL, err := url.Parse(c.serverURL)
	if err != nil {
		return fmt.Errorf("parse server URL: %w", err)
	}

	query := wsURL.Query()
	query.Set("user_id", userID)
	query.Set("contact_code", contactCode)
	wsURL.RawQuery = query.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}

	var previous *websocket.Conn

	c.mu.Lock()
	previous = c.conn
	c.conn = conn
	c.userID = userID
	c.contactCode = contactCode
	c.mu.Unlock()

	if previous != nil {
		_ = previous.Close()
	}

	go c.readLoop(conn)
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.userID = ""
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
		c.contactCode = ""
	}
}
