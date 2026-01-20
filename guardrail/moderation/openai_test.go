package moderation

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

func TestNewOpenAI(t *testing.T) {
	// Mock OpenAI Moderation API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/moderations" {
			t.Errorf("Expected path /moderations, got %s", r.URL.Path)
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}

		// Decode request to allow dynamic responses based on input if we wanted
		var req struct {
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		input := req.Input[0]

		// Mock response structure
		// Note: The actual structure of openai-go response types might differ slightly in serialization
		// but matching the JSON wire format is what matters for the mock server.
		response := map[string]interface{}{
			"id":      "modr-123456",
			"model":   "text-moderation-007",
			"results": []map[string]interface{}{},
		}

		result := map[string]interface{}{
			"flagged": false,
			"categories": map[string]bool{
				"harassment":             false,
				"harassment/threatening": false,
				"hate":                   false,
				"violence":               false,
			}, // Simplified for test
			"category_scores": map[string]float64{
				"harassment":             0.0,
				"harassment/threatening": 0.0,
				"hate":                   0.0,
				"violence":               0.0,
				"self-harm":              0.0,
				"sexual":                 0.0,
			},
		}

		// Simulate detection logic based on input keywords
		switch input {
		case "kill yourself":
			result["flagged"] = true
			result["categories"].(map[string]bool)["violence"] = true
			result["category_scores"].(map[string]float64)["violence"] = 0.99
		case "I hate you":
			// Below default threshold (0.5), above custom low threshold (0.1)
			result["category_scores"].(map[string]float64)["harassment"] = 0.4
		}

		response["results"] = append(response["results"].([]map[string]interface{}), result)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create OpenAI client pointing to mock server
	client := openai.NewClient(
		option.WithBaseURL(server.URL),
		option.WithAPIKey("test-key"),
	)

	tests := []struct {
		name             string
		opts             []Option
		input            string
		expectedPass     bool
		expectedTripwire bool
		expectedMsgPart  string
	}{
		{
			name:         "clean_text",
			opts:         nil,
			input:        "Hello world",
			expectedPass: true,
		},
		{
			name:             "violations_detected",
			opts:             []Option{WithModerationTripwire(true)},
			input:            "kill yourself",
			expectedPass:     false,
			expectedTripwire: true,
			expectedMsgPart:  "violence",
		},
		{
			name:         "score_below_default_threshold",
			opts:         nil,
			input:        "I hate you", // mock returns 0.4 harassment
			expectedPass: true,
		},
		{
			name:             "score_above_custom_low_threshold",
			opts:             []Option{WithModerationThreshold(0.3)},
			input:            "I hate you", // mock returns 0.4 harassment, threshold 0.3
			expectedPass:     false,
			expectedTripwire: false,
			expectedMsgPart:  "harassment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fix: passed &client instead of client
			guard := NewOpenAI(&client, tt.opts...)

			if guard == nil {
				t.Fatal("NewOpenAI returned nil")
			}

			ctx := context.Background()
			result, err := guard.Func(ctx, tt.input)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Result is nil")
			}

			if result.Passed != tt.expectedPass {
				t.Errorf("Expected Passed=%v, got %v. Message: %s",
					tt.expectedPass, result.Passed, result.Message)
			}

			if result.TripwireTriggered != tt.expectedTripwire {
				t.Errorf("Expected TripwireTriggered=%v, got %v",
					tt.expectedTripwire, result.TripwireTriggered)
			}

			if tt.expectedMsgPart != "" && !strings.Contains(result.Message, tt.expectedMsgPart) {
				t.Errorf("Expected message to contain '%s', got '%s'", tt.expectedMsgPart, result.Message)
			}
		})
	}
}

func TestOpenAIOptions(t *testing.T) {
	t.Run("WithModerationTripwire", func(t *testing.T) {
		c := &Config{}
		WithModerationTripwire(true)(c)
		if !c.Tripwire {
			t.Error("Tripwire should be true")
		}
	})

	t.Run("WithModerationThreshold", func(t *testing.T) {
		c := &Config{}
		WithModerationThreshold(0.8)(c)
		if c.Threshold != 0.8 {
			t.Errorf("Expected threshold 0.8, got %f", c.Threshold)
		}
	})

	t.Run("WithModerationCategories", func(t *testing.T) {
		c := &Config{}
		WithModerationCategories(CategoryViolence)(c)
		if len(c.Categories) != 1 || !c.Categories[CategoryViolence] {
			t.Error("Expected only Violence category to be set")
		}
	})
}
