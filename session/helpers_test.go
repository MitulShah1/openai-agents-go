package session

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
)

func verifyMessageContent(t *testing.T, msg openai.ChatCompletionMessageParamUnion, role, contentContains string) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	jsonStr := string(data)

	// Verify role
	if !strings.Contains(jsonStr, fmt.Sprintf(`"role":"%s"`, role)) {
		t.Errorf("Expected role '%s', got JSON: %s", role, jsonStr)
	}

	// Verify content
	if !strings.Contains(jsonStr, contentContains) {
		t.Errorf("Expected content to contain '%s', got JSON: %s", contentContains, jsonStr)
	}
}
