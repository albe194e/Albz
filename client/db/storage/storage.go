package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	db "github.com/albe194e/albz/client/db/sqlc/generated"

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

	return &Store{
		DB: conn,
		Q:  db.New(conn),
	}, nil
}

func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}

	return s.DB.Close()
}
