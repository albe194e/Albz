package app

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
	"github.com/albe194e/albz/shared/protocol"
	"github.com/google/uuid"
)

const (
	ConversationTypeDirect = "direct"
	ConversationTypeRoom   = "room"
)

func (c *Controller) GetConversationsByMe(ctx context.Context) ([]sql.Conversation, error) {
	return c.Store.Q.GetConversationsByUserID(ctx, c.State.CurrentUser.UserID)
}

func (c *Controller) CreateConversation(ctx context.Context, name string) error {
	conversationName := strings.TrimSpace(name)
	if conversationName == "" {
		conversationName = "New Conversation"
	}

	params := sql.CreateConversationParams{
		ID:   uuid.NewString(),
		Name: conversationName,
	}
	_, err := c.ensureConversationRecord(ctx, params.ID, conversationName, c.State.CurrentUser.UserID)
	return err
}

func (c *Controller) CreateConversationWithUsers(ctx context.Context, name string, userIDs ...string) (sql.Conversation, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return sql.Conversation{}, fmt.Errorf("no current user")
	}

	participantUserIDs, recipientUserIDs, err := c.normalizeConversationParticipantIDs(userIDs...)
	if err != nil {
		return sql.Conversation{}, err
	}

	conversationName := strings.TrimSpace(name)
	if conversationName == "" {
		conversationName = c.conversationDisplayName("", participantUserIDs)
	}

	params := sql.CreateConversationParams{
		ID:   uuid.NewString(),
		Name: conversationName,
	}
	conversation, err := c.ensureConversationRecord(ctx, params.ID, conversationName, participantUserIDs...)
	if err != nil {
		return sql.Conversation{}, err
	}

	c.broadcastConversationCreated(recipientUserIDs, params.ID, strings.TrimSpace(name))
	return conversation, nil
}

func (c *Controller) CreateConversationWithContact(ctx context.Context, contact sql.Contact) (sql.Conversation, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return sql.Conversation{}, fmt.Errorf("no current user")
	}

	existing, err := c.findDirectConversationWithUser(ctx, contact.UserID)
	if err != nil {
		return sql.Conversation{}, err
	}
	if existing != nil {
		return *existing, nil
	}

	return c.CreateConversationWithContacts(ctx, "", contact)
}

func (c *Controller) HandleConversationCreated(event protocol.Envelope[protocol.ConversationCreatedPayload]) {
	if c == nil || c.Store == nil || c.State == nil || c.State.CurrentUser == nil {
		return
	}

	participantUserIDs := event.Payload.ParticipantUserIDs
	if len(participantUserIDs) == 0 {
		participantUserIDs = []string{c.State.CurrentUser.UserID, event.Payload.FromUserID}
	}

	conversationName := c.conversationDisplayName(event.Payload.ConversationName, participantUserIDs)
	if _, err := c.ensureConversationRecord(
		context.Background(),
		event.Payload.ConversationID,
		conversationName,
		participantUserIDs...,
	); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("store incoming conversation: %v", err)
		c.notifyStateChanged()
		return
	}

	c.notifyStateChanged()
}

func (c *Controller) CreateConversationWithContacts(ctx context.Context, name string, contacts ...sql.Contact) (sql.Conversation, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return sql.Conversation{}, fmt.Errorf("no current user")
	}
	if len(contacts) == 0 {
		return sql.Conversation{}, fmt.Errorf("at least one contact is required")
	}

	userIDs := make([]string, 0, len(contacts))
	for _, contact := range contacts {
		userIDs = append(userIDs, contact.UserID)
	}

	participantUserIDs, recipientUserIDs, err := c.normalizeConversationParticipantIDs(userIDs...)
	if err != nil {
		return sql.Conversation{}, err
	}

	conversationName := c.conversationDisplayName(strings.TrimSpace(name), participantUserIDs)
	conversationID := uuid.NewString()
	conversation, err := c.ensureConversationRecord(ctx, conversationID, conversationName, participantUserIDs...)
	if err != nil {
		return sql.Conversation{}, err
	}

	c.broadcastConversationCreated(recipientUserIDs, conversationID, strings.TrimSpace(name))
	return conversation, nil
}

