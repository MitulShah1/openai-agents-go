package session

import (
	"context"
	"sync"

	"github.com/openai/openai-go"
)

// MemorySession stores conversations in memory (non-persistent).
// Ideal for development, testing, or temporary conversations.
type MemorySession struct {
	mu       sync.RWMutex
	sessions map[string][]openai.ChatCompletionMessageParamUnion
}

// NewMemorySession creates a new in-memory session store.
func NewMemorySession() *MemorySession {
	return &MemorySession{
		sessions: make(map[string][]openai.ChatCompletionMessageParamUnion),
	}
}

// Get retrieves messages for a session.
func (m *MemorySession) Get(_ context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages, exists := m.sessions[sessionID]
	if !exists {
		return nil, &NotFoundError{SessionID: sessionID}
	}

	// Return a copy to prevent external mutation
	result := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	copy(result, messages)
	return result, nil
}

// Append adds messages to a session.
func (m *MemorySession) Append(_ context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[sessionID]; !exists {
		m.sessions[sessionID] = make([]openai.ChatCompletionMessageParamUnion, 0)
	}

	m.sessions[sessionID] = append(m.sessions[sessionID], messages...)
	return nil
}

// Clear removes all messages from a session.
func (m *MemorySession) Clear(_ context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[sessionID]; !exists {
		return &NotFoundError{SessionID: sessionID}
	}

	m.sessions[sessionID] = make([]openai.ChatCompletionMessageParamUnion, 0)
	return nil
}

// Delete removes a session completely.
func (m *MemorySession) Delete(_ context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[sessionID]; !exists {
		return &NotFoundError{SessionID: sessionID}
	}

	delete(m.sessions, sessionID)
	return nil
}
