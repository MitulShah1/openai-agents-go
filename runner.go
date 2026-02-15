package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"

	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/internal/runner"
	"github.com/MitulShah1/openai-agents-go/jsonschema"
	"github.com/MitulShah1/openai-agents-go/models"
	"github.com/MitulShah1/openai-agents-go/session"
	"github.com/MitulShah1/openai-agents-go/tools"
	"github.com/MitulShah1/openai-agents-go/tracing"
	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// Runner manages the execution of agents.
// It handles the orchestration of OpenAI API calls, tool execution,
// session management, and guardrail validation.
type Runner struct {
	// Client is the OpenAI API client used to make completion requests.
	//
	// Deprecated: Use ModelProvider instead. Kept for backward compatibility.
	Client *openai.Client

	// ModelProvider is the default model provider for all agents.
	// When an agent does not have its own ModelProvider, this is used.
	ModelProvider models.ModelProvider
}

// NewRunner creates a new Runner with the given OpenAI client.
//
// Example:
//
//	client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
//	runner := agents.NewRunner(client)
func NewRunner(client *openai.Client) *Runner {
	return &Runner{
		Client:        client,
		ModelProvider: models.NewOpenAIProvider(client),
	}
}

// NewRunnerWithProvider creates a new Runner with the given model provider.
// This is the preferred constructor for using custom model providers.
//
// Example:
//
//	provider := models.NewOpenAIProvider(&client)
//	runner := agents.NewRunnerWithProvider(provider)
func NewRunnerWithProvider(provider models.ModelProvider) *Runner {
	return &Runner{
		ModelProvider: provider,
	}
}

