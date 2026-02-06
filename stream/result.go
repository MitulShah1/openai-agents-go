package stream

import (
	"context"
	"iter"
	"sync"

	"github.com/openai/openai-go/v3"
)

// CancelMode defines how a streaming run should be cancelled
type CancelMode string

const (
	// CancelImmediate stops immediately, cancels all tasks, clears queues
	CancelImmediate CancelMode = "immediate"

	// CancelAfterTurn completes current turn gracefully before stopping
	// - Allows LLM response to finish
	// - Executes pending tool calls
	// - Saves session state properly
	// - Tracks usage accurately
	// - Stops before next turn begins
	CancelAfterTurn CancelMode = "after_turn"
)

// Result represents a streaming agent run.
// It provides methods to stream events as they are generated and to cancel the run.
type Result struct {
	// Input is the original input messages
	Input []openai.ChatCompletionMessageParamUnion

	// NewItems contains the items generated during the run (any to avoid import cycle)
	NewItems []any

	// RawResponses contains the raw model responses
	RawResponses []ModelResponse

	// FinalOutput is the final output of the agent (nil until complete)
	FinalOutput any

	// CurrentAgent is the agent currently running (any to avoid import cycle)
	CurrentAgent any

	// CurrentTurn is the current turn number
	CurrentTurn int

	// MaxTurns is the maximum number of turns allowed
	MaxTurns int

	// IsComplete indicates whether the run has finished
	IsComplete bool

	// Internal fields
	eventChan  chan Event
	errChan    chan error
	cancelFunc context.CancelFunc
	cancelMode CancelMode
	mu         sync.RWMutex
}

// ModelResponse represents a response from the model
type ModelResponse struct {
	// ResponseID is the unique identifier for this response
	ResponseID string

	// Output contains the output items from the model
	Output []any

	// Usage contains token usage information
	Usage *Usage
}

// Usage tracks token usage for a model response
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// NewResult creates a new streaming result
func NewResult(
	input []openai.ChatCompletionMessageParamUnion,
	agent any,
	maxTurns int,
) *Result {
	return &Result{
		Input:        input,
		NewItems:     make([]any, 0),
		RawResponses: make([]ModelResponse, 0),
		CurrentAgent: agent,
		CurrentTurn:  0,
		MaxTurns:     maxTurns,
		IsComplete:   false,
		eventChan:    make(chan Event, 100), // Buffered channel
		errChan:      make(chan error, 1),
		cancelMode:   CancelImmediate,
	}
}

// StreamEvents returns an iterator for consuming streaming events.
// This uses Go 1.23+ iterator pattern for idiomatic iteration.
//
// Example usage:
//
//	for event, err := range result.StreamEvents(ctx) {
//	    if err != nil {
//	        return err
//	    }
//	    // Handle event
//	}
func (r *Result) StreamEvents(ctx context.Context) iter.Seq2[Event, error] {
	return func(yield func(Event, error) bool) {
		for {
			select {
			case <-ctx.Done():
				// Context cancelled
				yield(nil, ctx.Err())
				return

			case err := <-r.errChan:
				// Error occurred
				yield(nil, err)
				return

			case event, ok := <-r.eventChan:
				if !ok {
					// Channel closed, streaming complete
					return
				}

				// Yield the event
				if !yield(event, nil) {
					// Consumer stopped iteration
					return
				}

			}
		}
	}
}

// Cancel stops the streaming run.
//
// Args:
//   - mode: Cancellation strategy (CancelImmediate or CancelAfterTurn)
//
// Example:
//
//	result.Cancel(stream.CancelAfterTurn)  // Graceful
//	result.Cancel(stream.CancelImmediate)  // Immediate (default)
func (r *Result) Cancel(mode CancelMode) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cancelMode = mode

	if mode == CancelImmediate {
		// Immediate cancellation
		if r.cancelFunc != nil {
			r.cancelFunc()
		}
		r.IsComplete = true

		// Clear event queue
		for len(r.eventChan) > 0 {
			<-r.eventChan
		}
	}
	// For CancelAfterTurn, just set the flag
	// The runner will check this and stop gracefully
}

// IsCancelled returns true if cancellation has been requested
func (r *Result) IsCancelled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cancelMode != CancelImmediate || r.IsComplete
}

// GetCancelMode returns the current cancellation mode
func (r *Result) GetCancelMode() CancelMode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cancelMode
}

// EmitEvent sends an event to the stream
// This is called internally by the runner
func (r *Result) EmitEvent(event Event) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.IsComplete {
		select {
		case r.eventChan <- event:
		default:
			// Channel full, skip event (shouldn't happen with buffered channel)
		}
	}
}

// EmitError sends an error to the stream
// This is called internally by the runner
func (r *Result) EmitError(err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.IsComplete {
		select {
		case r.errChan <- err:
		default:
			// Error channel full, skip
		}
	}
}

// Complete marks the streaming run as complete and closes the event channel
// This is called internally by the runner
func (r *Result) Complete() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.IsComplete {
		r.IsComplete = true
		close(r.eventChan)
	}
}

// SetCancelFunc sets the context cancel function
// This is called internally by the runner
func (r *Result) SetCancelFunc(cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancelFunc = cancel
}

// AddItem adds a new item to the result
// This is called internally by the runner
func (r *Result) AddItem(item any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.NewItems = append(r.NewItems, item)
}

// AddResponse adds a model response to the result
// This is called internally by the runner
func (r *Result) AddResponse(response ModelResponse) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RawResponses = append(r.RawResponses, response)
}

// UpdateAgent updates the current agent
// This is called internally by the runner during handoffs
func (r *Result) UpdateAgent(agent any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentAgent = agent
}

// IncrementTurn increments the turn counter
// This is called internally by the runner
func (r *Result) IncrementTurn() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CurrentTurn++
}

// SetFinalOutput sets the final output
// This is called internally by the runner
func (r *Result) SetFinalOutput(output any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.FinalOutput = output
}
