package app

import (
	"context"
	dsql "database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
	"github.com/albe194e/albz/shared/protocol"
	"github.com/google/uuid"
)

const (
	MessageDirectionIncoming    = "incoming"
	MessageDirectionOutgoing    = "outgoing"
	MessageDeliveryStateSending = "sending"
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

	recipientUserIDs, err := c.resolveConversationRecipients(ctx, conversationID)
	if err != nil {
		return err
	}

	clientMessageID := uuid.NewString()
	messageID := uuid.NewString()
	now := time.Now().Unix()

	err = c.Store.Q.CreateMessage(ctx, sql.CreateMessageParams{
		ID:              messageID,
		ConversationID:  conversationID,
		SenderUserID:    c.State.CurrentUser.UserID,
		SenderDeviceID:  nullString(c.State.CurrentUser.DeviceID),
		ClientMessageID: clientMessageID,
		Body:            trimmedBody,
		CreatedAt:       now,
		ReceivedAt:      dsql.NullInt64{},
		Direction:       MessageDirectionOutgoing,
		DeliveryState:   MessageDeliveryStateSending,
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
		ToUserIDs:       recipientUserIDs,
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

func (c *Controller) resolveConversationRecipients(ctx context.Context, conversationID string) ([]string, error) {
	participantIDs, err := c.Store.Q.ListConversationParticipantIDs(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	recipientIDs := make([]string, 0, len(participantIDs))
	for _, participantID := range participantIDs {
		if participantID == c.State.CurrentUser.UserID {
			continue
		}

		recipientIDs = append(recipientIDs, participantID)
	}

	if len(recipientIDs) == 0 {
		return nil, fmt.Errorf("conversation %s has no recipient participant", conversationID)
	}

	return recipientIDs, nil
}
