# AGENT.md

This file provides guidance to AI coding assistants (like Claude Code, GitHub Copilot, etc.) when working with code in this repository.

## Overview

**OpenAI Agents Go SDK** - Community-maintained Go SDK for building AI agents with OpenAI's API. Provides multi-agent workflows, tool calling, handoffs, and structured outputs with full type safety.

- **Module**: `github.com/MitulShah1/openai-agents-go`
- **Package**: `agents`
- **Go Version**: 1.24+

## Mandatory Verification

After any code modification, run the full verification stack before considering work complete:

```bash
make check    # Runs fmt, vet, lint, and tests
go test -v -race ./...
```

Rerun checks after fixing failures. All checks must pass before pull requests.

## Build & Development Commands

```bash
# Build and test
go build ./...                    # Build all packages
go test ./...                     # Run all tests
go test -v -race ./...            # Race condition detection
go test -cover ./...              # Coverage analysis

# Run examples (requires OPENAI_API_KEY)
export OPENAI_API_KEY="your-key"
go run examples/01_basic/main.go
go run examples/06_structured_output/main.go

# Code quality (run before commits)
go fmt ./...                      # Format code
go vet ./...                      # Static analysis
golangci-lint run                 # Comprehensive linting

# Makefile targets (recommended)
make check                        # Run all checks (fmt, vet, lint)
make test                         # Run tests with coverage
```

## Repository Structure

```
.
├── agent.go               # Agent type and configuration
├── runner.go              # Agent execution orchestration
├── tool.go                # Tool interface and implementations
├── config.go              # Run configuration options
├── types.go               # Shared types (Result, Step, Usage, etc.)
├── errors.go              # Structured error types
├── guardrail/             # Input/output validation framework
│   └── builtin/          # Built-in guardrails (PII, URL, regex)
├── session/              # Conversation persistence
│   ├── memory.go         # In-memory session storage
│   └── file.go           # File-based session storage
├── internal/
│   └── jsonschema/        # JSON Schema builder for structured outputs
├── examples/              # Usage examples (numbered by complexity)
│   ├── 01_basic/          # Hello world agent
│   ├── 02_tools/          # Tool calling
│   ├── 03_handoffs/       # Agent handoffs
│   ├── 04_lifecycle_hooks/# OnBeforeRun/OnAfterRun hooks
│   ├── 05_config_usage/   # Run configuration
│   ├── 06_structured_output/  # JSON schema outputs
│   ├── 07_complex_schema/ # Nested schemas
│   ├── 08_guardrails_demo/    # Guardrails demonstration
│   ├── 09_sessions_demo/      # Sessions demonstration
│   └── 10_advanced_v02/       # Production chatbot (v0.2.0)
├── .github/workflows/     # CI/CD pipelines
├── AGENT.md              # This file
├── README.md             # User-facing documentation
└── ROADMAP.md            # Future features

**Data Flow**:
1. User → `Runner.Run()` → OpenAI API (via `github.com/openai/openai-go`)
2. API Response → Tool Execution → Agent Handoffs → Final Result
3. Structured Outputs: Schema → Validation → Type-safe JSON

## Code Conventions

- **Idiomatic Go**: Use `gofmt` formatting, standard naming conventions
- **Interface-driven**: All tools implement `Tool` interface
- **Error handling**: Use `fmt.Errorf` with `%w` verb for wrapping, include contextual information
- **Context-first**: All blocking functions accept `context.Context` as first parameter
- **Cyclomatic complexity**: Keep functions under complexity 30 (gocyclo threshold); higher acceptable for table-driven tests and orchestration code
- **Naming patterns**: 
  - Exported types use descriptive names (`Agent`, `Runner`, `Tool`)
  - Options use functional options pattern
  - Errors use `ErrXxx` or `XxxError` naming
- **No unnecessary exports**: Keep internal packages unexported unless needed by external consumers

## Key Design Patterns

1. **Functional Options Pattern**: Used throughout for configuration
   ```go
   agent := NewAgent("name")
   agent.Instructions = "helpful assistant"
   ```

2. **Tool Interface**: Abstract function execution
   ```go
   type Tool interface {
       ToParam() openai.ChatCompletionToolParam
       Execute(args string, ctx ContextVariables) (any, error)
   }
   ```

3. **Handoff Pattern**: Special tool result type for agent transfers
   ```go
   type Handoff struct { Agent *Agent }
   ```

4. **Structured Outputs**: Fluent schema builder
   ```go
   schema := jsonschema.Object().
       WithProperty("field", jsonschema.String()).
       WithRequired("field")
   ```

## Testing Practices

### Unit Tests
- **Mandatory**: Add or update unit tests for any code change unless truly infeasible; if tests can't be added, explain why in PR
- **Table-driven tests**: Use for multiple test cases
- **Mock external calls**: Don't call OpenAI API in tests (use fixtures if needed)
- **Test naming**: `TestFunctionName_Scenario` format
- **Coverage target**: Aim for >80% on core logic

### Running Tests
```bash
# All tests
go test -v ./...

