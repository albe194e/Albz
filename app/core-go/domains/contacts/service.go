package contacts

import (
	"context"

	dbsql "github.com/albe194e/albz/app/core-go/db"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ListContacts(ctx context.Context) ([]Contact, error) {
	rows, err := dbsql.Queries.ListContacts(ctx)
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
	return dbsql.Queries.UpsertContact(ctx, params)
}

func (s *Service) UpsertContactDevice(ctx context.Context, params UpsertContactDeviceParams) error {
	return dbsql.Queries.UpsertContactDevice(ctx, params)
}

func (s *Service) ListContactRequests(ctx context.Context) ([]ContactRequest, error) {
	rows, err := dbsql.Queries.ListContactRequests(ctx)
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
	return dbsql.Queries.UpsertContactRequest(ctx, params)
}

func (s *Service) DeleteContactRequestByFromUserID(ctx context.Context, userID string) error {
	return dbsql.Queries.DeleteContactRequestByFromUserID(ctx, userID)
}
