package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	fyne "fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	clientapp "github.com/albe194e/albz/client/app"
	"github.com/albe194e/albz/client/app/file"
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

	runtimeConfig, err := loadRuntimeConfig(fyneApp, os.Args[1:])
	if err != nil {
		fmt.Printf("failed to load runtime config: %v\n", err)
		return
	}
	fmt.Printf("using client database: %s\n", runtimeConfig.dbPath)

	store, err := storage.OpenSQLite(context.Background(), runtimeConfig.dbPath, schemaSQL)
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
	fileHandler, err := file.NewHandler(filepath.Dir(runtimeConfig.dbPath))
	if err != nil {
		fmt.Printf("failed to create file handler: %v\n", err)
		return
	}
	controller := &clientapp.Controller{
		State:       appState,
		Store:       store,
		FileHandler: fileHandler,
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

	ui.Run(fyneApp, controller, uiState, runtimeConfig.windowTitle)
}

type runtimeConfig struct {
	profileName string
	dbPath      string
	windowTitle string
}

func loadRuntimeConfig(app fyne.App, args []string) (runtimeConfig, error) {
	config := runtimeConfig{
		dbPath:      filepath.Join("dev-local-db", "local_storage", "albz.db"),
		windowTitle: "Albz",
	}

	flags := flag.NewFlagSet("albz-client", flag.ContinueOnError)
	profileFlag := flags.String("profile", "", "named desktop client profile")
	dataDirFlag := flags.String("data-dir", "", "custom directory for local client data")

	if err := flags.Parse(args); err != nil {
		return runtimeConfig{}, fmt.Errorf("parse flags: %w", err)
	}

	profileName := strings.TrimSpace(*profileFlag)
	if profileName == "" {
		profileName = strings.TrimSpace(os.Getenv("ALBZ_CLIENT_PROFILE"))
	}

	dataDir := strings.TrimSpace(*dataDirFlag)
	if dataDir == "" {
		dataDir = strings.TrimSpace(os.Getenv("ALBZ_CLIENT_DATA_DIR"))
	}

	if profileName != "" && dataDir != "" {
		return runtimeConfig{}, fmt.Errorf("use either profile or data-dir, not both")
	}

	if dataDir != "" {
		config.dbPath = filepath.Join(dataDir, "albz.db")
		return config, nil
	}

	if runtime.GOOS == "android" {
		rootPath, err := appStorageRootPath(app)
		if err != nil {
			return runtimeConfig{}, fmt.Errorf("resolve app storage root: %w", err)
		}
		config.dbPath = filepath.Join(rootPath, "dev-local-db", "local_storage", "albz.db")
		return config, nil
	}

	if profileName == "" {
		return config, nil
	}

	normalizedProfile, err := normalizeProfileName(profileName)
	if err != nil {
		return runtimeConfig{}, err
	}

	config.profileName = normalizedProfile
	config.dbPath = filepath.Join("dev-local-db", "local_storage", "profiles", normalizedProfile, "albz.db")
	config.windowTitle = fmt.Sprintf("Albz (%s)", normalizedProfile)

	return config, nil
}

func normalizeProfileName(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("profile name is required")
	}

	for _, char := range trimmed {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			continue
		}
		switch char {
		case '-', '_', '.':
			continue
		default:
			return "", fmt.Errorf("profile name %q contains unsupported character %q", trimmed, string(char))
		}
	}

	return trimmed, nil
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