# With race detection (required before PR)
go test -v -race ./...

# With coverage
go test -v -coverprofile=coverage.out ./...

# Specific package
go test -v ./internal/jsonschema/...
```

## Pull Request & Commit Guidelines

### Conventional Commits
Use conventional commit format for all commits:

```
<type>(<scope>): <short summary>

Optional longer description.
```

**Types**:
- `feat`: new feature
- `fix`: bug fix
- `docs`: documentation only
- `test`: adding or fixing tests
- `refactor`: code changes without feature or fix
- `perf`: performance improvement
- `chore`: build, CI, or tooling changes
- `ci`: CI configuration
- `style`: code style (formatting, etc.)

**Examples**:
```
feat(runner): add streaming support
fix(jsonschema): correct strict mode validation
docs(readme): add structured outputs example
test(agent): add lifecycle hooks coverage
refactor(runner): extract prepareRequest helper
```

Keep summary under 80 characters.

### Before Submitting PR
1. ✅ All automated checks pass (`make check`)
2. ✅ Tests cover new behavior and edge cases
3. ✅ Code is readable and maintainable
4. ✅ Public APIs have doc comments (godoc format)
5. ✅ Examples updated if behavior changes
6. ✅ README.md updated for user-facing changes
7. ✅ Commit history follows Conventional Commits

## Development Workflow

1. Sync with `main` branch:
   ```bash
   git checkout main && git pull origin main
   ```

2. Create feature branch:
   ```bash
   git checkout -b feat/short-description
   ```

3. Make changes and add/update tests

4. Run verification stack:
   ```bash
   make check
   go test -v -race ./...
   ```

5. Commit using Conventional Commits:
   ```bash
   git commit -m "feat(scope): add feature"
   ```

6. Push and open pull request:
   ```bash
   git push origin feat/short-description
   ```

## Review Process

Reviewers look for:
- ✅ All CI checks pass
- ✅ Tests are comprehensive and pass
- ✅ Code follows Go best practices
- ✅ Public API changes are documented
- ✅ Breaking changes are clearly marked
- ✅ Examples demonstrate new features
- ✅ Commit messages are clear and follow conventions

## CI/CD Pipeline

### CI Workflow (`.github/workflows/ci.yml`)
- **Triggers**: Push to main/develop, pull requests
- **Go versions**: 1.24.x, 1.25.x
- **Steps**:
  - Checkout and setup Go
  - Download and verify dependencies
  - Run tests with race detection and coverage
  - Upload coverage to Codecov (optional)
  - Run golangci-lint v2
  - Build all packages and examples

### Release Workflow (`.github/workflows/release.yml`)
- **Triggers**: Version tags (`v*`)
- **Requirements**: GoReleaser v2
- **Outputs**: 
  - Example binaries for multiple platforms
  - GitHub releases with changelog
  - Automatic pkg.go.dev update

## Common Issues & Solutions

### 1. OpenAI API Errors
- **Issue**: "Missing required parameter: 'response_format.json_schema'"
- **Fix**: Ensure `ResponseFormat` is properly constructed with manual SDK type creation
- **Reference**: `runner.go:prepareRequest()` method

### 2. Complexity Warnings
- **Issue**: gocyclo reports high complexity
- **Fix**: Extract helper methods (see `Runner.Run` refactoring into `prepareRequest` and `handleToolCalls`)
- **Threshold**: Functions should be under complexity 30

### 3. Lint Config Version Mismatch
- **Issue**: golangci-lint v1 vs v2 config
- **Fix**: Use `version: "2"` in `.golangci.yml` and ensure CI uses v2 action

### 4. Example Build Failures
- **Issue**: Examples fail to build after changes
- **Fix**: Update example imports and ensure all examples build:
  ```bash
  for dir in examples/*/; do (cd "$dir" && go build -v .); done
  ```

## Documentation Standards

### Keep Documentation Up-to-Date ⚠️

**CRITICAL**: Documentation must be updated alongside code changes. This is mandatory, not optional.

#### When to Update Documentation

After **ANY** code change, check if these need updates:

1. **CHANGELOG.md** - ALWAYS update for:
   - New features (Added section)
   - Bug fixes (Fixed section)
   - Breaking changes (Changed section)
   - Deprecated features (Deprecated section)
   - Add entry under "Unreleased" or upcoming version

2. **ROADMAP.md** - Update when:
   - Completing planned features (move from future → completed)
   - Adding new planned features
   - Changing timelines or priorities
   - Update "Current Version" and "Last Updated" in footer

3. **User Documentation** (`docs/`) - Update when:
   - Adding new features or APIs
   - Changing behavior of existing features
   - Adding new examples
   - Files to check: `docs/quickstart.md`, concept guides, API reference

4. **README.md** - Update when:
   - Adding major features (update "Supported Features" section)
   - Changing installation or quick start examples
   - Updating comparison table with other SDKs
   - Adding new examples directory

5. **Examples** - Update when:
   - API signatures change
   - Best practices change
   - New features are added

#### Documentation Update Checklist

Before submitting PR:
- [ ] CHANGELOG.md updated with changes
- [ ] ROADMAP.md updated if completing planned features
- [ ] User documentation updated if user-facing changes
- [ ] README.md updated if major feature or quick start affected
- [ ] Examples still build and run correctly
- [ ] Godoc comments added/updated for public APIs

#### AI Assistant Reminder

When making code changes:
1. Identify all affected documentation
2. Update documentation in the same commit/PR
3. Use conventional commit messages that indicate doc updates:
   - `feat(guardrails): add XYZ guardrail` → Update CHANGELOG, docs/guardrails.md
   - `fix(runner): correct timeout handling` → Update CHANGELOG, possibly README
   - `docs: update quickstart guide` → Documentation-only change

### Godoc Comments
All exported types, functions, and methods must have godoc comments:

```go
// Agent represents an AI agent with configured behavior.
// It can execute tools, hand off to other agents, and return structured outputs.
type Agent struct {
    Name string
    // ...
}

// NewAgent creates a new agent with the given name.
// The agent is initialized with default settings and no tools.
func NewAgent(name string) *Agent {
    // ...
}
```

### README Updates
When adding features:
1. Add to "Supported Features" section
2. Create example in Quick Start (use collapsible sections)
3. Update comparison table if relevant
4. Add to examples directory

## Dependencies

### Production
- **Only**: `github.com/openai/openai-go` (official OpenAI SDK)
- Keep core SDK dependency-free for maximum portability

### Development
- `golangci-lint`: Linting (v2.x)
- `goreleaser`: Release automation (v2.x)

## Tips for AI Assistants

1. **Always check tests**: Run tests after any code change
2. **Follow existing patterns**: Study similar code before adding new features
3. **Use type system**: Leverage Go's type safety for correctness
4. **Document thoughtfully**: Write clear godoc comments
5. **Keep it simple**: Prefer simple, readable code over clever solutions
6. **Test edge cases**: Think about nil pointers, empty slices, context cancellation
7. **Check examples**: Ensure examples still work after API changes

## Current Version: v0.2.0

**Completed Features:**
- ✅ Guardrails (PII detection, URL filtering, custom regex)
- ✅ Sessions (memory and file-based conversation persistence)
- ✅ Structured outputs with JSON schema
- ✅ Multi-agent workflows and handoffs
- ✅ Tool calling and lifecycle hooks

**Future Roadmap:**

See [ROADMAP.md](./ROADMAP.md) for planned features:
- 🔮 Database session backends (SQLite, Redis, PostgreSQL) - v0.3.0
- 🔮 Tracing and observability - v0.3.0
- 🔮 Streaming support - v0.4.0
- 🔮 Voice agent support - Future

---

**Note**: This SDK is community-maintained and not officially affiliated with OpenAI. For official SDKs, see:
- Python: https://github.com/openai/openai-agents-python
- JavaScript: https://github.com/openai/openai-agents-js
