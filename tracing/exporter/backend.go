package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

// DefaultIngestEndpoint is the default endpoint for ingesting traces.
const DefaultIngestEndpoint = "https://api.openai.com/v1/traces/ingest"

// BackendExporter exports traces/spans to the OpenAI ingest endpoint.
type BackendExporter struct {
	Endpoint string
	Client   *http.Client

	// Optional env-driven headers
	Organization string
	Project      string

	closeOnce sync.Once
}

// Option configures a BackendExporter.
type Option func(*BackendExporter)

// NewBackendExporter creates a new exporter with optimized HTTP client configuration.
func NewBackendExporter(opts ...Option) *BackendExporter {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
	}

	e := &BackendExporter{
		Endpoint: DefaultIngestEndpoint,
		Client: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		},
		Organization: os.Getenv("OPENAI_ORG_ID"),
		Project:      os.Getenv("OPENAI_PROJECT_ID"),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithEndpoint sets a custom ingest endpoint.
func WithEndpoint(endpoint string) Option {
	return func(e *BackendExporter) {
		e.Endpoint = endpoint
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(e *BackendExporter) {
		e.Client = client
	}
}

// WithOrganization sets the OpenAI organization header.
func WithOrganization(org string) Option {
	return func(e *BackendExporter) {
		e.Organization = org
	}
}

// WithProject sets the OpenAI project header.
func WithProject(project string) Option {
	return func(e *BackendExporter) {
		e.Project = project
	}
}

// Export sends traces and spans to the OpenAI ingest endpoint.
func (e *BackendExporter) Export(ctx context.Context, apiKey string, traces []schema.TraceExport, spans []schema.SpanExport) error {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("tracing export requires api key")
	}

	data := make([]any, len(traces)+len(spans))
	for i, t := range traces {
		data[i] = t
	}
	for i, s := range spans {
		data[len(traces)+i] = s
	}
	payload := schema.IngestPayload{Data: data}

	// Use streaming encoder for better performance
	var buf bytes.Buffer
	buf.Grow(estimatePayloadSize(traces, spans))
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return &ExportError{
			Err:       fmt.Errorf("encode payload: %w", err),
			Retryable: false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.Endpoint, &buf)
	if err != nil {
		return &ExportError{
			Err:       fmt.Errorf("create request: %w", err),
			Retryable: false,
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("OpenAI-Beta", "traces=v1")
	if e.Organization != "" {
		req.Header.Set("OpenAI-Organization", e.Organization)
	}
	if e.Project != "" {
		req.Header.Set("OpenAI-Project", e.Project)
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		// Determine if network error is retryable
		retryable := isRetryableNetworkError(err)
		return &ExportError{
			Err:       fmt.Errorf("http request: %w", err),
			Retryable: retryable,
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
	return &ExportError{
		StatusCode: resp.StatusCode,
		Body:       string(b),
		Retryable:  isRetryableStatusCode(resp.StatusCode),
	}
}

// Close closes idle connections in the HTTP client.
func (e *BackendExporter) Close() error {
	var err error
	e.closeOnce.Do(func() {
		if transport, ok := e.Client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	})
	return err
}

// estimatePayloadSize estimates the JSON payload size for buffer pre-allocation.
func estimatePayloadSize(traces []schema.TraceExport, spans []schema.SpanExport) int {
	// Rough estimate: 500 bytes per trace, 1KB per span
	return len(traces)*500 + len(spans)*1024
}

// isRetryableStatusCode returns true for HTTP status codes that should be retried.
func isRetryableStatusCode(statusCode int) bool {
	// Retry on rate limits (429) and server errors (5xx)
	if statusCode == 429 {
		return true
	}
	return statusCode >= 500 && statusCode <= 599
}

// isRetryableNetworkError returns true for network errors that should be retried.
func isRetryableNetworkError(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}
	return true // Default to retryable for unknown network errors
}
