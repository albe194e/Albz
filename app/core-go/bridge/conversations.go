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

func (b *Bridge) StartDirectConversation(friendUserID string) (string, error) {
	friend := b.controller().FindFriendByUserID(friendUserID)
	if friend == nil {
		return "", fmt.Errorf("friend %q was not found", friendUserID)
	}

	conversation, err := b.controller().CreateConversationWithFriend(context.Background(), *friend)
	if err != nil {
		return "", err
	}

	b.emitStateChanged()
	return conversation.ID, nil
}
