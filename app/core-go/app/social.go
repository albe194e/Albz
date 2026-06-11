package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/albe194e/albz/shared/protocol"
	"github.com/google/uuid"
)

func (c *Controller) LoadSocialState(ctx context.Context) error {
	if c == nil || c.Store == nil || c.State == nil {
		return nil
	}

	contacts, err := c.Store.Q.ListContacts(ctx)
	if err != nil {
		return err
	}

	contactRequests, err := c.Store.Q.ListContactRequests(ctx)
	if err != nil {
		return err
	}

	c.State.Contacts = contacts
	c.State.ContactRequests = contactRequests
	return nil
}

func (c *Controller) SendContactRequest(ctx context.Context, contactCode string) error {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return fmt.Errorf("no current user")
	}
	if c.Net == nil {
		return fmt.Errorf("network client is not configured")
	}

	trimmedContactCode := strings.ToUpper(strings.TrimSpace(contactCode))
	if trimmedContactCode == "" {
		return fmt.Errorf("contact code is required")
	}
	if trimmedContactCode == c.State.CurrentUser.ContactCode {
		return fmt.Errorf("cannot send a contact request to yourself")
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(ctx); err != nil {
			return err
		}
	}

	return c.Net.SendContactRequest(uuid.NewString(), protocol.ContactRequestSendPayload{
		ContactCode: trimmedContactCode,
		FromProfile: c.currentPublicContactProfile(),
	})
}

func (c *Controller) AcceptContactRequest(ctx context.Context, fromUserID string) error {
	if c == nil || c.Net == nil {
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

	return c.Net.AcceptContactRequest(uuid.NewString(), protocol.ContactRequestAcceptPayload{
		FromUserID:    fromUserID,
		AcceptProfile: c.currentPublicContactProfile(),
	})
}

func (c *Controller) RejectContactRequest(ctx context.Context, fromUserID string) error {
	if c == nil || c.Net == nil {
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

	return c.Net.RejectContactRequest(uuid.NewString(), protocol.ContactRequestRejectPayload{
		FromUserID: fromUserID,
	})
}

func (c *Controller) currentPublicContactProfile() protocol.PublicContactProfile {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return protocol.PublicContactProfile{}
	}

	return protocol.PublicContactProfile{
		UserID:      c.State.CurrentUser.ID,
		ContactCode: c.State.CurrentUser.ContactCode,
		Name:        c.State.CurrentUser.Name,
		Username:    c.State.CurrentUser.Username,
	}
}
