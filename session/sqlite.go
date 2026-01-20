package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	_ "modernc.org/sqlite" // SQLite driver
)

// SQLiteSession implements Session using SQLite database.
type SQLiteSession struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite session backend.
// The path parameter specifies the database file location.
// Use ":memory:" for an in-memory database.
func NewSQLite(path string) (Session, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	session := &SQLiteSession{db: db}

	// Initialize schema
	if err := session.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return session, nil
}

// initSchema creates the sessions table if it doesn't exist.
func (s *SQLiteSession) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			messages BLOB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_updated_at ON sessions(updated_at);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Get retrieves all messages for a session ID.
func (s *SQLiteSession) Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	var messagesJSON []byte
	err := s.db.QueryRowContext(ctx, "SELECT messages FROM sessions WHERE id = ?", sessionID).Scan(&messagesJSON)

	if err == sql.ErrNoRows {
		return nil, &NotFoundError{SessionID: sessionID}
	}
	if err != nil {
		return nil, &StorageError{SessionID: sessionID, Operation: "get", Err: err}
	}

	var messages []openai.ChatCompletionMessageParamUnion
	if err := json.Unmarshal(messagesJSON, &messages); err != nil {
		return nil, &StorageError{SessionID: sessionID, Operation: "unmarshal", Err: err}
	}

	return messages, nil
}

// Append adds messages to a session.
func (s *SQLiteSession) Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	// Get existing messages
	existing, err := s.Get(ctx, sessionID)
	if err != nil {
		// If session doesn't exist, create it
		if _, ok := err.(*NotFoundError); !ok {
			return err
		}
		existing = []openai.ChatCompletionMessageParamUnion{}
	}

	// Append new messages
	existing = append(existing, messages...)

	// Serialize to JSON
	messagesJSON, err := json.Marshal(existing)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "marshal", Err: err}
	}

	// Upsert
	query := `
		INSERT INTO sessions (id, messages, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			messages = excluded.messages,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err = s.db.ExecContext(ctx, query, sessionID, messagesJSON)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "append", Err: err}
	}

	return nil
}

// Clear removes all messages from a session.
func (s *SQLiteSession) Clear(ctx context.Context, sessionID string) error {
	// Check if session exists
	_, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// Clear messages (set to empty array)
	emptyJSON, _ := json.Marshal([]openai.ChatCompletionMessageParamUnion{})
	query := "UPDATE sessions SET messages = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err = s.db.ExecContext(ctx, query, emptyJSON, sessionID)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "clear", Err: err}
	}

	return nil
}

// Delete removes a session completely.
func (s *SQLiteSession) Delete(ctx context.Context, sessionID string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "delete", Err: err}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "delete", Err: err}
	}

	if rowsAffected == 0 {
		return &NotFoundError{SessionID: sessionID}
	}

	return nil
}

// Close closes the database connection.
func (s *SQLiteSession) Close() error {
	return s.db.Close()
}

func init() {
	// Register SQLite backend in the registry
	Register("sqlite", func(config map[string]any) (Session, error) {
		path, ok := config["path"].(string)
		if !ok {
			return nil, fmt.Errorf("sqlite backend requires 'path' config")
		}
		return NewSQLite(path)
	})
}
