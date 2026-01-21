package handoff

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
)

// HistoryMapper transforms conversation history for handoffs.
// This function type allows customization of how conversation history
// is processed when transferring between agents.
type HistoryMapper func(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error)

// DefaultHistoryMapper creates a summary message from the conversation transcript.
// This is used when WithHistoryNesting(true) is enabled.
//
// The mapper:
//  1. Collects all messages into a transcript
//  2. Creates a single assistant message summarizing the conversation
//  3. Returns just the summary message
//
// This reduces token usage for the next agent while preserving context.
func DefaultHistoryMapper(_ context.Context, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	// Build a summary of the conversation
	summary := buildConversationSummary(messages)

	// Create a single assistant message with the summary
	summaryMessage := openai.AssistantMessage(summary)

	return []openai.ChatCompletionMessageParamUnion{summaryMessage}, nil
}

// FlattenHistoryMapper returns all messages unchanged.
// This is the default behavior when history nesting is disabled.
func FlattenHistoryMapper(_ context.Context, messages []openai.ChatCompletionMessageParamUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	return messages, nil
}

// buildConversationSummary creates a textual summary of messages.
func buildConversationSummary(messages []openai.ChatCompletionMessageParamUnion) string {
	var parts []string
	parts = append(parts, "Previous conversation summary:")
	parts = append(parts, "")

	for i, msg := range messages {
		// Limit to last 10 messages to avoid huge summaries
		if i < len(messages)-10 {
			continue
		}

		summary := summarizeMessage(msg)
		if summary != "" {
			parts = append(parts, summary)
		}
	}

	return strings.Join(parts, "\n")
}

// summarizeMessage creates a one-line summary of a message.
func summarizeMessage(msg openai.ChatCompletionMessageParamUnion) string {
	// Use JSON marshaling to inspect the message
	data, err := json.Marshal(msg)
	if err != nil {
		return ""
	}

	// Parse the message structure
	var partial struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &partial); err != nil {
		return ""
	}

	// Extract content string
	var contentStr string
	if err := json.Unmarshal(partial.Content, &contentStr); err != nil {
		// Content might be an array or complex type, just use the raw JSON
		contentStr = string(partial.Content)
	}

	if contentStr == "" {
		return ""
	}

	// Format based on role (capitalize first letter)
	role := partial.Role
	if len(role) > 0 {
		role = strings.ToUpper(role[:1]) + strings.ToLower(role[1:])
	}
	return fmt.Sprintf("%s: %s", role, truncate(contentStr, 100))
}

// truncate truncates a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
