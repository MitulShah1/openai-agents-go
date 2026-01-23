// Package stream provides streaming support for agent execution.
//
// This package implements streaming functionality that matches the Python SDK's
// streaming capabilities, including:
//   - Raw response events from the LLM (OpenAI Responses API format)
//   - High-level semantic events (tool calls, handoffs, message outputs)
//   - Agent transition events
//   - Real-time function call argument streaming
//   - Reasoning content support
//
// # Event Types
//
// There are three main event types:
//
//  1. RawResponseEvent: Raw events from the LLM in OpenAI Responses API format
//  2. RunItemEvent: High-level semantic events for tool calls, messages, handoffs
//  3. AgentUpdatedEvent: Notifications when the current agent changes
//
// # Usage
//
// Basic streaming example:
//
//	result, err := runner.StreamWithResult(ctx, agent, messages)
//	if err != nil {
//	    return err
//	}
//
//	for event, err := range result.StreamEvents(ctx) {
//	    if err != nil {
//	        return err
//	    }
//
//	    switch e := event.(type) {
//	    case *stream.RawResponseEvent:
//	        // Handle raw LLM events
//	        if e.Type == "response.output_text.delta" {
//	            fmt.Print(e.Data)
//	        }
//	    case *stream.RunItemEvent:
//	        // Handle semantic events
//	        if e.Name == string(stream.ToolCalled) {
//	            fmt.Println("Tool called:", e.Item)
//	        }
//	    case *stream.AgentUpdatedEvent:
//	        fmt.Println("Agent changed to:", e.NewAgent.Name)
//	    }
//	}
package stream
