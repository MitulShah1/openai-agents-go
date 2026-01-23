package stream

import (
	"testing"
)

func TestRawResponseEvent_EventType(t *testing.T) {
	event := &RawResponseEvent{
		Type:           "response.created",
		Data:           map[string]any{"test": "data"},
		SequenceNumber: 1,
	}

	if got := event.EventType(); got != "raw_response_event" {
		t.Errorf("EventType() = %v, want %v", got, "raw_response_event")
	}
}

func TestRunItemEvent_EventType(t *testing.T) {
	event := &RunItemEvent{
		Name:           string(MessageOutputCreated),
		Item:           map[string]string{"type": "message"},
		SequenceNumber: 2,
	}

	if got := event.EventType(); got != "run_item_stream_event" {
		t.Errorf("EventType() = %v, want %v", got, "run_item_stream_event")
	}
}

func TestAgentUpdatedEvent_EventType(t *testing.T) {
	event := &AgentUpdatedEvent{
		NewAgent:       map[string]string{"name": "TestAgent"},
		SequenceNumber: 3,
	}

	if got := event.EventType(); got != "agent_updated_stream_event" {
		t.Errorf("EventType() = %v, want %v", got, "agent_updated_stream_event")
	}
}

func TestRunItemEventNames(t *testing.T) {
	tests := []struct {
		name      string
		eventName RunItemEventName
		want      string
	}{
		{"MessageOutputCreated", MessageOutputCreated, "message_output_created"},
		{"HandoffRequested", HandoffRequested, "handoff_requested"},
		{"HandoffOccurred", HandoffOccurred, "handoff_occured"}, //nolint:misspell // intentionally misspelled
		{"ToolCalled", ToolCalled, "tool_called"},
		{"ToolOutput", ToolOutput, "tool_output"},
		{"ReasoningItemCreated", ReasoningItemCreated, "reasoning_item_created"},
		{"MCPApprovalRequested", MCPApprovalRequested, "mcp_approval_requested"},
		{"MCPApprovalResponse", MCPApprovalResponse, "mcp_approval_response"},
		{"MCPListTools", MCPListTools, "mcp_list_tools"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.eventName); got != tt.want {
				t.Errorf("RunItemEventName = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventInterface(_ *testing.T) {
	// Verify all event types implement the Event interface
	var _ Event = (*RawResponseEvent)(nil)
	var _ Event = (*RunItemEvent)(nil)
	var _ Event = (*AgentUpdatedEvent)(nil)
}
