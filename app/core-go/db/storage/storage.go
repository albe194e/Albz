package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	db "github.com/albe194e/albz/app/core-go/db/sqlc/sql"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
	Q  *db.Queries
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

	schemaBytes := []byte(schemaSQL)

	if _, err := conn.ExecContext(ctx, string(schemaBytes)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	if err := migrateUsersTable(ctx, conn, schemaBytes); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate users table: %w", err)
	}

	if err := migrateConversationsTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate conversations table: %w", err)
	}

	if err := migrateMessagesTable(ctx, conn, schemaBytes); err != nil {
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
		Q:  db.New(conn),
	}, nil
}

func migrateUsersTable(ctx context.Context, conn *sql.DB, schemaBytes []byte) error {
	needsMigration, err := usersTableNeedsIDMigration(ctx, conn)
	if err != nil {
		return err
	}
	if needsMigration {
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin users migration tx: %w", err)
		}
		defer func() {
			_ = tx.Rollback()
		}()

		if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS users_legacy;`); err != nil {
			return fmt.Errorf("drop stale legacy users table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `ALTER TABLE users RENAME TO users_legacy;`); err != nil {
			return fmt.Errorf("rename legacy users table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, string(schemaBytes)); err != nil {
			return fmt.Errorf("create migrated users table: %w", err)
		}

		legacyColumns, err := tableColumnNames(ctx, tx, "users_legacy")
		if err != nil {
			return err
		}

		idExpr := "id"
		if _, ok := legacyColumns["id"]; !ok {
			if _, ok := legacyColumns["uuid"]; ok {
				idExpr = "uuid"
			} else {
				return fmt.Errorf("legacy users table is missing both id and uuid columns")
			}
		}

		nameExpr := "name"
		if _, ok := legacyColumns["name"]; !ok {
			return fmt.Errorf("legacy users table is missing required column %q", "name")
		}

		usernameExpr := "username"
		if _, ok := legacyColumns["username"]; !ok {
			usernameExpr = idExpr
		}

		hashedPasswordExpr := "hashed_password"
		if _, ok := legacyColumns["hashed_password"]; !ok {
			hashedPasswordExpr = "''"
		}

		contactCodeExpr := "contact_code"
		if _, ok := legacyColumns["contact_code"]; !ok {
			if _, ok := legacyColumns["friend_code"]; ok {
				contactCodeExpr = "friend_code"
			} else {
				contactCodeExpr = "'ALBZ-' || UPPER(HEX(RANDOMBLOB(6)))"
			}
		}

		profilePictureExpr := "profile_picture_url"
		if _, ok := legacyColumns["profile_picture_url"]; !ok {
			profilePictureExpr = "''"
		}

		copyQuery := fmt.Sprintf(`
INSERT INTO users (
  id,
  name,
  username,
  hashed_password,
  profile_picture_url,
  contact_code
)
SELECT
  %s,
  %s,
  %s,
  %s,
  %s,
  %s
FROM users_legacy;
`, idExpr, nameExpr, usernameExpr, hashedPasswordExpr, profilePictureExpr, contactCodeExpr)

		if _, err := tx.ExecContext(ctx, copyQuery); err != nil {
			return fmt.Errorf("copy legacy users: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `DROP TABLE users_legacy;`); err != nil {
			return fmt.Errorf("drop legacy users table: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit users migration: %w", err)
		}
	}

	columnTypes, err := tableColumnTypes(ctx, conn, "users")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["profile_picture_url"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE users ADD COLUMN profile_picture_url TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add users.profile_picture_url column: %w", err)
		}
	}

	return nil
}

func migrateMessagesTable(ctx context.Context, conn *sql.DB, schemaBytes []byte) error {
	needsMigration, err := messagesTableNeedsMigration(ctx, conn)
	if err != nil {
		return err
	}
	if !needsMigration {
		return nil
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS messages_legacy;`); err != nil {
		return fmt.Errorf("drop stale legacy table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `ALTER TABLE messages RENAME TO messages_legacy;`); err != nil {
		return fmt.Errorf("rename legacy messages table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, string(schemaBytes)); err != nil {
		return fmt.Errorf("create migrated messages table: %w", err)
	}

	legacyColumns, err := tableColumnNames(ctx, tx, "messages_legacy")
	if err != nil {
		return err
	}

	requiredColumns := []string{"conversation_id", "sender_id", "body", "created_at"}
	for _, column := range requiredColumns {
		if _, ok := legacyColumns[column]; !ok {
			return fmt.Errorf("legacy messages table is missing required column %q", column)
		}
	}

	clientMessageIDExpr := "'legacy-' || rowid"
	if _, ok := legacyColumns["client_message_id"]; ok {
		clientMessageIDExpr = "client_message_id"
	}

	deliveryStateExpr := "'delivered'"
	if _, ok := legacyColumns["delivery_state"]; ok {
		deliveryStateExpr = "delivery_state"
	}

	copyQuery := fmt.Sprintf(`
INSERT INTO messages (
  conversation_id,
  sender_id,
  client_message_id,
  body,
  created_at,
  delivery_state
)
SELECT
  conversation_id,
  sender_id,
  %s,
  body,
  created_at,
  %s
FROM messages_legacy
ORDER BY created_at ASC, rowid ASC;
`, clientMessageIDExpr, deliveryStateExpr)

	if _, err := tx.ExecContext(ctx, copyQuery); err != nil {
		return fmt.Errorf("copy legacy messages: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE messages_legacy;`); err != nil {
		return fmt.Errorf("drop legacy messages table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}

	return nil
}

func migrateConversationsTable(ctx context.Context, conn *sql.DB) error {
	columnTypes, err := tableColumnTypes(ctx, conn, "conversations")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["name"]; ok {
		return nil
	}

	if _, err := conn.ExecContext(ctx, `ALTER TABLE conversations ADD COLUMN name TEXT NOT NULL DEFAULT '';`); err != nil {
		return fmt.Errorf("add conversations.name column: %w", err)
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

	if _, ok := columnTypes["name"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contacts ADD COLUMN name TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add contacts.name column: %w", err)
		}
	}

	if _, ok := columnTypes["username"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contacts ADD COLUMN username TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add contacts.username column: %w", err)
		}
	}

	if _, ok := columnTypes["profile_picture_url"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contacts ADD COLUMN profile_picture_url TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add contacts.profile_picture_url column: %w", err)
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

	if _, ok := columnTypes["name"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN name TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add contact_requests.name column: %w", err)
		}
	}

	if _, ok := columnTypes["username"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE contact_requests ADD COLUMN username TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add contact_requests.username column: %w", err)
		}
	}

	return nil
}

func messagesTableNeedsMigration(ctx context.Context, conn *sql.DB) (bool, error) {
	columnTypes, err := tableColumnTypes(ctx, conn, "messages")
	if err != nil {
		return false, err
	}

	idType, ok := columnTypes["id"]
	if !ok {
		return true, nil
	}

	if !strings.EqualFold(strings.TrimSpace(idType), "INTEGER") {
		return true, nil
	}

	if _, ok := columnTypes["client_message_id"]; !ok {
		return true, nil
	}

	if _, ok := columnTypes["delivery_state"]; !ok {
		return true, nil
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

	nameExpr := "''"
	if _, ok := legacyColumns["name"]; ok {
		nameExpr = "name"
	}

	usernameExpr := "''"
	if _, ok := legacyColumns["username"]; ok {
		usernameExpr = "username"
	}

	profilePictureExpr := "''"
	if _, ok := legacyColumns["profile_picture_url"]; ok {
		profilePictureExpr = "profile_picture_url"
	}

	contactCodeExpr := "''"
	if _, ok := legacyColumns["contact_code"]; ok {
		contactCodeExpr = "contact_code"
	} else if _, ok := legacyColumns["friend_code"]; ok {
		contactCodeExpr = "friend_code"
	}

	copyQuery := fmt.Sprintf(`
INSERT INTO contacts (
  user_id,
  name,
  username,
  profile_picture_url,
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
  name = excluded.name,
  username = excluded.username,
  profile_picture_url = excluded.profile_picture_url,
  contact_code = excluded.contact_code,
  created_at = excluded.created_at;
`, nameExpr, usernameExpr, profilePictureExpr, contactCodeExpr)

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

	nameExpr := "''"
	if _, ok := legacyColumns["name"]; ok {
		nameExpr = "name"
	}

	usernameExpr := "''"
	if _, ok := legacyColumns["username"]; ok {
		usernameExpr = "username"
	}

	fromContactCodeExpr := "''"
	if _, ok := legacyColumns["from_contact_code"]; ok {
		fromContactCodeExpr = "from_contact_code"
	} else if _, ok := legacyColumns["from_friend_code"]; ok {
		fromContactCodeExpr = "from_friend_code"
	}

	copyQuery := fmt.Sprintf(`
INSERT INTO contact_requests (
  from_user_id,
  name,
  username,
  from_contact_code,
  created_at
)
SELECT
  from_user_id,
  %s,
  %s,
  %s,
  created_at
FROM friend_requests
ON CONFLICT(from_user_id) DO UPDATE SET
  name = excluded.name,
  username = excluded.username,
  from_contact_code = excluded.from_contact_code,
  created_at = excluded.created_at;
`, nameExpr, usernameExpr, fromContactCodeExpr)

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

func usersTableNeedsIDMigration(ctx context.Context, conn *sql.DB) (bool, error) {
	columnTypes, err := tableColumnTypes(ctx, conn, "users")
	if err != nil {
		return false, err
	}

	if _, ok := columnTypes["id"]; !ok {
		return true, nil
	}

	if _, ok := columnTypes["contact_code"]; !ok {
		return true, nil
	}

	return false, nil
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
