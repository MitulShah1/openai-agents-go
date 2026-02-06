package stream

import (
	"github.com/openai/openai-go/v3"
)

// Handler processes streaming chat completion chunks and converts them into
// structured stream events. It maintains state across chunks to reconstruct
// complete messages, function calls, and reasoning content.
type Handler struct {
	state  *streamingState
	seqNum *SequenceNumber
}

// NewHandler creates a new stream handler
func NewHandler() *Handler {
	return &Handler{
		state:  newStreamingState(),
		seqNum: NewSequenceNumber(),
	}
}

// streamingState tracks the current state of streaming across chunks
type streamingState struct {
	started bool

	// Content tracking
	textContentIndex      *textContentOutput
	refusalContentIndex   *refusalContentOutput
	reasoningContentIndex *reasoningContentOutput

	// Function call tracking
	functionCalls         map[int]*functionCallBuilder
	functionCallStreaming map[int]bool
	functionCallOutputIdx map[int]int

	// Provider-specific data
	providerData map[string]any
}

// textContentOutput tracks accumulated text content
type textContentOutput struct {
	index int
	text  string
}

// refusalContentOutput tracks refusal content
type refusalContentOutput struct {
	index   int
	refusal string
}

// reasoningContentOutput tracks reasoning content (reserved for future use)
type reasoningContentOutput struct {
	_ int // reserved
}

// functionCallBuilder accumulates function call data across chunks
type functionCallBuilder struct {
	index        int
	callID       string
	funcType     string
	name         string
	arguments    string
	providerData map[string]any
}

func (s *streamingState) reasoningOutputOffset() int {
	if s.reasoningContentIndex != nil {
		return 1
	}
	return 0
}

func newStreamingState() *streamingState {
	return &streamingState{
		functionCalls:         make(map[int]*functionCallBuilder),
		functionCallStreaming: make(map[int]bool),
		functionCallOutputIdx: make(map[int]int),
		providerData:          make(map[string]any),
	}
}

// ProcessChunk processes a single chat completion chunk and returns the generated events
func (h *Handler) ProcessChunk(chunk openai.ChatCompletionChunk) []Event {
	events := make([]Event, 0)

	// Emit response.created event on first chunk
	if !h.state.started {
		h.state.started = true
		events = append(events, &RawResponseEvent{
			Type:           "response.created",
			Data:           chunk,
			SequenceNumber: h.seqNum.Next(),
		})
	}

	// Update provider data
	if chunk.Model != "" {
		h.state.providerData["model"] = chunk.Model
	}
	if chunk.ID != "" {
		h.state.providerData["response_id"] = chunk.ID
	}

	// Process choices
	if len(chunk.Choices) == 0 {
		return events
	}

	delta := chunk.Choices[0].Delta

	// Handle text content
	if delta.Content != "" {
		contentEvents := h.processTextContent(delta.Content)
		events = append(events, contentEvents...)
	}

	// Handle refusal
	if delta.Refusal != "" {
		refusalEvents := h.processRefusal(delta.Refusal)
		events = append(events, refusalEvents...)
	}

	// Handle tool calls
	if len(delta.ToolCalls) > 0 {
		toolCallEvents := h.processToolCalls(delta.ToolCalls)
		events = append(events, toolCallEvents...)
	}

	return events
}

// processTextContent handles text content deltas
func (h *Handler) processTextContent(content string) []Event {
	events := make([]Event, 0)

	// Initialize text content tracking if needed
	if h.state.textContentIndex == nil {
		contentIndex := 0
		if h.state.reasoningContentIndex != nil {
			contentIndex++
		}
		if h.state.refusalContentIndex != nil {
			contentIndex++
		}

		h.state.textContentIndex = &textContentOutput{
			index: contentIndex,
			text:  "",
		}

		// Emit output_item.added event
		events = append(events, &RawResponseEvent{
			Type: "response.output_item.added",
			Data: map[string]any{
				"output_index": h.state.reasoningOutputOffset(),
				"item": map[string]any{
					"type":    "message",
					"role":    "assistant",
					"content": []any{},
					"status":  "in_progress",
				},
			},
			SequenceNumber: h.seqNum.Next(),
		})

		// Emit content_part.added event
		events = append(events, &RawResponseEvent{
			Type: "response.content_part.added",
			Data: map[string]any{
				"content_index": h.state.textContentIndex.index,
				"output_index":  h.state.reasoningOutputOffset(),
				"part": map[string]any{
					"type": "output_text",
					"text": "",
				},
			},
			SequenceNumber: h.seqNum.Next(),
		})
	}

	// Emit text delta event
	events = append(events, &RawResponseEvent{
		Type: "response.output_text.delta",
		Data: map[string]any{
			"content_index": h.state.textContentIndex.index,
			"output_index":  h.state.reasoningOutputOffset(),
			"delta":         content,
		},
		SequenceNumber: h.seqNum.Next(),
	})

	// Accumulate text
	h.state.textContentIndex.text += content

	return events
}

