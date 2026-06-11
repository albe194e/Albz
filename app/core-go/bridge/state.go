package bridge

import (
	"encoding/json"
	"fmt"

	clientapp "github.com/albe194e/albz/app/core-go/app"
	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
)

func (b *Bridge) SetEventSink(sink EventSink) error {
	if b == nil {
		return fmt.Errorf("bridge is nil")
	}

	b.sinkMu.Lock()
	b.sink = sink
	b.sinkMu.Unlock()

	b.emitStateChanged()
	return nil
}

func (b *Bridge) Config() Config {
	if b == nil || b.service == nil {
		return Config{}
	}

	return Config{
		ProfileName: b.service.Config.ProfileName,
		DataDir:     b.service.Config.DataDir,
		DBPath:      b.service.Config.DBPath,
		ServerURL:   b.service.Config.ServerURL,
	}
}

func (b *Bridge) ConfigJSON() (string, error) {
	data, err := json.Marshal(b.Config())
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (b *Bridge) Snapshot() Snapshot {
	if b == nil || b.service == nil || b.service.Controller == nil || b.service.Controller.State == nil {
		return Snapshot{}
	}

	return snapshotFromState(b.service.Controller.State)
}

func (b *Bridge) SnapshotJSON() (string, error) {
	data, err := json.Marshal(b.Snapshot())
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (b *Bridge) emitStateChanged() {
	b.sinkMu.RLock()
	sink := b.sink
	b.sinkMu.RUnlock()
	if sink == nil {
		return
	}

	data, err := json.Marshal(Event{
		Type:     EventTypeStateChanged,
		Snapshot: snapshotPointer(b.Snapshot()),
	})
	if err != nil {
		return
	}

	sink.OnCoreEvent(string(data))
}

func snapshotPointer(snapshot Snapshot) *Snapshot {
	return &snapshot
}

func snapshotFromState(state *clientapp.AppState) Snapshot {
	if state == nil {
		return Snapshot{}
	}

	snapshot := Snapshot{
		Messages:             make([]Message, 0, len(state.Messages)),
		Conversations:        make([]Conversation, 0, len(state.Conversations)),
		Contacts:             make([]Contact, 0, len(state.Contacts)),
		ContactRequests:      make([]ContactRequest, 0, len(state.ContactRequests)),
		LoadedConversationID: state.LoadedConversationID,
		ServerConnected:      state.ServerConnected,
		LastNetworkError:     state.LastNetworkError,
	}

	if state.CurrentUser != nil {
		snapshot.CurrentUser = &User{
			ID:                state.CurrentUser.ID,
			Name:              state.CurrentUser.Name,
			Username:          state.CurrentUser.Username,
			ProfilePictureUrl: state.CurrentUser.ProfilePictureUrl,
			ContactCode:       state.CurrentUser.ContactCode,
		}
	}

	for _, message := range state.Messages {
		snapshot.Messages = append(snapshot.Messages, Message{
			ID:              message.ID,
			ConversationID:  message.ConversationID,
			SenderID:        message.SenderID,
			ClientMessageID: message.ClientMessageID,
			Body:            message.Body,
			CreatedAt:       message.CreatedAt,
			DeliveryState:   message.DeliveryState,
		})
	}

	for _, conversation := range state.Conversations {
		snapshot.Conversations = append(snapshot.Conversations, Conversation{
			ID:   conversation.ID,
			Name: conversation.Name,
		})
	}

	for _, contact := range state.Contacts {
		snapshot.Contacts = append(snapshot.Contacts, mapContact(contact))
	}

	for _, request := range state.ContactRequests {
		snapshot.ContactRequests = append(snapshot.ContactRequests, ContactRequest{
			ID:              request.ID,
			FromUserID:      request.FromUserID,
			Name:            request.Name,
			Username:        request.Username,
			FromContactCode: request.FromContactCode,
			CreatedAt:       request.CreatedAt,
		})
	}

	return snapshot
}

func mapContact(contact sql.Contact) Contact {
	return Contact{
		ID:                contact.ID,
		UserID:            contact.UserID,
		Name:              contact.Name,
		Username:          contact.Username,
		ProfilePictureUrl: contact.ProfilePictureUrl,
		ContactCode:       contact.ContactCode,
		CreatedAt:         contact.CreatedAt,
	}
}