func (c *Controller) displayNameForConversationCreator(userID string, fallback string) string {
	contact := c.findContactByUserID(userID)
	if contact != nil {
		return ContactDisplayName(*contact)
	}

	if strings.TrimSpace(fallback) != "" {
		return fallback
	}

	return userID
}

func (c *Controller) conversationDisplayName(customName string, participantUserIDs []string) string {
	if trimmed := strings.TrimSpace(customName); trimmed != "" {
		return trimmed
	}

	otherNames := make([]string, 0, len(participantUserIDs))
	currentUserID := ""
	if c != nil && c.State != nil && c.State.CurrentUser != nil {
		currentUserID = c.State.CurrentUser.UserID
	}

	seen := make(map[string]struct{}, len(participantUserIDs))
	for _, participantUserID := range participantUserIDs {
		trimmed := strings.TrimSpace(participantUserID)
		if trimmed == "" || trimmed == currentUserID {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}

		seen[trimmed] = struct{}{}
		otherNames = append(otherNames, c.displayNameForConversationCreator(trimmed, trimmed))
	}

	switch len(otherNames) {
	case 0:
		return "New Conversation"
	case 1:
		return otherNames[0]
	case 2:
		return fmt.Sprintf("%s, %s", otherNames[0], otherNames[1])
	default:
		return fmt.Sprintf("%s, %s +%d", otherNames[0], otherNames[1], len(otherNames)-2)
	}
}

func (c *Controller) normalizeConversationParticipantIDs(userIDs ...string) ([]string, []string, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return nil, nil, fmt.Errorf("no current user")
	}

	participantUserIDs := []string{c.State.CurrentUser.UserID}
	recipientUserIDs := make([]string, 0, len(userIDs))
	seen := map[string]struct{}{
		c.State.CurrentUser.UserID: {},
	}

	for _, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}

		seen[trimmed] = struct{}{}
		participantUserIDs = append(participantUserIDs, trimmed)
		recipientUserIDs = append(recipientUserIDs, trimmed)
	}

	if len(recipientUserIDs) == 0 {
		return nil, nil, fmt.Errorf("at least one other participant is required")
	}

	slices.Sort(recipientUserIDs)
	return participantUserIDs, recipientUserIDs, nil
}

func (c *Controller) broadcastConversationCreated(recipientUserIDs []string, conversationID string, customName string) {
	if c == nil || c.Net == nil || c.State == nil {
		return
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(context.Background()); err != nil {
			c.State.LastNetworkError = err.Error()
			c.notifyStateChanged()
			return
		}
	}

	if err := c.Net.CreateConversation(uuid.NewString(), protocol.ConversationCreatePayload{
		ConversationID:   conversationID,
		ConversationName: customName,
		ToUserIDs:        recipientUserIDs,
	}); err != nil {
		c.State.LastNetworkError = err.Error()
		c.notifyStateChanged()
	}
}

func (c *Controller) ensureConversationRecord(ctx context.Context, conversationID string, name string, participantIDs ...string) (sql.Conversation, error) {
	if strings.TrimSpace(conversationID) == "" {
		return sql.Conversation{}, fmt.Errorf("conversation ID is required")
	}
	if strings.TrimSpace(name) == "" {
		return sql.Conversation{}, fmt.Errorf("conversation name is required")
	}

	conversationType := ConversationTypeRoom
	if len(participantIDs) == 2 {
		conversationType = ConversationTypeDirect
	}

	conversation, err := c.Store.Q.GetConversationByID(ctx, conversationID)
	if err != nil {
		if !errors.Is(err, dsql.ErrNoRows) {
			return sql.Conversation{}, err
		}

		conversation, err = c.Store.Q.CreateConversation(ctx, sql.CreateConversationParams{
			ID:        conversationID,
			Name:      name,
			Type:      conversationType,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: dsql.NullInt64{},
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
			UserID:         participantID,
			CreatedAt:      time.Now().Unix(),
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

	conversations, err := c.Store.Q.GetConversationsByUserID(ctx, c.State.CurrentUser.UserID)
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
			if participantID == c.State.CurrentUser.UserID {
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
