package bridge

import "context"

func (b *Bridge) SendMessage(conversationID, body string) error {
	return b.controller().AddMessage(context.Background(), conversationID, body)
}
