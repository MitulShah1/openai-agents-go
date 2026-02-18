// Package main demonstrates multi-agent handoffs with a shared trace context.
//
// When multiple agents collaborate on a single user request the entire workflow
// should appear as one trace in the observability dashboard, not as several
// disconnected traces. This example shows how to start a single trace before the
// first Run call and carry that trace through every subsequent Run so that every
// agent span, generation span, and handoff span are grouped under the same
// trace ID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/handoff"
	"github.com/MitulShah1/openai-agents-go/tools"
	"github.com/MitulShah1/openai-agents-go/tracing"
	"github.com/MitulShah1/openai-agents-go/tracing/exporter"
	"github.com/MitulShah1/openai-agents-go/tracing/processor"
)

// ─── Agent definitions ──────────────────────────────────────────────────────

func buildAgents() (triage, billing, technical *agents.Agent) {
	triage = agents.NewAgent("Triage")
	triage.Instructions = "You are a triage agent. Route the user to billing support for payment or invoice questions, or to technical support for product / API questions."

	billing = agents.NewAgent("Billing")
	billing.Instructions = "You are a billing support agent. Help customers with invoices, payments, and subscription questions."

	technical = agents.NewAgent("Technical")
	technical.Instructions = "You are a technical support agent. Help customers with API usage, SDK errors, and integration questions."

	// Triage can hand off to either specialist.
	triage.Tools = []tools.Tool{
		handoff.New(billing, handoff.WithDescription("Transfer to billing support")).ToTool(),
		handoff.New(technical, handoff.WithDescription("Transfer to technical support")).ToTool(),
	}

	return triage, billing, technical
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// ── 1. Set up tracing ────────────────────────────────────────────────────
	//
	// Use the console exporter so every span is printed to stdout.
	// Replace this with a backend exporter to send spans to your observability
	// platform.
	exp := exporter.NewConsoleExporter()
	proc := processor.NewBatch(exp)
	provider := tracing.NewProvider(proc)
	tracing.SetProvider(provider)

	defer func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			log.Printf("tracing shutdown error: %v", err)
		}
	}()

	// ── 2. Build agents and runner ───────────────────────────────────────────
	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)
	triageAgent, _, _ := buildAgents()

	// ── 3. Start ONE shared trace for the whole workflow ─────────────────────
	//
	// By calling provider.StartTrace before runner.Run we place a Trace in the
	// context.  The runner detects the existing trace (tracing.FromContext != nil)
	// and skips creating a new one.  Every agent run, generation span, and
	// handoff span that the runner emits will therefore share this single
	// trace ID.
	ctx := context.Background()
	ctx, sharedTrace, err := provider.StartTrace(ctx,
		tracing.WithWorkflowName("customer-support-workflow"),
		tracing.WithGroupID("group_000000000000000000000000"), // fixed group for demo
		tracing.WithMetadata(map[string]any{
			"user_id":   "user-42",
			"channel":   "web-chat",
			"demo_mode": true,
		}),
	)
	if err != nil {
		log.Printf("failed to start trace (non-fatal): %v", err)
	}
	defer sharedTrace.End(ctx)

	fmt.Printf("=== Shared trace ID: %s ===\n\n", sharedTrace.ID())

	// ── 4. Optional: add a custom "workflow" span that wraps everything ───────
	//
	// Any custom span started from the trace-bearing context is automatically
	// nested under the shared trace.
	ctx, workflowSpan, _ := tracing.CustomSpan(ctx, "customer-support-session")
	defer workflowSpan.End(ctx)

	// ── 5. Example A – automatic handoff via agent tools ────────────────────
	//
	// The user message triggers triage → billing handoff.  Both the triage
	// generation span and the billing generation span will appear under the
	// same trace ID because the context carries the shared trace.
	fmt.Println("=== Example A: Automatic handoff (triage → billing) ===")

	messagesA := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("I was charged twice for my invoice #INV-2024-001. Can you help?"),
	}

	resultA, err := runner.Run(ctx, triageAgent, messagesA)
	if err != nil {
		log.Fatalf("run A failed: %v", err)
	}
	fmt.Printf("Final agent : %s\n", resultA.Agent.Name)
	fmt.Printf("Response   : %s\n\n", resultA.FinalOutput)

	// ── 6. Example B – second turn, still under the same trace ──────────────
	//
	// A follow-up question from the same customer continues within the shared
	// trace context so both conversation turns are correlated.
	fmt.Println("=== Example B: Second turn (triage → technical) ===")

	messagesB := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("I'm also getting a 429 rate-limit error when calling the API. What should I do?"),
	}

	resultB, err := runner.Run(ctx, triageAgent, messagesB)
	if err != nil {
		log.Fatalf("run B failed: %v", err)
	}
	fmt.Printf("Final agent : %s\n", resultB.Agent.Name)
	fmt.Printf("Response   : %s\n\n", resultB.FinalOutput)

	// ── 7. Attach metadata to the shared trace after the runs ────────────────
	//
	// Use a custom span to record summary information for the whole workflow.
	_, summarySpan, _ := tracing.CustomSpan(ctx, "workflow-summary")
	summarySpan.SetAttributes(map[string]any{
		"turns_run":     2,
		"final_agent_a": resultA.Agent.Name,
		"final_agent_b": resultB.Agent.Name,
		"total_tokens":  resultA.Usage.TotalTokens + resultB.Usage.TotalTokens,
	})
	summarySpan.End(ctx)

	fmt.Printf("=== All spans above share trace ID: %s ===\n", sharedTrace.ID())
}
