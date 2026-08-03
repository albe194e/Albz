package contacts

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

func (s *Service) ListContacts(ctx context.Context) ([]Contact, error) {
	rows, err := s.queries.ListContacts(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Contact, 0, len(rows))
	for _, row := range rows {
		result = append(result, Contact{Contact: row})
	}
	return result, nil
}

func (s *Service) UpsertContact(ctx context.Context, params UpsertContactParams) error {
	return s.queries.UpsertContact(ctx, params)
}

func (s *Service) UpsertContactDevice(ctx context.Context, params UpsertContactDeviceParams) error {
	return s.queries.UpsertContactDevice(ctx, params)
}

func (s *Service) ListContactRequests(ctx context.Context) ([]ContactRequest, error) {
	rows, err := s.queries.ListContactRequests(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ContactRequest, 0, len(rows))
	for _, row := range rows {
		result = append(result, ContactRequest{ContactRequest: row})
	}
	return result, nil
}

func (s *Service) UpsertContactRequest(ctx context.Context, params UpsertContactRequestParams) error {
	return s.queries.UpsertContactRequest(ctx, params)
}

func (s *Service) DeleteContactRequestByFromUserID(ctx context.Context, userID string) error {
	return s.queries.DeleteContactRequestByFromUserID(ctx, userID)
}
