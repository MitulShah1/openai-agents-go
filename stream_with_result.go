package agents

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/internal/runner"
	"github.com/MitulShah1/openai-agents-go/models"
	"github.com/MitulShah1/openai-agents-go/session"
	"github.com/MitulShah1/openai-agents-go/stream"
	"github.com/MitulShah1/openai-agents-go/tools"
	"github.com/MitulShah1/openai-agents-go/tracing"
)

// StreamWithResult executes the agent loop with streaming and returns a Result object.
// This is the modern streaming API that provides full event streaming capabilities.
//
// The returned Result object provides an iterator for consuming events:
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
//	    // Handle event based on type
//	    switch e := event.(type) {
//	    case *stream.RawResponseEvent:
//	        // Handle raw LLM events
//	    case *stream.RunItemEvent:
//	        // Handle semantic events (tool calls, handoffs, etc.)
//	    case *stream.AgentUpdatedEvent:
//	        // Handle agent transitions
//	    }
//	}
func (r *Runner) StreamWithResult(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	opts ...RunOption,
) (*stream.Result, error) {
	// Validate input
	if len(messages) == 0 {
		return nil, ErrNoMessages
	}

	// Apply options
	options := &runOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Use default config if not provided
	config := options.config
	if config == nil {
		config = DefaultRunConfig()
	}

	// Create streaming result
	result := stream.NewResult(messages, agent, config.MaxTurns)

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	result.SetCancelFunc(cancel)

	// Start background execution
	go r.executeStreamWithResult(
		ctx,
		result,
		agent,
		messages,
		options.contextParams,
		config,
		options.sess,
		options.sessionID,
		options.approvalHandler,
	)

	return result, nil
}

// executeStreamWithResult is the internal implementation that runs in a goroutine
func (r *Runner) executeStreamWithResult(
	ctx context.Context,
	result *stream.Result,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	contextParams ContextVariables,
	config *RunConfig,
	sess session.Session,
	sessionID string,
	approvalHandler tools.ApprovalHandler,
) {
	defer result.Complete()

	// Ensure trace exists
	if tracing.FromContext(ctx) == nil && !tracing.IsTracingDisabled() {
		var tr tracing.Trace
		var err error
		ctx, tr, err = tracing.GetProvider().StartTrace(ctx, traceOptionsFromRunConfig(config)...)
		_ = err
		defer tr.End(ctx)
	}

	// Initialize context variables
	if contextParams == nil {
		contextParams = make(ContextVariables)
	}

	// Create executor
	executor := runner.NewExecutor(config.MaxTurns, config.Timeout)

	// Apply timeout
	var cancel context.CancelFunc
	ctx, cancel = executor.ApplyTimeout(ctx)
	defer cancel()

	// Execute OnBeforeRun hook
	if agent.OnBeforeRun != nil {
		if err := agent.OnBeforeRun(ctx, agent); err != nil {
			result.EmitError(fmt.Errorf("OnBeforeRun hook failed: %w", err))
			return
		}
	}

	// Run input guardrails
	if err := r.executeInputGuardrails(ctx, agent, messages); err != nil {
		result.EmitError(err)
		return
	}

	// Load session history
	sessionHandler := runner.NewSessionHandler(sess, sessionID)
	var err error
	newInputLen := len(messages)
	messages, err = sessionHandler.LoadHistory(ctx, messages)
	if err != nil {
		result.EmitError(err)
		return
	}
	sessionPrefixLen := len(messages) - newInputLen

	// Execute main agent loop with streaming
	finalResult, err := r.executeAgentLoopWithStreaming(
		ctx,
		result,
		agent,
		messages,
		contextParams,
		config,
		executor,
		approvalHandler,
	)
	if err != nil {
		result.EmitError(err)
		return
	}

	// Run output guardrails
	if err := r.executeOutputGuardrails(ctx, agent, finalResult.FinalOutput); err != nil {
		result.EmitError(err)
		return
	}

	// Save only new messages (skip the session prefix that was already persisted)
	if err := sessionHandler.SaveHistory(ctx, finalResult.Messages[sessionPrefixLen:]); err != nil {
		result.EmitError(err)
		return
	}

	// Execute OnAfterRun hook
	if agent.OnAfterRun != nil {
		if err := agent.OnAfterRun(ctx, agent); err != nil {
			result.EmitError(fmt.Errorf("OnAfterRun hook failed: %w", err))
			return
		}
	}

	// Set final output
	result.SetFinalOutput(finalResult.FinalOutput)
}

