package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/session"
)

// mockSession is a mock implementation of session.Session for testing
type mockSession struct {
	data      map[string][]openai.ChatCompletionMessageParamUnion
	getError  error
	saveError error
}

func newMockSession() *mockSession {
	return &mockSession{
		data: make(map[string][]openai.ChatCompletionMessageParamUnion),
	}
}

func (m *mockSession) Get(_ context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	history, exists := m.data[sessionID]
	if !exists {
		return nil, &session.NotFoundError{SessionID: sessionID}
	}
	return history, nil
}

func (m *mockSession) Append(_ context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.data[sessionID] = append(m.data[sessionID], messages...)
	return nil
}

func (m *mockSession) Clear(_ context.Context, sessionID string) error {
	delete(m.data, sessionID)
	return nil
}

func (m *mockSession) Delete(_ context.Context, sessionID string) error {
	delete(m.data, sessionID)
	return nil
}

func TestNewSessionHandler(t *testing.T) {
	tests := []struct {
		name      string
		sess      session.Session
		sessionID string
		wantNil   bool
	}{
		{
			name:      "valid session and ID",
			sess:      newMockSession(),
			sessionID: "test-session",
			wantNil:   false,
		},
		{
			name:      "nil session",
			sess:      nil,
			sessionID: "test-session",
			wantNil:   true,
		},
		{
			name:      "empty session ID",
			sess:      newMockSession(),
			sessionID: "",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSessionHandler(tt.sess, tt.sessionID)
			if (handler == nil) != tt.wantNil {
				t.Errorf("NewSessionHandler() = %v, wantNil %v", handler, tt.wantNil)
			}
		})
	}
}

func TestSessionHandler_LoadHistory(t *testing.T) {
	ctx := context.Background()
	newMessages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("new message"),
	}

	tests := []struct {
		name         string
		setupSession func() *mockSession
		sessionID    string
		messages     []openai.ChatCompletionMessageParamUnion
		wantErr      bool
		wantMsgCount int
	}{
		{
			name: "load existing session",
			setupSession: func() *mockSession {
				mock := newMockSession()
				mock.data["test-session"] = []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage("old message"),
				}
				return mock
			},
			sessionID:    "test-session",
			messages:     newMessages,
			wantErr:      false,
			wantMsgCount: 2, // 1 from session + 1 new
		},
		{
			name:         "new session - not found",
			setupSession: newMockSession,
			sessionID:    "new-session",
			messages:     newMessages,
			wantErr:      false,
			wantMsgCount: 1, // Only new message
		},
		{
			name: "session error - not NotFoundError",
			setupSession: func() *mockSession {
				mock := newMockSession()
				mock.getError = errors.New("connection error")
				return mock
			},
			sessionID:    "test-session",
			messages:     newMessages,
			wantErr:      true,
			wantMsgCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setupSession()
			handler := NewSessionHandler(mock, tt.sessionID)

			messages, err := handler.LoadHistory(ctx, tt.messages)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(messages) != tt.wantMsgCount {
				t.Errorf("LoadHistory() message count = %d, want %d", len(messages), tt.wantMsgCount)
			}
		})
	}
}

func TestSessionHandler_LoadHistory_NilHandler(t *testing.T) {
	ctx := context.Background()
	var handler *SessionHandler // nil handler
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("test"),
	}

	result, err := handler.LoadHistory(ctx, messages)

	if err != nil {
		t.Errorf("LoadHistory() with nil handler should not error, got: %v", err)
	}

	if len(result) != len(messages) {
		t.Errorf("LoadHistory() with nil handler should return original messages")
	}
}

func TestSessionHandler_SaveHistory(t *testing.T) {
	ctx := context.Background()
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("test message"),
	}

	tests := []struct {
		name    string
		setup   func() *mockSession
		wantErr bool
	}{
		{
			name:    "successful save",
			setup:   newMockSession,
			wantErr: false,
		},
		{
			name: "save error",
			setup: func() *mockSession {
				mock := newMockSession()
				mock.saveError = errors.New("storage error")
				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.setup()
			handler := NewSessionHandler(mock, "test-session")

			err := handler.SaveHistory(ctx, messages)

			if (err != nil) != tt.wantErr {
				t.Errorf("SaveHistory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSessionHandler_SaveHistory_NilHandler(t *testing.T) {
	ctx := context.Background()
	var handler *SessionHandler // nil handler
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("test"),
	}

	err := handler.SaveHistory(ctx, messages)

	if err != nil {
		t.Errorf("SaveHistory() with nil handler should not error, got: %v", err)
	}
}
