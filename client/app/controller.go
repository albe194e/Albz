package app

import "github.com/albe194e/albz/client/network"
import "github.com/albe194e/albz/client/db/storage"

type Controller struct {
	State          *AppState
	Store          *storage.Store
	Net            *network.Client
	OnStateChanged func()
}

func (c *Controller) notifyStateChanged() {
	if c != nil && c.OnStateChanged != nil {
		c.OnStateChanged()
	}
}
