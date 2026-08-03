package bridge

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/albe194e/albz/app/core-go/controllers"
	domaincontacts "github.com/albe194e/albz/app/core-go/domains/contacts"
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

func snapshotFromState(state *controllers.AppState) Snapshot {
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
		localHandle := nullableStringValue(state.CurrentUser.LocalHandle)
		profilePicturePath := nullableStringValue(state.CurrentUser.ProfilePicturePath)
		contactCode := nullableStringValue(state.CurrentUser.ContactCode)

		snapshot.CurrentUser = &User{
			ID:                 state.CurrentUser.UserID,
			UserID:             state.CurrentUser.UserID,
			DeviceID:           state.CurrentUser.DeviceID,
			Name:               state.CurrentUser.Name,
			LocalHandle:        localHandle,
			Username:           localHandle,
			ProfilePicturePath: profilePicturePath,
			ProfilePictureUrl:  profilePicturePath,
			ContactCode:        contactCode,
		}
	}

	for _, message := range state.Messages {
		snapshot.Messages = append(snapshot.Messages, Message{
			ID:              message.ID,
			ConversationID:  message.ConversationID,
			SenderUserID:    message.SenderUserID,
			SenderID:        message.SenderUserID,
			SenderDeviceID:  nullableStringValue(message.SenderDeviceID),
			ClientMessageID: message.ClientMessageID,
			Body:            message.Body,
			CreatedAt:       message.CreatedAt,
			ReceivedAt:      nullableInt64Value(message.ReceivedAt),
			Direction:       message.Direction,
			DeliveryState:   message.DeliveryState,
		})
	}

	for _, conversation := range state.Conversations {
		snapshot.Conversations = append(snapshot.Conversations, Conversation{
			ID:        conversation.ID,
			Name:      conversation.Name,
			Type:      conversation.Type,
			CreatedAt: conversation.CreatedAt,
			UpdatedAt: nullableInt64Value(conversation.UpdatedAt),
		})
	}

	for _, contact := range state.Contacts {
		snapshot.Contacts = append(snapshot.Contacts, mapContact(contact))
	}

	for _, request := range state.ContactRequests {
		localHandle := nullableStringValue(request.LocalHandle)
		fromContactCode := nullableStringValue(request.FromContactCode)
		profilePicturePath := nullableStringValue(request.ProfilePicturePath)

		snapshot.ContactRequests = append(snapshot.ContactRequests, ContactRequest{
			ID:                 request.ID,
			FromUserID:         request.FromUserID,
			FromDeviceID:       nullableStringValue(request.FromDeviceID),
			DisplayName:        request.DisplayName,
			Name:               request.DisplayName,
			LocalHandle:        localHandle,
			Username:           localHandle,
			ProfilePicturePath: profilePicturePath,
			FromContactCode:    fromContactCode,
			InvitePayload:      request.InvitePayload,
			State:              request.State,
			CreatedAt:          request.CreatedAt,
		})
	}

	return snapshot
}

func mapContact(contact domaincontacts.Contact) Contact {
	localHandle := nullableStringValue(contact.LocalHandle)
	profilePicturePath := nullableStringValue(contact.ProfilePicturePath)
	contactCode := nullableStringValue(contact.ContactCode)

	return Contact{
		ID:                 contact.ID,
		UserID:             contact.UserID,
		DisplayName:        contact.DisplayName,
		Name:               contact.DisplayName,
		LocalHandle:        localHandle,
		Username:           localHandle,
		ProfilePicturePath: profilePicturePath,
		ProfilePictureUrl:  profilePicturePath,
		ContactCode:        contactCode,
		CreatedAt:          contact.CreatedAt,
	}
}

func nullableStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func nullableInt64Value(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}

	return value.Int64
}
