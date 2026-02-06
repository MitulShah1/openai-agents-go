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
// It uses a normalized schema compatible with the OpenAI Agents Python SDK.
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

// initSchema creates the sessions and messages tables if they don't exist.
// This uses a normalized schema:
// - agent_sessions: Stores session metadata
// - agent_messages: Stores individual messages with structure
func (s *SQLiteSession) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS agent_sessions (
			id TEXT PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			metadata JSON
		);
		
		CREATE TABLE IF NOT EXISTS agent_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT,
			tool_calls JSON,
			function_call JSON,
			tool_call_id TEXT,
			name TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(session_id) REFERENCES agent_sessions(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_messages_session_id ON agent_messages(session_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_updated_at ON agent_sessions(updated_at);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Get retrieves all messages for a session ID.
func (s *SQLiteSession) Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	// First check if session exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM agent_sessions WHERE id = ?)", sessionID).Scan(&exists)
	if err != nil {
		return nil, &StorageError{SessionID: sessionID, Operation: "check_exists", Err: err}
	}
	if !exists {
		return nil, &NotFoundError{SessionID: sessionID}
	}

	// Query messages
	query := `
		SELECT role, content, tool_calls, function_call, tool_call_id, name
		FROM agent_messages
		WHERE session_id = ?
		ORDER BY id ASC
	`
	rows, err := s.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, &StorageError{SessionID: sessionID, Operation: "get_messages", Err: err}
	}
	defer func() { _ = rows.Close() }()

	var messages []openai.ChatCompletionMessageParamUnion
	for rows.Next() {
		var (
			role       string
			content    sql.NullString
			toolCalls  sql.NullString
			funcCall   sql.NullString
			toolCallID sql.NullString
			name       sql.NullString
		)

		if err := rows.Scan(&role, &content, &toolCalls, &funcCall, &toolCallID, &name); err != nil {
			return nil, &StorageError{SessionID: sessionID, Operation: "scan_message", Err: err}
		}

		msg, err := s.rowToMessage(role, content, toolCalls, funcCall, toolCallID, name)
		if err != nil {
			return nil, &StorageError{SessionID: sessionID, Operation: "parse_message", Err: err}
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// Append adds messages to a session.
func (s *SQLiteSession) Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Upsert Session
	_, err = tx.ExecContext(ctx, `
		INSERT INTO agent_sessions (id, created_at, updated_at)
		VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
	`, sessionID)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "upsert_session", Err: err}
	}

	// 2. Insert Messages
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO agent_messages (session_id, role, content, tool_calls, function_call, tool_call_id, name)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "prepare_insert", Err: err}
	}
	defer func() { _ = stmt.Close() }()

	for _, msg := range messages {
		role, content, toolCalls, funcCall, toolCallID, name, err := s.messageToRow(msg)
		if err != nil {
			return &StorageError{SessionID: sessionID, Operation: "parse_input_message", Err: err}
		}

		_, err = stmt.ExecContext(ctx, sessionID, role, content, toolCalls, funcCall, toolCallID, name)
		if err != nil {
			return &StorageError{SessionID: sessionID, Operation: "insert_message", Err: err}
		}
	}

	if err := tx.Commit(); err != nil {
		return &StorageError{SessionID: sessionID, Operation: "commit", Err: err}
	}

	return nil
}

// Clear removes all messages from a session but keeps the session entry.
func (s *SQLiteSession) Clear(ctx context.Context, sessionID string) error {
	// Check/Touch session
	res, err := s.db.ExecContext(ctx, "UPDATE agent_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", sessionID)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "touch_session", Err: err}
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return &NotFoundError{SessionID: sessionID}
	}

	// Delete messages
	_, err = s.db.ExecContext(ctx, "DELETE FROM agent_messages WHERE session_id = ?", sessionID)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "clear_messages", Err: err}
	}

	return nil
}

