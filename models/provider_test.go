package models

import (
	"context"
	"fmt"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/ssestream"
)

// mockModel implements Model for testing.
type mockModel struct {
	name string
}

func (m *mockModel) GetResponse(_ context.Context, _ openai.ChatCompletionNewParams, _ ModelSettings) (*ModelResponse, error) {
	return &ModelResponse{ResponseID: m.name}, nil
}
func (m *mockModel) StreamResponse(_ context.Context, _ openai.ChatCompletionNewParams, _ ModelSettings) (*ssestream.Stream[openai.ChatCompletionChunk], error) {
	return nil, nil
}
func (m *mockModel) ModelName() string { return m.name }

// mockProvider implements ModelProvider for testing.
type mockProvider struct {
	prefix string
}

func (p *mockProvider) GetModel(name string) (Model, error) {
	return &mockModel{name: fmt.Sprintf("%s:%s", p.prefix, name)}, nil
}

func TestMultiProvider_DefaultProvider(t *testing.T) {
	defaultProv := &mockProvider{prefix: "default"}
	mp := NewMultiProvider(defaultProv)

	model, err := mp.GetModel("gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ModelName() != "default:gpt-4o" {
		t.Errorf("expected default:gpt-4o, got %s", model.ModelName())
	}
}

func TestMultiProvider_PrefixRouting(t *testing.T) {
	defaultProv := &mockProvider{prefix: "openai"}
	anthropicProv := &mockProvider{prefix: "anthropic"}
	googleProv := &mockProvider{prefix: "google"}

	mp := NewMultiProvider(defaultProv,
		WithProviderPrefix("anthropic", anthropicProv),
		WithProviderPrefix("google", googleProv),
	)

	tests := []struct {
		input    string
		expected string
	}{
		{"gpt-4o", "openai:gpt-4o"},
		{"anthropic/claude-3-5-sonnet", "anthropic:claude-3-5-sonnet"},
		{"google/gemini-2.0-flash", "google:gemini-2.0-flash"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			model, err := mp.GetModel(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if model.ModelName() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, model.ModelName())
			}
		})
	}
}

func TestMultiProvider_UnknownPrefix(t *testing.T) {
	defaultProv := &mockProvider{prefix: "default"}
	mp := NewMultiProvider(defaultProv)

	_, err := mp.GetModel("unknown/some-model")
	if err == nil {
		t.Fatal("expected error for unknown prefix")
	}
}

func TestOpenAIProvider_GetModel(t *testing.T) {
	// Create a minimal client (won't make real API calls in unit tests)
	client := openai.NewClient()
	provider := NewOpenAIProvider(&client)

	model, err := provider.GetModel("gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ModelName() != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %s", model.ModelName())
	}
}

func TestOpenAIProvider_EmptyModelName(t *testing.T) {
	client := openai.NewClient()
	provider := NewOpenAIProvider(&client)

	_, err := provider.GetModel("")
	if err == nil {
		t.Fatal("expected error for empty model name")
	}
}

func TestOpenAIProvider_Client(t *testing.T) {
	client := openai.NewClient()
	provider := NewOpenAIProvider(&client)

	if provider.Client() != &client {
		t.Error("Client() should return the underlying client")
	}
}
