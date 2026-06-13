package runtime

import (
	"context"
	dsql "database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	clientapp "github.com/albe194e/albz/app/core-go/app"
	"github.com/albe194e/albz/app/core-go/app/file"
	dbsqlc "github.com/albe194e/albz/app/core-go/db/sqlc"
	"github.com/albe194e/albz/app/core-go/db/storage"
	"github.com/albe194e/albz/app/core-go/network"
)

type Options struct {
	ProfileName string
	DataDir     string
	ServerURL   string
}

type Config struct {
	ProfileName string
	DataDir     string
	DBPath      string
	ServerURL   string
}

type Service struct {
	Config      Config
	Store       *storage.Store
	Controller  *clientapp.Controller
	FileHandler *file.Handler
}

func New(ctx context.Context, options Options) (*Service, error) {
	config, err := resolveConfig(options)
	if err != nil {
		return nil, err
	}

	store, err := storage.OpenSQLite(ctx, config.DBPath, dbsqlc.SchemaSQL)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	fileHandler, err := file.NewHandler(filepath.Dir(config.DBPath))
	if err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("create file handler: %w", err)
	}

	appState := &clientapp.AppState{}
	controller := &clientapp.Controller{
		State:       appState,
		Store:       store,
		FileHandler: fileHandler,
	}
	controller.Net = network.NewClient(config.ServerURL, network.Handlers{
		OnConversationCreated:    controller.HandleConversationCreated,
		OnMessageCreated:         controller.HandleIncomingMessage,
		OnMessageDelivery:        controller.HandleMessageDelivery,
		OnContactRequestReceived: controller.HandleContactRequestReceived,
		OnContactRequestAccepted: controller.HandleContactRequestAccepted,
		OnContactRequestRejected: controller.HandleContactRequestRejected,
		OnError:                  controller.HandleNetworkError,
		OnDisconnect:             controller.HandleDisconnect,
	})

	return &Service{
		Config:      config,
		Store:       store,
		Controller:  controller,
		FileHandler: fileHandler,
	}, nil
}

func (s *Service) Close() error {
	if s == nil {
		return nil
	}

	var closeErr error
	if s.Controller != nil && s.Controller.Net != nil {
		closeErr = s.Controller.Net.Close()
	}
	if s.Store != nil {
		if err := s.Store.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}

	return closeErr
}

func (s *Service) TryLoadSession(ctx context.Context) (bool, error) {
	if s == nil || s.Controller == nil {
		return false, fmt.Errorf("service controller is required")
	}

	err := s.Controller.VerifySession(ctx)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, dsql.ErrNoRows) || err.Error() == "session has expired" {
		return false, nil
	}

	return false, err
}

func resolveConfig(options Options) (Config, error) {
	config := Config{
		ServerURL: defaultServerURL(strings.TrimSpace(options.ServerURL)),
	}

	profileName := strings.TrimSpace(options.ProfileName)
	dataDir := strings.TrimSpace(options.DataDir)

	if profileName != "" && dataDir != "" {
		return Config{}, fmt.Errorf("use either profile or data-dir, not both")
	}

	if dataDir != "" {
		absoluteDataDir, err := filepath.Abs(dataDir)
		if err != nil {
			return Config{}, fmt.Errorf("resolve data dir: %w", err)
		}
		config.DataDir = absoluteDataDir
		config.DBPath = filepath.Join(absoluteDataDir, "haddle.db")
		return config, nil
	}

	if runtime.GOOS == "android" {
		return Config{}, fmt.Errorf("android data dir must be provided by the host platform")
	}

	if profileName == "" {
		config.DataDir = filepath.Join("dev-local-db", "local_storage")
		config.DBPath = filepath.Join(config.DataDir, "haddle.db")
		return config, nil
	}

	normalizedProfile, err := normalizeProfileName(profileName)
	if err != nil {
		return Config{}, err
	}

	config.ProfileName = normalizedProfile
	config.DataDir = filepath.Join("dev-local-db", "local_storage", "profiles", normalizedProfile)
	config.DBPath = filepath.Join(config.DataDir, "haddle.db")
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

func defaultServerURL(value string) string {
	if value == "" {
		value = strings.TrimSpace(os.Getenv("HADDLE_SERVER_WS_URL"))
	}
	if value == "" {
		return "ws://localhost:8080/ws"
	}

	return value
}
