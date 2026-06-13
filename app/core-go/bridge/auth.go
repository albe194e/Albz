package bridge

import (
	"context"
	"fmt"
)

func (b *Bridge) TryLoadSession() (bool, error) {
	if b == nil || b.service == nil {
		return false, fmt.Errorf("bridge service is not initialized")
	}

	loaded, err := b.service.TryLoadSession(context.Background())
	if err == nil {
		b.emitStateChanged()
	}

	return loaded, err
}

func (b *Bridge) Register(name, username, password, profilePicturePath string) error {
	return b.controller().Register(context.Background(), name, username, password, profilePicturePath)
}

func (b *Bridge) Login(username, password string) error {
	return b.controller().Login(context.Background(), username, password)
}

func (b *Bridge) Logout() error {
	return b.controller().Logout(context.Background())
}
