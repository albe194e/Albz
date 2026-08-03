package identity

import dbsql "github.com/albe194e/albz/app/core-go/db/sqlc/generated"

type LocalIdentity struct {
	dbsql.LocalIdentity
}

type Session struct {
	dbsql.Session
}

type UpsertLocalIdentityParams = dbsql.UpsertLocalIdentityParams
type UpsertCurrentSessionParams = dbsql.UpsertCurrentSessionParams
