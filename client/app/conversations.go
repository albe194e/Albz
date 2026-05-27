package app

import (
	"context"

	"github.com/albe194e/albz/client/db/sqlc/sql"
	"github.com/google/uuid"
)

func (c *Controller) GetConversationsByMe(ctx context.Context) ([]sql.Conversation, error) {
	return c.Store.Q.GetConversationsByUserID(ctx, c.State.CurrentUser.ID)
}

func (c *Controller) CreateConversation(ctx context.Context, name string) error {
	params := sql.CreateConversationParams{
		ID:   uuid.New().String()[:16],
		Name: name,
	}
	err := c.Store.Q.CreateConversation(ctx, params)
	if err != nil {
		return err
	}

	// add the current user as a participant
	err = c.Store.Q.AddParticipant(ctx, sql.AddParticipantParams{
		ConversationID: params.ID,
		ParticipantID:  c.State.CurrentUser.ID,
	})
	if err != nil {
		return err
	}

	return nil
}
