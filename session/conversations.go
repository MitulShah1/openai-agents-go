package session

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/conversations"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

// ConversationsSession implements the Session interface using the OpenAI Conversations API.
// It maps the user's sessionID to an OpenAI conversation ID.
type ConversationsSession struct {
	client          *openai.Client
	mu              sync.RWMutex
	conversationIDs map[string]string // Maps sessionID -> OpenAI conversation ID
}

// NewConversationsSession creates a new ConversationsSession.
func NewConversationsSession(client *openai.Client) *ConversationsSession {
	return &ConversationsSession{
		client:          client,
		conversationIDs: make(map[string]string),
	}
}

// Get retrieves the conversation history from the OpenAI Conversations API.
func (s *ConversationsSession) Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	s.mu.RLock()
	convID, exists := s.conversationIDs[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, nil // Return empty history for new session
	}

	// List all items in the conversation with auto-paging
	pager := s.client.Conversations.Items.ListAutoPaging(ctx, convID, conversations.ItemListParams{})

	var items []conversations.ConversationItemUnion
	for pager.Next() {
		items = append(items, pager.Current())
	}
	if err := pager.Err(); err != nil {
		// Handle 404 (conversation deleted) by returning empty history
		var apiErr *openai.Error
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list conversation items: %w", err)
	}

	// Convert items back to ChatCompletion messages
	return s.convertItemsToMessages(items)
}

// Append adds new messages to the conversation.
func (s *ConversationsSession) Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	s.mu.Lock()
	convID, exists := s.conversationIDs[sessionID]
	s.mu.Unlock()

	inputItems, err := s.convertMessagesToInputItems(messages)
	if err != nil {
		return err
	}

	if len(inputItems) == 0 {
		return nil
	}

	if exists {
		// Try to add items to existing conversation
		_, err := s.client.Conversations.Items.New(ctx, convID, conversations.ItemNewParams{
			Items: inputItems,
		})

		// If success or not 404, we are done
		if err == nil {
			return nil
		}

		// If error is not 404, return it
		var apiErr *openai.Error
		if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
			return fmt.Errorf("failed to append items: %w", err)
		}

		// If 404, conversation was deleted remotely. Fall through to create new.
	}

	// Create new conversation
	conv, err := s.client.Conversations.New(ctx, conversations.ConversationNewParams{
		Items: inputItems,
		Metadata: shared.Metadata{
			"session_id": sessionID,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	s.mu.Lock()
	s.conversationIDs[sessionID] = conv.ID
	s.mu.Unlock()

	return nil
}

// Clear handles clearing the session history.
// Since Conversations API doesn't support "clearing" but keeping ID, we delete and remove mapping.
// A new one will be created on next Append.
func (s *ConversationsSession) Clear(ctx context.Context, sessionID string) error {
	return s.Delete(ctx, sessionID)
}

// Delete removes the session from the backend.
func (s *ConversationsSession) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	convID, exists := s.conversationIDs[sessionID]
	delete(s.conversationIDs, sessionID)
	s.mu.Unlock()

	if !exists {
		return nil
	}

	_, err := s.client.Conversations.Delete(ctx, convID)
	// Ignore 404s on delete
	var apiErr *openai.Error
	if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}

// convertMessagesToInputItems converts ChatCompletion messages to Conversations API input items
func (s *ConversationsSession) convertMessagesToInputItems(messages []openai.ChatCompletionMessageParamUnion) ([]responses.ResponseInputItemUnionParam, error) {
	var results []responses.ResponseInputItemUnionParam

	for _, msg := range messages {
		// ChatCompletionMessageParamUnion is a struct with pointer fields
		if p := msg.OfSystem; p != nil {
			text := extractTextFromUnion(p.Content)
			results = append(results, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Role:    responses.EasyInputMessageRoleSystem,
					Content: responses.EasyInputMessageContentUnionParam{OfString: openai.String(text)},
				},
			})
		} else if p := msg.OfUser; p != nil {
			text := extractTextFromUnion(p.Content)
			results = append(results, responses.ResponseInputItemUnionParam{
				OfMessage: &responses.EasyInputMessageParam{
					Role:    responses.EasyInputMessageRoleUser,
					Content: responses.EasyInputMessageContentUnionParam{OfString: openai.String(text)},
				},
			})
		} else if p := msg.OfAssistant; p != nil {
			// 1. Content
			text := extractTextFromUnion(p.Content)
			hasContent := text != ""
			// ToolCalls is pure slice
			hasTools := len(p.ToolCalls) > 0

			if hasContent || !hasTools {
				results = append(results, responses.ResponseInputItemUnionParam{
					OfMessage: &responses.EasyInputMessageParam{
						Role:    responses.EasyInputMessageRoleAssistant,
						Content: responses.EasyInputMessageContentUnionParam{OfString: openai.String(text)},
					},
				})
			}

			// 2. Tool Calls
			for _, toolCall := range p.ToolCalls {
				if f := toolCall.OfFunction; f != nil {
					results = append(results, responses.ResponseInputItemUnionParam{
						OfFunctionCall: &responses.ResponseFunctionToolCallParam{
							ID:        openai.String(f.ID), // Input param, optional ID wrapper
							CallID:    f.ID,
							Name:      f.Function.Name, // Function.Name is string
							Arguments: f.Function.Arguments,
						},
					})
				}
			}
		} else if p := msg.OfTool; p != nil {
			text := extractTextFromUnion(p.Content)
			results = append(results, responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: p.ToolCallID,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{OfString: openai.String(text)},
				},
			})
		} else {
			// Skip unknown or unsupported types (Developer, Function, etc not mapped yet)
			continue
		}
	}
	return results, nil
}

