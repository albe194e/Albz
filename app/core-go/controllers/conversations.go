package controllers

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/albe194e/albz/app/core-go/domains/contacts"
	"github.com/albe194e/albz/app/core-go/domains/conversations"
	"github.com/albe194e/albz/shared/protocol"
	"github.com/google/uuid"
)

func (c *Controller) GetConversationsByMe(ctx context.Context) ([]conversations.Conversation, error) {
	return c.ConversationService.ListByUserID(ctx, c.State.CurrentUser.UserID)
}

func (c *Controller) CreateConversation(ctx context.Context, name string) error {
	conversationName := strings.TrimSpace(name)
	if conversationName == "" {
		conversationName = "New Conversation"
	}

	params := conversations.CreateParams{
		ID:   uuid.NewString(),
		Name: conversationName,
	}
	_, err := c.ensureConversationRecord(ctx, params.ID, conversationName, c.State.CurrentUser.UserID)
	return err
}

func (c *Controller) CreateConversationWithUsers(ctx context.Context, name string, userIDs ...string) (conversations.Conversation, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return conversations.Conversation{}, fmt.Errorf("no current user")
	}

	participantUserIDs, recipientUserIDs, err := c.normalizeConversationParticipantIDs(userIDs...)
	if err != nil {
		return conversations.Conversation{}, err
	}

	conversationName := strings.TrimSpace(name)
	if conversationName == "" {
		conversationName = c.conversationDisplayName("", participantUserIDs)
	}

	params := conversations.CreateParams{
		ID:   uuid.NewString(),
		Name: conversationName,
	}
	conversation, err := c.ensureConversationRecord(ctx, params.ID, conversationName, participantUserIDs...)
	if err != nil {
		return conversations.Conversation{}, err
	}

	c.broadcastConversationCreated(recipientUserIDs, params.ID, strings.TrimSpace(name))
	return conversation, nil
}

func (c *Controller) CreateConversationWithContact(ctx context.Context, contact contacts.Contact) (conversations.Conversation, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return conversations.Conversation{}, fmt.Errorf("no current user")
	}

	existing, err := c.findDirectConversationWithUser(ctx, contact.UserID)
	if err != nil {
		return conversations.Conversation{}, err
	}
	if existing != nil {
		return *existing, nil
	}

	return c.CreateConversationWithContacts(ctx, "", contact)
}

func (c *Controller) HandleConversationCreated(event protocol.Envelope[protocol.ConversationCreatedPayload]) {
	if c == nil || c.ConversationService == nil || c.State == nil || c.State.CurrentUser == nil {
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

func (c *Controller) CreateConversationWithContacts(ctx context.Context, name string, contactList ...contacts.Contact) (conversations.Conversation, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return conversations.Conversation{}, fmt.Errorf("no current user")
	}
	if len(contactList) == 0 {
		return conversations.Conversation{}, fmt.Errorf("at least one contact is required")
	}

	userIDs := make([]string, 0, len(contactList))
	for _, contact := range contactList {
		userIDs = append(userIDs, contact.UserID)
	}

	participantUserIDs, recipientUserIDs, err := c.normalizeConversationParticipantIDs(userIDs...)
	if err != nil {
		return conversations.Conversation{}, err
	}

	conversationName := c.conversationDisplayName(strings.TrimSpace(name), participantUserIDs)
	conversationID := uuid.NewString()
	conversation, err := c.ensureConversationRecord(ctx, conversationID, conversationName, participantUserIDs...)
	if err != nil {
		return conversations.Conversation{}, err
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
	if c == nil || c.Relay == nil || c.State == nil {
		return
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(context.Background()); err != nil {
			c.State.LastNetworkError = err.Error()
			c.notifyStateChanged()
			return
		}
	}

	if err := c.Relay.CreateConversation(uuid.NewString(), protocol.ConversationCreatePayload{
		ConversationID:   conversationID,
		ConversationName: customName,
		ToUserIDs:        recipientUserIDs,
	}); err != nil {
		c.State.LastNetworkError = err.Error()
		c.notifyStateChanged()
	}
}

func (c *Controller) ensureConversationRecord(ctx context.Context, conversationID string, name string, participantIDs ...string) (conversations.Conversation, error) {
	if strings.TrimSpace(conversationID) == "" {
		return conversations.Conversation{}, fmt.Errorf("conversation ID is required")
	}
	if strings.TrimSpace(name) == "" {
		return conversations.Conversation{}, fmt.Errorf("conversation name is required")
	}

	conversationType := conversations.TypeRoom
	if len(participantIDs) == 2 {
		conversationType = conversations.TypeDirect
	}

	conversation, err := c.ConversationService.GetByID(ctx, conversationID)
	if err != nil {
		if !errors.Is(err, dsql.ErrNoRows) {
			return conversations.Conversation{}, err
		}

		conversation, err = c.ConversationService.Create(ctx, conversations.CreateParams{
			ID:        conversationID,
			Name:      name,
			Type:      conversationType,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: dsql.NullInt64{},
		})
		if err != nil {
			return conversations.Conversation{}, err
		}
	}

	existingParticipants, err := c.ConversationService.ListParticipantIDs(ctx, conversationID)
	if err != nil {
		return conversations.Conversation{}, err
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

		if err := c.ConversationService.AddParticipant(ctx, conversations.AddParticipantParams{
			ConversationID: conversationID,
			UserID:         participantID,
			CreatedAt:      time.Now().Unix(),
		}); err != nil {
			return conversations.Conversation{}, err
		}
		seen[participantID] = struct{}{}
	}

	if err := c.reloadConversations(ctx); err != nil {
		return conversations.Conversation{}, err
	}

	return conversation, nil
}

func (c *Controller) reloadConversations(ctx context.Context) error {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return nil
	}

	conversationList, err := c.ConversationService.ListByUserID(ctx, c.State.CurrentUser.UserID)
	if err != nil {
		return err
	}

	c.State.Conversations = conversationList
	return nil
}

func (c *Controller) findDirectConversationWithUser(ctx context.Context, otherUserID string) (*conversations.Conversation, error) {
	if strings.TrimSpace(otherUserID) == "" {
		return nil, fmt.Errorf("other user ID is required")
	}

	for _, conversation := range c.State.Conversations {
		participantIDs, err := c.ConversationService.ListParticipantIDs(ctx, conversation.ID)
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
