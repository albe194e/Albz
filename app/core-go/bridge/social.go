package bridge

import "context"

func (b *Bridge) SendContactRequest(contactCode string) error {
	return b.controller().SendContactRequest(context.Background(), contactCode)
}

func (b *Bridge) AcceptContactRequest(fromUserID string) error {
	return b.controller().AcceptContactRequest(context.Background(), fromUserID)
}

func (b *Bridge) RejectContactRequest(fromUserID string) error {
	return b.controller().RejectContactRequest(context.Background(), fromUserID)
}
