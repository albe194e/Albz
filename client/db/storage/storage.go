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

func OpenSQLite(ctx context.Context, dbPath string, schemaPath string) (*Store, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("db path is required")
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

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("read schema file %q: %w", schemaPath, err)
	}

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
	if !needsMigration {
		return nil
	}

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

	copyQuery := fmt.Sprintf(`
INSERT INTO users (
  id,
  name,
  username,
  hashed_password
)
SELECT
  %s,
  %s,
  %s,
  %s
FROM users_legacy;
`, idExpr, nameExpr, usernameExpr, hashedPasswordExpr)

	if _, err := tx.ExecContext(ctx, copyQuery); err != nil {
		return fmt.Errorf("copy legacy users: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE users_legacy;`); err != nil {
		return fmt.Errorf("drop legacy users table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit users migration: %w", err)
	}

	return nil
}

func migrateMessagesTable(ctx context.Context, conn *sql.DB, schemaBytes []byte) error {
	needsMigration, err := messagesTableNeedsIDMigration(ctx, conn)
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

	columnsToCopy := []string{"conversation_id", "sender_id", "body", "created_at"}
	for _, column := range columnsToCopy {
		if _, ok := legacyColumns[column]; !ok {
			return fmt.Errorf("legacy messages table is missing required column %q", column)
		}
	}

	copyQuery := fmt.Sprintf(`
INSERT INTO messages (
  %s
)
SELECT
  %s
FROM messages_legacy
ORDER BY created_at ASC, rowid ASC;
`, strings.Join(columnsToCopy, ", "), strings.Join(columnsToCopy, ", "))

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

func messagesTableNeedsIDMigration(ctx context.Context, conn *sql.DB) (bool, error) {
	columnTypes, err := tableColumnTypes(ctx, conn, "messages")
	if err != nil {
		return false, err
	}

	idType, ok := columnTypes["id"]
	if !ok {
		return true, nil
	}

	return !strings.EqualFold(strings.TrimSpace(idType), "INTEGER"), nil
}

func usersTableNeedsIDMigration(ctx context.Context, conn *sql.DB) (bool, error) {
	columnTypes, err := tableColumnTypes(ctx, conn, "users")
	if err != nil {
		return false, err
	}

	if _, ok := columnTypes["id"]; !ok {
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
