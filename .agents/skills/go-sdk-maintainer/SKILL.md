---
name: go-sdk-maintainer
description: Maintain and extend the OpenAI Agents Go SDK core packages with safe, tested changes. Use when adding SDK features, fixing bugs in agent/runner/tools/handoff/session packages, updating public API behavior, or preparing production-quality Go code changes.
---

# Go SDK Maintainer

Implement changes in small, reviewable commits.

## Follow this workflow

1. Read the touched package tests first.
2. Update code with backward compatibility in mind.
3. Add or update focused unit tests.
4. Run package tests first, then full repo tests.
5. Update docs when behavior changes.

## API safety rules

- Preserve exported API unless the task explicitly allows breaking changes.
- Prefer additive changes over mutating existing behavior.
- Keep defaults stable.
- Keep error messages actionable.

## Testing checklist

- Run targeted tests for touched packages.
- Run `go test ./...` before finishing.
- Ensure new code is formatted with `gofmt`.

Use `references/testing.md` for command patterns.
