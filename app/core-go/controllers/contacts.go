package controllers

import (
	"strings"

	"github.com/albe194e/albz/app/core-go/domains/contacts"
)

func ContactDisplayName(contact contacts.Contact) string {
	if strings.TrimSpace(contact.DisplayName) != "" {
		return contact.DisplayName
	}
	if strings.TrimSpace(nullStringValue(contact.LocalHandle)) != "" {
		return nullStringValue(contact.LocalHandle)
	}
	if strings.TrimSpace(nullStringValue(contact.ContactCode)) != "" {
		return nullStringValue(contact.ContactCode)
	}
	return contact.UserID
}

func ContactRequestDisplayName(request contacts.ContactRequest) string {
	if strings.TrimSpace(request.DisplayName) != "" {
		return request.DisplayName
	}
	if strings.TrimSpace(nullStringValue(request.LocalHandle)) != "" {
		return nullStringValue(request.LocalHandle)
	}
	if strings.TrimSpace(nullStringValue(request.FromContactCode)) != "" {
		return nullStringValue(request.FromContactCode)
	}
	return request.FromUserID
}

func (c *Controller) findContactByUserID(userID string) *contacts.Contact {
	if c == nil || c.State == nil {
		return nil
	}

	for _, contact := range c.State.Contacts {
		if contact.UserID == userID {
			contactCopy := contact
			return &contactCopy
		}
	}

	return nil
}

func (c *Controller) FindContactByUserID(userID string) *contacts.Contact {
	return c.findContactByUserID(userID)
}
