package main

import (
	"context"
	"fmt"
	"path/filepath"

	clientapp "github.com/albe194e/albz/client/app"
	"github.com/albe194e/albz/client/db/storage"
	"github.com/albe194e/albz/client/input"
	"github.com/albe194e/albz/client/ui"
	_ "github.com/albe194e/albz/client/ui/pages"
)

func main() {
	fmt.Println("starting Albz")

	dbPath := filepath.Join("db", "local_storage", "albz.db")
	schemaPath := filepath.Join("db", "sqlc", "schema.sql")

	store, err := storage.OpenSQLite(context.Background(), dbPath, schemaPath)
	if err != nil {
		fmt.Printf("failed to open SQLite database: %v\n", err)
		return
	}
	defer func() {
		if err := store.Close(); err != nil {
			fmt.Printf("failed to close SQLite database: %v\n", err)
		}
	}()
	appState := &clientapp.AppState{}
	controller := &clientapp.Controller{
		State: appState,
		Store: store,
	}

	userInput := &input.Input{}
	uiState := &ui.UIState{Page: ui.Landing}
	ui.Run(controller, userInput, uiState)

}
