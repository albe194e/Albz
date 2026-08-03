package controllers

import (
	"context"
	"fmt"

	"github.com/albe194e/albz/app/core-go/domains/contacts"
	"github.com/albe194e/albz/app/core-go/domains/conversations"
	"github.com/albe194e/albz/app/core-go/domains/identity"
	"github.com/albe194e/albz/app/core-go/domains/messages"
)

type AppState struct {
	CurrentUser          *identity.LocalIdentity
	Messages             []messages.Message
	Conversations        []conversations.Conversation
	Contacts             []contacts.Contact
	ContactRequests      []contacts.ContactRequest
	LoadedConversationID string
	ServerConnected      bool
	LastNetworkError     string
}

func (c *Controller) initStateFromDB(ctx context.Context) error {
	if c.State.CurrentUser == nil {
		return fmt.Errorf("CurrentUser is nil")
	}

	convs, err := c.ConversationService.ListByUserID(ctx, c.State.CurrentUser.UserID)
	if err != nil {
		return err
	}

	contactList, err := c.ContactService.ListContacts(ctx)
	if err != nil {
		return err
	}

	contactRequests, err := c.ContactService.ListContactRequests(ctx)
	if err != nil {
		return err
	}

	c.State.Conversations = convs
	c.State.Contacts = contactList
	c.State.ContactRequests = contactRequests
	c.State.Messages = nil
	c.State.LoadedConversationID = ""
	return nil
}
