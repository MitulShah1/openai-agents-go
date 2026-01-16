package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/internal/jsonschema"
	"github.com/MitulShah1/openai-agents-go/internal/runner"
	"github.com/MitulShah1/openai-agents-go/session"
)

// Runner manages the execution of agents.
// It handles the orchestration of OpenAI API calls, tool execution,
// session management, and guardrail validation.
type Runner struct {
	// Client is the OpenAI API client used to make completion requests
	Client *openai.Client
}

// NewRunner creates a new Runner with the given OpenAI client.
//
// Example:
//
//	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
//	runner := agents.NewRunner(client)
func NewRunner(client *openai.Client) *Runner {
	return &Runner{
		Client: client,
	}
}

// Run executes the agent loop with functional options.
//
// This method uses functional options for optional parameters.
// Only required parameters are positional arguments, while all
// optional configuration is provided via option functions.
//
// Required parameters:
//   - ctx: Context for cancellation and timeout control
//   - agent: The agent to execute
//   - messages: Initial conversation messages (must not be empty)
//
// Optional parameters (via options):
//   - Config: Runtime configuration via WithConfig() (uses defaults if not provided)
//   - Context Variables: Variables accessible to tools via WithContextVariables()
//   - Session: Conversation persistence via WithSession()
//
// Example (minimal):
//
//	result, err := runner.Run(
//	    ctx,
//	    myAgent,
//	    []openai.ChatCompletionMessageParamUnion{openai.UserMessage("Hello")},
//	)
//
// Example (with session):
//
//	result, err := runner.Run(
//	    ctx,
//	    myAgent,
//	    []openai.ChatCompletionMessageParamUnion{openai.UserMessage("Hello")},
//	    agents.WithSession(mySession, "user_123"),
//	)
//
// Example (with multiple options):
//
//	result, err := runner.Run(
//	    ctx,
//	    myAgent,
//	    []openai.ChatCompletionMessageParamUnion{openai.UserMessage("Hello")},
//	    agents.WithSession(mySession, "user_123"),
//	    agents.WithConfig(&agents.RunConfig{MaxTurns: 5}),
//	    agents.WithContextVariables(vars),
//	)
func (r *Runner) Run(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	opts ...RunOption,
) (*Result, error) {
	// Apply options
	options := &runOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Call the internal implementation
	return r.execute(
		ctx,
		agent,
		messages,
		options.contextParams,
		options.config,
		options.sess,
		options.sessionID,
	)
}

