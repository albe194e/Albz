package logging

import (
	"log"
	"os"
	"strings"
)

var devMode = parseDevMode(os.Getenv("HADDLE_DEV_MODE"))

func DevMode() bool {
	return devMode
}

func Debugf(format string, args ...any) {
	if !devMode {
		return
	}
	log.Printf("[haddle][core-go][debug] "+format, args...)
}

func Infof(format string, args ...any) {
	if !devMode {
		return
	}
	log.Printf("[haddle][core-go][info] "+format, args...)
}

func Errorf(format string, args ...any) {
	log.Printf("[haddle][core-go][error] "+format, args...)
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
