package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"

	"github.com/MitulShah1/openai-agents-go/internal/jsonschema"
)

// Session interface for conversation history persistence.
// Users should use implementations from github.com/MitulShah1/openai-agents-go/session
type Session interface {
	Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error)
	Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error
}

// NotFoundError is returned when a session doesn't exist.
type NotFoundError struct {
	SessionID string
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("session '%s' not found", e.SessionID)
}

// Runner manages the execution of agents.
type Runner struct {
	Client *openai.Client
}

// NewRunner creates a new Runner.
func NewRunner(client *openai.Client) *Runner {
	return &Runner{
		Client: client,
	}
}

// Run executes the agent loop with the given configuration.
// If session is provided, conversation history will be automatically loaded and saved.
func (r *Runner) Run(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	contextParams ContextVariables,
	config *RunConfig,
	session Session,
	sessionID string,
) (*Result, error) {
	if len(messages) == 0 {
		return nil, ErrNoMessages
	}

	// Use default config if not provided
	if config == nil {
		config = DefaultRunConfig()
	}

	// Apply timeout if specified
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Initialize context variables
	if contextParams == nil {
		contextParams = make(ContextVariables)
	}

	// Execute OnBeforeRun hook
	if agent.OnBeforeRun != nil {
		if err := agent.OnBeforeRun(ctx, agent); err != nil {
			return nil, fmt.Errorf("OnBeforeRun hook failed: %w", err)
		}
	}

	// Run input guardrails on the first agent (before any execution)
	if len(agent.InputGuardrails) > 0 && len(messages) > 0 {
		// Use string representation of messages for guardrail validation
		userInput := fmt.Sprintf("%v", messages[len(messages)-1])
		if err := r.runInputGuardrails(ctx, agent, userInput); err != nil {
			return nil, err
		}
	}

	// Load session history if session is provided
	if session != nil && sessionID != "" {
		sessionHistory, err := session.Get(ctx, sessionID)
		if err != nil {
			// If session not found, that's okay - we'll create it on save
			// Only fail on other errors
			if _, ok := err.(*NotFoundError); !ok {
				return nil, fmt.Errorf("failed to load session: %w", err)
			}
		} else {
			// Prepend session history to messages
			messages = append(sessionHistory, messages...)
		}
	}

	currentAgent := agent
	history := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	copy(history, messages)

	var usage Usage
	var steps []Step
	var lastMessage openai.ChatCompletionMessage
	turnCount := 0

	for {
		// Check max turns
		if config.MaxTurns > 0 && turnCount >= config.MaxTurns {
			return nil, ErrMaxTurnsExceeded
		}

		// Check context cancellation (timeout)
		if err := ctx.Err(); err != nil {
			if err == context.DeadlineExceeded {
				return nil, ErrTimeout
			}
			return nil, err
		}

		stepStart := time.Now()
		turnCount++

		// Prepare tools
		var tools []openai.ChatCompletionToolParam
		toolMap := make(map[string]Tool)
		for _, t := range currentAgent.Tools {
			tools = append(tools, t.ToParam())
			toolMap[t.Name] = t
		}

		// Prepare request
		req, err := r.prepareRequest(ctx, currentAgent, config, tools, history)
		if err != nil {
			return nil, err
		}

		// Call OpenAI
		completion, err := r.Client.Chat.Completions.New(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// Track usage
		if completion.Usage.PromptTokens > 0 {
			usage.Add(Usage{
				PromptTokens:     int(completion.Usage.PromptTokens),
				CompletionTokens: int(completion.Usage.CompletionTokens),
				TotalTokens:      int(completion.Usage.TotalTokens),
			})
		}

		message := completion.Choices[0].Message

		// Truncate tool call IDs in the assistant message if needed
		if len(message.ToolCalls) > 0 {
			for i := range message.ToolCalls {
				if len(message.ToolCalls[i].ID) > 40 {
					message.ToolCalls[i].ID = message.ToolCalls[i].ID[:40]
				}
			}
		}

		history = append(history, message.ToParam())

		// Record step
		step := Step{
			AgentName:  currentAgent.Name,
			StepNumber: turnCount,
			Duration:   time.Since(stepStart),
		}

		// Check for tool calls
		if len(message.ToolCalls) == 0 {
			// No tools called, save the final message and exit
			lastMessage = message
			steps = append(steps, step)
			break
		}

		// Handle Tool Calls
		toolMessages, recordedToolCalls, nextAgent := r.handleToolCalls(message.ToolCalls, toolMap, contextParams, currentAgent)

		step.ToolCalls = recordedToolCalls
		history = append(history, toolMessages...)

		if nextAgent != nil {
			currentAgent = nextAgent
		}

		step.Duration = time.Since(stepStart)
		steps = append(steps, step)

		// Continue loop
	}

	// Extract final output
	finalOutput := ""
	if len(history) > 0 {
		if len(lastMessage.Content) > 0 {
			finalOutput = lastMessage.Content
		} else if lastMessage.Refusal != "" {
			finalOutput = lastMessage.Refusal
		}
	}

	result := &Result{
		Messages:    history,
		Agent:       currentAgent,
		Usage:       usage,
		Steps:       steps,
		FinalOutput: finalOutput,
	}

	// Run output guardrails on the agent output
	if len(agent.OutputGuardrails) > 0 && finalOutput != "" {
		if err := r.runOutputGuardrails(ctx, agent, finalOutput); err != nil {
			return result, err
		}
	}

	// Save session history if session is provided
	if session != nil && sessionID != "" {
		if err := session.Append(ctx, sessionID, history); err != nil {
			return result, fmt.Errorf("failed to save session: %w", err)
		}
	}

	// Execute OnAfterRun hook
	if agent.OnAfterRun != nil {
		if err := agent.OnAfterRun(ctx, agent); err != nil {
			return result, fmt.Errorf("OnAfterRun hook failed: %w", err)
		}
	}

	return result, nil
}

func (r *Runner) prepareRequest(
	ctx context.Context,
	agent *Agent,
	config *RunConfig,
	tools []openai.ChatCompletionToolParam,
	history []openai.ChatCompletionMessageParamUnion,
) (openai.ChatCompletionNewParams, error) {
	req := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(agent.Model),
	}

	// Apply model settings
	if config.Temperature != nil {
		req.Temperature = openai.Float(*config.Temperature)
	} else if agent.Temperature != nil {
		req.Temperature = openai.Float(*agent.Temperature)
	}

	if config.MaxTokens != nil {
		req.MaxTokens = openai.Int(int64(*config.MaxTokens))
	} else if agent.MaxTokens != nil {
		req.MaxTokens = openai.Int(int64(*agent.MaxTokens))
	}

	if len(tools) > 0 {
		req.Tools = tools
		parallelCalls := agent.ParallelToolCalls
		if config.ParallelToolCalls != nil {
			parallelCalls = *config.ParallelToolCalls
		}
		if !parallelCalls {
			req.ParallelToolCalls = openai.Bool(false)
		}
	}

	// Apply response format
	var responseFormat *jsonschema.ResponseFormat
	if config.ResponseFormat != nil {
		responseFormat = config.ResponseFormat
	} else if agent.ResponseFormat != nil {
		responseFormat = agent.ResponseFormat
	}

	if responseFormat != nil {
		if responseFormat.Type == "text" {
			req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
				OfText: &openai.ResponseFormatTextParam{
					Type: "text",
				},
			}
		} else if responseFormat.Type == "json_schema" && responseFormat.JSONSchema != nil {
			js := responseFormat.JSONSchema
			schemaMap, err := js.Schema.ToMap()
			if err != nil {
				return req, fmt.Errorf("invalid schema: %w", err)
			}

			params := openai.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:   js.Name,
				Schema: schemaMap,
				Strict: openai.Bool(js.Strict),
			}
			if js.Description != "" {
				params.Description = openai.String(js.Description)
			}

			req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
					Type:       "json_schema",
					JSONSchema: params,
				},
			}
		}
	}

	// Inject system instructions
	instructions := agent.GetInstructions(ctx)
	messagesForTurn := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(instructions),
	}
	messagesForTurn = append(messagesForTurn, history...)
	req.Messages = messagesForTurn

	return req, nil
}

