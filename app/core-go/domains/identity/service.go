package identity

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

func (s *Service) GetLocalIdentity(ctx context.Context) (LocalIdentity, error) {
	row, err := s.queries.GetLocalIdentity(ctx)
	return LocalIdentity{LocalIdentity: row}, err
}

func (s *Service) GetLocalIdentityByUserID(ctx context.Context, userID string) (LocalIdentity, error) {
	row, err := s.queries.GetLocalIdentityByUserID(ctx, userID)
	return LocalIdentity{LocalIdentity: row}, err
}

func (s *Service) UpsertLocalIdentity(ctx context.Context, params UpsertLocalIdentityParams) error {
	return s.queries.UpsertLocalIdentity(ctx, params)
}

func (s *Service) GetCurrentSession(ctx context.Context) (Session, error) {
	row, err := s.queries.GetCurrentSession(ctx)
	return Session{Session: row}, err
}

func (s *Service) UpsertCurrentSession(ctx context.Context, params UpsertCurrentSessionParams) error {
	return s.queries.UpsertCurrentSession(ctx, params)
}

func (s *Service) DeleteCurrentSession(ctx context.Context) error {
	return s.queries.DeleteCurrentSession(ctx)
}
