package app

import (
	"strings"

	"github.com/albe194e/albz/client/db/sqlc/sql"
)

func FriendDisplayName(friend sql.Friend) string {
	if strings.TrimSpace(friend.Name) != "" {
		return friend.Name
	}
	if strings.TrimSpace(friend.Username) != "" {
		return friend.Username
	}
	if strings.TrimSpace(friend.FriendCode) != "" {
		return friend.FriendCode
	}
	return friend.UserID
}

func FriendRequestDisplayName(request sql.FriendRequest) string {
	if strings.TrimSpace(request.Name) != "" {
		return request.Name
	}
	if strings.TrimSpace(request.Username) != "" {
		return request.Username
	}
	if strings.TrimSpace(request.FromFriendCode) != "" {
		return request.FromFriendCode
	}
	return request.FromUserID
}

func (c *Controller) findFriendByUserID(userID string) *sql.Friend {
	if c == nil || c.State == nil {
		return nil
	}

	for _, friend := range c.State.Friends {
		if friend.UserID == userID {
			friendCopy := friend
			return &friendCopy
		}
	}

	return nil
}
