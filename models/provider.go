package models

import "fmt"

// ModelProvider is a factory interface that creates Model instances by name.
//
// The runner resolves which provider to use in this order:
//  1. Agent.ModelProvider (if set)
//  2. Runner-level default provider
//  3. Built-in OpenAI provider (backward compatibility)
//
// Built-in implementations:
//   - [OpenAIProvider]: Creates models using OpenAI's Chat Completions API
//   - [MultiProvider]: Routes model names by prefix to different providers
type ModelProvider interface {
	// GetModel returns a Model for the given model name.
	// The name follows the format used by the provider (e.g., "gpt-4o" for OpenAI).
	// Returns an error if the model name is not supported.
	GetModel(modelName string) (Model, error)
}

// MultiProvider routes model names to different providers based on a prefix.
//
// Model names are expected in the format "prefix/model-name". If no prefix
// is present, the default provider is used.
//
// Example:
//
//	mp := models.NewMultiProvider(
//	    openaiProvider,
//	    models.WithProviderPrefix("anthropic", anthropicProvider),
//	)
//	model, _ := mp.GetModel("gpt-4o")                        // → openaiProvider
//	model, _ := mp.GetModel("anthropic/claude-3-5-sonnet")    // → anthropicProvider
type MultiProvider struct {
	defaultProvider ModelProvider
	providers       map[string]ModelProvider
}

// MultiProviderOption configures a MultiProvider.
type MultiProviderOption func(*MultiProvider)

// WithProviderPrefix registers a provider for a given prefix.
// Panics if provider is nil.
func WithProviderPrefix(prefix string, provider ModelProvider) MultiProviderOption {
	if provider == nil {
		panic("models: WithProviderPrefix called with nil provider")
	}
	return func(mp *MultiProvider) {
		mp.providers[prefix] = provider
	}
}

// NewMultiProvider creates a MultiProvider with a default provider and optional
// prefix-based routing. Panics if defaultProvider is nil.
func NewMultiProvider(defaultProvider ModelProvider, opts ...MultiProviderOption) *MultiProvider {
	if defaultProvider == nil {
		panic("models: NewMultiProvider called with nil defaultProvider")
	}
	mp := &MultiProvider{
		defaultProvider: defaultProvider,
		providers:       make(map[string]ModelProvider),
	}
	for _, opt := range opts {
		opt(mp)
	}
	return mp
}

// GetModel routes the model name to the appropriate provider.
// Names with a "prefix/" are routed to the registered provider for that prefix.
// Names without a prefix use the default provider.
func (mp *MultiProvider) GetModel(modelName string) (Model, error) {
	// Look for prefix separator
	for i, ch := range modelName {
		if ch == '/' {
			prefix := modelName[:i]
			name := modelName[i+1:]
			if provider, ok := mp.providers[prefix]; ok {
				return provider.GetModel(name)
			}
			return nil, fmt.Errorf("models: unknown provider prefix %q in model name %q", prefix, modelName)
		}
	}

	// No prefix — use default provider
	return mp.defaultProvider.GetModel(modelName)
}