// execute is the internal implementation of Run.
func (r *Runner) execute(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	contextParams ContextVariables,
	config *RunConfig,
	sess session.Session,
	sessionID string,
) (*Result, error) {
	// Validate input
	if len(messages) == 0 {
		return nil, ErrNoMessages
	}

	// Use default config if not provided
	if config == nil {
		config = DefaultRunConfig()
	}

	// Initialize context variables
	if contextParams == nil {
		contextParams = make(ContextVariables)
	}

	// Create executor with timeout configuration
	executor := runner.NewExecutor(config.MaxTurns, config.Timeout)

	// Apply timeout if specified
	var cancel context.CancelFunc
	ctx, cancel = executor.ApplyTimeout(ctx)
	defer cancel()

	// Execute OnBeforeRun hook
	if agent.OnBeforeRun != nil {
		if err := agent.OnBeforeRun(ctx, agent); err != nil {
			return nil, fmt.Errorf("OnBeforeRun hook failed: %w", err)
		}
	}

	// Run input guardrails
	if err := r.executeInputGuardrails(ctx, agent, messages); err != nil {
		return nil, err
	}

	// Load session history
	sessionHandler := runner.NewSessionHandler(sess, sessionID)
	var err error
	messages, err = sessionHandler.LoadHistory(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Execute main agent loop
	result, err := r.executeAgentLoop(ctx, agent, messages, contextParams, config, executor)
	if err != nil {
		return result, err
	}

	// Run output guardrails
	if err := r.executeOutputGuardrails(ctx, agent, result.FinalOutput); err != nil {
		return result, err
	}

	// Save session history
	if err := sessionHandler.SaveHistory(ctx, result.Messages); err != nil {
		return result, err
	}

	// Execute OnAfterRun hook
	if agent.OnAfterRun != nil {
		if err := agent.OnAfterRun(ctx, agent); err != nil {
			return result, fmt.Errorf("OnAfterRun hook failed: %w", err)
		}
	}

	return result, nil
}

// executeAgentLoop runs the main agent execution loop
func (r *Runner) executeAgentLoop(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	contextParams ContextVariables,
	config *RunConfig,
	executor *runner.Executor,
) (*Result, error) {
	currentAgent := agent
	history := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	copy(history, messages)

	var usage Usage
	var steps []Step
	var lastMessage openai.ChatCompletionMessage
	turnCount := 0

	for {
		// Check if execution should continue
		shouldContinue, err := executor.ShouldContinueExecution(ctx, turnCount)
		if err != nil {
			if strings.Contains(err.Error(), "max turns") {
				return nil, ErrMaxTurnsExceeded
			}
			if strings.Contains(err.Error(), "timeout") {
				return nil, ErrTimeout
			}
			return nil, err
		}
		if !shouldContinue {
			break
		}

		stepStart := time.Now()
		turnCount++

		// Prepare tools
		tools, toolMap := r.prepareTools(currentAgent)

		// Prepare and execute request
		req, err := r.prepareRequest(ctx, currentAgent, config, tools, history)
		if err != nil {
			return nil, err
		}

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

		// Truncate tool call IDs
		runner.TruncateToolCallIDs(&message)

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

		// Handle tool calls
		toolMessages, recordedToolCalls, nextAgent := r.handleToolCalls(
			message.ToolCalls,
			toolMap,
			contextParams,
			currentAgent,
		)

		step.ToolCalls = recordedToolCalls
		history = append(history, toolMessages...)

		if nextAgent != nil {
			currentAgent = nextAgent
		}

		step.Duration = time.Since(stepStart)
		steps = append(steps, step)
	}

	// Extract final output
	finalOutput := extractFinalOutput(lastMessage)

	return &Result{
		Messages:    history,
		Agent:       currentAgent,
		Usage:       usage,
		Steps:       steps,
		FinalOutput: finalOutput,
	}, nil
}

// prepareTools builds the tool map and parameter list
func (r *Runner) prepareTools(agent *Agent) ([]openai.ChatCompletionToolUnionParam, runner.ToolMap) {
	tools := make([]openai.ChatCompletionToolUnionParam, 0, len(agent.Tools))
	toolMap := make(runner.ToolMap, len(agent.Tools))

	for i := range agent.Tools {
		t := agent.Tools[i]
		tools = append(tools, t.ToParam())
		// Create an adapter to bridge Tool.Execute to ToolExecutor.Execute
		toolMap[t.Name] = &toolAdapter{tool: &t}
	}

	return tools, toolMap
}

// toolAdapter adapts a Tool to implement runner.ToolExecutor
type toolAdapter struct {
	tool *Tool
}

// Execute implements runner.ToolExecutor
func (ta *toolAdapter) Execute(arguments string, contextVariables map[string]any) (any, error) {
	// ContextVariables is just a type alias for map[string]any, so this is safe
	return ta.tool.Execute(arguments, ContextVariables(contextVariables))
}

// prepareRequest builds the OpenAI API request
func (r *Runner) prepareRequest(
	ctx context.Context,
	agent *Agent,
	config *RunConfig,
	tools []openai.ChatCompletionToolUnionParam,
	history []openai.ChatCompletionMessageParamUnion,
) (openai.ChatCompletionNewParams, error) {
	// Merge configurations
	merger := &runner.ConfigMerger{
		AgentTemperature:       agent.Temperature,
		AgentMaxTokens:         agent.MaxTokens,
		AgentParallelToolCalls: agent.ParallelToolCalls,
		AgentResponseFormat:    agent.ResponseFormat,
		RunTemperature:         config.Temperature,
		RunMaxTokens:           config.MaxTokens,
		RunParallelToolCalls:   config.ParallelToolCalls,
		RunResponseFormat:      config.ResponseFormat,
	}

	if config.Debug {
		fmt.Printf("DEBUG prepareRequest: agent.ResponseFormat=%v, config.ResponseFormat=%v\n",
			agent.ResponseFormat, config.ResponseFormat)
		fmt.Printf("DEBUG prepareRequest: merger.AgentResponseFormat=%v, merger.RunResponseFormat=%v\n",
			merger.AgentResponseFormat, merger.RunResponseFormat)
		fmt.Printf("DEBUG prepareRequest: merger.GetResponseFormat()=%v\n", merger.GetResponseFormat())
	}

	parallelToolCalls := merger.GetParallelToolCalls()
	requestConfig := &runner.RequestConfig{
		Model:              agent.Model,
		Temperature:        merger.GetTemperature(),
		MaxTokens:          merger.GetMaxTokens(),
		ParallelToolCalls:  &parallelToolCalls,
		ResponseFormat:     merger.GetResponseFormat(),
		SystemInstructions: agent.GetInstructions(ctx),
	}

	return runner.PrepareRequest(ctx, requestConfig, tools, history, r.convertResponseFormat)
}

// convertResponseFormat converts response format to OpenAI parameter format
func (r *Runner) convertResponseFormat(format any) (openai.ChatCompletionNewParamsResponseFormatUnion, error) {
	responseFormat, ok := format.(*jsonschema.ResponseFormat)
	if !ok || responseFormat == nil {
		return openai.ChatCompletionNewParamsResponseFormatUnion{}, nil
	}

	if responseFormat.Type == "text" {
		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfText: &openai.ResponseFormatTextParam{
				Type: "text",
			},
		}, nil
	}

	if responseFormat.Type == "json_schema" && responseFormat.JSONSchema != nil {
		js := responseFormat.JSONSchema
		schemaMap, err := js.Schema.ToMap()
		if err != nil {
			return openai.ChatCompletionNewParamsResponseFormatUnion{}, fmt.Errorf("invalid schema: %w", err)
		}

		params := openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:   js.Name,
			Schema: schemaMap,
			Strict: openai.Bool(js.Strict),
		}
		if js.Description != "" {
			params.Description = openai.String(js.Description)
		}

		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				Type:       "json_schema",
				JSONSchema: params,
			},
		}, nil
	}

	return openai.ChatCompletionNewParamsResponseFormatUnion{}, nil
}

