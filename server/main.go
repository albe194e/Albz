package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/albe194e/albz/shared/protocol"
	"github.com/gorilla/websocket"
)

const defaultAddr = ":8080"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type relayServer struct {
	mu                  sync.RWMutex
	clientsByUserID     map[string]*clientConn
	clientsByFriendCode map[string]*clientConn
}

type clientConn struct {
	userID     string
	friendCode string
	conn       *websocket.Conn
	writeM     sync.Mutex
}

type rawEnvelope struct {
	Type      protocol.EventType `json:"type"`
	RequestID string             `json:"request_id,omitempty"`
	Payload   json.RawMessage    `json:"payload"`
}

func main() {
	addr := os.Getenv("ALBZ_SERVER_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	server := &relayServer{
		clientsByUserID:     make(map[string]*clientConn),
		clientsByFriendCode: make(map[string]*clientConn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/ws", server.handleWebSocket)

	log.Printf("relay server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *relayServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	friendCode := strings.TrimSpace(r.URL.Query().Get("friend_code"))
	if userID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}
	if friendCode == "" {
		http.Error(w, "missing friend_code", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &clientConn{
		userID:     userID,
		friendCode: friendCode,
		conn:       conn,
	}

	for _, previous := range s.registerClient(client) {
		_ = previous.close()
	}

	defer func() {
		s.unregisterClient(client)
		_ = client.close()
	}()

	for {
		if err := s.readAndHandleMessage(client); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected websocket close for %s: %v", client.userID, err)
			}
			return
		}
	}
}

func (s *relayServer) registerClient(client *clientConn) []*clientConn {
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
	appendPrevious(s.clientsByFriendCode[client.friendCode])

	s.clientsByUserID[client.userID] = client
	s.clientsByFriendCode[client.friendCode] = client

	return previous
}

func (s *relayServer) unregisterClient(client *clientConn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if current, ok := s.clientsByUserID[client.userID]; ok && current == client {
		delete(s.clientsByUserID, client.userID)
	}
	if current, ok := s.clientsByFriendCode[client.friendCode]; ok && current == client {
		delete(s.clientsByFriendCode, client.friendCode)
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

	switch envelope.Type {
	case protocol.EventMessageSend:
		return s.handleMessageSend(client, envelope)
	case protocol.EventConversationCreate:
		return s.handleConversationCreate(client, envelope)
	case protocol.EventFriendRequestSend:
		return s.handleFriendRequestSend(client, envelope)
	case protocol.EventFriendRequestAccept:
		return s.handleFriendRequestAccept(client, envelope)
	case protocol.EventFriendRequestReject:
		return s.handleFriendRequestReject(client, envelope)
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

func (s *relayServer) getClientByFriendCode(friendCode string) *clientConn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clientsByFriendCode[friendCode]
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
