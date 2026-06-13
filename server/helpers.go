package main

import (
	"errors"
	"strings"
)

func normalizeRecipientUserIDs(senderUserID string, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, errors.New("at least one recipient is required")
	}

	normalized := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			return nil, errors.New("recipient user IDs must not be empty")
		}
		if trimmed == senderUserID {
			return nil, errors.New("cannot create a conversation with yourself")
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}

		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	if len(normalized) == 0 {
		return nil, errors.New("at least one recipient is required")
	}

	return normalized, nil
}
