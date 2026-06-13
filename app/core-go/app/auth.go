package app

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/albe194e/albz/app/core-go/app/crypt"
	"github.com/albe194e/albz/app/core-go/db/sqlc/sql"
	corelog "github.com/albe194e/albz/app/core-go/logging"
	"github.com/google/uuid"
)

func (c *Controller) Register(ctx context.Context, name, username, password, profilePath string) error {
	trimmedName := strings.TrimSpace(name)
	trimmedHandle := strings.TrimSpace(username)
	if trimmedName == "" {
		return fmt.Errorf("name is required")
	}
	if trimmedHandle == "" {
		return fmt.Errorf("local handle is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if _, err := c.Store.Q.GetLocalIdentity(ctx); err == nil {
		return fmt.Errorf("local identity already exists on this device")
	} else if !errors.Is(err, dsql.ErrNoRows) {
		return err
	}

	privateKey, publicKey, err := crypt.GenerateDeviceKeyPair()
	if err != nil {
		return fmt.Errorf("generate device key pair: %w", err)
	}

	kdfParams := crypt.DefaultDeviceKeyKDFParams()

	kdfSalt, err := crypt.GenerateDeviceKeySalt()
	if err != nil {
		return fmt.Errorf("generate KDF salt: %w", err)
	}

	encryptedPrivateKey, err := crypt.EncryptDevicePrivateKey(privateKey, password, kdfSalt, kdfParams)
	if err != nil {
		return fmt.Errorf("encrypt device private key: %w", err)
	}

	kdfParamsJSON, err := crypt.MarshalKDFParams(kdfParams)
	if err != nil {
		return err
	}

	err = c.Store.Q.UpsertLocalIdentity(ctx, sql.UpsertLocalIdentityParams{
		UserID:                    uuid.NewString(),
		DeviceID:                  uuid.NewString(),
		DevicePublicKey:           publicKey,
		EncryptedDevicePrivateKey: encryptedPrivateKey,
		KdfSalt:                   kdfSalt,
		KdfParams:                 kdfParamsJSON,
		Name:                      trimmedName,
		LocalHandle:               nullString(trimmedHandle),
		ProfilePicturePath:        nullString(strings.TrimSpace(profilePath)),
		ContactCode:               nullString(newContactCode()),
		CreatedAt:                 time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	return c.Login(ctx, trimmedHandle, password)
}

func (c *Controller) Login(ctx context.Context, username, password string) error {
	if c == nil || c.Store == nil || c.State == nil || c.Net == nil {
		return fmt.Errorf("controller is not fully initialized")
	}

	identity, err := c.Store.Q.GetLocalIdentity(ctx)
	if err != nil {
		return err
	}

	if !localIdentityMatchesLogin(identity, strings.TrimSpace(username)) {
		return fmt.Errorf("invalid username or password")
	}

	decryptedPrivateKey, err := crypt.DecryptDevicePrivateKey(crypt.EncryptedDeviceKeyRecord{
		EncryptedPrivateKey: identity.EncryptedDevicePrivateKey,
		KDFSalt:             identity.KdfSalt,
		KDFParams:           identity.KdfParams,
	}, password)
	if err != nil {
		return fmt.Errorf("invalid username or password")
	}

	previousCurrentUser := c.State.CurrentUser
	previousMessages := c.State.Messages
	previousConversations := c.State.Conversations
	previousContacts := c.State.Contacts
	previousContactRequests := c.State.ContactRequests
	previousLoadedConversationID := c.State.LoadedConversationID
	previousServerConnected := c.State.ServerConnected
	previousLastNetworkError := c.State.LastNetworkError
	previousUnlockedDevicePrivateKey := append([]byte(nil), c.UnlockedDevicePrivateKey...)
	restorePreviousState := func() {
		c.State.CurrentUser = previousCurrentUser
		c.State.Messages = previousMessages
		c.State.Conversations = previousConversations
		c.State.Contacts = previousContacts
		c.State.ContactRequests = previousContactRequests
		c.State.LoadedConversationID = previousLoadedConversationID
		c.State.ServerConnected = previousServerConnected
		c.State.LastNetworkError = previousLastNetworkError
		c.UnlockedDevicePrivateKey = append([]byte(nil), previousUnlockedDevicePrivateKey...)
	}

	if err := c.Net.Connect(
		identity.UserID,
		identity.DeviceID,
		nullStringValue(identity.ContactCode),
		identity.DevicePublicKey,
		decryptedPrivateKey,
	); err != nil {
		c.State.ServerConnected = false
		c.State.LastNetworkError = err.Error()
		c.notifyStateChanged()
		return err
	}

	c.State.CurrentUser = &identity
	c.UnlockedDevicePrivateKey = append([]byte(nil), decryptedPrivateKey...)
	c.State.ServerConnected = true
	c.State.LastNetworkError = ""

	if err := c.State.InitStateFromDB(ctx, c.Store); err != nil {
		if !previousServerConnected {
			_ = c.Net.Close()
		}
		restorePreviousState()
		corelog.Errorf("failed to initialize state after login: %v", err)
		return err
	}

	_, err = c.CreateOrUpdateSession(ctx, identity.UserID, identity.DeviceID)
	if err != nil {
		if !previousServerConnected {
			_ = c.Net.Close()
		}
		restorePreviousState()
		return err
	}

	c.notifyStateChanged()
	return nil
}

func newContactCode() string {
	value := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
	return "HADDLE-" + value[:12]
}

func (c *Controller) CreateOrUpdateSession(ctx context.Context, userID string, deviceID string) (string, error) {
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
		DeviceID:  deviceID,
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
		return c.invalidateSession(ctx)
	}

	identity, err := c.Store.Q.GetLocalIdentity(ctx)
	if err != nil {
		if errors.Is(err, dsql.ErrNoRows) {
			return c.invalidateSession(ctx)
		}
		return err
	}
	if !localIdentityIsUsable(identity) {
		return c.invalidateSession(ctx)
	}

	if session.UserID != identity.UserID || session.DeviceID != identity.DeviceID {
		return c.invalidateSession(ctx)
	}

	c.State.CurrentUser = &identity
	if len(c.UnlockedDevicePrivateKey) > 0 {
		c.UnlockedDevicePrivateKey = append([]byte(nil), c.UnlockedDevicePrivateKey...)
	}
	if err := c.State.InitStateFromDB(ctx, c.Store); err != nil {
		corelog.Errorf("failed to initialize state during session verification: %v", err)
		return err
	}
	if len(c.UnlockedDevicePrivateKey) > 0 {
		if err := c.ConnectToServer(ctx); err != nil {
			corelog.Debugf("failed to reconnect to relay during session verification: %v", err)
		}
	}
	return nil
}

func (c *Controller) invalidateSession(ctx context.Context) error {
	if err := c.DeleteSession(ctx); err != nil && !errors.Is(err, dsql.ErrNoRows) {
		return err
	}

	if c != nil && c.State != nil {
		c.State.CurrentUser = nil
		c.State.Messages = nil
		c.State.Conversations = nil
		c.State.Contacts = nil
		c.State.ContactRequests = nil
		c.State.LoadedConversationID = ""
		c.State.ServerConnected = false
		c.State.LastNetworkError = ""
	}
	if c != nil {
		c.UnlockedDevicePrivateKey = nil
	}

	return dsql.ErrNoRows
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
	c.UnlockedDevicePrivateKey = nil

	c.notifyStateChanged()
	return nil
}

func localIdentityMatchesLogin(identity sql.LocalIdentity, login string) bool {
	if login == "" {
		return false
	}

	if localHandle := strings.TrimSpace(nullStringValue(identity.LocalHandle)); localHandle != "" && localHandle == login {
		return true
	}
	if identity.UserID == login {
		return true
	}
	return strings.TrimSpace(nullStringValue(identity.ContactCode)) == login
}

func localIdentityIsUsable(identity sql.LocalIdentity) bool {
	if strings.TrimSpace(identity.UserID) == "" {
		return false
	}
	if strings.TrimSpace(identity.DeviceID) == "" {
		return false
	}
	if strings.TrimSpace(identity.Name) == "" {
		return false
	}
	if len(identity.DevicePublicKey) == 0 {
		return false
	}
	if len(identity.EncryptedDevicePrivateKey) == 0 {
		return false
	}
	if len(identity.KdfSalt) == 0 {
		return false
	}
	return strings.TrimSpace(identity.KdfParams) != ""
}
