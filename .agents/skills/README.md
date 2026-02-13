# Project Skills

This repository provides project-local agent skills in `.agents/skills`, similar to the layout used by `openai/openai-agents-python`.

Each skill folder contains a `SKILL.md` with trigger metadata and instructions.

These are project-local Codex workflow skills (not the SDK runtime `agents.Skill` type in Go code).

## Included skills

- `go-sdk-maintainer`: Core SDK feature and bugfix workflow.
- `example-author`: Authoring and validating runnable examples.
- `docs-maintainer`: README + MkDocs documentation maintenance.
