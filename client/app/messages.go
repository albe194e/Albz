package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/albe194e/albz/client/db/sqlc/sql"
	"github.com/albe194e/albz/shared/protocol"
	"github.com/google/uuid"
)

const (
	MessageDeliveryStatePending = "pending"
	MessageDeliveryStateFailed  = "failed"
)

func (c *Controller) LoadConversationMessages(ctx context.Context, conversationID string) ([]sql.Message, error) {
	if strings.TrimSpace(conversationID) == "" {
		c.State.Messages = nil
		c.State.LoadedConversationID = ""
		c.notifyStateChanged()
		return nil, nil
	}

	messages, err := c.Store.Q.ListMessagesByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	c.State.Messages = messages
	c.State.LoadedConversationID = conversationID
	c.notifyStateChanged()
	return messages, nil
}

func (c *Controller) AddMessage(ctx context.Context, conversationID string, body string) error {
	if c.State.CurrentUser == nil {
		return fmt.Errorf("no current user")
	}
	if strings.TrimSpace(conversationID) == "" {
		return fmt.Errorf("no active conversation selected")
	}

	trimmedBody := strings.TrimSpace(body)
	if trimmedBody == "" {
		return fmt.Errorf("message body is required")
	}

	recipientID, err := c.resolveConversationRecipient(ctx, conversationID)
	if err != nil {
		return err
	}

	clientMessageID := uuid.NewString()
	now := time.Now().Unix()

	err = c.Store.Q.CreateMessage(ctx, sql.CreateMessageParams{
		ConversationID:  conversationID,
		SenderID:        c.State.CurrentUser.ID,
		ClientMessageID: clientMessageID,
		Body:            trimmedBody,
		CreatedAt:       now,
		DeliveryState:   MessageDeliveryStatePending,
	})
	if err != nil {
		return err
	}

	if _, err := c.LoadConversationMessages(ctx, conversationID); err != nil {
		return err
	}

	if c.Net == nil {
		return fmt.Errorf("network client is not configured")
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(ctx); err != nil {
			_ = c.Store.Q.UpdateMessageDeliveryStateByClientMessageID(ctx, sql.UpdateMessageDeliveryStateByClientMessageIDParams{
				DeliveryState:   MessageDeliveryStateFailed,
				ClientMessageID: clientMessageID,
			})
			_, _ = c.LoadConversationMessages(ctx, conversationID)
			return err
		}
	}

	err = c.Net.SendMessage(uuid.NewString(), protocol.MessageSendPayload{
		ClientMessageID: clientMessageID,
		ConversationID:  conversationID,
		ToUserID:        recipientID,
		Body:            trimmedBody,
		SentAt:          now,
	})
	if err != nil {
		_ = c.Store.Q.UpdateMessageDeliveryStateByClientMessageID(ctx, sql.UpdateMessageDeliveryStateByClientMessageIDParams{
			DeliveryState:   MessageDeliveryStateFailed,
			ClientMessageID: clientMessageID,
		})
		_, _ = c.LoadConversationMessages(ctx, conversationID)
		return err
	}

	return nil
}

func (c *Controller) resolveConversationRecipient(ctx context.Context, conversationID string) (string, error) {
	participantIDs, err := c.Store.Q.ListConversationParticipantIDs(ctx, conversationID)
	if err != nil {
		return "", err
	}

	recipientID := ""
	for _, participantID := range participantIDs {
		if participantID == c.State.CurrentUser.ID {
			continue
		}

		if recipientID != "" {
			return "", fmt.Errorf("conversation %s has multiple recipients; group chat is not supported yet", conversationID)
		}

		recipientID = participantID
	}

	if recipientID == "" {
		return "", fmt.Errorf("conversation %s has no recipient participant", conversationID)
	}

	return recipientID, nil
}
