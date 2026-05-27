package app

import "github.com/albe194e/albz/client/db/storage"

type Controller struct {
	State *AppState
	Store *storage.Store
}
