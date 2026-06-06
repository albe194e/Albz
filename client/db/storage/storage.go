package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	db "github.com/albe194e/albz/client/db/sqlc/sql"

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

	if err := migrateFriendsTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate friends table: %w", err)
	}

	if err := migrateFriendRequestsTable(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migrate friend requests table: %w", err)
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

		friendCodeExpr := "friend_code"
		if _, ok := legacyColumns["friend_code"]; !ok {
			friendCodeExpr = "'ALBZ-' || UPPER(HEX(RANDOMBLOB(6)))"
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
  friend_code
)
SELECT
  %s,
  %s,
  %s,
  %s,
  %s,
  %s
FROM users_legacy;
`, idExpr, nameExpr, usernameExpr, hashedPasswordExpr, profilePictureExpr, friendCodeExpr)

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

func migrateFriendsTable(ctx context.Context, conn *sql.DB) error {
	columnTypes, err := tableColumnTypes(ctx, conn, "friends")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["name"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE friends ADD COLUMN name TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add friends.name column: %w", err)
		}
	}

	if _, ok := columnTypes["username"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE friends ADD COLUMN username TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add friends.username column: %w", err)
		}
	}

	if _, ok := columnTypes["profile_picture_url"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE friends ADD COLUMN profile_picture_url TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add friends.profile_picture_url column: %w", err)
		}
	}

	return nil
}

func migrateFriendRequestsTable(ctx context.Context, conn *sql.DB) error {
	columnTypes, err := tableColumnTypes(ctx, conn, "friend_requests")
	if err != nil {
		return err
	}

	if _, ok := columnTypes["name"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE friend_requests ADD COLUMN name TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add friend_requests.name column: %w", err)
		}
	}

	if _, ok := columnTypes["username"]; !ok {
		if _, err := conn.ExecContext(ctx, `ALTER TABLE friend_requests ADD COLUMN username TEXT NOT NULL DEFAULT '';`); err != nil {
			return fmt.Errorf("add friend_requests.username column: %w", err)
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

func usersTableNeedsIDMigration(ctx context.Context, conn *sql.DB) (bool, error) {
	columnTypes, err := tableColumnTypes(ctx, conn, "users")
	if err != nil {
		return false, err
	}

	if _, ok := columnTypes["id"]; !ok {
		return true, nil
	}

	if _, ok := columnTypes["friend_code"]; !ok {
		return true, nil
	}

	return false, nil
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
