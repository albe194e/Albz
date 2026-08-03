package contacts

import dbsql "github.com/albe194e/albz/app/core-go/db/sqlc/generated"

type Contact struct {
	dbsql.Contact
}

type ContactDevice struct {
	dbsql.ContactDevice
}

type ContactRequest struct {
	dbsql.ContactRequest
}

type UpsertContactParams = dbsql.UpsertContactParams
type UpsertContactDeviceParams = dbsql.UpsertContactDeviceParams
type UpsertContactRequestParams = dbsql.UpsertContactRequestParams
