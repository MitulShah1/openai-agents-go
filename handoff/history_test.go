package handoff

import (
	"context"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
)

// Test history summarization

func TestBuildConversationSummary(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello, I need help"),
		openai.AssistantMessage("Hi! How can I assist you today?"),
		openai.UserMessage("I have a billing question"),
		openai.AssistantMessage("I can help with that"),
	}

	summary := buildConversationSummary(messages)

	// Check it has a header
	if !strings.Contains(summary, "Previous conversation summary") {
		t.Errorf("Expected summary to contain header, got: %s", summary)
	}

	// Should contain all message roles
	expectedRoles := []string{"User:", "Assistant:"}
	for _, role := range expectedRoles {
		if !strings.Contains(summary, role) {
			t.Errorf("Expected summary to contain %q, got: %s", role, summary)
		}
	}
}

func TestBuildConversationSummary_LongHistory(t *testing.T) {
	// Create 20 messages
	var messages []openai.ChatCompletionMessageParamUnion
	for i := 0; i < 20; i++ {
		messages = append(messages, openai.UserMessage("Message "+string(rune('A'+i))))
	}

	summary := buildConversationSummary(messages)

	// Should only include last 10 messages (letters K-T)
	if !strings.Contains(summary, "K") {
		t.Error("Expected summary to include message K (11th message)")
	}

	// Should not include first message
	if strings.Contains(summary, "Message A") {
		t.Error("Should not include early messages when history is long")
	}
}

func TestSummarizeMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  openai.ChatCompletionMessageParamUnion
		contains []string
	}{
		{
			"user message",
			openai.UserMessage("Hello world"),
			[]string{"User:", "Hello world"},
		},
		{
			"assistant message",
			openai.AssistantMessage("How can I help?"),
			[]string{"Assistant:", "How can I help?"},
		},
		{
			"system message",
			openai.SystemMessage("You are a helpful assistant"),
			[]string{"System:", "helpful assistant"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := summarizeMessage(tt.message)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Expected summary to contain %q, got %q", substr, result)
				}
			}
		})
	}
}

func TestSummarizeMessage_LongContent(t *testing.T) {
	longMessage := openai.UserMessage(strings.Repeat("A", 200))

	result := summarizeMessage(longMessage)

	// Should be truncated
	if len(result) > 120 { // "User: " + 100 chars + "..."
		t.Errorf("Expected message to be truncated, got length %d", len(result))
	}

	// Should end with "..."
	if !strings.HasSuffix(result, "...") {
		t.Errorf("Expected truncated message to end with '...', got %q", result)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncated", "hello world", 8, "hello..."},
		{"empty string", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}

			if len(result) > tt.maxLen {
				t.Errorf("Result length %d exceeds max length %d", len(result), tt.maxLen)
			}
		})
	}
}

// Test empty/nil cases

func TestDefaultHistoryMapper_EmptyMessages(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{}

	result, err := DefaultHistoryMapper(context.Background(), messages)

	if err != nil {
		t.Errorf("Should not error on empty messages: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d messages", len(result))
	}
}

func TestDefaultHistoryMapper_SingleMessage(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}

	result, err := DefaultHistoryMapper(context.Background(), messages)

	if err != nil {
		t.Errorf("Should not error on single message: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 summary message, got %d", len(result))
	}
}

func TestDefaultHistoryMapper_MultipleMessages(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
		openai.AssistantMessage("Hi!"),
		openai.UserMessage("How are you?"),
		openai.AssistantMessage("I'm doing well, thanks!"),
	}

	result, err := DefaultHistoryMapper(context.Background(), messages)

	if err != nil {
		t.Errorf("Should not error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 summary message, got %d", len(result))
	}

	// The summary should be an assistant message
	// We can't easily inspect the content without marshaling,
	// but we can verify it exists and has length
}

func TestFlattenHistoryMapper_PreservesAllMessages(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Message 1"),
		openai.AssistantMessage("Message 2"),
		openai.UserMessage("Message 3"),
	}

	result, err := FlattenHistoryMapper(context.Background(), messages)

	if err != nil {
		t.Errorf("Should not error: %v", err)
	}

	if len(result) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(result))
	}

	// Messages should be the same (not modified)
	for i := range messages {
		if &messages[i] != &result[i] {
			t.Errorf("Message at index %d was modified", i)
		}
	}
}

// Test context handling

func TestHistoryMappers_WithContext(t *testing.T) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Test"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Both mappers should accept context even if they don't use it
	_, err := DefaultHistoryMapper(ctx, messages)
	if err != nil {
		t.Errorf("DefaultHistoryMapper should not error with context: %v", err)
	}

	_, err = FlattenHistoryMapper(ctx, messages)
	if err != nil {
		t.Errorf("FlattenHistoryMapper should not error with context: %v", err)
	}
}

// Test role capitalization

func TestRoleCapitalization(t *testing.T) {
	tests := []struct {
		role     string
		expected string
	}{
		{"user", "User:"},
		{"assistant", "Assistant:"},
		{"system", "System:"},
		{"tool", "Tool:"},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			var msg openai.ChatCompletionMessageParamUnion
			switch tt.role {
			case "user":
				msg = openai.UserMessage("content")
			case "assistant":
				msg = openai.AssistantMessage("content")
			case "system":
				msg = openai.SystemMessage("content")
			case "tool":
				msg = openai.ToolMessage("content", "tool-id")
			}

			result := summarizeMessage(msg)

			if !strings.HasPrefix(result, tt.expected) {
				t.Errorf("Expected message to start with %q, got %q", tt.expected, result)
			}
		})
	}
}
