package controllers

import (
	"github.com/albe194e/albz/app/core-go/domains/contacts"
	"github.com/albe194e/albz/app/core-go/domains/conversations"
	"github.com/albe194e/albz/app/core-go/domains/identity"
	"github.com/albe194e/albz/app/core-go/domains/messages"
	"github.com/albe194e/albz/app/core-go/file"
	"github.com/albe194e/albz/app/core-go/relay"
)

type Options struct {
	FileHandler *file.Handler
	ServerURL   string
}

type Controller struct {
	State                    *AppState
	IdentityService          *identity.Service
	ContactService           *contacts.Service
	ConversationService      *conversations.Service
	MessageService           *messages.Service
	Relay                    *relay.Client
	FileHandler              *file.Handler
	UnlockedDevicePrivateKey []byte
	OnStateChanged           func()
}

func NewController(options Options) *Controller {
	c := &Controller{
		State:               &AppState{},
		IdentityService:     identity.NewService(),
		ContactService:      contacts.NewService(),
		ConversationService: conversations.NewService(),
		MessageService:      messages.NewService(),
		FileHandler:         options.FileHandler,
	}

	c.Relay = relay.NewClient(options.ServerURL, relay.Handlers{
		OnConversationCreated:    c.HandleConversationCreated,
		OnMessageCreated:         c.HandleIncomingMessage,
		OnMessageDelivery:        c.HandleMessageDelivery,
		OnContactRequestReceived: c.HandleContactRequestReceived,
		OnContactRequestAccepted: c.HandleContactRequestAccepted,
		OnContactRequestRejected: c.HandleContactRequestRejected,
		OnError:                  c.HandleNetworkError,
		OnDisconnect:             c.HandleDisconnect,
	})

	return c
}

func (c *Controller) notifyStateChanged() {
	if c != nil && c.OnStateChanged != nil {
		c.OnStateChanged()
	}
}
