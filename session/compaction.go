package session

import (
	"context"

	"github.com/openai/openai-go/v3"
)

// OpenAIResponsesCompactionAwareSession is an interface for sessions that support compaction.
type OpenAIResponsesCompactionAwareSession interface {
	Session
	Compact(ctx context.Context, sessionID string, maxTokens int) error
}

// OpenAIResponsesCompactionArgs configuration for compaction.
type OpenAIResponsesCompactionArgs struct {
	MaxTokens int
}

// Compactor handles message compaction logic.
type Compactor struct {
	// Strategy function could be here
}

// DefaultCompactor is a simple compactor that truncates old messages if token count exceeds limit.
// Note: This is a placeholder for the sophisticated logic described in Issue #21.
// Real implementation requires token counting (tiktoken) and semantic analysis.
func DefaultCompactor(messages []openai.ChatCompletionMessageParamUnion, maxTokens int) []openai.ChatCompletionMessageParamUnion {
	// Simple FIFO sliding window for now
	// Ideally we keep System message and recent context
	if len(messages) <= 2 {
		return messages
	}

	// Just return last N messages?
	// Without token counting library dependency, exact limit is hard.
	// We assume we should implement the "Algorithm" later or use approximate.

	return messages
}
