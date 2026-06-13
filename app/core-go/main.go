package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	corelog "github.com/albe194e/albz/app/core-go/logging"
	coreruntime "github.com/albe194e/albz/app/core-go/runtime"
)

func main() {
	options, err := loadOptions(os.Args[1:])
	if err != nil {
		corelog.Errorf("failed to load runtime options: %v", err)
		return
	}

	service, err := coreruntime.New(context.Background(), options)
	if err != nil {
		corelog.Errorf("failed to start core runtime: %v", err)
		return
	}
	defer func() {
		if err := service.Close(); err != nil {
			corelog.Errorf("failed to close core runtime: %v", err)
		}
	}()

	corelog.Infof("core-go database: %s", service.Config.DBPath)
	loaded, err := service.TryLoadSession(context.Background())
	if err != nil {
		corelog.Errorf("failed to verify local session: %v", err)
		return
	}
	if loaded {
		corelog.Infof("loaded existing local session")
	} else {
		corelog.Infof("no active local session")
	}

	corelog.Infof("core-go runtime ready")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}

func loadOptions(args []string) (coreruntime.Options, error) {
	flags := flag.NewFlagSet("core-go", flag.ContinueOnError)
	profileFlag := flags.String("profile", "", "named desktop client profile")
	dataDirFlag := flags.String("data-dir", "", "custom directory for local client data")
	serverURLFlag := flags.String("server-url", "", "relay websocket URL")

	if err := flags.Parse(args); err != nil {
		return coreruntime.Options{}, err
	}

	serverURL := strings.TrimSpace(*serverURLFlag)
	if serverURL == "" {
		serverURL = strings.TrimSpace(os.Getenv("HADDLE_SERVER_WS_URL"))
	}

	return coreruntime.Options{
		ProfileName: strings.TrimSpace(*profileFlag),
		DataDir:     strings.TrimSpace(*dataDirFlag),
		ServerURL:   serverURL,
	}, nil
}
