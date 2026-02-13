package agents

import (
	"strings"

	"github.com/MitulShah1/openai-agents-go/guardrail"
	"github.com/MitulShah1/openai-agents-go/tools"
)

// Skill is a reusable runtime capability bundle that can be attached to an agent.
//
// This runtime type is separate from project-local Codex skills stored in
// .agents/skills, which are development workflow instructions.
type Skill struct {
	// Name identifies the skill.
	Name string

	// Description explains what the skill does.
	Description string

	// Instructions are appended to the agent's core instructions.
	Instructions string

	// Tools are added to the agent toolset when the skill is applied.
	Tools []tools.Tool

	// InputGuardrails are appended to agent input guardrails.
	InputGuardrails []*guardrail.Guardrail

	// OutputGuardrails are appended to agent output guardrails.
	OutputGuardrails []*guardrail.Guardrail
}

// AddSkill applies a runtime skill to an agent.
func (a *Agent) AddSkill(skill Skill) {
	a.Skills = append(a.Skills, skill)
	a.Tools = append(a.Tools, skill.Tools...)
	a.InputGuardrails = append(a.InputGuardrails, skill.InputGuardrails...)
	a.OutputGuardrails = append(a.OutputGuardrails, skill.OutputGuardrails...)
}

// AddSkills applies multiple runtime skills to an agent.
func (a *Agent) AddSkills(skills ...Skill) {
	for _, skill := range skills {
		a.AddSkill(skill)
	}
}

func (a *Agent) instructionsWithSkills(base string) string {
	if len(a.Skills) == 0 {
		return base
	}

	parts := make([]string, 0, len(a.Skills)+1)
	base = strings.TrimSpace(base)
	if base != "" {
		parts = append(parts, base)
	}

	for _, skill := range a.Skills {
		if s := strings.TrimSpace(skill.Instructions); s != "" {
			parts = append(parts, s)
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n")
}
