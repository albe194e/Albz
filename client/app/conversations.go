package app

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/albe194e/albz/client/db/sqlc/sql"
	"github.com/albe194e/albz/shared/protocol"
	"github.com/google/uuid"
)

func (c *Controller) GetConversationsByMe(ctx context.Context) ([]sql.Conversation, error) {
	return c.Store.Q.GetConversationsByUserID(ctx, c.State.CurrentUser.ID)
}

func (c *Controller) CreateConversation(ctx context.Context, name string) error {
	params := sql.CreateConversationParams{
		ID:   uuid.New().String()[:16],
		Name: name,
	}
	_, err := c.ensureConversationRecord(ctx, params.ID, name, c.State.CurrentUser.ID)
	return err
}

func (c *Controller) CreateConversationWithFriend(ctx context.Context, friend sql.Friend) (sql.Conversation, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return sql.Conversation{}, fmt.Errorf("no current user")
	}

	existing, err := c.findDirectConversationWithUser(ctx, friend.UserID)
	if err != nil {
		return sql.Conversation{}, err
	}
	if existing != nil {
		return *existing, nil
	}

	conversationID := uuid.New().String()[:16]
	conversation, err := c.ensureConversationRecord(ctx, conversationID, FriendDisplayName(friend), c.State.CurrentUser.ID, friend.UserID)
	if err != nil {
		return sql.Conversation{}, err
	}

	if c.Net != nil {
		if !c.State.ServerConnected {
			if err := c.ConnectToServer(ctx); err != nil {
				c.State.LastNetworkError = err.Error()
				c.notifyStateChanged()
				return conversation, nil
			}
		}

		if err := c.Net.CreateConversation(uuid.NewString(), protocol.ConversationCreatePayload{
			ConversationID: conversationID,
			ToUserID:       friend.UserID,
		}); err != nil {
			c.State.LastNetworkError = err.Error()
			c.notifyStateChanged()
			return conversation, nil
		}
	}

	return conversation, nil
}

func (c *Controller) HandleConversationCreated(event protocol.Envelope[protocol.ConversationCreatedPayload]) {
	if c == nil || c.Store == nil || c.State == nil || c.State.CurrentUser == nil {
		return
	}

	if _, err := c.ensureConversationRecord(
		context.Background(),
		event.Payload.ConversationID,
		c.displayNameForConversationCreator(event.Payload.FromUserID, event.Payload.FromFriendCode),
		c.State.CurrentUser.ID,
		event.Payload.FromUserID,
	); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("store incoming conversation: %v", err)
		c.notifyStateChanged()
		return
	}

	c.notifyStateChanged()
}

func (c *Controller) displayNameForConversationCreator(userID string, fallback string) string {
	friend := c.findFriendByUserID(userID)
	if friend != nil {
		return FriendDisplayName(*friend)
	}

	if strings.TrimSpace(fallback) != "" {
		return fallback
	}

	return userID
}

func (c *Controller) ensureConversationRecord(ctx context.Context, conversationID string, name string, participantIDs ...string) (sql.Conversation, error) {
	if strings.TrimSpace(conversationID) == "" {
		return sql.Conversation{}, fmt.Errorf("conversation ID is required")
	}
	if strings.TrimSpace(name) == "" {
		return sql.Conversation{}, fmt.Errorf("conversation name is required")
	}

	conversation, err := c.Store.Q.GetConversationByID(ctx, conversationID)
	if err != nil {
		if !errors.Is(err, dsql.ErrNoRows) {
			return sql.Conversation{}, err
		}

		conversation, err = c.Store.Q.CreateConversation(ctx, sql.CreateConversationParams{
			ID:   conversationID,
			Name: name,
		})
		if err != nil {
			return sql.Conversation{}, err
		}
	}

	existingParticipants, err := c.Store.Q.ListConversationParticipantIDs(ctx, conversationID)
	if err != nil {
		return sql.Conversation{}, err
	}

	seen := make(map[string]struct{}, len(existingParticipants))
	for _, participantID := range existingParticipants {
		seen[participantID] = struct{}{}
	}

	for _, participantID := range participantIDs {
		if strings.TrimSpace(participantID) == "" {
			continue
		}
		if _, ok := seen[participantID]; ok {
			continue
		}

		if err := c.Store.Q.AddParticipant(ctx, sql.AddParticipantParams{
			ConversationID: conversationID,
			ParticipantID:  participantID,
		}); err != nil {
			return sql.Conversation{}, err
		}
		seen[participantID] = struct{}{}
	}

	if err := c.reloadConversations(ctx); err != nil {
		return sql.Conversation{}, err
	}

	return conversation, nil
}

func (c *Controller) reloadConversations(ctx context.Context) error {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return nil
	}

	conversations, err := c.Store.Q.GetConversationsByUserID(ctx, c.State.CurrentUser.ID)
	if err != nil {
		return err
	}

	c.State.Conversations = conversations
	return nil
}

func (c *Controller) findDirectConversationWithUser(ctx context.Context, otherUserID string) (*sql.Conversation, error) {
	if strings.TrimSpace(otherUserID) == "" {
		return nil, fmt.Errorf("other user ID is required")
	}

	for _, conversation := range c.State.Conversations {
		participantIDs, err := c.Store.Q.ListConversationParticipantIDs(ctx, conversation.ID)
		if err != nil {
			return nil, err
		}

		if len(participantIDs) != 2 {
			continue
		}

		hasCurrentUser := false
		hasOtherUser := false
		for _, participantID := range participantIDs {
			if participantID == c.State.CurrentUser.ID {
				hasCurrentUser = true
			}
			if participantID == otherUserID {
				hasOtherUser = true
			}
		}

		if hasCurrentUser && hasOtherUser {
			conversationCopy := conversation
			return &conversationCopy, nil
		}
	}

	return nil, nil
}
