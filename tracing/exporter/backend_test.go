package exporter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MitulShah1/openai-agents-go/tracing/schema"
)

func TestNewBackendExporter(t *testing.T) {
	exporter := NewBackendExporter()

	if exporter == nil {
		t.Fatal("NewBackendExporter returned nil")
	}

	if exporter.Endpoint != DefaultIngestEndpoint {
		t.Errorf("Expected default endpoint %q, got %q", DefaultIngestEndpoint, exporter.Endpoint)
	}

	if exporter.Client == nil {
		t.Error("HTTP client is nil")
	}
}

func TestNewBackendExporterWithOptions(t *testing.T) {
	customEndpoint := "https://custom.endpoint.com/traces"
	customOrg := "org-123"
	customProject := "proj-456"

	exporter := NewBackendExporter(
		WithEndpoint(customEndpoint),
		WithOrganization(customOrg),
		WithProject(customProject),
	)

	if exporter.Endpoint != customEndpoint {
		t.Errorf("Expected endpoint %q, got %q", customEndpoint, exporter.Endpoint)
	}

	if exporter.Organization != customOrg {
		t.Errorf("Expected organization %q, got %q", customOrg, exporter.Organization)
	}

	if exporter.Project != customProject {
		t.Errorf("Expected project %q, got %q", customProject, exporter.Project)
	}
}

func TestBackendExporterExportSuccess(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type: application/json")
		}

		if r.Header.Get("OpenAI-Beta") != "traces=v1" {
			t.Error("Expected OpenAI-Beta header")
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-key" {
			t.Errorf("Expected Authorization header with test-key, got %q", authHeader)
		}

		// Return success
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exporter := NewBackendExporter(
		WithEndpoint(server.URL),
	)

	ctx := context.Background()
	traces := []schema.TraceExport{
		{
			Object:       "trace",
			ID:           "trace_123",
			WorkflowName: "test",
		},
	}
	spans := []schema.SpanExport{
		{
			Object:  "trace.span",
			ID:      "span_123",
			TraceID: "trace_123",
		},
	}

	err := exporter.Export(ctx, "test-key", traces, spans)
	if err != nil {
		t.Errorf("Export failed: %v", err)
	}
}

func TestBackendExporterExportWithOrgAndProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("OpenAI-Organization") != "org-123" {
			t.Error("Expected OpenAI-Organization header")
		}
		if r.Header.Get("OpenAI-Project") != "proj-456" {
			t.Error("Expected OpenAI-Project header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exporter := NewBackendExporter(
		WithEndpoint(server.URL),
		WithOrganization("org-123"),
		WithProject("proj-456"),
	)

	ctx := context.Background()
	err := exporter.Export(ctx, "test-key", []schema.TraceExport{}, []schema.SpanExport{})
	if err != nil {
		t.Errorf("Export failed: %v", err)
	}
}

func TestBackendExporterExportNoAPIKey(t *testing.T) {
	exporter := NewBackendExporter()

	ctx := context.Background()
	err := exporter.Export(ctx, "", []schema.TraceExport{}, []schema.SpanExport{})

	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestBackendExporterExportRetryableError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		// Return 429 (rate limit) - retryable
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	exporter := NewBackendExporter(
		WithEndpoint(server.URL),
	)

	ctx := context.Background()
	err := exporter.Export(ctx, "test-key", []schema.TraceExport{}, []schema.SpanExport{})

	if err == nil {
		t.Error("Expected error for 429 response")
	}

	exportErr, ok := err.(*ExportError)
	if !ok {
		t.Fatalf("Expected ExportError, got %T", err)
	}

	if !exportErr.Retryable {
		t.Error("Expected error to be retryable")
	}

	if exportErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status code 429, got %d", exportErr.StatusCode)
	}
}

func TestBackendExporterExportNonRetryableError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return 400 (bad request) - not retryable
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	exporter := NewBackendExporter(
		WithEndpoint(server.URL),
	)

	ctx := context.Background()
	err := exporter.Export(ctx, "test-key", []schema.TraceExport{}, []schema.SpanExport{})

	if err == nil {
		t.Error("Expected error for 400 response")
	}

	exportErr, ok := err.(*ExportError)
	if !ok {
		t.Fatalf("Expected ExportError, got %T", err)
	}

	if exportErr.Retryable {
		t.Error("Expected error to be non-retryable")
	}
}

func TestBackendExporterClose(t *testing.T) {
	exporter := NewBackendExporter()

	err := exporter.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Should be idempotent
	err = exporter.Close()
	if err != nil {
		t.Errorf("Second Close failed: %v", err)
	}
}

func TestIsRetryableStatusCode(t *testing.T) {
	tests := []struct {
		code      int
		retryable bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, true}, // Rate limit
		{500, true}, // Server error
		{502, true}, // Bad gateway
		{503, true}, // Service unavailable
		{504, true}, // Gateway timeout
	}

	for _, tt := range tests {
		result := isRetryableStatusCode(tt.code)
		if result != tt.retryable {
			t.Errorf("isRetryableStatusCode(%d) = %v, want %v", tt.code, result, tt.retryable)
		}
	}
}

func TestEstimatePayloadSize(t *testing.T) {
	traces := make([]schema.TraceExport, 10)
	spans := make([]schema.SpanExport, 20)

	size := estimatePayloadSize(traces, spans)

	// Should estimate: 10*500 + 20*1024 = 5000 + 20480 = 25480
	expected := 25480
	if size != expected {
		t.Errorf("Expected payload size %d, got %d", expected, size)
	}
}

func TestBackendExporterPayloadUsesDataKey(t *testing.T) {
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exporter := NewBackendExporter(WithEndpoint(server.URL))

	traces := []schema.TraceExport{
		{Object: "trace", ID: "trace_1", WorkflowName: "wf"},
	}
	spans := []schema.SpanExport{
		{Object: "trace.span", ID: "span_1", TraceID: "trace_1"},
	}

	if err := exporter.Export(context.Background(), "test-key", traces, spans); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(capturedBody, &payload); err != nil {
		t.Fatalf("Failed to parse request body: %v", err)
	}

	if _, ok := payload["data"]; !ok {
		t.Error("Expected top-level 'data' key in ingest payload, but it was missing")
	}
	if _, ok := payload["traces"]; ok {
		t.Error("Unexpected 'traces' key in ingest payload; data should be in 'data'")
	}
	if _, ok := payload["spans"]; ok {
		t.Error("Unexpected 'spans' key in ingest payload; data should be in 'data'")
	}

	var items []json.RawMessage
	if err := json.Unmarshal(payload["data"], &items); err != nil {
		t.Fatalf("Failed to parse 'data' array: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items in 'data', got %d", len(items))
	}
}
