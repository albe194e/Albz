package app

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
	"github.com/google/uuid"
)

func (c *Controller) Register(ctx context.Context, name, username, password, profile_url string) error {
	createUserParams := sql.CreateUserParams{
		ID:                uuid.New().String(),
		Username:          username,
		Name:              name,
		HashedPassword:    password,
		ProfilePictureUrl: profile_url,
		ContactCode:       newContactCode(),
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

	if err := c.ConnectToServer(ctx); err != nil {
		fmt.Printf("failed to connect to server: %v\n", err)
	}

	return nil
}

func newContactCode() string {
	value := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
	return "ALBZ-" + value[:12]
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
	if err := c.ConnectToServer(ctx); err != nil {
		fmt.Printf("failed to connect to server: %v\n", err)
	}
	return nil
}

func (c *Controller) Logout(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("controller is nil")
	}

	if err := c.DeleteSession(ctx); err != nil {
		return err
	}

	if c.Net != nil {
		if err := c.Net.Close(); err != nil {
			return err
		}
	}

	if c.State != nil {
		c.State.CurrentUser = nil
		c.State.Messages = nil
		c.State.Conversations = nil
		c.State.Contacts = nil
		c.State.ContactRequests = nil
		c.State.LoadedConversationID = ""
		c.State.ServerConnected = false
		c.State.LastNetworkError = ""
	}

	c.notifyStateChanged()
	return nil
}
