package tracing

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	signalHandlerOnce     sync.Once
	shutdownHandlers      []func()
	shutdownHandlersMutex sync.Mutex
)

// RegisterShutdownHandler registers a function to be called on graceful shutdown.
// This is useful for custom cleanup logic that should run when the process receives
// SIGTERM or SIGINT.
//
// Example:
//
//	tracing.RegisterShutdownHandler(func() {
//	    log.Println("Flushing custom metrics...")
//	    metrics.Flush()
//	})
func RegisterShutdownHandler(handler func()) {
	shutdownHandlersMutex.Lock()
	defer shutdownHandlersMutex.Unlock()
	shutdownHandlers = append(shutdownHandlers, handler)
}

// initSignalHandler sets up signal handling for graceful shutdown.
// It listens for SIGTERM and SIGINT and flushes all traces before exiting.
func initSignalHandler() {
	// Check if signal handling is disabled
	if os.Getenv("OPENAI_AGENTS_DISABLE_SIGNAL_HANDLER") != "" {
		return
	}

	signalHandlerOnce.Do(func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

		go func() {
			sig := <-sigChan

			// Create a context with timeout for shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Shutdown the trace provider
			if err := GetProvider().Shutdown(ctx); err != nil {
				// Log error but don't fail - we're shutting down anyway
				_ = err
			}

			// Call custom shutdown handlers
			shutdownHandlersMutex.Lock()
			handlers := make([]func(), len(shutdownHandlers))
			copy(handlers, shutdownHandlers)
			shutdownHandlersMutex.Unlock()

			for _, handler := range handlers {
				handler()
			}

			// Exit with appropriate code
			if sig == syscall.SIGTERM {
				os.Exit(143) // 128 + 15 (SIGTERM)
			}
			os.Exit(130) // 128 + 2 (SIGINT)
		}()
	})
}
