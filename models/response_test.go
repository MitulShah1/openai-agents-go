package models

import "testing"

func TestModelUsage_Fields(t *testing.T) {
	usage := ModelUsage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	if usage.PromptTokens != 100 {
		t.Errorf("expected PromptTokens 100, got %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 50 {
		t.Errorf("expected CompletionTokens 50, got %d", usage.CompletionTokens)
	}
	if usage.TotalTokens != 150 {
		t.Errorf("expected TotalTokens 150, got %d", usage.TotalTokens)
	}
}

func TestModelResponse_Fields(t *testing.T) {
	resp := &ModelResponse{
		Usage: ModelUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		ResponseID: "resp_abc123",
	}

	if resp.ResponseID != "resp_abc123" {
		t.Errorf("expected ResponseID resp_abc123, got %s", resp.ResponseID)
	}
	if resp.Completion != nil {
		t.Error("expected nil Completion for test response")
	}
	if resp.Usage.TotalTokens != 30 {
		t.Errorf("expected TotalTokens 30, got %d", resp.Usage.TotalTokens)
	}
}