// handleToolCalls executes tool calls and returns results
func (r *Runner) handleToolCalls(
	toolCalls []openai.ChatCompletionMessageToolCallUnion,
	toolMap runner.ToolMap,
	contextParams ContextVariables,
	_ *Agent,
) ([]openai.ChatCompletionMessageParamUnion, []ToolCall, *Agent) {
	// Use the internal tool handler
	messages, results, nextAgentAny := runner.HandleToolCalls(
		toolCalls,
		toolMap,
		contextParams,
		r.isHandoffFunc,
	)

	// Convert results to public ToolCall type
	recordedToolCalls := make([]ToolCall, len(results))
	for i, result := range results {
		recordedToolCalls[i] = ToolCall{
			ToolName:  result.ToolName,
			Arguments: result.Arguments,
			Result:    result.Result,
			Error:     result.Error,
			Duration:  result.Duration,
		}
	}

	// Handle agent handoff
	var nextAgent *Agent
	if nextAgentAny != nil {
		if extractedAgent, ok := nextAgentAny.(*Agent); ok {
			nextAgent = extractedAgent
			// Update the last message to indicate the transfer
			// The HandleToolCalls function already created the message with truncated ID
			// We just need to update the content, preserving the already-truncated ID
			if len(messages) > 0 {
				lastIdx := len(messages) - 1
				// Extract the truncated ID that HandleToolCalls already created
				var truncatedID string
				// Get the truncated tool call ID from the last processed tool call
				if lastIdx < len(toolCalls) {
					truncatedID = toolCalls[lastIdx].ID
					if len(truncatedID) > runner.MaxToolCallIDLength {
						truncatedID = truncatedID[:runner.MaxToolCallIDLength]
					}
				}
				messages[lastIdx] = openai.ToolMessage(
					fmt.Sprintf("Transferred to %s", nextAgent.Name),
					truncatedID,
				)
			}
		}
	}

	return messages, recordedToolCalls, nextAgent
}

