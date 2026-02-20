package processor

import (
	"context"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"

	"github.com/MitulShah1/openai-agents-go/tracing/exporter"
	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// BatchOption configures a BatchProcessor.
type BatchOption func(*BatchProcessor)

// WithMaxBatchSize sets the maximum number of items per batch.
func WithMaxBatchSize(n int) BatchOption {
	return func(p *BatchProcessor) {
		if n > 0 {
			p.maxBatchSize = n
		}
	}
}

// WithFlushInterval sets how often to flush batches.
func WithFlushInterval(d time.Duration) BatchOption {
	return func(p *BatchProcessor) {
		if d > 0 {
			p.flushInterval = d
		}
	}
}

// WithMaxQueueSize sets the maximum queue size before dropping items.
func WithMaxQueueSize(n int) BatchOption {
	return func(p *BatchProcessor) {
		if n > 0 {
			p.maxQueueSize = n
		}
	}
}

// batchItem represents a single trace or span to be batched.
type batchItem struct {
	apiKey string
	trace  *schema.TraceExport
	span   *schema.SpanExport
}

// BatchProcessor batches spans/traces and exports them periodically.
type BatchProcessor struct {
	exporter exporter.Exporter

	maxBatchSize  int
	maxQueueSize  int
	flushInterval time.Duration

	itemCh chan batchItem
	stopCh chan struct{}
	wg     sync.WaitGroup

	mu    sync.Mutex
	byKey map[string]*keyQueue

	batchPool sync.Pool
	shutdown  sync.Once
}

type keyQueue struct {
	traces []schema.TraceExport
	spans  []schema.SpanExport
}

// NewBatch creates a new batch processor.
func NewBatch(exporter exporter.Exporter, opts ...BatchOption) *BatchProcessor {
	p := &BatchProcessor{
		exporter:      exporter,
		maxBatchSize:  100,
		maxQueueSize:  10_000,
		flushInterval: 2 * time.Second,
		byKey:         make(map[string]*keyQueue),
		stopCh:        make(chan struct{}),
	}

	for _, opt := range opts {
		opt(p)
	}

	p.itemCh = make(chan batchItem, p.maxQueueSize)

	p.batchPool.New = func() any {
		return &batch{
			traces: make([]schema.TraceExport, 0, p.maxBatchSize),
			spans:  make([]schema.SpanExport, 0, p.maxBatchSize),
		}
	}

	p.wg.Add(2)
	go p.batchWorker()
	go p.flushWorker()

	return p
}

// OnTraceStart is called when a trace starts.
func (p *BatchProcessor) OnTraceStart(trace schema.TraceExport, apiKey string) {
	p.enqueue(batchItem{apiKey: apiKey, trace: &trace})
}

// OnSpanEnd is called when a span ends.
func (p *BatchProcessor) OnSpanEnd(span schema.SpanExport, apiKey string) {
	p.enqueue(batchItem{apiKey: apiKey, span: &span})
}

// OnTraceEnd is called when a trace ends.
func (p *BatchProcessor) OnTraceEnd(trace schema.TraceExport, apiKey string) {
	p.enqueue(batchItem{apiKey: apiKey, trace: &trace})
}

// Shutdown shuts down the processor.
func (p *BatchProcessor) Shutdown(_ context.Context) error {
	p.shutdown.Do(func() {
		close(p.stopCh)
		p.wg.Wait()
		if p.exporter != nil {
			_ = p.exporter.Close()
		}
	})
	return nil
}

func (p *BatchProcessor) enqueue(item batchItem) {
	select {
	case p.itemCh <- item:
	case <-p.stopCh:
		return
	default:
		// Drop item (log if metrics available)
	}
}

func (p *BatchProcessor) batchWorker() {
	defer p.wg.Done()
	for {
		select {
		case item := <-p.itemCh:
			p.addToBatch(item)
		case <-p.stopCh:
			p.drainQueue()
			p.flushAll(context.Background())
			return
		}
	}
}

func (p *BatchProcessor) flushWorker() {
	defer p.wg.Done()
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.flushAll(context.Background())
		case <-p.stopCh:
			return
		}
	}
}

func (p *BatchProcessor) addToBatch(item batchItem) {
	p.mu.Lock()

	q := p.byKey[item.apiKey]
	if q == nil {
		q = &keyQueue{
			traces: make([]schema.TraceExport, 0, p.maxBatchSize),
			spans:  make([]schema.SpanExport, 0, p.maxBatchSize),
		}
		p.byKey[item.apiKey] = q
	}

	if item.trace != nil {
		q.traces = append(q.traces, *item.trace)
	}
	if item.span != nil {
		q.spans = append(q.spans, *item.span)
	}

	// Build a batch to export while holding the lock, then export outside the lock.
	// This avoids the fragile unlock-inside-lock pattern.
	var toFlush *batch
	if len(q.traces)+len(q.spans) >= p.maxBatchSize {
		toFlush = p.createBatch(item.apiKey, q)
	}
	p.mu.Unlock()

	if toFlush != nil {
		p.exportBatch(context.Background(), toFlush)
	}
}

func (p *BatchProcessor) drainQueue() {
	for {
		select {
		case item := <-p.itemCh:
			p.addToBatch(item)
		default:
			return
		}
	}
}

type batch struct {
	apiKey string
	traces []schema.TraceExport
	spans  []schema.SpanExport
}

func (p *BatchProcessor) flushAll(ctx context.Context) {
	p.mu.Lock()
	batches := make([]*batch, 0, len(p.byKey))
	for apiKey, q := range p.byKey {
		if len(q.traces) == 0 && len(q.spans) == 0 {
			continue
		}
		b := p.createBatch(apiKey, q)
		batches = append(batches, b)
	}
	p.mu.Unlock()

	for _, b := range batches {
		p.exportBatch(ctx, b)
	}
}

func (p *BatchProcessor) createBatch(apiKey string, q *keyQueue) *batch {
	b := p.batchPool.Get().(*batch)
	b.apiKey = apiKey

	// Move data
	remaining := p.maxBatchSize

	// Tricky: we prioritize spans or traces? Doesn't matter much.
	// Let's just take what fits.

	// Take spans first
	nSpans := min(len(q.spans), remaining)
	if nSpans > 0 {
		b.spans = append(b.spans[:0], q.spans[:nSpans]...)
		q.spans = q.spans[nSpans:]
		remaining -= nSpans
	} else {
		b.spans = b.spans[:0]
	}

	// Take traces
	nTraces := min(len(q.traces), remaining)
	if nTraces > 0 {
		b.traces = append(b.traces[:0], q.traces[:nTraces]...)
		q.traces = q.traces[nTraces:]
	} else {
		b.traces = b.traces[:0]
	}

	return b
}

func (p *BatchProcessor) exportBatch(ctx context.Context, b *batch) {
	defer p.batchPool.Put(b)

	if p.exporter == nil {
		return
	}

	// Retry logic
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 200 * time.Millisecond
	bo.MaxInterval = 5 * time.Second

	_, _ = backoff.Retry(ctx, func() (struct{}, error) {
		err := p.exporter.Export(ctx, b.apiKey, b.traces, b.spans)
		if err == nil {
			return struct{}{}, nil
		}
		// If permanent error, don't retry.
		// For now assume all are retryable except if context done.
		return struct{}{}, err
	}, backoff.WithBackOff(bo))
}
