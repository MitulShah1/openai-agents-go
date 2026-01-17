package session

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// MockConversationsServer handles simulated OpenAI Conversations API requests
type MockConversationsServer struct {
	conversations map[string][]map[string]any
}

func NewMockConversationsServer() *MockConversationsServer {
	return &MockConversationsServer{
		conversations: make(map[string][]map[string]any),
	}
}

func (s *MockConversationsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Verify auth
	if r.Header.Get("Authorization") != "Bearer test-key" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Set JSON content type for all successful responses
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	switch {
	// Create Conversation
	case path == "/v1/conversations" && r.Method == http.MethodPost:
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Extract items if any
		convID := "conv_123"
		s.conversations[convID] = []map[string]any{}

		// Add initial items from request
		if items, ok := req["items"].([]any); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]any); ok {
					s.conversations[convID] = append(s.conversations[convID], itemMap)
				}
			}
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         convID,
			"object":     "conversation",
			"created_at": 1234567890,
		})

	// Add Items to Conversation
	case strings.Contains(path, "/items") && r.Method == http.MethodPost:
		// Extract convID from path /v1/conversations/{id}/items
		parts := strings.Split(path, "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		convID := parts[3]

		if _, exists := s.conversations[convID]; !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if items, ok := req["items"].([]any); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]any); ok {
					s.conversations[convID] = append(s.conversations[convID], itemMap)
				}
			}
		}

		// Returns ConversationItemList (simplified)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data":   []any{}, // We don't verify return value of Append yet
		})

	// List Items
	case strings.Contains(path, "/items") && r.Method == http.MethodGet:
		parts := strings.Split(path, "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		convID := parts[3]

		items, exists := s.conversations[convID]
		if !exists {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"message": "Conversation not found",
					"type":    "invalid_request_error",
					"code":    "resource_not_found",
				},
			})
			return
		}

		// Map stored input items to output items structure
		var outputItems []map[string]any
		for _, item := range items {
			outItem := make(map[string]any)
			outItem["id"] = "item_id"
			outItem["type"] = "message" // Default

			if role, ok := item["role"].(string); ok {
				outItem["role"] = role
			}

			// Handle content
			if content, ok := item["content"].(map[string]any); ok {
				_ = content // Mute unused variable error if we don't use it yet

				var text string
				// In JSON unmarshal, it might be string or map
				if s, ok := item["content"].(string); ok {
					text = s
				} else if _, ok := item["content"].(map[string]any); ok {
					// Complex content
					text = "complex_content" // Simplified
				}

				outItem["content"] = []map[string]any{
					{
						"type": "text",
						"text": text,
					},
				}
			}

			// Handle function call inputs
			if _, ok := item["call_id"]; ok {
				// It's a tool call (function_call) or output
				outItem["call_id"] = item["call_id"]
				outItem["type"] = "function_call"
				if args, ok := item["arguments"].(string); ok {
					outItem["arguments"] = args
				}
				// Note: output doesn't have name in item...
			}

			outputItems = append(outputItems, outItem)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":   "list",
			"data":     outputItems,
			"has_more": false,
		})

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// setupMockServer creates a test server and a client configured to use it.
func setupMockServer(_ *testing.T, handler http.Handler) (*openai.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := openai.NewClient(
		option.WithBaseURL(server.URL+"/v1/"),
		option.WithAPIKey("test-key"),
	)
	return &client, server
}

func TestConversationsSession_AppendAndGet(t *testing.T) {
	handler := NewMockConversationsServer()
	client, server := setupMockServer(t, handler)
	defer server.Close()

	session := NewConversationsSession(client)
	ctx := context.Background()
	sessionID := "test-session"

	// 1. Append (Create)
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("Hello"),
	}
	err := session.Append(ctx, sessionID, msgs) // creates conversation
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// 2. Get
	got, err := session.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(got))
	}
	if userMsg := got[0].OfUser; userMsg != nil {
		_ = userMsg
	} else {
		t.Errorf("Expected UserMessage, got %v", got[0])
	}

	// 3. Append (Add Items)
	msgs2 := []openai.ChatCompletionMessageParamUnion{
		openai.AssistantMessage("Hi"),
	}
	err = session.Append(ctx, sessionID, msgs2)
	if err != nil {
		t.Fatalf("Append 2 failed: %v", err)
	}

	// 4. Get again
	got, err = session.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("Get 2 failed: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(got))
	}
}

func TestConversationsSession_Delete(t *testing.T) {
	deleted := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set JSON content type
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/conversations/") {
			deleted = true
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":      "conv_123",
				"deleted": true,
			})
			return
		}
		// Allow creation for setup
		if r.Method == "POST" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "conv_123",
			})
			return
		}
	})

	client, server := setupMockServer(t, handler)
	defer server.Close()

	session := NewConversationsSession(client)
	ctx := context.Background()

	// Seed session
	_ = session.Append(ctx, "sess1", []openai.ChatCompletionMessageParamUnion{openai.UserMessage("hi")})

	// Delete
	err := session.Delete(ctx, "sess1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if !deleted {
		t.Error("Expected DELETE request to API")
	}

	// Verify mapping removed
	// Verify mapping removed
	if _, err := session.Get(ctx, "sess1"); err != nil {
		// Get should return nil, nil if known to be missing locally
		// But implementation calls List if missing? No, Get returns nil, nil if map missing.
		// This block is reachable if errors wrap properly.
		// But Get() returns (nil, nil) on not found.
		// So err should be nil if successfully "not found".
		// If err != nil, it's a real error.
		t.Errorf("Unexpected error: %v", err)
	}
}