// isHandoffFunc checks if a result is an agent handoff
func (r *Runner) isHandoffFunc(result any) (any, bool) {
	return IsHandoff(result)
}

// executeInputGuardrails runs input guardrails on the agent
func (r *Runner) executeInputGuardrails(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
) error {
	if len(agent.InputGuardrails) == 0 || len(messages) == 0 {
		return nil
	}

	// Convert guardrails to internal format
	guardrails := make([]*runner.Guardrail, len(agent.InputGuardrails))
	for i, gr := range agent.InputGuardrails {
		guardrails[i] = &runner.Guardrail{
			Name: gr.Name,
			Func: func(ctx context.Context, text string) (runner.GuardrailResult, error) {
				result, err := gr.Func(ctx, text)
				return runner.GuardrailResult{
					TripwireTriggered: result.TripwireTriggered,
					Message:           result.Message,
				}, err
			},
		}
	}

	executor := runner.NewGuardrailExecutor(guardrails, "input")
	userInput := fmt.Sprintf("%v", messages[len(messages)-1])

	if err := executor.Execute(ctx, userInput); err != nil {
		// Convert internal error to public error type
		if strings.Contains(err.Error(), "guardrail") {
			// Extract guardrail name and message from error
			return &guardrail.InputGuardrailTripwireError{
				GuardrailName: extractGuardrailName(err.Error()),
				Message:       err.Error(),
			}
		}
		return err
	}

	return nil
}

// executeOutputGuardrails runs output guardrails on the agent output
func (r *Runner) executeOutputGuardrails(
	ctx context.Context,
	agent *Agent,
	finalOutput string,
) error {
	if len(agent.OutputGuardrails) == 0 || finalOutput == "" {
		return nil
	}

	// Convert guardrails to internal format
	guardrails := make([]*runner.Guardrail, len(agent.OutputGuardrails))
	for i, gr := range agent.OutputGuardrails {
		guardrails[i] = &runner.Guardrail{
			Name: gr.Name,
			Func: func(ctx context.Context, text string) (runner.GuardrailResult, error) {
				result, err := gr.Func(ctx, text)
				return runner.GuardrailResult{
					TripwireTriggered: result.TripwireTriggered,
					Message:           result.Message,
				}, err
			},
		}
	}

	executor := runner.NewGuardrailExecutor(guardrails, "output")

	if err := executor.Execute(ctx, finalOutput); err != nil {
		// Convert internal error to public error type
		if strings.Contains(err.Error(), "guardrail") {
			return &guardrail.OutputGuardrailTripwireError{
				GuardrailName: extractGuardrailName(err.Error()),
				Message:       err.Error(),
			}
		}
		return err
	}

	return nil
}

// extractGuardrailName extracts the guardrail name from an error message
func extractGuardrailName(errMsg string) string {
	// Simple extraction from error message format
	// Format: "input guardrail 'name' triggered: message"
	start := strings.Index(errMsg, "'")
	if start == -1 {
		return "unknown"
	}
	end := strings.Index(errMsg[start+1:], "'")
	if end == -1 {
		return "unknown"
	}
	return errMsg[start+1 : start+1+end]
}

// extractFinalOutput extracts the final output from the last message
func extractFinalOutput(lastMessage openai.ChatCompletionMessage) string {
	if len(lastMessage.Content) > 0 {
		return lastMessage.Content
	}
	if lastMessage.Refusal != "" {
		return lastMessage.Refusal
	}
	return ""
}