// Delete removes a session completely.
func (s *SQLiteSession) Delete(ctx context.Context, sessionID string) error {
	// Enable foreign keys for cascade delete (SQLite specific pragma usually needed per conn,
	// but we can just delete from sessions and rely on schema or delete manually)
	// For safety/portability in tight loop, let's just delete both if cascade isn't on.
	// But let's try standard delete from parent.

	// Actually, standard SQLite driver might not enforce FKs unless PRAGMA foreign_keys = ON;
	// Let's do explicit delete to be safe.

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx, "DELETE FROM agent_sessions WHERE id = ?", sessionID)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "delete_session", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "rows_affected", Err: err}
	}

	if rows == 0 {
		return &NotFoundError{SessionID: sessionID}
	}

	// Delete messages (orphans if FK not enforced)
	_, err = tx.ExecContext(ctx, "DELETE FROM agent_messages WHERE session_id = ?", sessionID)
	if err != nil {
		return &StorageError{SessionID: sessionID, Operation: "delete_messages", Err: err}
	}

	return tx.Commit()
}

// Close closes the database connection.
func (s *SQLiteSession) Close() error {
	return s.db.Close()
}

// --- Helpers for mapping between OpenAI types and SQL rows ---

func (s *SQLiteSession) rowToMessage(role string, content, toolCalls, _, toolCallID, _ sql.NullString) (openai.ChatCompletionMessageParamUnion, error) {
	switch role {
	case "user":
		return openai.UserMessage(content.String), nil
	case "system":
		return openai.SystemMessage(content.String), nil
	case "assistant":
		// Reconstruct Assistant message
		// Note: We need to handle tool calls and content
		var tCalls []sqlToolCall
		if toolCalls.Valid && toolCalls.String != "" && toolCalls.String != "null" {
			if err := json.Unmarshal([]byte(toolCalls.String), &tCalls); err != nil {

				return openai.ChatCompletionMessageParamUnion{}, err
			}
		}

		// Create base assistant message
		msg := openai.AssistantMessage(content.String)
		if len(tCalls) > 0 {
			// Add tool calls
			modelMsg := msg.OfAssistant
			// We need to convert ToolCallParam back to UnionParam for the slice
			var unionCalls []openai.ChatCompletionMessageToolCallUnionParam
			for _, tc := range tCalls {
				unionCalls = append(unionCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
						Type: openai.ToolType("function"),
					},
				})
			}
			modelMsg.ToolCalls = unionCalls
		}
		return msg, nil

	case "tool":
		return openai.ToolMessage(toolCallID.String, content.String), nil
	default:
		// Fallback for unknown roles (e.g. function)
		// For now treating as minimal assistant-like or skipping?
		// Let's return simple user message structure as fallback or error.
		return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("unsupported role: %s", role)
	}
}

func (s *SQLiteSession) messageToRow(msg openai.ChatCompletionMessageParamUnion) (role string, content, toolCalls, funcCall, toolCallID, name *string, err error) {
	// Helper to ptr
	strPtr := func(s string) *string { return &s }
	jsonPtr := func(v any) (*string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		s := string(b)
		return &s, nil
	}

	if p := msg.OfSystem; p != nil {
		return "system", strPtr(extractTextFromUnion(p.Content)), nil, nil, nil, nil, nil
	}
	if p := msg.OfUser; p != nil {
		return "user", strPtr(extractTextFromUnion(p.Content)), nil, nil, nil, nil, nil
	}
	if p := msg.OfAssistant; p != nil {
		role = "assistant"
		content = strPtr(extractTextFromUnion(p.Content))

		if len(p.ToolCalls) > 0 {
			// Extract tool calls
			var calls []sqlToolCall
			for _, tc := range p.ToolCalls {
				if f := tc.OfFunction; f != nil {
					calls = append(calls, sqlToolCall{
						ID:   f.ID,
						Type: string(f.Type),
						Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{
							Name:      f.Function.Name,
							Arguments: f.Function.Arguments,
						},
					})
				}
			}

			toolCalls, err = jsonPtr(calls)
			if err != nil {
				return
			}
		}
		return
	}
	if p := msg.OfTool; p != nil {
		return "tool", strPtr(extractTextFromUnion(p.Content)), nil, nil, strPtr(p.ToolCallID), nil, nil
	}

	// TODO: Handle other types
	return "unknown", nil, nil, nil, nil, nil, fmt.Errorf("unsupported message type")
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

// sqlToolCall matches the JSON structure for tool calls in the database
type sqlToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}