func (r *Runner) handleToolCalls(
	toolCalls []openai.ChatCompletionMessageToolCall,
	toolMap map[string]Tool,
	contextParams ContextVariables,
	currentAgent *Agent,
) ([]openai.ChatCompletionMessageParamUnion, []ToolCall, *Agent) {
	var messages []openai.ChatCompletionMessageParamUnion
	var recordedToolCalls []ToolCall
	nextAgent := currentAgent

	for _, toolCall := range toolCalls {
		toolStart := time.Now()
		toolName := toolCall.Function.Name
		args := toolCall.Function.Arguments

		tool, found := toolMap[toolName]
		var result any
		var err error

		if !found {
			// Provide helpful error with available tools
			available := make([]string, 0, len(toolMap))
			for name := range toolMap {
				available = append(available, name)
			}
			result = fmt.Sprintf("Error: Tool %s not found. Available tools: %v", toolName, available)
			err = fmt.Errorf("tool %s not found (available: %v)", toolName, available)
		} else {
			result, err = tool.Execute(args, contextParams)
			if err != nil {
				result = fmt.Sprintf("Error executing tool %s: %v", toolName, err)
				err = NewToolExecutionError(toolName, err)
			}
		}

		// Record tool call
		recordedToolCalls = append(recordedToolCalls, ToolCall{
			ToolName:  toolName,
			Arguments: args,
			Result:    result,
			Error:     err,
			Duration:  time.Since(toolStart),
		})

		// Check for Handoff
		if extractedAgent, ok := IsHandoff(result); ok {
			nextAgent = extractedAgent
			result = fmt.Sprintf("Transferred to %s", nextAgent.Name)
		}

		// Add tool output to history
		toolCallID := toolCall.ID
		if len(toolCallID) > 40 {
			toolCallID = toolCallID[:40]
		}
		resultStr := fmt.Sprintf("%v", result)
		messages = append(messages, openai.ToolMessage(resultStr, toolCallID))
	}

	return messages, recordedToolCalls, nextAgent
}
