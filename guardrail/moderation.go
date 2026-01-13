// Package guardrail provides input/output validation for agents.
package guardrail

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
)

// ModerationCategory represents a content moderation category.
type ModerationCategory string

// Available moderation categories from OpenAI's Moderation API.
const (
	CategoryHarassment            ModerationCategory = "harassment"
	CategoryHarassmentThreatening ModerationCategory = "harassment/threatening"
	CategoryHate                  ModerationCategory = "hate"
	CategoryHateThreatening       ModerationCategory = "hate/threatening"
	CategoryIllicit               ModerationCategory = "illicit"
	CategoryIllicitViolent        ModerationCategory = "illicit/violent"
	CategorySelfHarm              ModerationCategory = "self-harm"
	CategorySelfHarmInstructions  ModerationCategory = "self-harm/instructions"
	CategorySelfHarmIntent        ModerationCategory = "self-harm/intent"
	CategorySexual                ModerationCategory = "sexual"
	CategorySexualMinors          ModerationCategory = "sexual/minors"
	CategoryViolence              ModerationCategory = "violence"
	CategoryViolenceGraphic       ModerationCategory = "violence/graphic"
)

// ModerationConfig configures the OpenAI Moderation guardrail.
type ModerationConfig struct {
	// Tripwire determines if flagged content should halt execution
	Tripwire bool

	// Threshold is the minimum score (0.0-1.0) to flag content
	// Default is 0.5. Lower values are more strict.
	Threshold float64

	// Categories specifies which categories to check
	// If empty, all categories are checked
	Categories map[ModerationCategory]bool
}

// ModerationOption is a functional option for configuring the moderation guardrail.
type ModerationOption func(*ModerationConfig)

// WithModerationTripwire enables tripwire mode (halts execution on detection).
func WithModerationTripwire(enabled bool) ModerationOption {
	return func(c *ModerationConfig) {
		c.Tripwire = enabled
	}
}

// WithModerationThreshold sets the score threshold for flagging content.
// Default is 0.5. Lower values are more strict (0.0-1.0).
func WithModerationThreshold(threshold float64) ModerationOption {
	return func(c *ModerationConfig) {
		c.Threshold = threshold
	}
}

// WithModerationCategories enables specific moderation categories.
// If not called, all categories are checked by default.
func WithModerationCategories(categories ...ModerationCategory) ModerationOption {
	return func(c *ModerationConfig) {
		if c.Categories == nil {
			c.Categories = make(map[ModerationCategory]bool)
		}
		for _, cat := range categories {
			c.Categories[cat] = true
		}
	}
}

