package agents

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/tracing"
	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

type recordingProcessor struct {
	mu     sync.Mutex
	traces []schema.TraceExport
	spans  []schema.SpanExport

	done chan struct{}
}

func newRecordingProcessor() *recordingProcessor {
	return &recordingProcessor{done: make(chan struct{})}
}

func (p *recordingProcessor) OnTraceStart(_ schema.TraceExport, _ string) {
	// Traces are no longer emitted on start
}

func (p *recordingProcessor) OnSpanEnd(span schema.SpanExport, _ string) {
	p.mu.Lock()
	p.spans = append(p.spans, span)
	p.mu.Unlock()
}

func (p *recordingProcessor) OnTraceEnd(trace schema.TraceExport, _ string) {
	p.mu.Lock()
	p.traces = append(p.traces, trace)
	p.mu.Unlock()

	select {
	case <-p.done:
	default:
		close(p.done)
	}
}

func (p *recordingProcessor) Shutdown(context.Context) error { return nil }

func TestTracingSingleRunCreatesSingleTrace(t *testing.T) {
	proc := newRecordingProcessor()
	tracing.SetProvider(tracing.NewProvider(proc))

	r := NewRunner(&openai.Client{})
	a := NewAgent("TestAgent")
	msgs := []openai.ChatCompletionMessageParamUnion{openai.UserMessage("hi")}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // force early failure before any OpenAI call

	_, _ = r.Run(ctx, a, msgs)

	select {
	case <-proc.done:
	case <-time.After(250 * time.Millisecond):
		// Even on early failure, trace should end quickly.
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()
	if len(proc.traces) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(proc.traces))
	}
}

func TestTracingMultipleRunsCreateMultipleTraces(t *testing.T) {
	proc := newRecordingProcessor()
	tracing.SetProvider(tracing.NewProvider(proc))

	r := NewRunner(&openai.Client{})
	a := NewAgent("TestAgent")
	msgs := []openai.ChatCompletionMessageParamUnion{openai.UserMessage("hi")}

	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _ = r.Run(ctx, a, msgs)
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()
	if len(proc.traces) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(proc.traces))
	}
}

func TestTracingExplicitTraceWrapsMultipleRuns(t *testing.T) {
	proc := newRecordingProcessor()
	p := tracing.NewProvider(proc)
	tracing.SetProvider(p)

	base := context.Background()
	ctx, tr, _ := p.StartTrace(base, tracing.WithWorkflowName("explicit"))

	r := NewRunner(&openai.Client{})
	a := NewAgent("TestAgent")
	msgs := []openai.ChatCompletionMessageParamUnion{openai.UserMessage("hi")}

	for i := 0; i < 2; i++ {
		ctx2, cancel := context.WithCancel(ctx)
		cancel()
		_, _ = r.Run(ctx2, a, msgs)
	}

	tr.End(ctx)

	proc.mu.Lock()
	defer proc.mu.Unlock()
	if len(proc.traces) != 1 {
		t.Fatalf("expected 1 trace when explicitly wrapped, got %d", len(proc.traces))
	}
}

func TestTracingDisabledConfigDisablesEmission(t *testing.T) {
	proc := newRecordingProcessor()
	tracing.SetProvider(tracing.NewProvider(proc))

	r := NewRunner(&openai.Client{})
	a := NewAgent("TestAgent")
	msgs := []openai.ChatCompletionMessageParamUnion{openai.UserMessage("hi")}

	enabled := false
	cfg := &RunConfig{TraceEnabled: &enabled}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _ = r.Run(ctx, a, msgs, WithConfig(cfg))

	proc.mu.Lock()
	defer proc.mu.Unlock()
	if len(proc.traces) != 0 {
		t.Fatalf("expected 0 traces when disabled, got %d", len(proc.traces))
	}
}

func TestTracingStreamCreatesTraceEvenIfNotDrained(t *testing.T) {
	proc := newRecordingProcessor()
	tracing.SetProvider(tracing.NewProvider(proc))

	r := NewRunner(&openai.Client{})
	a := NewAgent("TestAgent")
	msgs := []openai.ChatCompletionMessageParamUnion{openai.UserMessage("hi")}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.Stream(ctx, a, msgs)
	if err != nil {
		t.Fatalf("expected Stream to return channel, got err: %v", err)
	}

	// Do not read from the channel; we only care that trace started/ended.
	select {
	case <-proc.done:
	case <-time.After(250 * time.Millisecond):
		// best effort: goroutine may be scheduled slightly later
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()
	if len(proc.traces) != 1 {
		t.Fatalf("expected 1 trace for Stream, got %d", len(proc.traces))
	}
}
