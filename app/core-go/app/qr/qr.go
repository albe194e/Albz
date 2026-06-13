package qr

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/skip2/go-qrcode"
)

type QRPayload struct {
}

const (
	contactCodePrefix = "HADDLE-"
	deepLinkScheme    = "haddle"
	addContactHost    = "add-contact"
)

func GenerateContactQRCode(contactCode string) ([]byte, error) {
	deepLink, err := contactDeepLink(contactCode)
	if err != nil {
		return nil, err
	}

	return generatePNGBytes(deepLink)
}

func contactDeepLink(contactCode string) (string, error) {
	normalizedContactCode := strings.ToUpper(strings.TrimSpace(contactCode))
	if normalizedContactCode == "" {
		return "", fmt.Errorf("contact code is required")
	}
	if !strings.HasPrefix(normalizedContactCode, contactCodePrefix) {
		return "", fmt.Errorf("contact code must start with %s", contactCodePrefix)
	}

	return fmt.Sprintf(
		"%s://%s?code=%s",
		deepLinkScheme,
		addContactHost,
		url.QueryEscape(normalizedContactCode),
	), nil
}

func generatePNGBytes(data string) ([]byte, error) {
	png, err := qrcode.Encode(data, qrcode.Highest, 256)
	if err != nil {
		return nil, err
	}

	return png, nil
}
