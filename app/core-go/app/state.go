package app

import (
	"context"
	"fmt"

	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
	"github.com/albe194e/albz/app/core-go/db/storage"
)

type AppState struct {
	CurrentUser          *sql.User
	Messages             []sql.Message
	Conversations        []sql.Conversation
	Friends              []sql.Friend
	FriendRequests       []sql.FriendRequest
	LoadedConversationID string
	ServerConnected      bool
	LastNetworkError     string
}

func (s *AppState) InitStateFromDB(ctx context.Context, store *storage.Store) error {
	if s.CurrentUser == nil {
		return fmt.Errorf("CurrentUser is nil")
	}

	convs, err := store.Q.GetConversationsByUserID(ctx, s.CurrentUser.ID)
	if err != nil {
		return err
	}

	friends, err := store.Q.ListFriends(ctx)
	if err != nil {
		return err
	}

	friendRequests, err := store.Q.ListFriendRequests(ctx)
	if err != nil {
		return err
	}

	s.Conversations = convs
	s.Friends = friends
	s.FriendRequests = friendRequests
	s.Messages = nil
	s.LoadedConversationID = ""
	return nil
}
