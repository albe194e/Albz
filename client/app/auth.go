package app

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/albe194e/albz/client/db/sqlc/sql"
	"github.com/google/uuid"
)

func (c *Controller) Register(ctx context.Context, name, username, password string) error {
	createUserParams := sql.CreateUserParams{
		ID:             uuid.New().String(),
		Username:       username,
		Name:           name,
		HashedPassword: password,
	}

	err := c.Store.Q.CreateUser(ctx, createUserParams)
	if err != nil {
		return err
	}

	return c.Login(ctx, username, password)
}

func (c *Controller) Login(ctx context.Context, username, password string) error {
	user, err := c.Store.Q.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}

	if user.HashedPassword != password {
		return fmt.Errorf("invalid username or password")
	}

	c.State.CurrentUser = &user
	_, err = c.CreateOrUpdateSession(ctx, user.ID)
	if err != nil {
		return err
	}

	// Init state
	err = c.State.InitStateFromDB(ctx, c.Store)
	if err != nil {
		fmt.Printf("failed to initialize state: %v\n", err)
		return err
	}

	return nil
}

func (c *Controller) CreateOrUpdateSession(ctx context.Context, userID string) (string, error) {
	sessionID := uuid.New().String()
	currentSession, err := c.Store.Q.GetCurrentSession(ctx)
	if err != nil && !errors.Is(err, dsql.ErrNoRows) {
		return "", err
	}
	if err == nil {
		sessionID = currentSession.SessionID
	}

	now := time.Now()
	err = c.Store.Q.UpsertCurrentSession(ctx, sql.UpsertCurrentSessionParams{
		SessionID: sessionID,
		UserID:    userID,
		CreatedAt: now.Unix(),
		ExpiresAt: now.Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func (c *Controller) DeleteSession(ctx context.Context) error {
	return c.Store.Q.DeleteCurrentSession(ctx)
}

func (c *Controller) VerifySession(ctx context.Context) error {
	session, err := c.Store.Q.GetCurrentSession(ctx)
	if err != nil {
		return err
	}

	if session.ExpiresAt < time.Now().Unix() {
		if err := c.DeleteSession(ctx); err != nil {
			return err
		}
		return fmt.Errorf("session has expired")
	}

	user, err := c.Store.Q.GetUserByID(ctx, session.UserID)
	if err != nil {
		return err
	}

	c.State.CurrentUser = &user
	// Init state
	err = c.State.InitStateFromDB(ctx, c.Store)
	if err != nil {
		fmt.Printf("failed to initialize state: %v\n", err)
		return err
	}
	return nil
}
