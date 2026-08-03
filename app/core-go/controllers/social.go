package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/albe194e/albz/app/core-go/qr"
	"github.com/albe194e/albz/shared/protocol"
	"github.com/google/uuid"
)

func (c *Controller) LoadSocialState(ctx context.Context) error {
	if c == nil || c.ContactService == nil || c.State == nil {
		return nil
	}

	contactList, err := c.ContactService.ListContacts(ctx)
	if err != nil {
		return err
	}

	contactRequests, err := c.ContactService.ListContactRequests(ctx)
	if err != nil {
		return err
	}

	c.State.Contacts = contactList
	c.State.ContactRequests = contactRequests
	return nil
}

func (c *Controller) SendContactRequest(ctx context.Context, contactCode string) error {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return fmt.Errorf("no current user")
	}
	if c.Relay == nil {
		return fmt.Errorf("network client is not configured")
	}

	trimmedContactCode := strings.ToUpper(strings.TrimSpace(contactCode))
	if trimmedContactCode == "" {
		return fmt.Errorf("contact code is required")
	}
	if trimmedContactCode == nullStringValue(c.State.CurrentUser.ContactCode) {
		return fmt.Errorf("cannot send a contact request to yourself")
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(ctx); err != nil {
			return err
		}
	}

	fromProfile, err := c.currentPublicContactProfile(true)
	if err != nil {
		return err
	}

	return c.Relay.SendContactRequest(uuid.NewString(), protocol.ContactRequestSendPayload{
		ContactCode: trimmedContactCode,
		FromProfile: fromProfile,
	})
}

func (c *Controller) AcceptContactRequest(ctx context.Context, fromUserID string) error {
	if c == nil || c.Relay == nil {
		return fmt.Errorf("network client is not configured")
	}
	if strings.TrimSpace(fromUserID) == "" {
		return fmt.Errorf("from user ID is required")
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(ctx); err != nil {
			return err
		}
	}

	acceptProfile, err := c.currentPublicContactProfile(true)
	if err != nil {
		return err
	}

	return c.Relay.AcceptContactRequest(uuid.NewString(), protocol.ContactRequestAcceptPayload{
		FromUserID:    fromUserID,
		AcceptProfile: acceptProfile,
	})
}

func (c *Controller) RejectContactRequest(ctx context.Context, fromUserID string) error {
	if c == nil || c.Relay == nil {
		return fmt.Errorf("network client is not configured")
	}
	if strings.TrimSpace(fromUserID) == "" {
		return fmt.Errorf("from user ID is required")
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(ctx); err != nil {
			return err
		}
	}

	return c.Relay.RejectContactRequest(uuid.NewString(), protocol.ContactRequestRejectPayload{
		FromUserID: fromUserID,
	})
}

func (c *Controller) GetContactQRCode() ([]byte, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return nil, fmt.Errorf("no current user")
	}

	return qr.GenerateContactQRCode(nullStringValue(c.State.CurrentUser.ContactCode))
}

func (c *Controller) currentPublicContactProfile(includeProfilePicture bool) (protocol.PublicContactProfile, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return protocol.PublicContactProfile{}, fmt.Errorf("no current user")
	}

	var profilePicture []byte
	if includeProfilePicture {
		var err error
		profilePicture, err = c.loadCurrentProfilePictureBytes()
		if err != nil {
			return protocol.PublicContactProfile{}, err
		}
	}

	return protocol.PublicContactProfile{
		UserID:          c.State.CurrentUser.UserID,
		DeviceID:        c.State.CurrentUser.DeviceID,
		DevicePublicKey: append([]byte(nil), c.State.CurrentUser.DevicePublicKey...),
		ProfilePicture:  profilePicture,
		ContactCode:     nullStringValue(c.State.CurrentUser.ContactCode),
		Name:            c.State.CurrentUser.Name,
		Username:        nullStringValue(c.State.CurrentUser.LocalHandle),
	}, nil
}

func (c *Controller) loadCurrentProfilePictureBytes() ([]byte, error) {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return nil, fmt.Errorf("no current user")
	}

	profilePicturePath := strings.TrimSpace(
		nullStringValue(c.State.CurrentUser.ProfilePicturePath),
	)
	if profilePicturePath == "" {
		return nil, nil
	}
	if c.FileHandler == nil {
		return nil, fmt.Errorf("file handler is not configured")
	}

	loadedImage, err := c.FileHandler.LoadImageFromPath(profilePicturePath)
	if err != nil {
		return nil, fmt.Errorf("load shared profile picture: %w", err)
	}

	return append([]byte(nil), loadedImage.File.Data...), nil
}
