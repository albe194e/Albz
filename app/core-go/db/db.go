package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	dbsql "github.com/albe194e/albz/app/core-go/db/sqlc/generated"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
	Q  *dbsql.Queries
}

func OpenSQLite(ctx context.Context, dbPath string, schemaSQL string) (*Store, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("db path is required")
	}
	if strings.TrimSpace(schemaSQL) == "" {
		return nil, fmt.Errorf("schema SQL is required")
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	conn.SetMaxOpenConns(1)

	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if _, err := conn.ExecContext(ctx, schemaSQL); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	if err := migrateLocalIdentityTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate local identity table: %w", err)
	}

	if err := migrateSessionsTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate sessions table: %w", err)
	}

	if err := migrateConversationsTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate conversations table: %w", err)
	}

	if err := migrateConversationParticipantsTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate conversation participants table: %w", err)
	}

	if err := migrateMessagesTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate messages table: %w", err)
	}

	if err := migrateContactsTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate contacts table: %w", err)
	}

	if err := migrateContactRequestsTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate contact requests table: %w", err)
	}

	return &Store{
		DB: conn,
		Q:  dbsql.New(conn),
	}, nil
}

func migrateLocalIdentityTable(ctx context.Context, conn *sql.DB) error {
	columnTypes, err := tableColumnTypes(ctx, conn, "local_identity")
	if err != nil {
		return err
	}

	requiredColumns := []string{
		"user_id",
		"device_id",
		"device_public_key",
		"encrypted_device_private_key",
		"kdf_salt",
		"kdf_params",
		"name",
		"created_at",
	}
	for _, column := range requiredColumns {
		if _, ok := columnTypes[column]; !ok {
			return fmt.Errorf("local_identity table is missing required column %q", column)
		}
	}

	return nil
}

func migrateSessionsTable(ctx context.Context, conn *sql.DB) error {
	columnTypes, err := tableColumnTypes(ctx, conn, "sessions")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["device_id"]; ok {
		return nil
	}

	if _, err := conn.ExecContext(ctx, `ALTER TABLE sessions ADD COLUMN device_id TEXT NOT NULL DEFAULT '';`); err != nil {
		return fmt.Errorf("add sessions.device_id column: %w", err)
	}

	return nil
}

