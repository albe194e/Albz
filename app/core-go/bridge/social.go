package bridge

import "context"

func (b *Bridge) SendFriendRequest(friendCode string) error {
	return b.controller().SendFriendRequest(context.Background(), friendCode)
}

func (b *Bridge) AcceptFriendRequest(fromUserID string) error {
	return b.controller().AcceptFriendRequest(context.Background(), fromUserID)
}

func (b *Bridge) RejectFriendRequest(fromUserID string) error {
	return b.controller().RejectFriendRequest(context.Background(), fromUserID)
}
