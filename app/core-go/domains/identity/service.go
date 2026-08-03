package identity

import (
	"context"

	dbsql "github.com/albe194e/albz/app/core-go/db"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetLocalIdentity(ctx context.Context) (LocalIdentity, error) {
	row, err := dbsql.Queries.GetLocalIdentity(ctx)
	return LocalIdentity{LocalIdentity: row}, err
}

func (s *Service) GetLocalIdentityByUserID(ctx context.Context, userID string) (LocalIdentity, error) {
	row, err := dbsql.Queries.GetLocalIdentityByUserID(ctx, userID)
	return LocalIdentity{LocalIdentity: row}, err
}

func (s *Service) UpsertLocalIdentity(ctx context.Context, params UpsertLocalIdentityParams) error {
	return dbsql.Queries.UpsertLocalIdentity(ctx, params)
}

func (s *Service) GetCurrentSession(ctx context.Context) (Session, error) {
	row, err := dbsql.Queries.GetCurrentSession(ctx)
	return Session{Session: row}, err
}

func (s *Service) UpsertCurrentSession(ctx context.Context, params UpsertCurrentSessionParams) error {
	return dbsql.Queries.UpsertCurrentSession(ctx, params)
}

func (s *Service) DeleteCurrentSession(ctx context.Context) error {
	return dbsql.Queries.DeleteCurrentSession(ctx)
}
