package app

import (
	"strings"

	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
)

func ContactDisplayName(contact sql.Contact) string {
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

func ContactRequestDisplayName(request sql.ContactRequest) string {
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

func (c *Controller) findContactByUserID(userID string) *sql.Contact {
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

func (c *Controller) FindContactByUserID(userID string) *sql.Contact {
	return c.findContactByUserID(userID)
}
