package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/openai/openai-go"
)

// FileSession stores conversations as JSON files.
// Provides persistent storage without external dependencies.
type FileSession struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileSession creates a new file-based session store.
// basePath is the directory where session files will be stored.
func NewFileSession(basePath string) (*FileSession, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	return &FileSession{
		basePath: basePath,
	}, nil
}

// sessionPath returns the file path for a session.
func (f *FileSession) sessionPath(sessionID string) string {
	return filepath.Join(f.basePath, sessionID+".json")
}

// Get retrieves messages for a session.
func (f *FileSession) Get(_ context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	path := f.sessionPath(sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &NotFoundError{SessionID: sessionID}
		}
		return nil, &StorageError{
			SessionID: sessionID,
			Operation: "read",
			Err:       err,
		}
	}

	var messages []openai.ChatCompletionMessageParamUnion
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, &StorageError{
			SessionID: sessionID,
			Operation: "unmarshal",
			Err:       err,
		}
	}

	return messages, nil
}

// Append adds messages to a session.
func (f *FileSession) Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Load existing messages
	var existing []openai.ChatCompletionMessageParamUnion
	path := f.sessionPath(sessionID)

	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return &StorageError{
				SessionID: sessionID,
				Operation: "unmarshal",
				Err:       err,
			}
		}
	} else if !os.IsNotExist(err) {
		return &StorageError{
			SessionID: sessionID,
			Operation: "read",
			Err:       err,
		}
	}

	// Append new messages
	existing = append(existing, messages...)

	// Write atomically using temp file + rename
	return f.writeAtomic(sessionID, existing)
}

// Clear removes all messages from a session.
func (f *FileSession) Clear(_ context.Context, sessionID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	path := f.sessionPath(sessionID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &NotFoundError{SessionID: sessionID}
	}

	return f.writeAtomic(sessionID, []openai.ChatCompletionMessageParamUnion{})
}

// Delete removes a session completely.
func (f *FileSession) Delete(_ context.Context, sessionID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	path := f.sessionPath(sessionID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &NotFoundError{SessionID: sessionID}
	}

	if err := os.Remove(path); err != nil {
		return &StorageError{
			SessionID: sessionID,
			Operation: "delete",
			Err:       err,
		}
	}

	return nil
}

// writeAtomic writes data atomically using temp file + rename.
func (f *FileSession) writeAtomic(sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return &StorageError{
			SessionID: sessionID,
			Operation: "marshal",
			Err:       err,
		}
	}

	// Write to temp file
	tempPath := f.sessionPath(sessionID) + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return &StorageError{
			SessionID: sessionID,
			Operation: "write",
			Err:       err,
		}
	}

	// Atomic rename
	finalPath := f.sessionPath(sessionID)
	if err := os.Rename(tempPath, finalPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return &StorageError{
			SessionID: sessionID,
			Operation: "rename",
			Err:       err,
		}
	}

	return nil
}
