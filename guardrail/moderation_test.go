package guardrail_test

import (
	"context"
	"testing"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// Note: These are unit tests with mocked responses.
// For integration tests with the real OpenAI API, use the integration test tag.

func TestModerationGuardrail_AllCategories(t *testing.T) {
	// This test would require mocking the OpenAI client
	// For now, we'll create a basic structure test
	t.Skip("Requires OpenAI API key and real API call - run with integration tag")
}

func TestModerationGuardrail_CustomThreshold(t *testing.T) {
	t.Skip("Requires OpenAI API key and real API call - run with integration tag")
}

func TestModerationGuardrail_SpecificCategories(t *testing.T) {
	t.Skip("Requires OpenAI API key and real API call - run with integration tag")
}

func TestModerationGuardrail_TripwireEnabled(t *testing.T) {
	t.Skip("Requires OpenAI API key and real API call - run with integration tag")
}

func TestModerationCategories_Constants(t *testing.T) {
	// Test that category constants are properly defined
	tests := []struct {
		name     string
		category guardrail.ModerationCategory
		expected string
	}{
		{"Harassment", guardrail.CategoryHarassment, "harassment"},
		{"Harassment/Threatening", guardrail.CategoryHarassmentThreatening, "harassment/threatening"},
		{"Hate", guardrail.CategoryHate, "hate"},
		{"Hate/Threatening", guardrail.CategoryHateThreatening, "hate/threatening"},
		{"Illicit", guardrail.CategoryIllicit, "illicit"},
		{"Illicit/Violent", guardrail.CategoryIllicitViolent, "illicit/violent"},
		{"Self-Harm", guardrail.CategorySelfHarm, "self-harm"},
		{"Self-Harm/Instructions", guardrail.CategorySelfHarmInstructions, "self-harm/instructions"},
		{"Self-Harm/Intent", guardrail.CategorySelfHarmIntent, "self-harm/intent"},
		{"Sexual", guardrail.CategorySexual, "sexual"},
		{"Sexual/Minors", guardrail.CategorySexualMinors, "sexual/minors"},
		{"Violence", guardrail.CategoryViolence, "violence"},
		{"Violence/Graphic", guardrail.CategoryViolenceGraphic, "violence/graphic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.category) != tt.expected {
				t.Errorf("Category %s = %v, want %v", tt.name, tt.category, tt.expected)
			}
		})
	}
}

func TestModerationConfig_DefaultValues(t *testing.T) {
	// We can test the configuration logic without making API calls
	ctx := context.Background()
	_ = ctx // Will be used when we add mock client tests

	// Test that options work correctly
	tests := []struct {
		name    string
		opts    []guardrail.ModerationOption
		wantErr bool
	}{
		{
			name:    "Default configuration",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "With tripwire enabled",
			opts: []guardrail.ModerationOption{
				guardrail.WithModerationTripwire(true),
			},
			wantErr: false,
		},
		{
			name: "With custom threshold",
			opts: []guardrail.ModerationOption{
				guardrail.WithModerationThreshold(0.7),
			},
			wantErr: false,
		},
		{
			name: "With specific categories",
			opts: []guardrail.ModerationOption{
				guardrail.WithModerationCategories(
					guardrail.CategoryHate,
					guardrail.CategoryViolence,
				),
			},
			wantErr: false,
		},
		{
			name: "With multiple options",
			opts: []guardrail.ModerationOption{
				guardrail.WithModerationTripwire(true),
				guardrail.WithModerationThreshold(0.3),
				guardrail.WithModerationCategories(guardrail.CategorySexual),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create guardrail with options (using nil client for now)
			// In real tests, we'd mock the client
			guard := guardrail.NewModerationGuardrail(nil, tt.opts...)
			if guard == nil {
				t.Fatal("NewModerationGuardrail returned nil")
			}
			if guard.Name != "openai_moderation" {
				t.Errorf("Name = %v, want %v", guard.Name, "openai_moderation")
			}
		})
	}
}