// processRefusal handles refusal content
func (h *Handler) processRefusal(refusal string) []Event {
	events := make([]Event, 0)

	// Initialize refusal tracking if needed
	if h.state.refusalContentIndex == nil {
		refusalIndex := 0
		if h.state.reasoningContentIndex != nil {
			refusalIndex++
		}
		if h.state.textContentIndex != nil {
			refusalIndex++
		}

		h.state.refusalContentIndex = &refusalContentOutput{
			index:   refusalIndex,
			refusal: "",
		}

		// Emit output_item.added and content_part.added events
		events = append(events, &RawResponseEvent{
			Type: "response.output_item.added",
			Data: map[string]any{
				"output_index": h.state.reasoningOutputOffset(),
				"item": map[string]any{
					"type":    "message",
					"role":    "assistant",
					"content": []any{},
					"status":  "in_progress",
				},
			},
			SequenceNumber: h.seqNum.Next(),
		})

		events = append(events, &RawResponseEvent{
			Type: "response.content_part.added",
			Data: map[string]any{
				"content_index": h.state.refusalContentIndex.index,
				"output_index":  h.state.reasoningOutputOffset(),
				"part": map[string]any{
					"type":    "refusal",
					"refusal": "",
				},
			},
			SequenceNumber: h.seqNum.Next(),
		})
	}

	// Emit refusal delta event
	events = append(events, &RawResponseEvent{
		Type: "response.refusal.delta",
		Data: map[string]any{
			"content_index": h.state.refusalContentIndex.index,
			"output_index":  h.state.reasoningOutputOffset(),
			"delta":         refusal,
		},
		SequenceNumber: h.seqNum.Next(),
	})

	h.state.refusalContentIndex.refusal += refusal

	return events
}

// processToolCalls handles tool call deltas
func (h *Handler) processToolCalls(toolCalls []openai.ChatCompletionChunkChoiceDeltaToolCall) []Event {
	events := make([]Event, 0)

	for _, tc := range toolCalls {
		idx := int(tc.Index)

		// Initialize function call builder if needed
		if _, exists := h.state.functionCalls[idx]; !exists {
			h.state.functionCalls[idx] = &functionCallBuilder{
				index:        idx,
				arguments:    "",
				providerData: make(map[string]any),
			}
			h.state.functionCallStreaming[idx] = false
		}

		builder := h.state.functionCalls[idx]

		// Accumulate data
		if tc.ID != "" {
			builder.callID = tc.ID
		}
		if tc.Type != "" {
			builder.funcType = tc.Type
		}
		if tc.Function.Name != "" {
			builder.name = tc.Function.Name
		}
		if tc.Function.Arguments != "" {
			builder.arguments += tc.Function.Arguments
		}

		// Start streaming if we have name and call_id
		if !h.state.functionCallStreaming[idx] && builder.name != "" && builder.callID != "" {
			// Calculate output index
			outputIndex := 0
			if h.state.reasoningContentIndex != nil {
				outputIndex++
			}
			if h.state.textContentIndex != nil {
				outputIndex++
			}
			if h.state.refusalContentIndex != nil {
				outputIndex++
			}

			// Add offset for already started function calls
			for _, streaming := range h.state.functionCallStreaming {
				if streaming {
					outputIndex++
				}
			}

			h.state.functionCallStreaming[idx] = true
			h.state.functionCallOutputIdx[idx] = outputIndex

			// Emit output_item.added event
			events = append(events, &RawResponseEvent{
				Type: "response.output_item.added",
				Data: map[string]any{
					"output_index": outputIndex,
					"item": map[string]any{
						"type":      "function_call",
						"call_id":   builder.callID,
						"name":      builder.name,
						"arguments": "",
					},
				},
				SequenceNumber: h.seqNum.Next(),
			})
		}

		// Stream arguments if streaming has started
		if h.state.functionCallStreaming[idx] && tc.Function.Arguments != "" {
			outputIndex := h.state.functionCallOutputIdx[idx]
			events = append(events, &RawResponseEvent{
				Type: "response.function_call_arguments.delta",
				Data: map[string]any{
					"output_index": outputIndex,
					"delta":        tc.Function.Arguments,
				},
				SequenceNumber: h.seqNum.Next(),
			})
		}
	}

	return events
}

