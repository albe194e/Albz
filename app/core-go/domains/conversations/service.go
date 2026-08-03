package conversations

import (
	"context"

	dbsql "github.com/albe194e/albz/app/core-go/db"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ListByUserID(ctx context.Context, userID string) ([]Conversation, error) {
	rows, err := dbsql.Queries.GetConversationsByUserID(ctx, userID)
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
	row, err := dbsql.Queries.GetConversationByID(ctx, conversationID)
	return Conversation{Conversation: row}, err
}

func (s *Service) Create(ctx context.Context, params CreateParams) (Conversation, error) {
	row, err := dbsql.Queries.CreateConversation(ctx, params)
	return Conversation{Conversation: row}, err
}

func (s *Service) AddParticipant(ctx context.Context, params AddParticipantParams) error {
	return dbsql.Queries.AddParticipant(ctx, params)
}

func (s *Service) ListParticipantIDs(ctx context.Context, conversationID string) ([]string, error) {
	return dbsql.Queries.ListConversationParticipantIDs(ctx, conversationID)
}
