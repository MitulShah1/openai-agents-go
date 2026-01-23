package stream

import (
	"context"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
)

func TestNewResult(t *testing.T) {
	agent := map[string]string{"name": "TestAgent"}
	input := []openai.ChatCompletionMessageParamUnion{}
	maxTurns := 10

	result := NewResult(input, agent, maxTurns)

	if result.CurrentAgent == nil {
		t.Error("CurrentAgent should not be nil")
	}
	if result.MaxTurns != maxTurns {
		t.Errorf("MaxTurns = %v, want %v", result.MaxTurns, maxTurns)
	}
	if result.CurrentTurn != 0 {
		t.Errorf("CurrentTurn = %v, want 0", result.CurrentTurn)
	}
	if result.IsComplete {
		t.Error("IsComplete should be false initially")
	}
}

func TestResult_EmitEvent(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)

	event := &RawResponseEvent{
		Type:           "test",
		SequenceNumber: 1,
	}

	result.EmitEvent(event)

	// Check event was added to channel
	select {
	case received := <-result.eventChan:
		if received != event {
			t.Errorf("Received event = %v, want %v", received, event)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event not received within timeout")
	}
}

func TestResult_StreamEvents(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)
	ctx := context.Background()

	// Emit some events
	go func() {
		result.EmitEvent(&RawResponseEvent{Type: "event1", SequenceNumber: 1})
		result.EmitEvent(&RawResponseEvent{Type: "event2", SequenceNumber: 2})
		result.EmitEvent(&RawResponseEvent{Type: "event3", SequenceNumber: 3})
		result.Complete()
	}()

	// Consume events
	count := 0
	for event, err := range result.StreamEvents(ctx) {
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if raw, ok := event.(*RawResponseEvent); ok {
			count++
			if raw.SequenceNumber != count {
				t.Errorf("Event %d has sequence number %d", count, raw.SequenceNumber)
			}
		}
	}

	if count != 3 {
		t.Errorf("Received %d events, want 3", count)
	}
}

func TestResult_StreamEvents_ContextCancellation(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Emit events continuously
	go func() {
		for i := 0; i < 100; i++ {
			result.EmitEvent(&RawResponseEvent{Type: "event", SequenceNumber: i})
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Cancel after receiving a few events
	count := 0
	for _, err := range result.StreamEvents(ctx) {
		if err != nil {
			// Context cancelled
			if err != context.Canceled {
				t.Errorf("Expected context.Canceled, got %v", err)
			}
			break
		}
		count++
		if count == 3 {
			cancel()
		}
	}

	if count < 3 {
		t.Errorf("Received %d events before cancellation, expected at least 3", count)
	}
}

func TestResult_Cancel_Immediate(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)

	// Add some events
	result.EmitEvent(&RawResponseEvent{Type: "event1"})
	result.EmitEvent(&RawResponseEvent{Type: "event2"})

	// Cancel immediately
	result.Cancel(CancelImmediate)

	if !result.IsComplete {
		t.Error("IsComplete should be true after immediate cancel")
	}

	// Event channel should be cleared
	if len(result.eventChan) != 0 {
		t.Errorf("Event channel has %d events, want 0", len(result.eventChan))
	}
}

func TestResult_Cancel_AfterTurn(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)

	result.Cancel(CancelAfterTurn)

	if result.GetCancelMode() != CancelAfterTurn {
		t.Errorf("CancelMode = %v, want %v", result.GetCancelMode(), CancelAfterTurn)
	}
}

func TestResult_AddItem(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)

	item := map[string]string{"type": "message", "content": "test"}
	result.AddItem(item)

	if len(result.NewItems) != 1 {
		t.Errorf("NewItems length = %d, want 1", len(result.NewItems))
	}
	if result.NewItems[0] == nil {
		t.Error("Added item should not be nil")
	}
}

func TestResult_AddResponse(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)

	response := ModelResponse{ResponseID: "test-123"}
	result.AddResponse(response)

	if len(result.RawResponses) != 1 {
		t.Errorf("RawResponses length = %d, want 1", len(result.RawResponses))
	}
	if result.RawResponses[0].ResponseID != "test-123" {
		t.Error("Added response not found in RawResponses")
	}
}

func TestResult_UpdateAgent(t *testing.T) {
	agent1 := map[string]string{"name": "Agent1"}
	agent2 := map[string]string{"name": "Agent2"}

	result := NewResult(nil, agent1, 10)

	result.UpdateAgent(agent2)

	if result.CurrentAgent == nil {
		t.Error("CurrentAgent should not be nil after update")
	}
}

func TestResult_IncrementTurn(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)

	result.IncrementTurn()
	if result.CurrentTurn != 1 {
		t.Errorf("CurrentTurn = %d, want 1", result.CurrentTurn)
	}

	result.IncrementTurn()
	if result.CurrentTurn != 2 {
		t.Errorf("CurrentTurn = %d, want 2", result.CurrentTurn)
	}
}

func TestResult_SetFinalOutput(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)

	output := "final output"
	result.SetFinalOutput(output)

	if result.FinalOutput != output {
		t.Errorf("FinalOutput = %v, want %v", result.FinalOutput, output)
	}
}

func TestResult_ThreadSafety(t *testing.T) {
	result := NewResult(nil, map[string]string{"name": "test"}, 10)

	// Concurrent operations
	done := make(chan bool)

	// Goroutine 1: Emit events
	go func() {
		for i := 0; i < 100; i++ {
			result.EmitEvent(&RawResponseEvent{SequenceNumber: i})
		}
		done <- true
	}()

	// Goroutine 2: Add items
	go func() {
		for i := 0; i < 100; i++ {
			result.AddItem(map[string]string{"type": "test"})
		}
		done <- true
	}()

	// Goroutine 3: Increment turns
	go func() {
		for i := 0; i < 100; i++ {
			result.IncrementTurn()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify state
	if result.CurrentTurn != 100 {
		t.Errorf("CurrentTurn = %d, want 100", result.CurrentTurn)
	}
	if len(result.NewItems) != 100 {
		t.Errorf("NewItems length = %d, want 100", len(result.NewItems))
	}
}
