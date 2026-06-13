package main

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	dbsqlc "github.com/albe194e/albz/server/db"
	dbstorage "github.com/albe194e/albz/server/db/storage"
	"github.com/albe194e/albz/shared/protocol"
	"github.com/gorilla/websocket"
)

const (
	defaultAddr       = ":8080"
	defaultDBPath     = "dev-local-db/server/relay.db"
	authProofContext  = "haddle-relay-auth-v1"
	authChallengeSize = 32
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type relayServer struct {
	store                *dbstorage.Store
	mu                   sync.RWMutex
	clientsByUserID      map[string]*clientConn
	clientsByContactCode map[string]*clientConn
}

type pendingAuth struct {
	ServerPrivateKey *ecdh.PrivateKey
	Nonce            []byte
}

type clientConn struct {
	userID          string
	deviceID        string
	contactCode     string
	devicePublicKey []byte
	authenticated   bool
	pendingAuth     *pendingAuth
	conn            *websocket.Conn
	writeM          sync.Mutex
}

type rawEnvelope struct {
	Type      protocol.EventType `json:"type"`
	RequestID string             `json:"request_id,omitempty"`
	Payload   json.RawMessage    `json:"payload"`
}

func main() {
	addr := os.Getenv("HADDLE_SERVER_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	store, err := dbstorage.OpenSQLite(context.Background(), resolveServerDBPath(), dbsqlc.SchemaSQL)
	if err != nil {
		logFatalf("open relay db: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	server := &relayServer{
		store:                store,
		clientsByUserID:      make(map[string]*clientConn),
		clientsByContactCode: make(map[string]*clientConn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ws", server.handleWebSocket)

	logInfof("relay server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logFatalf("relay server stopped: %v", err)
	}
}

func resolveServerDBPath() string {
	if configured := strings.TrimSpace(os.Getenv("HADDLE_SERVER_DB_PATH")); configured != "" {
		return configured
	}

	return filepath.Clean(defaultDBPath)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *relayServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &clientConn{
		conn: conn,
	}

	defer func() {
		s.unregisterClient(client)
		_ = client.close()
	}()

	for {
		if err := s.readAndHandleMessage(client); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logDebugf(
					"unexpected websocket close for user=%s device=%s: %v",
					redactID(client.userID),
					redactID(client.deviceID),
					err,
				)
			}
			return
		}
	}
}

func (s *relayServer) registerAuthenticatedClient(client *clientConn) []*clientConn {
	s.mu.Lock()
	defer s.mu.Unlock()

	var previous []*clientConn
	appendPrevious := func(conn *clientConn) {
		if conn == nil || conn == client {
			return
		}
		for _, existing := range previous {
			if existing == conn {
				return
			}
		}
		previous = append(previous, conn)
	}

	appendPrevious(s.clientsByUserID[client.userID])
	appendPrevious(s.clientsByContactCode[client.contactCode])

	s.clientsByUserID[client.userID] = client
	s.clientsByContactCode[client.contactCode] = client

	return previous
}

func (s *relayServer) unregisterClient(client *clientConn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if current, ok := s.clientsByUserID[client.userID]; ok && current == client {
		delete(s.clientsByUserID, client.userID)
	}
	if current, ok := s.clientsByContactCode[client.contactCode]; ok && current == client {
		delete(s.clientsByContactCode, client.contactCode)
	}
}

func (s *relayServer) readAndHandleMessage(client *clientConn) error {
	_, data, err := client.conn.ReadMessage()
	if err != nil {
		return err
	}

	var envelope rawEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return s.sendError(client, "", protocol.ErrorCodeInvalidMessage, "message must be valid JSON")
	}

	if !client.authenticated {
		switch envelope.Type {
		case protocol.EventDeviceRegister:
			return s.handleDeviceRegister(client, envelope)
		case protocol.EventAuthRespond:
			return s.handleAuthRespond(client, envelope)
		default:
			return s.sendError(client, envelope.RequestID, protocol.ErrorCodeUnauthenticated, "client must authenticate before sending relay events")
		}
	}

	switch envelope.Type {
	case protocol.EventMessageSend:
		return s.handleMessageSend(client, envelope)
	case protocol.EventConversationCreate:
		return s.handleConversationCreate(client, envelope)
	case protocol.EventContactRequestSend:
		return s.handleContactRequestSend(client, envelope)
	case protocol.EventContactRequestAccept:
		return s.handleContactRequestAccept(client, envelope)
	case protocol.EventContactRequestReject:
		return s.handleContactRequestReject(client, envelope)
	default:
		return s.sendError(client, envelope.RequestID, protocol.ErrorCodeUnsupported, "unsupported event type")
	}
}

func (s *relayServer) getClientByUserID(userID string) *clientConn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clientsByUserID[userID]
}

func (s *relayServer) getClientsByUserIDs(userIDs []string) []*clientConn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*clientConn, 0, len(userIDs))
	for _, userID := range userIDs {
		client := s.clientsByUserID[userID]
		if client == nil {
			continue
		}
		clients = append(clients, client)
	}

	return clients
}

func (s *relayServer) getClientByContactCode(contactCode string) *clientConn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clientsByContactCode[contactCode]
}

func (s *relayServer) sendError(client *clientConn, requestID, code, message string) error {
	return client.writeJSON(protocol.Envelope[protocol.ErrorPayload]{
		Type:      protocol.EventError,
		EventID:   newID(),
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
		Payload: protocol.ErrorPayload{
			Code:    code,
			Message: message,
		},
	})
}

func (s *relayServer) sendAuthFailure(client *clientConn, code, message string) error {
	return client.writeJSON(protocol.Envelope[protocol.AuthFailurePayload]{
		Type:      protocol.EventAuthFailure,
		EventID:   newID(),
		Timestamp: time.Now().Unix(),
		Payload: protocol.AuthFailurePayload{
			Code:    code,
			Message: message,
		},
	})
}

func (c *clientConn) writeJSON(payload any) error {
	c.writeM.Lock()
	defer c.writeM.Unlock()

	if err := c.conn.WriteJSON(payload); err != nil {
		return err
	}
	return nil
}

func (c *clientConn) close() error {
	c.writeM.Lock()
	defer c.writeM.Unlock()

	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil && !errors.Is(err, websocket.ErrCloseSent) {
		return c.conn.Close()
	}
	return c.conn.Close()
}

func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}