// NewModerationGuardrail creates a guardrail that uses OpenAI's Moderation API
// to detect harmful content across multiple categories.
func NewModerationGuardrail(client *openai.Client, opts ...ModerationOption) *Guardrail {
	config := &ModerationConfig{
		Tripwire:  false, // Default to non-blocking
		Threshold: 0.5,
		Categories: map[ModerationCategory]bool{
			// Check all categories by default
			CategoryHarassment:            true,
			CategoryHarassmentThreatening: true,
			CategoryHate:                  true,
			CategoryHateThreatening:       true,
			CategoryIllicit:               true,
			CategoryIllicitViolent:        true,
			CategorySelfHarm:              true,
			CategorySelfHarmInstructions:  true,
			CategorySelfHarmIntent:        true,
			CategorySexual:                true,
			CategorySexualMinors:          true,
			CategoryViolence:              true,
			CategoryViolenceGraphic:       true,
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	return NewGuardrail("openai_moderation", func(ctx context.Context, input string) (*Result, error) {
		// Call OpenAI Moderation API
		resp, err := client.Moderations.New(ctx, openai.ModerationNewParams{
			Input: openai.ModerationNewParamsInputUnion{
				OfStringArray: []string{input},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("moderation API call failed: %w", err)
		}

		if len(resp.Results) == 0 {
			return nil, fmt.Errorf("moderation API returned no results")
		}

		result := resp.Results[0]

		// Check enabled categories
		violations, scores := checkModerationCategories(config, result.CategoryScores)

		if len(violations) > 0 {
			return &Result{
				Passed:            false,
				TripwireTriggered: config.Tripwire,
				Message:           fmt.Sprintf("Content flagged by moderation API: %s", strings.Join(violations, ", ")),
				Metadata: map[string]any{
					"flagged":    result.Flagged,
					"violations": violations,
					"scores":     scores,
				},
			}, nil
		}

		return &Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "Content passed moderation check",
		}, nil
	})
}

// checkModerationCategories validates scores against configured categories and threshold.
func checkModerationCategories(config *ModerationConfig, scores openai.ModerationCategoryScores) ([]string, []string) {
	var violations []string
	var scoreDetails []string

	// Check each category
	if config.Categories[CategoryHarassment] && scores.Harassment >= config.Threshold {
		violations = append(violations, "harassment")
		scoreDetails = append(scoreDetails, fmt.Sprintf("harassment: %.2f", scores.Harassment))
	}
	if config.Categories[CategoryHarassmentThreatening] && scores.HarassmentThreatening >= config.Threshold {
		violations = append(violations, "harassment/threatening")
		scoreDetails = append(scoreDetails, fmt.Sprintf("harassment/threatening: %.2f", scores.HarassmentThreatening))
	}
	if config.Categories[CategoryHate] && scores.Hate >= config.Threshold {
		violations = append(violations, "hate")
		scoreDetails = append(scoreDetails, fmt.Sprintf("hate: %.2f", scores.Hate))
	}
	if config.Categories[CategoryHateThreatening] && scores.HateThreatening >= config.Threshold {
		violations = append(violations, "hate/threatening")
		scoreDetails = append(scoreDetails, fmt.Sprintf("hate/threatening: %.2f", scores.HateThreatening))
	}
	if config.Categories[CategoryIllicit] && scores.Illicit >= config.Threshold {
		violations = append(violations, "illicit")
		scoreDetails = append(scoreDetails, fmt.Sprintf("illicit: %.2f", scores.Illicit))
	}
	if config.Categories[CategoryIllicitViolent] && scores.IllicitViolent >= config.Threshold {
		violations = append(violations, "illicit/violent")
		scoreDetails = append(scoreDetails, fmt.Sprintf("illicit/violent: %.2f", scores.IllicitViolent))
	}
	if config.Categories[CategorySelfHarm] && scores.SelfHarm >= config.Threshold {
		violations = append(violations, "self-harm")
		scoreDetails = append(scoreDetails, fmt.Sprintf("self-harm: %.2f", scores.SelfHarm))
	}
	if config.Categories[CategorySelfHarmInstructions] && scores.SelfHarmInstructions >= config.Threshold {
		violations = append(violations, "self-harm/instructions")
		scoreDetails = append(scoreDetails, fmt.Sprintf("self-harm/instructions: %.2f", scores.SelfHarmInstructions))
	}
	if config.Categories[CategorySelfHarmIntent] && scores.SelfHarmIntent >= config.Threshold {
		violations = append(violations, "self-harm/intent")
		scoreDetails = append(scoreDetails, fmt.Sprintf("self-harm/intent: %.2f", scores.SelfHarmIntent))
	}
	if config.Categories[CategorySexual] && scores.Sexual >= config.Threshold {
		violations = append(violations, "sexual")
		scoreDetails = append(scoreDetails, fmt.Sprintf("sexual: %.2f", scores.Sexual))
	}
	if config.Categories[CategorySexualMinors] && scores.SexualMinors >= config.Threshold {
		violations = append(violations, "sexual/minors")
		scoreDetails = append(scoreDetails, fmt.Sprintf("sexual/minors: %.2f", scores.SexualMinors))
	}
	if config.Categories[CategoryViolence] && scores.Violence >= config.Threshold {
		violations = append(violations, "violence")
		scoreDetails = append(scoreDetails, fmt.Sprintf("violence: %.2f", scores.Violence))
	}
	if config.Categories[CategoryViolenceGraphic] && scores.ViolenceGraphic >= config.Threshold {
		violations = append(violations, "violence/graphic")
		scoreDetails = append(scoreDetails, fmt.Sprintf("violence/graphic: %.2f", scores.ViolenceGraphic))
	}

	return violations, scoreDetails
}
