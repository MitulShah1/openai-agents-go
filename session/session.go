// Package session provides conversation history persistence for agents.
package session

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
)

// Session manages conversation history persistence across agent runs.
type Session interface {
	// Get retrieves all messages for a session ID
	Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error)

	// Append adds messages to a session
	Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error

	// Clear removes all messages from a session
	Clear(ctx context.Context, sessionID string) error

	// Delete removes a session completely
	Delete(ctx context.Context, sessionID string) error
}

// NotFoundError is returned when a session doesn't exist.
type NotFoundError struct {
	SessionID string
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("session '%s' not found", e.SessionID)
}

// StorageError wraps storage-related errors.
type StorageError struct {
	SessionID string
	Operation string
	Err       error
}

// Error implements the error interface.
func (e *StorageError) Error() string {
	return fmt.Sprintf("session '%s' %s failed: %v", e.SessionID, e.Operation, e.Err)
}

// Unwrap returns the underlying error.
func (e *StorageError) Unwrap() error {
	return e.Err
}
