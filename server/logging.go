package main

import (
	"log"
	"os"
	"strings"
)

var relayDevMode = parseDevMode(os.Getenv("HADDLE_DEV_MODE"))

func logDebugf(format string, args ...any) {
	if !relayDevMode {
		return
	}
	log.Printf("[haddle][relay][debug] "+format, args...)
}

func logInfof(format string, args ...any) {
	if !relayDevMode {
		return
	}
	log.Printf("[haddle][relay][info] "+format, args...)
}

func logErrorf(format string, args ...any) {
	log.Printf("[haddle][relay][error] "+format, args...)
}

func logFatalf(format string, args ...any) {
	log.Fatalf("[haddle][relay][fatal] "+format, args...)
}

func parseDevMode(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func redactID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "(empty)"
	}
	if len(trimmed) <= 8 {
		return trimmed
	}
	return trimmed[:4] + "..." + trimmed[len(trimmed)-4:]
}
