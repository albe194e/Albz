package app

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
	"github.com/albe194e/albz/shared/protocol"
)

func (c *Controller) ConnectToServer(ctx context.Context) error {
	if c == nil || c.Net == nil || c.State == nil || c.State.CurrentUser == nil {
		return nil
	}
	if len(c.UnlockedDevicePrivateKey) == 0 {
		return fmt.Errorf("device key is locked; login is required before connecting to the relay")
	}

	if err := c.Net.Connect(
		c.State.CurrentUser.UserID,
		c.State.CurrentUser.DeviceID,
		nullStringValue(c.State.CurrentUser.ContactCode),
		c.State.CurrentUser.DevicePublicKey,
		c.UnlockedDevicePrivateKey,
	); err != nil {
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

		participantUserIDs := event.Payload.ParticipantUserIDs
		if len(participantUserIDs) == 0 {
			participantUserIDs = []string{c.State.CurrentUser.UserID, event.Payload.FromUserID}
		}

		conversationName := c.conversationDisplayName("", participantUserIDs)

		if _, err := c.ensureConversationRecord(
			context.Background(),
			event.Payload.ConversationID,
			conversationName,
			participantUserIDs...,
		); err != nil {
			c.State.LastNetworkError = fmt.Sprintf("create incoming conversation: %v", err)
			c.notifyStateChanged()
			return
		}
	}

	err := c.Store.Q.CreateMessage(context.Background(), sql.CreateMessageParams{
		ID:              event.Payload.MessageID,
		ConversationID:  event.Payload.ConversationID,
		SenderUserID:    event.Payload.FromUserID,
		SenderDeviceID:  nullString(strings.TrimSpace(event.Payload.FromDeviceID)),
		ClientMessageID: event.Payload.MessageID,
		Body:            event.Payload.Body,
		CreatedAt:       event.Payload.SentAt,
		ReceivedAt: dsql.NullInt64{
			Int64: event.Timestamp,
			Valid: event.Timestamp > 0,
		},
		Direction:     MessageDirectionIncoming,
		DeliveryState: string(protocol.DeliveryStatusDelivered),
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

func (c *Controller) HandleContactRequestReceived(event protocol.Envelope[protocol.ContactRequestReceivedPayload]) {
	if c == nil || c.Store == nil {
		return
	}

	contactCode := strings.TrimSpace(event.Payload.FromProfile.ContactCode)
	err := c.Store.Q.UpsertContactRequest(context.Background(), sql.UpsertContactRequestParams{
		FromUserID:      event.Payload.FromProfile.UserID,
		FromDeviceID:    nullString(strings.TrimSpace(event.Payload.FromProfile.DeviceID)),
		DisplayName:     event.Payload.FromProfile.Name,
		LocalHandle:     nullString(strings.TrimSpace(event.Payload.FromProfile.Username)),
		FromPublicKey:   append([]byte(nil), event.Payload.FromProfile.DevicePublicKey...),
		FromContactCode: nullString(contactCode),
		InvitePayload:   contactCode,
		State:           "pending",
		CreatedAt:       event.Timestamp,
	})
	if err != nil {
		c.State.LastNetworkError = fmt.Sprintf("store contact request: %v", err)
		c.notifyStateChanged()
		return
	}

	if err := c.LoadSocialState(context.Background()); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("reload social state: %v", err)
	}

	c.notifyStateChanged()
}

func (c *Controller) HandleContactRequestAccepted(event protocol.Envelope[protocol.ContactRequestAcceptedPayload]) {
	if c == nil || c.Store == nil {
		return
	}

	profile := event.Payload.Profile
	if profile.UserID == "" {
		c.State.LastNetworkError = "accepted contact is missing user ID"
		c.notifyStateChanged()
		return
	}

	if profile.Name == "" || profile.Username == "" || profile.ContactCode == "" {
		request, err := c.findContactRequestByUserID(profile.UserID)
		if err != nil {
			c.State.LastNetworkError = fmt.Sprintf("load accepted contact request: %v", err)
			c.notifyStateChanged()
			return
		}
		if request != nil {
			if profile.Name == "" {
				profile.Name = request.DisplayName
			}
			if profile.Username == "" {
				profile.Username = nullStringValue(request.LocalHandle)
			}
			if profile.ContactCode == "" {
				profile.ContactCode = nullStringValue(request.FromContactCode)
			}
		}
	}

	if err := c.Store.Q.UpsertContact(context.Background(), sql.UpsertContactParams{
		UserID:             profile.UserID,
		DisplayName:        profile.Name,
		LocalHandle:        nullString(strings.TrimSpace(profile.Username)),
		ProfilePicturePath: dsql.NullString{},
		ContactCode:        nullString(strings.TrimSpace(profile.ContactCode)),
		CreatedAt:          event.Timestamp,
	}); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("store contact: %v", err)
		c.notifyStateChanged()
		return
	}

	_ = c.Store.Q.DeleteContactRequestByFromUserID(context.Background(), profile.UserID)
	_ = c.upsertContactDeviceFromProfile(context.Background(), profile, event.Timestamp)

	if err := c.LoadSocialState(context.Background()); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("reload social state: %v", err)
	}

	c.notifyStateChanged()
}

func (c *Controller) HandleContactRequestRejected(event protocol.Envelope[protocol.ContactRequestRejectedPayload]) {
	if c == nil || c.Store == nil {
		return
	}

	if err := c.Store.Q.DeleteContactRequestByFromUserID(context.Background(), event.Payload.UserID); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("delete rejected contact request: %v", err)
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

func (c *Controller) findContactRequestByUserID(userID string) (*sql.ContactRequest, error) {
	if c == nil || c.Store == nil {
		return nil, nil
	}

	requests, err := c.Store.Q.ListContactRequests(context.Background())
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

func (c *Controller) upsertContactDeviceFromProfile(ctx context.Context, profile protocol.PublicContactProfile, createdAt int64) error {
	if c == nil || c.Store == nil {
		return nil
	}
	if strings.TrimSpace(profile.UserID) == "" || strings.TrimSpace(profile.DeviceID) == "" || len(profile.DevicePublicKey) == 0 {
		return nil
	}

	return c.Store.Q.UpsertContactDevice(ctx, sql.UpsertContactDeviceParams{
		ContactUserID: profile.UserID,
		DeviceID:      profile.DeviceID,
		PublicKey:     append([]byte(nil), profile.DevicePublicKey...),
		CreatedAt:     createdAt,
		RevokedAt:     dsql.NullInt64{},
	})
}