// executeAgentLoopWithStreaming runs the main agent loop with full streaming support
func (r *Runner) executeAgentLoopWithStreaming(
	ctx context.Context,
	result *stream.Result,
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

	for result.GetCancelMode() != stream.CancelAfterTurn || turnCount == 0 {

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
		result.IncrementTurn()

		// Create agent span
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

		// Prepare request
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

		// Start generation span
		ctxGen, genSpan, _ := tracing.StartGenerationSpan(ctx, tracing.WithModel(currentAgent.Model))

		// Create streaming handler
		streamHandler := stream.NewHandler()

		// Enable streaming
		streamObj, err := model.StreamResponse(ctxGen, req, models.ModelSettings{})
		if err != nil {
			genSpan.RecordError(err)
			genSpan.End(ctxGen)
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, fmt.Errorf("stream creation failed: %w", err)
		}

		// Process stream chunks
		for streamObj.Next() {
			chunk := streamObj.Current()

			// Process chunk through handler
			events := streamHandler.ProcessChunk(chunk)

			// Emit all generated events
			for _, event := range events {
				result.EmitEvent(event)
			}
		}

		if err := streamObj.Err(); err != nil {
			genSpan.RecordError(err)
			genSpan.End(ctxGen)
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, fmt.Errorf("stream failed: %w", err)
		}

		_ = streamObj.Close()

		// Finalize handler to get completion events
		finalEvents := streamHandler.Finalize()
		for _, event := range finalEvents {
			result.EmitEvent(event)
		}

		// Construct message from accumulated content
		message := openai.ChatCompletionMessage{
			Role:    "assistant",
			Content: streamHandler.GetAccumulatedText(),
		}

		// Add tool calls if any
		toolCalls := streamHandler.GetFunctionCalls()
		if len(toolCalls) > 0 {
			message.ToolCalls = toolCalls
		}

		// Truncate tool call IDs
		runner.TruncateToolCallIDs(&message)

		genSpan.SetAttributes(map[string]any{
			"request":  req,
			"response": message,
		})
		genSpan.End(ctxGen)

		history = append(history, message.ToParam())

		step := Step{
			AgentName:  currentAgent.Name,
			StepNumber: turnCount,
			Duration:   time.Since(stepStart),
		}

		if len(message.ToolCalls) == 0 {
			lastMessage = message
			steps = append(steps, step)

			// Emit message output created event
			result.EmitEvent(&stream.RunItemEvent{
				Name: string(stream.MessageOutputCreated),
				Item: &RunItem{
					Type:    "message",
					Content: message.Content,
					Agent:   currentAgent,
				},
				SequenceNumber: 0, // Will be set by handler
			})
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
			if approvalReqErr, ok := approvalErr.(*ToolApprovalRequiredError); ok {
				result.EmitEvent(&stream.ApprovalRequiredEvent{
					Requests:       approvalReqErr.Requests,
					SequenceNumber: 0,
				})
			}
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

		// Emit tool call events
		for _, tc := range recordedToolCalls {
			result.EmitEvent(&stream.RunItemEvent{
				Name: string(stream.ToolCalled),
				Item: &RunItem{
					Type: "tool_call",
					ToolCall: &ToolCall{
						ToolName:  tc.ToolName,
						Arguments: tc.Arguments,
						Result:    tc.Result,
						Error:     tc.Error,
						Duration:  tc.Duration,
					},
					Agent: currentAgent,
				},
				SequenceNumber: 0,
			})

			// Emit tool output event
			result.EmitEvent(&stream.RunItemEvent{
				Name: string(stream.ToolOutput),
				Item: &RunItem{
					Type: "tool_output",
					ToolOutput: &ToolOutput{
						ToolCallID: "", // Would need to track this
						Output:     fmt.Sprintf("%v", tc.Result),
						Error:      tc.Error,
					},
					Agent: currentAgent,
				},
				SequenceNumber: 0,
			})
		}

		if nextAgent != nil {
			if tracing.FromContext(ctx) != nil && nextAgent.Name != "" {
				ctx2, hs, _ := tracing.StartHandoffSpan(ctx, currentAgent.Name, nextAgent.Name, "")
				hs.End(ctx2)
			}

			// Emit handoff events
			result.EmitEvent(&stream.RunItemEvent{
				Name: string(stream.HandoffRequested),
				Item: &RunItem{
					Type:  "handoff",
					Agent: currentAgent,
				},
				SequenceNumber: 0,
			})

			result.EmitEvent(&stream.RunItemEvent{
				Name: string(stream.HandoffOccurred),
				Item: &RunItem{
					Type:  "handoff",
					Agent: nextAgent,
				},
				SequenceNumber: 0,
			})

			result.EmitEvent(&stream.AgentUpdatedEvent{
				NewAgent:       nextAgent,
				SequenceNumber: 0,
			})

			currentAgent = nextAgent
			result.UpdateAgent(nextAgent)
		}

		step.Duration = time.Since(stepStart)
		steps = append(steps, step)

		if agentSpan != nil {
			agentSpan.End(ctx)
		}
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
