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

	friends, err := c.Store.Q.ListFriends(ctx)
	if err != nil {
		return err
	}

	friendRequests, err := c.Store.Q.ListFriendRequests(ctx)
	if err != nil {
		return err
	}

	c.State.Friends = friends
	c.State.FriendRequests = friendRequests
	return nil
}

func (c *Controller) SendFriendRequest(ctx context.Context, friendCode string) error {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return fmt.Errorf("no current user")
	}
	if c.Net == nil {
		return fmt.Errorf("network client is not configured")
	}

	trimmedFriendCode := strings.ToUpper(strings.TrimSpace(friendCode))
	if trimmedFriendCode == "" {
		return fmt.Errorf("friend code is required")
	}
	if trimmedFriendCode == c.State.CurrentUser.FriendCode {
		return fmt.Errorf("cannot send a friend request to yourself")
	}

	if !c.State.ServerConnected {
		if err := c.ConnectToServer(ctx); err != nil {
			return err
		}
	}

	return c.Net.SendFriendRequest(uuid.NewString(), protocol.FriendRequestSendPayload{
		FriendCode:  trimmedFriendCode,
		FromProfile: c.currentPublicFriendProfile(),
	})
}

func (c *Controller) AcceptFriendRequest(ctx context.Context, fromUserID string) error {
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

	return c.Net.AcceptFriendRequest(uuid.NewString(), protocol.FriendRequestAcceptPayload{
		FromUserID:    fromUserID,
		AcceptProfile: c.currentPublicFriendProfile(),
	})
}

func (c *Controller) RejectFriendRequest(ctx context.Context, fromUserID string) error {
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

	return c.Net.RejectFriendRequest(uuid.NewString(), protocol.FriendRequestRejectPayload{
		FromUserID: fromUserID,
	})
}

func (c *Controller) currentPublicFriendProfile() protocol.PublicFriendProfile {
	if c == nil || c.State == nil || c.State.CurrentUser == nil {
		return protocol.PublicFriendProfile{}
	}

	return protocol.PublicFriendProfile{
		UserID:     c.State.CurrentUser.ID,
		FriendCode: c.State.CurrentUser.FriendCode,
		Name:       c.State.CurrentUser.Name,
		Username:   c.State.CurrentUser.Username,
	}
}
