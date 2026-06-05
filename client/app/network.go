package app

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"

	"github.com/albe194e/albz/client/db/sqlc/sql"
	"github.com/albe194e/albz/shared/protocol"
)

func (c *Controller) ConnectToServer(ctx context.Context) error {
	if c == nil || c.Net == nil || c.State == nil || c.State.CurrentUser == nil {
		return nil
	}

	if err := c.Net.Connect(c.State.CurrentUser.ID, c.State.CurrentUser.FriendCode); err != nil {
		c.State.ServerConnected = false
		c.State.LastNetworkError = err.Error()
		c.notifyStateChanged()
		return err
	}

	c.State.ServerConnected = true
	c.State.LastNetworkError = ""
	c.notifyStateChanged()
	return nil
}

func (c *Controller) HandleIncomingMessage(event protocol.Envelope[protocol.MessageCreatedPayload]) {
	if c == nil || c.Store == nil {
		return
	}

	if _, err := c.Store.Q.GetConversationByID(context.Background(), event.Payload.ConversationID); err != nil {
		if !errors.Is(err, dsql.ErrNoRows) {
			c.State.LastNetworkError = fmt.Sprintf("load incoming conversation: %v", err)
			c.notifyStateChanged()
			return
		}

		conversationName := c.displayNameForConversationCreator(event.Payload.FromUserID, "")

		if _, err := c.ensureConversationRecord(
			context.Background(),
			event.Payload.ConversationID,
			conversationName,
			c.State.CurrentUser.ID,
			event.Payload.FromUserID,
		); err != nil {
			c.State.LastNetworkError = fmt.Sprintf("create incoming conversation: %v", err)
			c.notifyStateChanged()
			return
		}
	}

	err := c.Store.Q.CreateMessage(context.Background(), sql.CreateMessageParams{
		ConversationID:  event.Payload.ConversationID,
		SenderID:        event.Payload.FromUserID,
		ClientMessageID: event.Payload.MessageID,
		Body:            event.Payload.Body,
		CreatedAt:       event.Payload.SentAt,
		DeliveryState:   string(protocol.DeliveryStatusDelivered),
	})
	if err != nil {
		c.State.LastNetworkError = fmt.Sprintf("store incoming message: %v", err)
		c.notifyStateChanged()
		return
	}

	if c.State.LoadedConversationID == event.Payload.ConversationID {
		if _, err := c.LoadConversationMessages(context.Background(), event.Payload.ConversationID); err != nil {
			c.State.LastNetworkError = fmt.Sprintf("reload conversation messages: %v", err)
		}
	}

	c.notifyStateChanged()
}

func (c *Controller) HandleMessageDelivery(event protocol.Envelope[protocol.MessageDeliveryPayload]) {
	if c == nil || c.Store == nil {
		return
	}

	err := c.Store.Q.UpdateMessageDeliveryStateByClientMessageID(context.Background(), sql.UpdateMessageDeliveryStateByClientMessageIDParams{
		DeliveryState:   string(event.Payload.Status),
		ClientMessageID: event.Payload.ClientMessageID,
	})
	if err != nil {
		c.State.LastNetworkError = fmt.Sprintf("update delivery state: %v", err)
		c.notifyStateChanged()
		return
	}

	if c.State.LoadedConversationID != "" {
		if _, err := c.LoadConversationMessages(context.Background(), c.State.LoadedConversationID); err != nil {
			c.State.LastNetworkError = fmt.Sprintf("reload messages after delivery update: %v", err)
		}
	}

	c.notifyStateChanged()
}

func (c *Controller) HandleFriendRequestReceived(event protocol.Envelope[protocol.FriendRequestReceivedPayload]) {
	if c == nil || c.Store == nil {
		return
	}

	err := c.Store.Q.UpsertFriendRequest(context.Background(), sql.UpsertFriendRequestParams{
		FromUserID:     event.Payload.FromProfile.UserID,
		Name:           event.Payload.FromProfile.Name,
		Username:       event.Payload.FromProfile.Username,
		FromFriendCode: event.Payload.FromProfile.FriendCode,
		CreatedAt:      event.Timestamp,
	})
	if err != nil {
		c.State.LastNetworkError = fmt.Sprintf("store friend request: %v", err)
		c.notifyStateChanged()
		return
	}

	if err := c.LoadSocialState(context.Background()); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("reload social state: %v", err)
	}

	c.notifyStateChanged()
}

func (c *Controller) HandleFriendRequestAccepted(event protocol.Envelope[protocol.FriendRequestAcceptedPayload]) {
	if c == nil || c.Store == nil {
		return
	}

	profile := event.Payload.Profile
	if profile.UserID == "" {
		c.State.LastNetworkError = "accepted friend is missing user ID"
		c.notifyStateChanged()
		return
	}

	if profile.Name == "" || profile.Username == "" || profile.FriendCode == "" {
		request, err := c.findFriendRequestByUserID(profile.UserID)
		if err != nil {
			c.State.LastNetworkError = fmt.Sprintf("load accepted friend request: %v", err)
			c.notifyStateChanged()
			return
		}
		if request != nil {
			if profile.Name == "" {
				profile.Name = request.Name
			}
			if profile.Username == "" {
				profile.Username = request.Username
			}
			if profile.FriendCode == "" {
				profile.FriendCode = request.FromFriendCode
			}
		}
	}

	if err := c.Store.Q.UpsertFriend(context.Background(), sql.UpsertFriendParams{
		UserID:     profile.UserID,
		Name:       profile.Name,
		Username:   profile.Username,
		FriendCode: profile.FriendCode,
		CreatedAt:  event.Timestamp,
	}); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("store friend: %v", err)
		c.notifyStateChanged()
		return
	}

	_ = c.Store.Q.DeleteFriendRequestByFromUserID(context.Background(), profile.UserID)

	if err := c.LoadSocialState(context.Background()); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("reload social state: %v", err)
	}

	c.notifyStateChanged()
}

func (c *Controller) HandleFriendRequestRejected(event protocol.Envelope[protocol.FriendRequestRejectedPayload]) {
	if c == nil || c.Store == nil {
		return
	}

	if err := c.Store.Q.DeleteFriendRequestByFromUserID(context.Background(), event.Payload.UserID); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("delete rejected friend request: %v", err)
		c.notifyStateChanged()
		return
	}

	if err := c.LoadSocialState(context.Background()); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("reload social state: %v", err)
	}

	c.notifyStateChanged()
}

func (c *Controller) HandleNetworkError(event protocol.Envelope[protocol.ErrorPayload]) {
	if c == nil || c.State == nil {
		return
	}

	c.State.LastNetworkError = event.Payload.Message
	c.notifyStateChanged()
}

func (c *Controller) HandleDisconnect(err error) {
	if c == nil || c.State == nil {
		return
	}

	c.State.ServerConnected = false
	if err != nil {
		c.State.LastNetworkError = err.Error()
	}
	c.notifyStateChanged()
}

func (c *Controller) findFriendRequestByUserID(userID string) (*sql.FriendRequest, error) {
	if c == nil || c.Store == nil {
		return nil, nil
	}

	requests, err := c.Store.Q.ListFriendRequests(context.Background())
	if err != nil {
		return nil, err
	}

	for _, request := range requests {
		if request.FromUserID == userID {
			requestCopy := request
			return &requestCopy, nil
		}
	}

	return nil, nil
}
