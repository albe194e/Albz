package bridge

import (
	"context"
	"fmt"
	"strings"
)

func (b *Bridge) OpenConversation(conversationID string) error {
	_, err := b.controller().LoadConversationMessages(context.Background(), conversationID)
	return err
}

func (b *Bridge) CreateConversation(name string, participantUserIDs []string) (string, error) {
	if len(participantUserIDs) == 0 {
		return "", fmt.Errorf("at least one other participant is required")
	}

	conversation, err := b.controller().CreateConversationWithUsers(
		context.Background(),
		strings.TrimSpace(name),
		participantUserIDs...,
	)
	if err != nil {
		return "", err
	}

	b.emitStateChanged()
	return conversation.ID, nil
}

func (b *Bridge) StartDirectConversation(contactUserID string) (string, error) {
	contact := b.controller().FindContactByUserID(contactUserID)
	if contact == nil {
		return "", fmt.Errorf("contact %q was not found", contactUserID)
	}

	conversation, err := b.controller().CreateConversationWithContact(context.Background(), *contact)
	if err != nil {
		return "", err
	}

	b.emitStateChanged()
	return conversation.ID, nil
}
