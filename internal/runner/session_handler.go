package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/session"
)

// SessionHandler manages session history loading and saving
type SessionHandler struct {
	session   session.Session
	sessionID string
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sess session.Session, sessionID string) *SessionHandler {
	if sess == nil || sessionID == "" {
		return nil
	}
	return &SessionHandler{
		session:   sess,
		sessionID: sessionID,
	}
}

// LoadHistory loads conversation history from the session
// Returns the combined history (session history + new messages)
// If session is not found, returns only the new messages without error
func (sh *SessionHandler) LoadHistory(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
) ([]openai.ChatCompletionMessageParamUnion, error) {
	if sh == nil {
		return messages, nil
	}

	sessionHistory, err := sh.session.Get(ctx, sh.sessionID)
	if err != nil {
		// Check if it's a NotFoundError - that's acceptable, we'll create the session on save
		var notFoundErr *session.NotFoundError
		if errors.As(err, &notFoundErr) {
			return messages, nil
		}
		// Other errors should be returned
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Prepend session history to new messages
	combined := make([]openai.ChatCompletionMessageParamUnion, 0, len(sessionHistory)+len(messages))
	combined = append(combined, sessionHistory...)
	combined = append(combined, messages...)

	return combined, nil
}

// SaveHistory saves the conversation history to the session
func (sh *SessionHandler) SaveHistory(
	ctx context.Context,
	history []openai.ChatCompletionMessageParamUnion,
) error {
	if sh == nil {
		return nil
	}

	if err := sh.session.Append(ctx, sh.sessionID, history); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}
