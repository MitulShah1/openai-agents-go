package agents

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/internal/runner"
	"github.com/MitulShah1/openai-agents-go/models"
	"github.com/MitulShah1/openai-agents-go/session"
	"github.com/MitulShah1/openai-agents-go/tools"
	"github.com/MitulShah1/openai-agents-go/tracing"
)

// Stream executes the agent loop with streaming responses.
// It returns a channel that receives events (tokens, tool calls, errors, result).
// The channel is closed when execution is complete.
func (r *Runner) Stream(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	opts ...RunOption,
) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 100)

	// Apply options
	options := &runOptions{}
	for _, opt := range opts {
		opt(options)
	}

	go func() {
		defer close(ch)

		// Call the internal streaming implementation
		result, err := r.executeStream(
			ctx,
			agent,
			messages,
			options.contextParams,
			options.config,
			options.sess,
			options.sessionID,
			options.approvalHandler,
			ch,
		)
		if err != nil {
			ch <- StreamEvent{Type: StreamEventError, Error: err}
		} else {
			ch <- StreamEvent{Type: StreamEventResult, Result: result}
		}
	}()

	return ch, nil
}

// executeStream is the internal implementation of Stream.
func (r *Runner) executeStream(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	contextParams ContextVariables,
	config *RunConfig,
	sess session.Session,
	sessionID string,
	approvalHandler tools.ApprovalHandler,
	ch chan<- StreamEvent,
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

	// Create executor
	executor := runner.NewExecutor(config.MaxTurns, config.Timeout)

	// Apply timeout
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

	// Execute main agent loop with streaming
	result, err := r.executeAgentLoopStream(ctx, agent, messages, contextParams, config, executor, approvalHandler, ch)
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

// executeAgentLoopStream runs the main agent execution loop with streaming
//
//nolint:gocyclo
func (r *Runner) executeAgentLoopStream(
	ctx context.Context,
	agent *Agent,
	messages []openai.ChatCompletionMessageParamUnion,
	contextParams ContextVariables,
	config *RunConfig,
	executor *runner.Executor,
	approvalHandler tools.ApprovalHandler,
	ch chan<- StreamEvent,
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

		// Start generation span (redaction handled automatically)
		ctxGen, genSpan, _ := tracing.StartGenerationSpan(ctx, tracing.WithModel(currentAgent.Model))

		// Enable streaming
		stream, err := model.StreamResponse(ctxGen, req, models.ModelSettings{Prompt: resolvedPrompt})
		if err != nil {
			genSpan.RecordError(err)
			genSpan.End(ctxGen)
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, fmt.Errorf("stream creation failed: %w", err)
		}

		var accumulatedContent strings.Builder
		var toolCalls []openai.ChatCompletionChunkChoiceDeltaToolCall

		// Process stream
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta

				// Emit content token
				if delta.Content != "" {
					ch <- StreamEvent{
						Type:    StreamEventToken,
						Content: delta.Content,
					}
					accumulatedContent.WriteString(delta.Content)
				}

				// Accumulate tool calls
				if len(delta.ToolCalls) > 0 {
					toolCalls = append(toolCalls, delta.ToolCalls...)
				}
			}
		}

		if err := stream.Err(); err != nil {
			genSpan.RecordError(err)
			genSpan.End(ctxGen)
			if agentSpan != nil {
				agentSpan.End(ctx)
			}
			return nil, fmt.Errorf("stream failed: %w", err)
		}

		_ = stream.Close()

		// Construct message from accumulated content
		message := openai.ChatCompletionMessage{
			Role:    "assistant",
			Content: accumulatedContent.String(),
		}

		// Reconstruct tool calls
		if len(toolCalls) > 0 {
			finalToolCalls := r.reconstructToolCalls(toolCalls)
			message.ToolCalls = finalToolCalls
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
				ch <- StreamEvent{
					Type:             StreamEventApprovalRequired,
					ApprovalRequests: approvalReqErr.Requests,
				}
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
			tcCopy := tc // Create copy to safely take address
			ch <- StreamEvent{
				Type:     StreamEventToolCall,
				ToolCall: &tcCopy,
			}
		}

		if nextAgent != nil {
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

	finalOutput := extractFinalOutput(lastMessage)

	return &Result{
		Messages:    history,
		Agent:       currentAgent,
		Usage:       usage,
		Steps:       steps,
		FinalOutput: finalOutput,
	}, nil
}

func (r *Runner) reconstructToolCalls(deltas []openai.ChatCompletionChunkChoiceDeltaToolCall) []openai.ChatCompletionMessageToolCallUnion {
	type tcBuilder struct {
		Index int64
		ID    string
		Type  string
		Name  string
		Args  string
	}

	builders := make(map[int64]*tcBuilder)

	for _, delta := range deltas {
		idx := delta.Index
		b, exists := builders[idx]
		if !exists {
			b = &tcBuilder{Index: idx}
			builders[idx] = b
		}

		if delta.ID != "" {
			b.ID = delta.ID
		}
		if delta.Type != "" {
			b.Type = delta.Type
		}
		if delta.Function.Name != "" {
			b.Name += delta.Function.Name
		}
		if delta.Function.Arguments != "" {
			b.Args += delta.Function.Arguments
		}
	}

	result := make([]openai.ChatCompletionMessageToolCallUnion, 0, len(builders))

	// Iterate through map and find max index
	var maxIdx int64 = -1
	for idx := range builders {
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	for i := int64(0); i <= maxIdx; i++ {
		if b, ok := builders[i]; ok {
			result = append(result, openai.ChatCompletionMessageToolCallUnion{
				ID:   b.ID,
				Type: b.Type,
				Function: openai.ChatCompletionMessageFunctionToolCallFunction{
					Name:      b.Name,
					Arguments: b.Args,
				},
			})
		}
	}
	return result
}