// Finalize emits completion events for all accumulated content
func (h *Handler) Finalize() []Event {
	events := make([]Event, 0)

	// Emit content_part.done for text content
	if h.state.textContentIndex != nil {
		events = append(events, &RawResponseEvent{
			Type: "response.content_part.done",
			Data: map[string]any{
				"content_index": h.state.textContentIndex.index,
				"output_index":  h.state.reasoningOutputOffset(),
				"part": map[string]any{
					"type": "output_text",
					"text": h.state.textContentIndex.text,
				},
			},
			SequenceNumber: h.seqNum.Next(),
		})
	}

	// Emit content_part.done for refusal
	if h.state.refusalContentIndex != nil {
		events = append(events, &RawResponseEvent{
			Type: "response.content_part.done",
			Data: map[string]any{
				"content_index": h.state.refusalContentIndex.index,
				"output_index":  h.state.reasoningOutputOffset(),
				"part": map[string]any{
					"type":    "refusal",
					"refusal": h.state.refusalContentIndex.refusal,
				},
			},
			SequenceNumber: h.seqNum.Next(),
		})
	}

	// Emit output_item.done for function calls
	for idx, builder := range h.state.functionCalls {
		if h.state.functionCallStreaming[idx] {
			outputIndex := h.state.functionCallOutputIdx[idx]
			events = append(events, &RawResponseEvent{
				Type: "response.output_item.done",
				Data: map[string]any{
					"output_index": outputIndex,
					"item": map[string]any{
						"type":      "function_call",
						"call_id":   builder.callID,
						"name":      builder.name,
						"arguments": builder.arguments,
					},
				},
				SequenceNumber: h.seqNum.Next(),
			})
		}
	}

	// Emit output_item.done for message
	if h.state.textContentIndex != nil || h.state.refusalContentIndex != nil {
		content := make([]map[string]any, 0)
		if h.state.textContentIndex != nil {
			content = append(content, map[string]any{
				"type": "output_text",
				"text": h.state.textContentIndex.text,
			})
		}
		if h.state.refusalContentIndex != nil {
			content = append(content, map[string]any{
				"type":    "refusal",
				"refusal": h.state.refusalContentIndex.refusal,
			})
		}

		events = append(events, &RawResponseEvent{
			Type: "response.output_item.done",
			Data: map[string]any{
				"output_index": h.state.reasoningOutputOffset(),
				"item": map[string]any{
					"type":    "message",
					"role":    "assistant",
					"content": content,
					"status":  "completed",
				},
			},
			SequenceNumber: h.seqNum.Next(),
		})
	}

	// Emit response.completed event
	events = append(events, &RawResponseEvent{
		Type:           "response.completed",
		Data:           map[string]any{"status": "completed"},
		SequenceNumber: h.seqNum.Next(),
	})

	return events
}

// GetFunctionCalls returns the accumulated function calls
func (h *Handler) GetFunctionCalls() []openai.ChatCompletionMessageToolCallUnion {
	calls := make([]openai.ChatCompletionMessageToolCallUnion, 0, len(h.state.functionCalls))

	// Sort by index to maintain order
	maxIdx := -1
	for idx := range h.state.functionCalls {
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	for i := 0; i <= maxIdx; i++ {
		if builder, ok := h.state.functionCalls[i]; ok {
			calls = append(calls, openai.ChatCompletionMessageToolCallUnion{
				ID:   builder.callID,
				Type: builder.funcType,
				Function: openai.ChatCompletionMessageFunctionToolCallFunction{
					Name:      builder.name,
					Arguments: builder.arguments,
				},
			})
		}
	}

	return calls
}

// GetAccumulatedText returns the accumulated text content
func (h *Handler) GetAccumulatedText() string {
	if h.state.textContentIndex != nil {
		return h.state.textContentIndex.text
	}
	return ""
}
