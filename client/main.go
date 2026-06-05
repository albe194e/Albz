package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	fyne "fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	clientapp "github.com/albe194e/albz/client/app"
	"github.com/albe194e/albz/client/db/storage"
	"github.com/albe194e/albz/client/network"
	"github.com/albe194e/albz/client/ui"
	_ "github.com/albe194e/albz/client/ui/pages"
)

//go:embed db/sqlc/schema.sql
var schemaSQL string

func main() {
	fmt.Println("starting Albz")
	fyneApp := fyneapp.New()

	dbPath := filepath.Join("db", "local_storage", "albz.db")
	if runtime.GOOS == "android" {
		rootPath, err := appStorageRootPath(fyneApp)
		if err != nil {
			fmt.Printf("failed to resolve app storage root: %v\n", err)
			return
		}
		dbPath = filepath.Join(rootPath, "db", "local_storage", "albz.db")
	}

	store, err := storage.OpenSQLite(context.Background(), dbPath, schemaSQL)
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
	controller.Net = network.NewClient(serverURL(), network.Handlers{
		OnConversationCreated:   controller.HandleConversationCreated,
		OnMessageCreated:        controller.HandleIncomingMessage,
		OnMessageDelivery:       controller.HandleMessageDelivery,
		OnFriendRequestReceived: controller.HandleFriendRequestReceived,
		OnFriendRequestAccepted: controller.HandleFriendRequestAccepted,
		OnFriendRequestRejected: controller.HandleFriendRequestRejected,
		OnError:                 controller.HandleNetworkError,
		OnDisconnect:            controller.HandleDisconnect,
	})

	uiState := &ui.UIState{}
	uiState.Init()

	ui.Run(fyneApp, controller, uiState)
}

func appStorageRootPath(app fyne.App) (string, error) {
	if app == nil {
		return "", fmt.Errorf("app is required")
	}

	rootURI := app.Storage().RootURI()
	if rootURI == nil {
		return "", fmt.Errorf("app storage root URI is nil")
	}

	rootPath := strings.TrimSpace(rootURI.Path())
	if rootPath == "" {
		return "", fmt.Errorf("app storage root path is empty")
	}

	return rootPath, nil
}

func serverURL() string {
	value := strings.TrimSpace(os.Getenv("ALBZ_SERVER_WS_URL"))
	if value == "" {
		return "ws://localhost:8080/ws"
	}

	return value
}
