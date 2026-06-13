package app

import (
	"context"
	"fmt"

	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
	"github.com/albe194e/albz/app/core-go/db/storage"
)

type AppState struct {
	CurrentUser          *sql.LocalIdentity
	Messages             []sql.Message
	Conversations        []sql.Conversation
	Contacts             []sql.Contact
	ContactRequests      []sql.ContactRequest
	LoadedConversationID string
	ServerConnected      bool
	LastNetworkError     string
}

func (s *AppState) InitStateFromDB(ctx context.Context, store *storage.Store) error {
	if s.CurrentUser == nil {
		return fmt.Errorf("CurrentUser is nil")
	}

	convs, err := store.Q.GetConversationsByUserID(ctx, s.CurrentUser.UserID)
	if err != nil {
		return err
	}

	contacts, err := store.Q.ListContacts(ctx)
	if err != nil {
		return err
	}

	contactRequests, err := store.Q.ListContactRequests(ctx)
	if err != nil {
		return err
	}

	s.Conversations = convs
	s.Contacts = contacts
	s.ContactRequests = contactRequests
	s.Messages = nil
	s.LoadedConversationID = ""
	return nil
}
