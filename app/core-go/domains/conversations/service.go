package conversations

import (
	"context"

	dbsql "github.com/albe194e/albz/app/core-go/db/sqlc/generated"
)

type Service struct {
	queries *dbsql.Queries
}

func NewService(queries *dbsql.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) ListByUserID(ctx context.Context, userID string) ([]Conversation, error) {
	rows, err := s.queries.GetConversationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]Conversation, 0, len(rows))
	for _, row := range rows {
		result = append(result, Conversation{Conversation: row})
	}
	return result, nil
}

func (s *Service) GetByID(ctx context.Context, conversationID string) (Conversation, error) {
	row, err := s.queries.GetConversationByID(ctx, conversationID)
	return Conversation{Conversation: row}, err
}

func (s *Service) Create(ctx context.Context, params CreateParams) (Conversation, error) {
	row, err := s.queries.CreateConversation(ctx, params)
	return Conversation{Conversation: row}, err
}

func (s *Service) AddParticipant(ctx context.Context, params AddParticipantParams) error {
	return s.queries.AddParticipant(ctx, params)
}

func (s *Service) ListParticipantIDs(ctx context.Context, conversationID string) ([]string, error) {
	return s.queries.ListConversationParticipantIDs(ctx, conversationID)
}
