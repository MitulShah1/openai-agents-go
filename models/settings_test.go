package models

import (
	"testing"

	"github.com/MitulShah1/openai-agents-go/jsonschema"
)

func ptrFloat64(v float64) *float64 { return &v }
func ptrInt(v int) *int             { return &v }
func ptrBool(v bool) *bool          { return &v }

func TestModelSettings_Resolve_Empty(t *testing.T) {
	base := ModelSettings{}
	override := ModelSettings{}
	result := base.Resolve(override)

	if result.Temperature != nil {
		t.Errorf("expected nil Temperature, got %v", *result.Temperature)
	}
	if result.MaxTokens != nil {
		t.Errorf("expected nil MaxTokens, got %v", *result.MaxTokens)
	}
}

func TestModelSettings_Resolve_BaseOnly(t *testing.T) {
	base := ModelSettings{
		Temperature:       ptrFloat64(0.7),
		MaxTokens:         ptrInt(100),
		ParallelToolCalls: ptrBool(true),
	}
	override := ModelSettings{}
	result := base.Resolve(override)

	if result.Temperature == nil || *result.Temperature != 0.7 {
		t.Errorf("expected Temperature 0.7, got %v", result.Temperature)
	}
	if result.MaxTokens == nil || *result.MaxTokens != 100 {
		t.Errorf("expected MaxTokens 100, got %v", result.MaxTokens)
	}
	if result.ParallelToolCalls == nil || *result.ParallelToolCalls != true {
		t.Errorf("expected ParallelToolCalls true, got %v", result.ParallelToolCalls)
	}
}

func TestModelSettings_Resolve_OverrideTakesPrecedence(t *testing.T) {
	base := ModelSettings{
		Temperature: ptrFloat64(0.7),
		MaxTokens:   ptrInt(100),
		TopP:        ptrFloat64(0.9),
		Stop:        []string{"stop1"},
	}
	override := ModelSettings{
		Temperature: ptrFloat64(1.0),
		MaxTokens:   ptrInt(200),
		Stop:        []string{"stop2", "stop3"},
	}
	result := base.Resolve(override)

	if *result.Temperature != 1.0 {
		t.Errorf("expected Temperature 1.0, got %v", *result.Temperature)
	}
	if *result.MaxTokens != 200 {
		t.Errorf("expected MaxTokens 200, got %v", *result.MaxTokens)
	}
	// TopP should come from base (not overridden)
	if result.TopP == nil || *result.TopP != 0.9 {
		t.Errorf("expected TopP 0.9, got %v", result.TopP)
	}
	if len(result.Stop) != 2 || result.Stop[0] != "stop2" {
		t.Errorf("expected Stop [stop2 stop3], got %v", result.Stop)
	}
}

func TestModelSettings_Resolve_AllFields(t *testing.T) {
	base := ModelSettings{}
	override := ModelSettings{
		Temperature:       ptrFloat64(0.5),
		MaxTokens:         ptrInt(500),
		TopP:              ptrFloat64(0.95),
		FrequencyPenalty:  ptrFloat64(0.1),
		PresencePenalty:   ptrFloat64(0.2),
		Stop:              []string{"END"},
		ParallelToolCalls: ptrBool(false),
		ResponseFormat: &jsonschema.ResponseFormat{
			Type: "json_schema",
		},
	}
	result := base.Resolve(override)

	if *result.Temperature != 0.5 {
		t.Errorf("expected Temperature 0.5, got %v", *result.Temperature)
	}
	if *result.MaxTokens != 500 {
		t.Errorf("expected MaxTokens 500, got %v", *result.MaxTokens)
	}
	if *result.TopP != 0.95 {
		t.Errorf("expected TopP 0.95, got %v", *result.TopP)
	}
	if *result.FrequencyPenalty != 0.1 {
		t.Errorf("expected FrequencyPenalty 0.1, got %v", *result.FrequencyPenalty)
	}
	if *result.PresencePenalty != 0.2 {
		t.Errorf("expected PresencePenalty 0.2, got %v", *result.PresencePenalty)
	}
	if result.Stop[0] != "END" {
		t.Errorf("expected Stop [END], got %v", result.Stop)
	}
	if *result.ParallelToolCalls != false {
		t.Errorf("expected ParallelToolCalls false, got %v", *result.ParallelToolCalls)
	}
	if result.ResponseFormat == nil || result.ResponseFormat.Type != "json_schema" {
		t.Errorf("expected ResponseFormat json_schema, got %v", result.ResponseFormat)
	}
}