// resolveModel returns the Model to use for the given agent.
// Resolution order:
//  1. Agent.ModelProvider (if set)
//  2. Runner.ModelProvider (if set)
//  3. Fallback: wrap Runner.Client in OpenAIProvider (backward compatibility)
func (r *Runner) resolveModel(agent *Agent) (models.Model, error) {
	provider := agent.ModelProvider
	if provider == nil {
		provider = r.ModelProvider
	}
	if provider == nil {
		if r.Client != nil {
			provider = models.NewOpenAIProvider(r.Client)
		} else {
			return nil, fmt.Errorf("no model provider configured: set Runner.ModelProvider or Agent.ModelProvider")
		}
	}
	return provider.GetModel(agent.Model)
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
		options.approvalHandler,
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
	approvalHandler tools.ApprovalHandler,
) (*Result, error) {
	// Validate input
	if len(messages) == 0 {
		return nil, ErrNoMessages
	}

	// Use default config if not provided
	if config == nil {
		config = DefaultRunConfig()
	}

	// Ensure trace exists (unless already inside a trace)
	if tracing.FromContext(ctx) == nil && !tracing.IsTracingDisabled() {
		var tr tracing.Trace
		var err error
		// Use functional options from config
		ctx, tr, err = tracing.GetProvider().StartTrace(ctx, traceOptionsFromRunConfig(config)...)
		_ = err // Ignore error, SDK continues with or without trace
		defer tr.End(ctx)
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
	newInputLen := len(messages)
	messages, err = sessionHandler.LoadHistory(ctx, messages)
	if err != nil {
		return nil, err
	}
	sessionPrefixLen := len(messages) - newInputLen

	// Execute main agent loop
	result, err := r.executeAgentLoop(ctx, agent, messages, contextParams, config, executor, approvalHandler)
	if err != nil {
		return result, err
	}

	// Run output guardrails
	if err := r.executeOutputGuardrails(ctx, agent, result.FinalOutput); err != nil {
		return result, err
	}

	// Save only new messages (skip the session prefix that was already persisted)
	if err := sessionHandler.SaveHistory(ctx, result.Messages[sessionPrefixLen:]); err != nil {
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
	approvalHandler tools.ApprovalHandler,
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
			if errors.Is(err, runner.ErrMaxTurnsExceeded) {
				return nil, ErrMaxTurnsExceeded
			}
			if errors.Is(err, runner.ErrTimeout) {
				return nil, ErrTimeout
			}
			return nil, err
		}
		if !shouldContinue {
			break
		}

		stepStart := time.Now()
		turnCount++

		// Create agent span for this iteration (no-op if no current trace)
		var agentSpan tracing.Span
		if !tracing.IsTracingDisabled() {
			var err error
			ctx, agentSpan, err = tracing.StartAgentSpan(ctx, currentAgent.Name,
				tracing.WithModel(currentAgent.Model),
				tracing.WithInstructions(currentAgent.GetInstructions(ctx)))
			if err != nil {
				agentSpan = nil
			}
		}

		// Prepare tools
		tools, toolMap := r.prepareTools(currentAgent)

		// Prepare and execute request
		req, err := r.prepareRequest(ctx, currentAgent, config, tools, history)
		if err != nil {
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, err
		}

		// Resolve the model for this agent
		model, err := r.resolveModel(currentAgent)
		if err != nil {
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, fmt.Errorf("failed to resolve model: %w", err)
		}

		// Resolve prompt if configured
		resolvedPrompt, err := currentAgent.GetPrompt(contextParams)
		if err != nil {
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, fmt.Errorf("prompt resolution failed: %w", err)
		}

		// Start generation span
		// Note: Sensitive data redaction is handled automatically by the exporter based on trace config
		ctxGen, genSpan, _ := tracing.StartGenerationSpan(ctx, tracing.WithModel(currentAgent.Model))

		resp, err := model.GetResponse(ctxGen, req, models.ModelSettings{Prompt: resolvedPrompt})
		if err != nil {
			genSpan.RecordError(err)
			genSpan.End(ctxGen)
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// Validate response
		if resp.Completion == nil || len(resp.Completion.Choices) == 0 {
			genSpan.RecordError(ErrEmptyModelResponse)
			genSpan.End(ctxGen)
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, ErrEmptyModelResponse
		}

		completion := resp.Completion

		genSpan.SetAttributes(map[string]any{
			"request":  req,
			"response": completion,
			"usage": &schema.Usage{
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				TotalTokens:      resp.Usage.TotalTokens,
			},
		})
		genSpan.End(ctxGen)

		// Track usage
		if resp.Usage.PromptTokens > 0 || resp.Usage.CompletionTokens > 0 || resp.Usage.TotalTokens > 0 {
			usage.Add(Usage{
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				TotalTokens:      resp.Usage.TotalTokens,
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
			lastMessage = message
			steps = append(steps, step)
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			break
		}

		// Check tool approvals before execution
		approvalErr := r.checkToolApprovals(
			message.ToolCalls, currentAgent, contextParams, approvalHandler,
			history, turnCount, config,
		)
		if approvalErr != nil {
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, approvalErr
		}

		// Handle tool calls
		toolMessages, recordedToolCalls, nextAgent := r.handleToolCalls(
			ctx,
			message.ToolCalls,
			toolMap,
			contextParams,
			currentAgent,
			resolveParallelToolCalls(currentAgent, config),
			config.MaxToolConcurrency,
		)

		step.ToolCalls = recordedToolCalls
		history = append(history, toolMessages...)

		if nextAgent != nil {
			// Record handoff span (Python parity).
			if tracing.FromContext(ctx) != nil && nextAgent.Name != "" {
				ctx2, hs, _ := tracing.StartHandoffSpan(ctx, currentAgent.Name, nextAgent.Name, "")
				hs.End(ctx2)
			}
			currentAgent = nextAgent
		}

		step.Duration = time.Since(stepStart)
		steps = append(steps, step)

		if agentSpan != nil {
			agentSpan.End(ctx)
		}
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

// checkToolApprovals evaluates whether any tool calls require approval.
// If an ApprovalHandler is provided, it is called synchronously for each tool
// needing approval. If no handler is provided, a ToolApprovalRequiredError is
// returned with a RunState snapshot for pause/resume.
func (r *Runner) checkToolApprovals(
	toolCalls []openai.ChatCompletionMessageToolCallUnion,
	agent *Agent,
	contextParams ContextVariables,
	approvalHandler tools.ApprovalHandler,
	history []openai.ChatCompletionMessageParamUnion,
	turnCount int,
	config *RunConfig,
) error {
	// Build a map of tool name -> Tool for approval checking
	toolIndex := make(map[string]*tools.Tool, len(agent.Tools))
	for i := range agent.Tools {
		toolIndex[agent.Tools[i].Name] = &agent.Tools[i]
	}

	var pendingRequests []tools.ApprovalRequest
	for _, tc := range toolCalls {
		t, ok := toolIndex[tc.Function.Name]
		if !ok {
			continue
		}

		args := parseToolArgs(tc.Function.Arguments)
		needs, err := t.RequiresApproval(args, tc.ID, contextParams)
		if err != nil {
			return fmt.Errorf("approval check failed for tool %s: %w", tc.Function.Name, err)
		}
		if !needs {
			continue
		}

		req := tools.ApprovalRequest{
			ToolName: tc.Function.Name,
			CallID:   tc.ID,
			Args:     args,
			Context:  contextParams,
		}

		if approvalHandler != nil {
			resp, err := approvalHandler(req)
			if err != nil {
				return fmt.Errorf("approval handler error for tool %s: %w", tc.Function.Name, err)
			}
			if !resp.Approved {
				return &ToolApprovalRequiredError{
					Requests: []tools.ApprovalRequest{req},
					State: &RunState{
						Agent:            agent,
						Messages:         history,
						TurnCount:        turnCount,
						PendingToolCalls: toolCalls,
						ContextVariables: contextParams,
						Config:           config,
					},
				}
			}
			continue
		}

		pendingRequests = append(pendingRequests, req)
	}

	if len(pendingRequests) > 0 {
		return &ToolApprovalRequiredError{
			Requests: pendingRequests,
			State: &RunState{
				Agent:            agent,
				Messages:         history,
				TurnCount:        turnCount,
				PendingToolCalls: toolCalls,
				ContextVariables: contextParams,
				Config:           config,
			},
		}
	}

	return nil
}

// parseToolArgs parses JSON arguments into a map, returning an empty map on failure.
func parseToolArgs(argsJSON string) map[string]any {
	if argsJSON == "" {
		return map[string]any{}
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]any{}
	}
	return args
}

// Resume continues execution after tool approval decisions have been made.
//
// This method is used in conjunction with ToolApprovalRequiredError to implement
// human-in-the-loop approval workflows. When Run returns a ToolApprovalRequiredError,
// the caller makes approval decisions and passes them to Resume along with the
// captured RunState.
//
// Approved tools are executed normally. Rejected tools produce a rejection message
// that is sent back to the model so it can respond appropriately.
//
// Example:
//
//	result, err := runner.Run(ctx, agent, messages)
//	var approvalErr *agents.ToolApprovalRequiredError
//	if errors.As(err, &approvalErr) {
//	    approvals := map[string]*tools.ApprovalResponse{
//	        "call_123": {Approved: true},
//	        "call_456": {Approved: false, Reason: "not allowed"},
//	    }
//	    result, err = runner.Resume(ctx, approvalErr.State, approvals)
//	}
func (r *Runner) Resume(
	ctx context.Context,
	state *RunState,
	approvals map[string]*tools.ApprovalResponse,
	opts ...RunOption,
) (*Result, error) {
	if state == nil {
		return nil, fmt.Errorf("resume: state is nil")
	}

	// Apply options
	options := &runOptions{}
	for _, opt := range opts {
		opt(options)
	}

	config := state.Config
	if options.config != nil {
		config = config.Merge(options.config)
	}
	if config == nil {
		config = DefaultRunConfig()
	}

	contextParams := state.ContextVariables
	if options.contextParams != nil {
		contextParams = options.contextParams
	}

	approvalHandler := options.approvalHandler

	// Rebuild tool map
	_, toolMap := r.prepareTools(state.Agent)

	// Process pending tool calls based on approval decisions
	var toolMessages []openai.ChatCompletionMessageParamUnion
	for _, tc := range state.PendingToolCalls {
		callID := tc.ID
		if len(callID) > runner.MaxToolCallIDLength {
			callID = callID[:runner.MaxToolCallIDLength]
		}

		resp, hasDecision := approvals[tc.ID]
		if !hasDecision || !resp.Approved {
			reason := "tool call rejected"
			if hasDecision && resp.Reason != "" {
				reason = resp.Reason
			}
			toolMessages = append(toolMessages, openai.ToolMessage(reason, callID))
			continue
		}

		// Execute approved tool
		executor, found := toolMap[tc.Function.Name]
		if !found {
			toolMessages = append(toolMessages, openai.ToolMessage(
				fmt.Sprintf("Error: Tool %s not found", tc.Function.Name), callID,
			))
			continue
		}
		result, err := executor.Execute(tc.Function.Arguments, contextParams)
		if err != nil {
			toolMessages = append(toolMessages, openai.ToolMessage(
				fmt.Sprintf("Error executing tool %s: %s", tc.Function.Name, err.Error()), callID,
			))
			continue
		}

		// Check for handoff
		if extractedAgent, ok := IsHandoff(result); ok {
			state.Agent = extractedAgent
			result = fmt.Sprintf("Transferred to %s", extractedAgent.Name)
		}

		toolMessages = append(toolMessages, openai.ToolMessage(fmt.Sprintf("%v", result), callID))
	}

	// Append tool messages to history and continue agent loop
	history := make([]openai.ChatCompletionMessageParamUnion, len(state.Messages))
	copy(history, state.Messages)
	history = append(history, toolMessages...)

	executor := runner.NewExecutor(config.MaxTurns, config.Timeout)
	var cancel context.CancelFunc
	ctx, cancel = executor.ApplyTimeout(ctx)
	defer cancel()

	// Continue the agent loop from the saved turn count
	result, err := r.executeAgentLoopResume(
		ctx, state.Agent, history, contextParams, config,
		executor, state.TurnCount, approvalHandler,
	)
	if err != nil {
		return result, err
	}

	// Run output guardrails
	if err := r.executeOutputGuardrails(ctx, state.Agent, result.FinalOutput); err != nil {
		return result, err
	}

	// Save session if provided
	if options.sess != nil {
		sessionHandler := runner.NewSessionHandler(options.sess, options.sessionID)
		if err := sessionHandler.SaveHistory(ctx, result.Messages); err != nil {
			return result, err
		}
	}

	return result, nil
}

// executeAgentLoopResume continues the agent loop from a saved state.
func (r *Runner) executeAgentLoopResume(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	contextParams ContextVariables,
	config *RunConfig,
	executor *runner.Executor,
	startTurnCount int,
	approvalHandler tools.ApprovalHandler,
) (*Result, error) {
	currentAgent := agent
	history := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	copy(history, messages)

	var usage Usage
	var steps []Step
	var lastMessage openai.ChatCompletionMessage
	turnCount := startTurnCount

	for {
		shouldContinue, err := executor.ShouldContinueExecution(ctx, turnCount)
		if err != nil {
			if errors.Is(err, runner.ErrMaxTurnsExceeded) {
				return nil, ErrMaxTurnsExceeded
			}
			if errors.Is(err, runner.ErrTimeout) {
				return nil, ErrTimeout
			}
			return nil, err
		}
		if !shouldContinue {
			break
		}

		stepStart := time.Now()
		turnCount++

		agentTools, toolMap := r.prepareTools(currentAgent)

		req, err := r.prepareRequest(ctx, currentAgent, config, agentTools, history)
		if err != nil {
			return nil, err
		}

		model, err := r.resolveModel(currentAgent)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve model: %w", err)
		}

		// Resolve prompt if configured
		resolvedPrompt, err := currentAgent.GetPrompt(contextParams)
		if err != nil {
			return nil, fmt.Errorf("prompt resolution failed: %w", err)
		}

		ctxGen, genSpan, _ := tracing.StartGenerationSpan(ctx, tracing.WithModel(currentAgent.Model))

		resp, err := model.GetResponse(ctxGen, req, models.ModelSettings{Prompt: resolvedPrompt})
		if err != nil {
			genSpan.RecordError(err)
			genSpan.End(ctxGen)
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// Validate response
		if resp.Completion == nil || len(resp.Completion.Choices) == 0 {
			genSpan.RecordError(ErrEmptyModelResponse)
			genSpan.End(ctxGen)
			return nil, ErrEmptyModelResponse
		}

		completion := resp.Completion

		genSpan.SetAttributes(map[string]any{
			"request":  req,
			"response": completion,
			"usage": &schema.Usage{
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				TotalTokens:      resp.Usage.TotalTokens,
			},
		})
		genSpan.End(ctxGen)

		if resp.Usage.PromptTokens > 0 || resp.Usage.CompletionTokens > 0 || resp.Usage.TotalTokens > 0 {
			usage.Add(Usage{
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				TotalTokens:      resp.Usage.TotalTokens,
			})
		}

		message := completion.Choices[0].Message
		runner.TruncateToolCallIDs(&message)
		history = append(history, message.ToParam())

		step := Step{
			AgentName:  currentAgent.Name,
			StepNumber: turnCount,
			Duration:   time.Since(stepStart),
		}

		if len(message.ToolCalls) == 0 {
			lastMessage = message
			steps = append(steps, step)
			break
		}

		approvalErr := r.checkToolApprovals(
			message.ToolCalls, currentAgent, contextParams, approvalHandler,
			history, turnCount, config,
		)
		if approvalErr != nil {
			return nil, approvalErr
		}

		toolMessages, recordedToolCalls, nextAgent := r.handleToolCalls(
			ctx,
			message.ToolCalls,
			toolMap,
			contextParams,
			currentAgent,
			resolveParallelToolCalls(currentAgent, config),
			config.MaxToolConcurrency,
		)

		step.ToolCalls = recordedToolCalls
		history = append(history, toolMessages...)

		if nextAgent != nil {
			currentAgent = nextAgent
		}

		step.Duration = time.Since(stepStart)
		steps = append(steps, step)
	}

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
	tool *tools.Tool
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

	// Debug block removed

	requestConfig := &runner.RequestConfig{
		Model:              agent.Model,
		Temperature:        merger.GetTemperature(),
		MaxTokens:          merger.GetMaxTokens(),
		ParallelToolCalls:  merger.GetParallelToolCallsPtr(),
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
	ctx context.Context,
	toolCalls []openai.ChatCompletionMessageToolCallUnion,
	toolMap runner.ToolMap,
	contextParams ContextVariables,
	currentAgent *Agent,
	parallel bool,
	maxConcurrency int,
) ([]openai.ChatCompletionMessageParamUnion, []ToolCall, *Agent) {
	// Use the internal tool handler
	messages, results, nextAgentAny := runner.HandleToolCalls(
		ctx,
		toolCalls,
		toolMap,
		contextParams,
		r.isHandoffFunc,
		parallel,
		maxConcurrency,
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
			// Emit handoff span
			if !tracing.IsTracingDisabled() {
				fromName := ""
				if currentAgent != nil {
					fromName = currentAgent.Name
				}
				ctx2, sp, _ := tracing.StartHandoffSpan(ctx, fromName, nextAgent.Name, "")
				sp.End(ctx2)
			}
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
		gr := gr
		guardrails[i] = &runner.Guardrail{
			Name: gr.Name,
			Func: func(ctx context.Context, text string) (runner.GuardrailResult, error) {
				// Span per guardrail execution (Python parity).
				ctx2, sp, _ := tracing.StartGuardrailSpan(ctx, gr.Name, "input")
				result, err := gr.Func(ctx2, text)
				if err != nil {
					sp.RecordError(err)
				}
				tripped := result.TripwireTriggered
				sp.SetAttributes(map[string]any{
					"tripwire_triggered": tripped,
					"message":            result.Message,
				})
				sp.End(ctx2)
				return runner.GuardrailResult{
					TripwireTriggered: result.TripwireTriggered,
					Message:           result.Message,
				}, err
			},
		}
	}

	executor := runner.NewGuardrailExecutor(guardrails, "input")
	userInput := extractUserInput(messages[len(messages)-1])

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
		gr := gr
		guardrails[i] = &runner.Guardrail{
			Name: gr.Name,
			Func: func(ctx context.Context, text string) (runner.GuardrailResult, error) {
				// Span per guardrail execution (Python parity).
				ctx2, sp, _ := tracing.StartGuardrailSpan(ctx, gr.Name, "output")
				result, err := gr.Func(ctx2, text)
				if err != nil {
					sp.RecordError(err)
				}
				tripped := result.TripwireTriggered
				sp.SetAttributes(map[string]any{
					"tripwire_triggered": tripped,
					"message":            result.Message,
				})
				sp.End(ctx2)
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

// resolveParallelToolCalls determines the effective parallel tool calls setting.
// Config overrides agent setting when explicitly set.
func resolveParallelToolCalls(agent *Agent, config *RunConfig) bool {
	if config != nil && config.ParallelToolCalls != nil {
		return *config.ParallelToolCalls
	}
	return agent.ParallelToolCalls
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

// AsyncResult represents the outcome of an asynchronous run
type AsyncResult struct {
	Result *Result
	Error  error
}

// RunAsync executes the agent concurrently in a goroutine.
// It returns a channel that will receive exactly one AsyncResult when execution completes.
//
// This enables non-blocking execution patterns, allowing the caller to continue
// other work while the agent runs, or to manage multiple running agents simultaneously.
//
// Example:
//
//	resultChan := runner.RunAsync(ctx, agent, messages)
//	// Do other work...
//	params := <-resultChan
//	if params.Error != nil {
//	    // handle error
//	}
func (r *Runner) RunAsync(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	opts ...RunOption,
) <-chan AsyncResult {
	ch := make(chan AsyncResult, 1)
	go func() {
		defer close(ch)
		res, err := r.Run(ctx, agent, messages, opts...)
		ch <- AsyncResult{Result: res, Error: err}
	}()
	return ch
}

func extractUserInput(msg openai.ChatCompletionMessageParamUnion) string {
	if p := msg.OfUser; p != nil {
		if s := p.Content.OfString; !param.IsOmitted(s) {
			return s.Value
		}
		var parts []string
		for _, part := range p.Content.OfArrayOfContentParts {
			if text := part.OfText; text != nil {
				parts = append(parts, text.Text)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, " ")
		}
	}
	if p := msg.OfSystem; p != nil {
		if s := p.Content.OfString; !param.IsOmitted(s) {
			return s.Value
		}
	}
	if p := msg.OfAssistant; p != nil {
		if s := p.Content.OfString; !param.IsOmitted(s) {
			return s.Value
		}
	}
	return fmt.Sprintf("%v", msg)
}
