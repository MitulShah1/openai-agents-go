package session

import (
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestDefaultCompactor(t *testing.T) {
	// Create dummy messages
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("System"),
		openai.UserMessage("User 1"),
		openai.AssistantMessage("Assistant 1"),
		openai.UserMessage("User 2"),
	}

	// Test no compaction needed (len 4 > 2, but logic is naive)
	compacted := DefaultCompactor(msgs, 100)
	if len(compacted) != 4 {
		t.Errorf("expected 4 messages, got %d", len(compacted))
	}

	// Test logic if we implemented truncation.
	// Since DefaultCompactor is a placeholder that currently just returns messages unless < 2 (which is weird, logic is: if len <= 2 return msgs).
	// Wait, the logic I wrote was:
	// if len(messages) <= 2 { return messages }
	// return messages
	// So it does NOTHING.
	// This is consistent with "placeholder".
}
