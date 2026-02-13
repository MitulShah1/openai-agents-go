---
name: example-author
description: Create or update runnable examples under examples/ for OpenAI Agents Go SDK features. Use when adding feature demos, improving developer onboarding snippets, or validating API ergonomics with end-to-end example code.
---

# Example Author

Create examples that compile and teach one concept clearly.

## Authoring rules

- Keep each example focused on one primary concept.
- Use clear names and concise comments.
- Prefer realistic tool schemas and messages.
- Avoid unnecessary dependencies.

## Validation steps

1. Build or test the example package.
2. Verify imports and API usage match current SDK.
3. Keep output/logging simple and readable.

## Placement

- Add new examples in `examples/<nn>_<topic>/main.go`.
- Reuse patterns from nearby examples to keep consistency.
