package app

import (
	"github.com/albe194e/albz/client/app/file"
	"github.com/albe194e/albz/client/db/storage"
	"github.com/albe194e/albz/client/network"
)

type Controller struct {
	State          *AppState
	Store          *storage.Store
	Net            *network.Client
	FileHandler    *file.Handler
	OnStateChanged func()
}

func (c *Controller) notifyStateChanged() {
	if c != nil && c.OnStateChanged != nil {
		c.OnStateChanged()
	}
}
