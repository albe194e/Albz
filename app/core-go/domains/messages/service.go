package messages

import (
	"context"

	dbsql "github.com/albe194e/albz/app/core-go/db"
	generated "github.com/albe194e/albz/app/core-go/db/sqlc/generated"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Create(ctx context.Context, params CreateParams) error {
	return dbsql.Queries.CreateMessage(ctx, params)
}

func (s *Service) Get(ctx context.Context, messageID string) (Message, error) {
	row, err := dbsql.Queries.GetMessage(ctx, messageID)
	return Message{Message: row}, err
}

func (s *Service) List(ctx context.Context) ([]Message, error) {
	rows, err := dbsql.Queries.GetMessages(ctx)
	if err != nil {
		return nil, err
	}
	return wrapMessages(rows), nil
}

func (s *Service) ListByConversation(ctx context.Context, conversationID string) ([]Message, error) {
	rows, err := dbsql.Queries.ListMessagesByConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	return wrapMessages(rows), nil
}

func (s *Service) UpdateDeliveryState(ctx context.Context, params UpdateDeliveryStateParams) error {
	return dbsql.Queries.UpdateMessageDeliveryStateByClientMessageID(ctx, params)
}

func wrapMessages(rows []generated.Message) []Message {
	result := make([]Message, 0, len(rows))
	for _, row := range rows {
		result = append(result, Message{Message: row})
	}
	return result
}
