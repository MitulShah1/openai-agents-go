// Package moderation provides AI content moderation guardrails.
package moderation

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"

	"github.com/MitulShah1/openai-agents-go/guardrail"
)

// Category represents a content moderation category from OpenAI's Moderation API.
// These categories are used to classify and flag potentially harmful content.
type Category string

// Moderation categories supported by OpenAI's Moderation API.
// Each category represents a different type of potentially harmful content.
//
// Categories include:
//   - Harassment: Content targeting or demeaning individuals
//   - Hate: Content promoting hate based on protected characteristics
//   - Illicit: Instructions for illegal activities
//   - Self-Harm: Content promoting or encouraging self-harm
//   - Sexual: Sexual content (excluding minors)
//   - Violence: Content depicting or promoting violence
//
// Some categories have sub-categories (e.g., "hate/threatening") for more severe violations.
const (
	CategoryHarassment            Category = "harassment"
	CategoryHarassmentThreatening Category = "harassment/threatening"
	CategoryHate                  Category = "hate"
	CategoryHateThreatening       Category = "hate/threatening"
	CategoryIllicit               Category = "illicit"
	CategoryIllicitViolent        Category = "illicit/violent"
	CategorySelfHarm              Category = "self-harm"
	CategorySelfHarmInstructions  Category = "self-harm/instructions"
	CategorySelfHarmIntent        Category = "self-harm/intent"
	CategorySexual                Category = "sexual"
	CategorySexualMinors          Category = "sexual/minors"
	CategoryViolence              Category = "violence"
	CategoryViolenceGraphic       Category = "violence/graphic"
)

// Config configures the OpenAI Moderation guardrail behavior.
// Use functional options (WithModerationX) to customize the configuration.
type Config struct {
	// Tripwire determines if flagged content should halt execution
	Tripwire bool

	// Threshold is the minimum score (0.0-1.0) to flag content
	// Default is 0.5. Lower values are more strict.
	Threshold float64

	// Categories specifies which categories to check
	// If empty, all categories are checked
	Categories map[Category]bool
}

// Option is a functional option for configuring the moderation guardrail.
type Option func(*Config)

// WithModerationTripwire enables tripwire mode (halts execution on detection).
func WithModerationTripwire(enabled bool) Option {
	return func(c *Config) {
		c.Tripwire = enabled
	}
}

// WithModerationThreshold sets the score threshold for flagging content.
// Default is 0.5. Lower values are more strict (0.0-1.0).
func WithModerationThreshold(threshold float64) Option {
	return func(c *Config) {
		c.Threshold = threshold
	}
}

// WithModerationCategories enables specific moderation categories.
// If not called, all categories are checked by default.
func WithModerationCategories(categories ...Category) Option {
	return func(c *Config) {
		if c.Categories == nil {
			c.Categories = make(map[Category]bool)
		}
		for _, cat := range categories {
			c.Categories[cat] = true
		}
	}
}

// NewOpenAI creates a guardrail that uses OpenAI's Moderation API
// to detect harmful content across multiple categories.
//
// By default, all 13 moderation categories are checked with a threshold of 0.5.
// Content scoring above the threshold is flagged as a violation.
//
// Example usage:
//
//	client := openai.NewClient(option.WithAPIKey("your-key"))
//
//	// Default configuration (all categories, threshold 0.5)
//	guard := guardrail.NewOpenAI(client)
//
//	// Custom threshold (stricter)
//	guard = guardrail.NewOpenAI(client,
//		guardrail.WithModerationThreshold(0.3),
//	)
//
//	// Specific categories only
//	guard = guardrail.NewOpenAI(client,
//		guardrail.WithModerationCategories(
//			guardrail.CategoryHate,
//			guardrail.CategoryViolence,
//		),
//	)
//
//	// With tripwire (halts agent on violation)
//	guard = guardrail.NewOpenAI(client,
//		guardrail.WithModerationTripwire(true),
//	)
func NewOpenAI(client *openai.Client, opts ...Option) *guardrail.Guardrail {
	config := &Config{
		Tripwire:  false, // Default to non-blocking
		Threshold: 0.5,
		Categories: map[Category]bool{
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

	return guardrail.NewGuardrail("openai_moderation", func(ctx context.Context, input string) (*guardrail.Result, error) {
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
			return &guardrail.Result{
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

		return &guardrail.Result{
			Passed:            true,
			TripwireTriggered: false,
			Message:           "Content passed moderation check",
		}, nil
	})
}

// checkModerationCategories validates scores against configured categories and threshold.
func checkModerationCategories(config *Config, scores openai.ModerationCategoryScores) ([]string, []string) {
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
