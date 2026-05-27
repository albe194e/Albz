package app

import (
	"context"
	"fmt"

	"github.com/albe194e/albz/client/db/sqlc/sql"
	"github.com/albe194e/albz/client/db/storage"
)

type AppState struct {
	CurrentUser          *sql.User
	ActiveConversationID string
	Messages             []sql.Message
	Conversations        []sql.Conversation
}

func (s *AppState) InitStateFromDB(ctx context.Context, store *storage.Store) error {
	if s.CurrentUser == nil {
		return fmt.Errorf("CurrentUser is nil")
	}

	convs, err := store.Q.GetConversationsByUserID(ctx, s.CurrentUser.ID)
	if err != nil {
		return err
	}

	s.Conversations = convs
	return nil
}