func migrateConversationsTable(ctx context.Context, conn *sql.DB) error {
	columnTypes, err := tableColumnTypes(ctx, conn, "conversations")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["type"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE conversations ADD COLUMN type TEXT NOT NULL DEFAULT 'direct';`); err != nil {
			return fmt.Errorf("add conversations.type column: %w", err)
		}
	}

	if _, ok := columnTypes["created_at"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE conversations ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;`); err != nil {
			return fmt.Errorf("add conversations.created_at column: %w", err)
		}
	}

	if _, ok := columnTypes["updated_at"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE conversations ADD COLUMN updated_at INTEGER;`); err != nil {
			return fmt.Errorf("add conversations.updated_at column: %w", err)
		}
	}

	return nil
}

func migrateConversationParticipantsTable(ctx context.Context, conn *sql.DB) error {
	columnTypes, err := tableColumnTypes(ctx, conn, "conversation_participants")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["user_id"]; ok {
		if _, ok := columnTypes["created_at"]; ok {
			return nil
		}

		if _, err := conn.ExecContext(ctx, `ALTER TABLE conversation_participants ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;`); err != nil {
			return fmt.Errorf("add conversation_participants.created_at column: %w", err)
		}

		return nil
	}

	if _, ok := columnTypes["participant_id"]; !ok {
		return nil
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin conversation participants migration tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS conversation_participants_legacy;`); err != nil {
		return fmt.Errorf("drop stale legacy conversation participants table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `ALTER TABLE conversation_participants RENAME TO conversation_participants_legacy;`); err != nil {
		return fmt.Errorf("rename legacy conversation participants table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
CREATE TABLE conversation_participants (
	id INTEGER PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
`); err != nil {
		return fmt.Errorf("create migrated conversation participants table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO conversation_participants (
	id,
	conversation_id,
	user_id,
	created_at
)
SELECT
	id,
	conversation_id,
	participant_id,
	0
FROM conversation_participants_legacy;
`); err != nil {
		return fmt.Errorf("copy legacy conversation participants: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE conversation_participants_legacy;`); err != nil {
		return fmt.Errorf("drop legacy conversation participants table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
CREATE UNIQUE INDEX IF NOT EXISTS idx_conversation_participants_unique
ON conversation_participants (conversation_id, user_id);
`); err != nil {
		return fmt.Errorf("create conversation participants unique index: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit conversation participants migration: %w", err)
	}

	return nil
}

func migrateMessagesTable(ctx context.Context, conn *sql.DB) error {
	needsMigration, err := messagesTableNeedsMigration(ctx, conn)
	if err != nil {
		return err
	}
	if !needsMigration {
		return nil
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin messages migration tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS messages_legacy;`); err != nil {
		return fmt.Errorf("drop stale legacy messages table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `ALTER TABLE messages RENAME TO messages_legacy;`); err != nil {
		return fmt.Errorf("rename legacy messages table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
CREATE TABLE messages (
	id TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	sender_user_id TEXT NOT NULL,
	sender_device_id TEXT,
	client_message_id TEXT NOT NULL,
	body TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	received_at INTEGER,
	direction TEXT NOT NULL,
	delivery_state TEXT NOT NULL,
	FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
`); err != nil {
		return fmt.Errorf("create migrated messages table: %w", err)
	}

	legacyColumns, err := tableColumnNames(ctx, tx, "messages_legacy")
	if err != nil {
		return err
	}

	requiredColumns := []string{"conversation_id", "body", "created_at"}
	for _, column := range requiredColumns {
		if _, ok := legacyColumns[column]; !ok {
			return fmt.Errorf("legacy messages table is missing required column %q", column)
		}
	}

	senderUserExpr := "sender_user_id"
	if _, ok := legacyColumns["sender_user_id"]; !ok {
		if _, ok := legacyColumns["sender_id"]; ok {
			senderUserExpr = "sender_id"
		} else {
			return fmt.Errorf("legacy messages table is missing sender_user_id/sender_id")
		}
	}

	messageIDExpr := "'legacy-' || rowid"
	if _, ok := legacyColumns["id"]; ok {
		messageIDExpr = "CAST(id AS TEXT)"
	}

	clientMessageIDExpr := "'legacy-client-' || rowid"
	if _, ok := legacyColumns["client_message_id"]; ok {
		clientMessageIDExpr = "client_message_id"
	}

	receivedAtExpr := "NULL"
	if _, ok := legacyColumns["received_at"]; ok {
		receivedAtExpr = "received_at"
	}

	directionExpr := fmt.Sprintf(
		"CASE WHEN %s = COALESCE((SELECT user_id FROM local_identity WHERE id = 1), (SELECT user_id FROM sessions WHERE id = 1)) THEN 'outgoing' ELSE 'incoming' END",
		senderUserExpr,
	)
	if _, ok := legacyColumns["direction"]; ok {
		directionExpr = "direction"
	}

	deliveryStateExpr := "'delivered'"
	if _, ok := legacyColumns["delivery_state"]; ok {
		deliveryStateExpr = "delivery_state"
	}

	copyQuery := fmt.Sprintf(`
INSERT INTO messages (
	id,
	conversation_id,
	sender_user_id,
	sender_device_id,
	client_message_id,
	body,
	created_at,
	received_at,
	direction,
	delivery_state
)
SELECT
	%s,
	conversation_id,
	%s,
	NULL,
	%s,
	body,
	created_at,
	%s,
	%s,
	%s
FROM messages_legacy
ORDER BY created_at ASC, rowid ASC;
`, messageIDExpr, senderUserExpr, clientMessageIDExpr, receivedAtExpr, directionExpr, deliveryStateExpr)

	if _, err := tx.ExecContext(ctx, copyQuery); err != nil {
		return fmt.Errorf("copy legacy messages: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE messages_legacy;`); err != nil {
		return fmt.Errorf("drop legacy messages table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
CREATE INDEX IF NOT EXISTS idx_messages_conversation_created
ON messages (conversation_id, created_at);
`); err != nil {
		return fmt.Errorf("create messages conversation index: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_client_message_id
ON messages (client_message_id);
`); err != nil {
		return fmt.Errorf("create messages client message index: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit messages migration: %w", err)
	}

	return nil
}

func migrateContactsTable(ctx context.Context, conn *sql.DB) error {
	hasLegacyTable, err := tableExists(ctx, conn, "friends")
	if err != nil {
		return err
	}
	if hasLegacyTable {
		if err := migrateLegacyContactsTable(ctx, conn); err != nil {
			return err
		}
	}

	columnTypes, err := tableColumnTypes(ctx, conn, "contacts")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["display_name"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contacts ADD COLUMN display_name TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add contacts.display_name column: %w", err)
		}
	}

	if _, ok := columnTypes["local_handle"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contacts ADD COLUMN local_handle TEXT;`); err != nil {
			return fmt.Errorf("add contacts.local_handle column: %w", err)
		}
	}

	if _, ok := columnTypes["profile_picture_path"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contacts ADD COLUMN profile_picture_path TEXT;`); err != nil {
			return fmt.Errorf("add contacts.profile_picture_path column: %w", err)
		}
	}

	columnTypes, err = tableColumnTypes(ctx, conn, "contacts")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["name"]; ok {
		if _, err := conn.ExecContext(ctx, `
UPDATE contacts
SET display_name = name
WHERE COALESCE(display_name, '') = '';
`); err != nil {
			return fmt.Errorf("backfill contacts.display_name: %w", err)
		}
	}

	if _, ok := columnTypes["username"]; ok {
		if _, err := conn.ExecContext(ctx, `
UPDATE contacts
SET local_handle = username
WHERE COALESCE(local_handle, '') = '';
`); err != nil {
			return fmt.Errorf("backfill contacts.local_handle: %w", err)
		}
	}

	if _, ok := columnTypes["profile_picture_url"]; ok {
		if _, err := conn.ExecContext(ctx, `
UPDATE contacts
SET profile_picture_path = profile_picture_url
WHERE COALESCE(profile_picture_path, '') = '';
`); err != nil {
			return fmt.Errorf("backfill contacts.profile_picture_path: %w", err)
		}
	}

	return nil
}

func migrateContactRequestsTable(ctx context.Context, conn *sql.DB) error {
	hasLegacyTable, err := tableExists(ctx, conn, "friend_requests")
	if err != nil {
		return err
	}
	if hasLegacyTable {
		if err := migrateLegacyContactRequestsTable(ctx, conn); err != nil {
			return err
		}
	}

	columnTypes, err := tableColumnTypes(ctx, conn, "contact_requests")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["from_device_id"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN from_device_id TEXT;`); err != nil {
			return fmt.Errorf("add contact_requests.from_device_id column: %w", err)
		}
	}

	if _, ok := columnTypes["display_name"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN display_name TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add contact_requests.display_name column: %w", err)
		}
	}

	if _, ok := columnTypes["local_handle"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN local_handle TEXT;`); err != nil {
			return fmt.Errorf("add contact_requests.local_handle column: %w", err)
		}
	}

	if _, ok := columnTypes["profile_picture_path"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN profile_picture_path TEXT;`); err != nil {
			return fmt.Errorf("add contact_requests.profile_picture_path column: %w", err)
		}
	}

	if _, ok := columnTypes["from_public_key"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN from_public_key BLOB;`); err != nil {
			return fmt.Errorf("add contact_requests.from_public_key column: %w", err)
		}
	}

	if _, ok := columnTypes["invite_payload"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN invite_payload TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add contact_requests.invite_payload column: %w", err)
		}
	}

	if _, ok := columnTypes["state"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN state TEXT NOT NULL DEFAULT 'pending';`); err != nil {
			return fmt.Errorf("add contact_requests.state column: %w", err)
		}
	}

	columnTypes, err = tableColumnTypes(ctx, conn, "contact_requests")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["name"]; ok {
		if _, err := conn.ExecContext(ctx, `
UPDATE contact_requests
SET display_name = name
WHERE COALESCE(display_name, '') = '';
`); err != nil {
			return fmt.Errorf("backfill contact_requests.display_name: %w", err)
		}
	}

	if _, ok := columnTypes["username"]; ok {
		if _, err := conn.ExecContext(ctx, `
UPDATE contact_requests
SET local_handle = username
WHERE COALESCE(local_handle, '') = '';
`); err != nil {
			return fmt.Errorf("backfill contact_requests.local_handle: %w", err)
		}
	}

	if _, ok := columnTypes["from_contact_code"]; ok {
		if _, err := conn.ExecContext(ctx, `
UPDATE contact_requests
SET invite_payload = COALESCE(from_contact_code, '')
WHERE COALESCE(invite_payload, '') = '';
`); err != nil {
			return fmt.Errorf("backfill contact_requests.invite_payload: %w", err)
		}
	}

	return nil
}

func messagesTableNeedsMigration(ctx context.Context, conn *sql.DB) (bool, error) {
	columnTypes, err := tableColumnTypes(ctx, conn, "messages")
	if err != nil {
		return false, err
	}

	requiredColumns := []string{
		"id",
		"sender_user_id",
		"client_message_id",
		"received_at",
		"direction",
		"delivery_state",
	}
	for _, column := range requiredColumns {
		if _, ok := columnTypes[column]; !ok {
			return true, nil
		}
	}

	return false, nil
}

func migrateLegacyContactsTable(ctx context.Context, conn *sql.DB) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin contacts migration tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	legacyColumns, err := tableColumnNames(ctx, tx, "friends")
	if err != nil {
		return err
	}

	requiredColumns := []string{"user_id", "created_at"}
	for _, column := range requiredColumns {
		if _, ok := legacyColumns[column]; !ok {
			return fmt.Errorf("legacy friends table is missing required column %q", column)
		}
	}

	displayNameExpr := "''"
	if _, ok := legacyColumns["name"]; ok {
		displayNameExpr = "name"
	}

	localHandleExpr := "NULL"
	if _, ok := legacyColumns["username"]; ok {
		localHandleExpr = "username"
	}

	profilePictureExpr := "NULL"
	if _, ok := legacyColumns["profile_picture_url"]; ok {
		profilePictureExpr = "profile_picture_url"
	}

	contactCodeExpr := "NULL"
	if _, ok := legacyColumns["contact_code"]; ok {
		contactCodeExpr = "contact_code"
	} else if _, ok := legacyColumns["friend_code"]; ok {
		contactCodeExpr = "friend_code"
	}

	copyQuery := fmt.Sprintf(`
INSERT INTO contacts (
	user_id,
	display_name,
	local_handle,
	profile_picture_path,
	contact_code,
	created_at
)
SELECT
	user_id,
	%s,
	%s,
	%s,
	%s,
	created_at
FROM friends
ON CONFLICT(user_id) DO UPDATE SET
	display_name = excluded.display_name,
	local_handle = excluded.local_handle,
	profile_picture_path = excluded.profile_picture_path,
	contact_code = excluded.contact_code,
	created_at = excluded.created_at;
`, displayNameExpr, localHandleExpr, profilePictureExpr, contactCodeExpr)

	if _, err := tx.ExecContext(ctx, copyQuery); err != nil {
		return fmt.Errorf("copy legacy contacts: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE friends;`); err != nil {
		return fmt.Errorf("drop legacy friends table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit contacts migration: %w", err)
	}

	return nil
}

func migrateLegacyContactRequestsTable(ctx context.Context, conn *sql.DB) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin contact requests migration tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	legacyColumns, err := tableColumnNames(ctx, tx, "friend_requests")
	if err != nil {
		return err
	}

	requiredColumns := []string{"from_user_id", "created_at"}
	for _, column := range requiredColumns {
		if _, ok := legacyColumns[column]; !ok {
			return fmt.Errorf("legacy friend_requests table is missing required column %q", column)
		}
	}

	displayNameExpr := "''"
	if _, ok := legacyColumns["name"]; ok {
		displayNameExpr = "name"
	}

	localHandleExpr := "NULL"
	if _, ok := legacyColumns["username"]; ok {
		localHandleExpr = "username"
	}

	fromContactCodeExpr := "NULL"
	if _, ok := legacyColumns["from_contact_code"]; ok {
		fromContactCodeExpr = "from_contact_code"
	} else if _, ok := legacyColumns["from_friend_code"]; ok {
		fromContactCodeExpr = "from_friend_code"
	}

	copyQuery := fmt.Sprintf(`
INSERT INTO contact_requests (
	from_user_id,
	display_name,
	local_handle,
	from_contact_code,
	invite_payload,
	state,
	created_at
)
SELECT
	from_user_id,
	%s,
	%s,
	%s,
	COALESCE(%s, ''),
	'pending',
	created_at
FROM friend_requests
ON CONFLICT(from_user_id) DO UPDATE SET
	display_name = excluded.display_name,
	local_handle = excluded.local_handle,
	from_contact_code = excluded.from_contact_code,
	invite_payload = excluded.invite_payload,
	state = excluded.state,
	created_at = excluded.created_at;
`, displayNameExpr, localHandleExpr, fromContactCodeExpr, fromContactCodeExpr)

	if _, err := tx.ExecContext(ctx, copyQuery); err != nil {
		return fmt.Errorf("copy legacy contact requests: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE friend_requests;`); err != nil {
		return fmt.Errorf("drop legacy friend_requests table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit contact requests migration: %w", err)
	}

	return nil
}

func tableExists(ctx context.Context, conn *sql.DB, tableName string) (bool, error) {
	row := conn.QueryRowContext(
		ctx,
		`SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?;`,
		tableName,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check %s table: %w", tableName, err)
	}

	return count > 0, nil
}

func tableColumnNames(ctx context.Context, conn *sql.Tx, tableName string) (map[string]struct{}, error) {
	rows, err := conn.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s);`, tableName))
	if err != nil {
		return nil, fmt.Errorf("inspect %s table: %w", tableName, err)
	}
	defer rows.Close()

	columns := make(map[string]struct{})
	for rows.Next() {
		var (
			cid     int
			name    string
			colType string
			notNull int
			pk      int
			dflt    sql.NullString
		)

		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return nil, fmt.Errorf("scan %s table info: %w", tableName, err)
		}

		columns[name] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s table info: %w", tableName, err)
	}

	return columns, nil
}

func tableColumnTypes(ctx context.Context, conn *sql.DB, tableName string) (map[string]string, error) {
	rows, err := conn.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s);`, tableName))
	if err != nil {
		return nil, fmt.Errorf("inspect %s table: %w", tableName, err)
	}
	defer rows.Close()

	columns := make(map[string]string)
	for rows.Next() {
		var (
			cid     int
			name    string
			colType string
			notNull int
			pk      int
			dflt    sql.NullString
		)

		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return nil, fmt.Errorf("scan %s table info: %w", tableName, err)
		}

		columns[name] = colType
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s table info: %w", tableName, err)
	}

	return columns, nil
}

func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}

	return s.DB.Close()
}
