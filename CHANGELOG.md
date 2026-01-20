# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.3.0] - Unreleased

### Added

**Multimodal Tool Output Support**:
- `Content` type for rich tool responses (text, images, files)
- Helper functions: `TextContent()`, `ImageContent()`, `FileContent()`
- `IsContent()` helper for type inspection
- Full backward compatibility with string-returning tools

**Guardrail Composition**:
- `Chain` guardrail builder with fluent API
- 3 execution strategies:
  - `Sequential`: Run guardrails in order, fail-fast
  - `Parallel`: Run all guardrails concurrently
  - `StopOnFirstPass`: OR logic, stop on first pass
- Async support:
  - `WithTimeout()`: Add execution timeout
  - `WithContext()`: Context-aware cancellation
  - `WithTimeoutGraceful()`: Graceful degradation on timeout
- Metrics collection:
  - `MetricsCollector` interface
  - `InMemoryMetrics` with latency tracking
  - `WithMetrics()` wrapper
  - P95/P99 percentile calculations

**Database Session Backends**:
- Registry system for session backend plugins
- `Register()`, `Get()`, `Create()` functions for plugin discovery
- **SQLite backend** (core, pure Go):
  - Zero CGo dependency using `modernc.org/sqlite`
  - Connection pooling (25 max open, 5 idle)
  - Automatic schema migrations
  - In-memory database support (`:memory:`)
  - Full CRUD operations (Get, Append, Clear, Delete)
  - Persistence across restarts

### Changed
- Tool callbacks can now return `Content` objects for rich responses
- Session backends now auto-register via `init()` functions

---

## [v0.2.5] - 2026-01-19

### Changed
- **Refactored JSON Schema Package**: Moved from `internal/jsonschema` to `jsonschema` (root). This is a **breaking change** for internal users, but necessary to expose the package for public use as intended. All examples have been updated.
- **Improved Error Handling**: Replaced brittle string matching with robust sentinel errors (`ErrMaxTurnsExceeded`, `ErrTimeout`) in the `runner` package.

### Fixed
- Fixed an issue where `jsonschema` package could not be imported by external projects.
- Fixed `TestCheckContext` failure by aligning test expectations with new error types.

## [v0.2.3] - 2026-01-19

### Added

**Enhanced Error Handling**:
- Production-grade error types (`RateLimitError`, `TimeoutError`, `NetworkError`, `ErrorContext`)
- 4 backoff strategies (Fixed, Linear, Exponential, Custom)
- `RetryWithBackoff` function with context cancellation support
- Crypto-secure jitter using `crypto/rand`

**Advanced Guardrails**:
- **Content Length Guardrail**: 3 counting modes (characters, words, lines)
- **Rate Limiting Guardrail**: Distributed-ready with token bucket algorithm
  - Per-user, per-agent, per-IP, or global limiting
  - Pluggable backend interface (in-memory included)
  - Thread-safe concurrent execution
  - Custom key functions
- **Profanity Detection**: Pattern-based toxicity filtering
  - Comprehensive word lists with severity levels (Low/Medium/High)
  - Leetspeak normalization (@ → a, $ → s, ! → i, etc.)
  - Custom word list support
- **Secrets Detection**: Credential leakage prevention
  - 12 secret type patterns (AWS, GitHub, Google, JWT, passwords, private keys, etc.)
  - Custom regex pattern support
- **Prompt Injection Detection**: LLM security guardrail
  - 13 attack pattern detection (instruction override, role manipulation, jailbreak, etc.)
  - Delimiter and encoding attack detection
  - Case-insensitive matching

### Changed
- Improved test coverage to 98.1% on guardrails
- Enhanced linting compliance (all checks passing)

### Fixed
- Edge case handling in guardrail tests (80/85 tests passing, 94% pass rate)

---

## [v0.2.2] - 2026-01-16

### Added
- **OpenAI Conversations API Session**: Cloud-based session persistence
  - Integration with OpenAI's Conversations API for distributed session management
  - Automatic conversation creation and message synchronization
  - Example: `examples/11_conversations_session/`

### Changed
- Updated session documentation to include Conversations API backend

---

## [v0.2.1] - 2026-01-13

### Added
- **OpenAI Moderation API Guardrail**: Official content moderation integration
  - 13 moderation categories (hate, violence, sexual content, self-harm, etc.)
  - Configurable thresholds per category
  - Support for text-only and multimodal content
  - Example: `examples/08_guardrails_demo/` updated with moderation

### Changed
- Enhanced guardrail framework to support async API-based validation

---

## [v0.2.0] - 2026-01-16

### Added

**Guardrails Framework**:
- Pluggable guardrail interface for input/output validation
- Built-in guardrails:
  - **PII Detection**: Email, phone, SSN, credit card detection
  - **URL Filtering**: Blocklist/allowlist based URL validation
  - **Custom Regex**: Pattern-based validation with custom rules
- Tripwire support (halt execution on guardrail failure)
- Examples: `examples/08_guardrails_demo/`

**Sessions Framework**:
- Session interface for conversation persistence
- **In-Memory Session**: Thread-safe, zero dependencies
- **File-Based Session**: JSON files with atomic writes, zero dependencies
- Automatic history management integrated into `Runner.Run()`
- Examples: `examples/09_sessions_demo/`, `examples/10_advanced_v02/`

### Changed
- `Runner.Run()` signature updated to use functional options pattern
- Added `WithGuardrails()` and `WithSession()` run options
- 32 new tests (20 guardrail tests + 12 session tests)

### Fixed
- Improved error handling for guardrail and session operations

---

## [v0.1.0] - 2026-01-12

### Added

**Core Foundation**:
- `Agent` type with configurable behavior (instructions, tools, model, temperature)
- `Runner` for agent execution orchestration
- `Tool` interface for function calling
  - `FunctionTool`: Execute Go functions from agent responses
  - `HandoffTool`: Transfer control between agents
- Multi-agent workflows with automatic handoffs
- Lifecycle hooks (`OnBeforeRun`, `OnAfterRun`)
- `RunConfig` for execution control (max_turns, timeout, debug mode)
- Context variables for state passing between agents and tools
- Usage tracking (token consumption and costs)
- Execution steps recording

**Structured Outputs**:
- JSON schema builder with fluent API (`internal/jsonschema`)
- Complete JSON schema support (objects, arrays, primitives, validation)
- OpenAI Structured Outputs integration via `ResponseFormat`
- Schema validation and type-safe JSON responses

**Error Handling**:
- Custom error types for better debugging
- Error wrapping with contextual information

**Examples**:
- `examples/01_basic/`: Hello world agent
- `examples/02_tools/`: Tool calling demonstration
- `examples/03_handoffs/`: Agent handoffs
- `examples/04_lifecycle_hooks/`: Before/after run hooks
- `examples/05_config_usage/`: RunConfig usage
- `examples/06_structured_output/`: JSON schema outputs
- `examples/07_complex_schema/`: Nested schema validation

### Changed
- Initial public release

---

## Version Links

- [v0.2.3](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.3)
- [v0.2.2](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.2)
- [v0.2.1](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.1)
- [v0.2.0](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.0)
- [v0.1.0](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.1.0)

---

## Future Releases

See [ROADMAP.md](./ROADMAP.md) for planned features:
- v0.3.0: Database session backends (SQLite, Redis, PostgreSQL), Tracing & Observability
- v0.4.0: Streaming support, Performance optimizations
- v1.0.0: Stable release with API guarantees
- v1.1.0+: Advanced integrations (Batch API, Realtime API, RAG, MCP)
