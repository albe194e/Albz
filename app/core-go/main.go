package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	coreruntime "github.com/albe194e/albz/app/core-go/runtime"
)

func main() {
	options, err := loadOptions(os.Args[1:])
	if err != nil {
		fmt.Printf("failed to load runtime options: %v\n", err)
		return
	}

	service, err := coreruntime.New(context.Background(), options)
	if err != nil {
		fmt.Printf("failed to start core runtime: %v\n", err)
		return
	}
	defer func() {
		if err := service.Close(); err != nil {
			fmt.Printf("failed to close core runtime: %v\n", err)
		}
	}()

	fmt.Printf("core-go database: %s\n", service.Config.DBPath)
	loaded, err := service.TryLoadSession(context.Background())
	if err != nil {
		fmt.Printf("failed to verify local session: %v\n", err)
		return
	}
	if loaded {
		fmt.Println("loaded existing local session")
	} else {
		fmt.Println("no active local session")
	}

	fmt.Println("core-go runtime ready")

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
		serverURL = strings.TrimSpace(os.Getenv("ALBZ_SERVER_WS_URL"))
	}

	return coreruntime.Options{
		ProfileName: strings.TrimSpace(*profileFlag),
		DataDir:     strings.TrimSpace(*dataDirFlag),
		ServerURL:   serverURL,
	}, nil
}
