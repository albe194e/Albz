package controllers

import (
	"context"
	dsql "database/sql"
	"fmt"
	"strings"
	"time"

	domainmessages "github.com/albe194e/albz/app/core-go/domains/messages"
	"github.com/albe194e/albz/shared/protocol"
	"github.com/google/uuid"
)

func (c *Controller) LoadConversationMessages(ctx context.Context, conversationID string) ([]domainmessages.Message, error) {
	if strings.TrimSpace(conversationID) == "" {
		c.State.Messages = nil
		c.State.LoadedConversationID = ""
		c.notifyStateChanged()
		return nil, nil
	}

	messageList, err := c.MessageService.ListByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	c.State.Messages = messageList
	c.State.LoadedConversationID = conversationID
	c.notifyStateChanged()
	return messageList, nil
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

	err = c.MessageService.Create(ctx, domainmessages.CreateParams{
		ID:              messageID,
		ConversationID:  conversationID,
		SenderUserID:    c.State.CurrentUser.UserID,
		SenderDeviceID:  nullString(c.State.CurrentUser.DeviceID),
		ClientMessageID: clientMessageID,
		Body:            trimmedBody,
		CreatedAt:       now,
		ReceivedAt:      dsql.NullInt64{},
		Direction:       domainmessages.DirectionOutgoing,
		DeliveryState:   domainmessages.DeliveryStateSending,
	})
	if err != nil {
		return err
	}

	if _, err := c.LoadConversationMessages(ctx, conversationID); err != nil {
		return err
	}

	if c.Relay == nil {
		return fmt.Errorf("network client is not configured")
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(ctx); err != nil {
			_ = c.MessageService.UpdateDeliveryState(ctx, domainmessages.UpdateDeliveryStateParams{
				DeliveryState:   domainmessages.DeliveryStateFailed,
				ClientMessageID: clientMessageID,
			})
			_, _ = c.LoadConversationMessages(ctx, conversationID)
			return err
		}
	}

	err = c.Relay.SendMessage(uuid.NewString(), protocol.MessageSendPayload{
		ClientMessageID: clientMessageID,
		ConversationID:  conversationID,
		ToUserIDs:       recipientUserIDs,
		Body:            trimmedBody,
		SentAt:          now,
	})
	if err != nil {
		_ = c.MessageService.UpdateDeliveryState(ctx, domainmessages.UpdateDeliveryStateParams{
			DeliveryState:   domainmessages.DeliveryStateFailed,
			ClientMessageID: clientMessageID,
		})
		_, _ = c.LoadConversationMessages(ctx, conversationID)
		return err
	}

	return nil
}

func (c *Controller) resolveConversationRecipients(ctx context.Context, conversationID string) ([]string, error) {
	participantIDs, err := c.ConversationService.ListParticipantIDs(ctx, conversationID)
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