// convertItemsToMessages converts Conversations API items back to ChatCompletion messages
// It handles coalescing tool calls back into assistant messages.
func (s *ConversationsSession) convertItemsToMessages(items []conversations.ConversationItemUnion) ([]openai.ChatCompletionMessageParamUnion, error) {
	var messages []openai.ChatCompletionMessageParamUnion

	var currentAssistantMsg *openai.ChatCompletionAssistantMessageParam

	// Helper to flush accumulated assistant message
	flushAssistant := func() {
		if currentAssistantMsg != nil {
			messages = append(messages, openai.ChatCompletionMessageParamUnion{
				OfAssistant: currentAssistantMsg,
			})
			currentAssistantMsg = nil
		}
	}

	for _, item := range items {
		switch item.Type {
		case "message":
			flushAssistant()

			text := extractTextFromContent(item.Content)

			switch item.Role {
			case conversations.MessageRoleUser:
				u := openai.UserMessage(text)
				messages = append(messages, u) // u is Union
			case conversations.MessageRoleSystem:
				messages = append(messages, openai.SystemMessage(text))
			case conversations.MessageRoleAssistant:
				// Start accumulating
				// We need "naked" struct to append tool calls later
				// But openai.AssistantMessage returns Union.
				// We need to extract the underlying param struct.
				u := openai.AssistantMessage(text)
				currentAssistantMsg = u.OfAssistant // Extract pointer
			}

		case "function_call":
			if currentAssistantMsg == nil {
				u := openai.AssistantMessage("")
				currentAssistantMsg = u.OfAssistant
			}

			// Map ToolCall
			// Note: Name is missing from ConversationItemUnion but required for ToolCall logic usually.
			// We use placeholder "unknown" or empty.
			call := openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID: item.CallID, // Strict string
					// Type elided (defaults to function)
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      "unknown",
						Arguments: item.Arguments,
					},
				},
			}

			// Append to helper slices if they exist?
			// ChatCompletionAssistantMessageParam has ToolCalls which is []ChatCompletionMessageToolCallUnionParam

			// We need to dereference/manipulate slice
			// p.ToolCalls might be nil or empty.
			// The field IS `ToolCalls param.Field[[]ChatCompletionMessageToolCallUnionParam]`
			// We need to use `openai.F(...)` to set it or `Append` logic.

			// Since we have a pointer to struct, we can modify it.
			currentTools := currentAssistantMsg.ToolCalls
			currentTools = append(currentTools, call)
			currentAssistantMsg.ToolCalls = currentTools

		case "function_call_output":
			flushAssistant()

			// Tool Output
			// Content is the output
			text := extractTextFromContent(item.Content)
			messages = append(messages, openai.ToolMessage(item.CallID, text))

		default:
			// Ignore unsupported types (like evaluations, etc)
			flushAssistant()
		}
	}
	flushAssistant()

	return messages, nil
}

func extractTextFromContent(content conversations.ConversationItemUnionContent) string {
	var sb strings.Builder
	for _, part := range content.OfMessageContentArray {
		if part.Type == "text" {
			sb.WriteString(part.Text)
		}
	}
	return sb.String()
}

// extractTextFromUnion handles extracting text from various ChatCompletion Content Unions.
// Uses reflection or specific type checks is hard without generics.
// We implement specific checks for likely types or default to string representation.
func extractTextFromUnion(v interface{}) string {
	// 1. Check for standard param.Opt[string] or similar via reflection-like approach?
	// Using fmt.Sprint is safest fallback if we can't access fields.
	// However, we know specific types:
	// openai.ChatCompletionSystemMessageParamContentUnion
	// openai.ChatCompletionUserMessageParamContentUnion
	// openai.ChatCompletionAssistantMessageParamContentUnion
	// openai.ChatCompletionToolMessageParamContentUnion

	// Each has parameters. But for input params, the fields are accessible?
	// Actually, earlier error `type ... has no field Value`.
	// We should check if `OfString` field exists.
	// If not accessible via interface, we can't do much without reflection.
	// But since this file is in `session` package, it imports `openai`.
	// We can type switch!

	switch u := v.(type) {
	case openai.ChatCompletionSystemMessageParamContentUnion:
		return u.OfString.Value
	case openai.ChatCompletionUserMessageParamContentUnion:
		return u.OfString.Value
	case openai.ChatCompletionAssistantMessageParamContentUnion:
		return u.OfString.Value
	case openai.ChatCompletionToolMessageParamContentUnion:
		return u.OfString.Value
	}
	return ""
}
