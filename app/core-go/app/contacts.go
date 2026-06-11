package app

import (
	"strings"

	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
)

func ContactDisplayName(contact sql.Contact) string {
	if strings.TrimSpace(contact.Name) != "" {
		return contact.Name
	}
	if strings.TrimSpace(contact.Username) != "" {
		return contact.Username
	}
	if strings.TrimSpace(contact.ContactCode) != "" {
		return contact.ContactCode
	}
	return contact.UserID
}

func ContactRequestDisplayName(request sql.ContactRequest) string {
	if strings.TrimSpace(request.Name) != "" {
		return request.Name
	}
	if strings.TrimSpace(request.Username) != "" {
		return request.Username
	}
	if strings.TrimSpace(request.FromContactCode) != "" {
		return request.FromContactCode
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
