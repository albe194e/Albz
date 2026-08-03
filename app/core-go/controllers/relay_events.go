package controllers

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"strings"

	domaincontacts "github.com/albe194e/albz/app/core-go/domains/contacts"
	domainmessages "github.com/albe194e/albz/app/core-go/domains/messages"
	"github.com/albe194e/albz/shared/protocol"
)

func (c *Controller) ConnectToServer(ctx context.Context) error {
	if c == nil || c.Relay == nil || c.State == nil || c.State.CurrentUser == nil {
		return nil
	}
	if len(c.UnlockedDevicePrivateKey) == 0 {
		return fmt.Errorf("device key is locked; login is required before connecting to the relay")
	}

	if err := c.Relay.Connect(
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
	if c == nil || c.ConversationService == nil || c.MessageService == nil {
		return
	}

	if _, err := c.ConversationService.GetByID(context.Background(), event.Payload.ConversationID); err != nil {
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

	err := c.MessageService.Create(context.Background(), domainmessages.CreateParams{
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
		Direction:     domainmessages.DirectionIncoming,
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
	if c == nil || c.MessageService == nil {
		return
	}

	err := c.MessageService.UpdateDeliveryState(context.Background(), domainmessages.UpdateDeliveryStateParams{
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
	if c == nil || c.ContactService == nil {
		return
	}

	contactCode := strings.TrimSpace(event.Payload.FromProfile.ContactCode)
	profilePicturePath, err := c.saveSharedContactProfilePicture(event.Payload.FromProfile)
	if err != nil {
		c.State.LastNetworkError = fmt.Sprintf("save contact request profile picture: %v", err)
		c.notifyStateChanged()
		return
	}

	err = c.ContactService.UpsertContactRequest(context.Background(), domaincontacts.UpsertContactRequestParams{
		FromUserID:         event.Payload.FromProfile.UserID,
		FromDeviceID:       nullString(strings.TrimSpace(event.Payload.FromProfile.DeviceID)),
		DisplayName:        event.Payload.FromProfile.Name,
		LocalHandle:        nullString(strings.TrimSpace(event.Payload.FromProfile.Username)),
		ProfilePicturePath: profilePicturePath,
		FromPublicKey:      append([]byte(nil), event.Payload.FromProfile.DevicePublicKey...),
		FromContactCode:    nullString(contactCode),
		InvitePayload:      contactCode,
		State:              "pending",
		CreatedAt:          event.Timestamp,
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
	if c == nil || c.ContactService == nil {
		return
	}

	profile := event.Payload.Profile
	if profile.UserID == "" {
		c.State.LastNetworkError = "accepted contact is missing user ID"
		c.notifyStateChanged()
		return
	}

	request, err := c.findContactRequestByUserID(profile.UserID)
	if err != nil {
		c.State.LastNetworkError = fmt.Sprintf("load accepted contact request: %v", err)
		c.notifyStateChanged()
		return
	}

	if profile.Name == "" || profile.Username == "" || profile.ContactCode == "" {
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

	profilePicturePath, err := c.saveSharedContactProfilePicture(profile)
	if err != nil {
		c.State.LastNetworkError = fmt.Sprintf("save accepted contact profile picture: %v", err)
		c.notifyStateChanged()
		return
	}
	if !profilePicturePath.Valid && request != nil && request.ProfilePicturePath.Valid {
		profilePicturePath = request.ProfilePicturePath
	}

	if err := c.ContactService.UpsertContact(context.Background(), domaincontacts.UpsertContactParams{
		UserID:             profile.UserID,
		DisplayName:        profile.Name,
		LocalHandle:        nullString(strings.TrimSpace(profile.Username)),
		ProfilePicturePath: profilePicturePath,
		ContactCode:        nullString(strings.TrimSpace(profile.ContactCode)),
		CreatedAt:          event.Timestamp,
	}); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("store contact: %v", err)
		c.notifyStateChanged()
		return
	}

	_ = c.ContactService.DeleteContactRequestByFromUserID(context.Background(), profile.UserID)
	_ = c.upsertContactDeviceFromProfile(context.Background(), profile, event.Timestamp)

	if err := c.LoadSocialState(context.Background()); err != nil {
		c.State.LastNetworkError = fmt.Sprintf("reload social state: %v", err)
	}

	c.notifyStateChanged()
}

func (c *Controller) HandleContactRequestRejected(event protocol.Envelope[protocol.ContactRequestRejectedPayload]) {
	if c == nil || c.ContactService == nil {
		return
	}

	if err := c.ContactService.DeleteContactRequestByFromUserID(context.Background(), event.Payload.UserID); err != nil {
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

func (c *Controller) findContactRequestByUserID(userID string) (*domaincontacts.ContactRequest, error) {
	if c == nil || c.ContactService == nil {
		return nil, nil
	}

	requests, err := c.ContactService.ListContactRequests(context.Background())
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
	if c == nil || c.ContactService == nil {
		return nil
	}
	if strings.TrimSpace(profile.UserID) == "" || strings.TrimSpace(profile.DeviceID) == "" || len(profile.DevicePublicKey) == 0 {
		return nil
	}

	return c.ContactService.UpsertContactDevice(ctx, domaincontacts.UpsertContactDeviceParams{
		ContactUserID: profile.UserID,
		DeviceID:      profile.DeviceID,
		PublicKey:     append([]byte(nil), profile.DevicePublicKey...),
		CreatedAt:     createdAt,
		RevokedAt:     dsql.NullInt64{},
	})
}

func (c *Controller) saveSharedContactProfilePicture(profile protocol.PublicContactProfile) (dsql.NullString, error) {
	if len(profile.ProfilePicture) == 0 {
		return dsql.NullString{}, nil
	}
	if c == nil || c.FileHandler == nil {
		return dsql.NullString{}, fmt.Errorf("file handler is not configured")
	}

	filename := strings.TrimSpace(profile.UserID)
	if filename == "" {
		filename = "contact-profile-picture"
	} else {
		filename += "-profile-picture"
	}

	path, err := c.FileHandler.SaveImageToStorage(profile.ProfilePicture, filename)
	if err != nil {
		return dsql.NullString{}, err
	}

	return nullString(path), nil
}
