package session

import (
	"fmt"
	"sync"

	"github.com/openai/openai-go/v3"
)

// BackendFactory is a function that creates a session backend from configuration.
type BackendFactory func(config map[string]any) (Session, error)

var (
	// Global registry of session backends
	registry     = make(map[string]BackendFactory)
	registryLock sync.RWMutex
)

// Register registers a session backend factory with the given name.
// This allows plugins to register themselves at init time.
func Register(name string, factory BackendFactory) {
	registryLock.Lock()
	defer registryLock.Unlock()
	registry[name] = factory
}

// Get retrieves a backend factory by name.
// Returns the factory and true if found, nil and false otherwise.
func Get(name string) (BackendFactory, bool) {
	registryLock.RLock()
	defer registryLock.RUnlock()
	factory, ok := registry[name]
	return factory, ok
}

// Create creates a session instance using the named backend.
// The config map is passed to the backend factory.
func Create(name string, config map[string]any) (Session, error) {
	factory, ok := Get(name)
	if !ok {
		return nil, fmt.Errorf("session backend %q not found", name)
	}
	return factory(config)
}

// ListBackends returns the names of all registered backends.
func ListBackends() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

func init() {
	// Register built-in backends
	Register("memory", func(config map[string]any) (Session, error) {
		return NewMemorySession(), nil
	})

	Register("file", func(config map[string]any) (Session, error) {
		dir, ok := config["directory"].(string)
		if !ok {
			return nil, fmt.Errorf("file backend requires 'directory' config")
		}
		return NewFileSession(dir)
	})

	Register("conversations", func(config map[string]any) (Session, error) {
		client, ok := config["client"].(*openai.Client)
		if !ok {
			return nil, fmt.Errorf("conversations backend requires 'client' (*openai.Client) in config")
		}
		return NewConversationsSession(client), nil
	})
}
