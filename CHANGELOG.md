# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

## [v0.5.0] - 2026-02-05

### Added

**MCP (Model Context Protocol)**:
- `pkg/mcp`: Client, server, and tool adapters using mark3labs/mcp-go
- `ToAgentTool()` adapter to convert MCP tools to SDK tools

**OpenAI API Clients**:
- `pkg/files`: Files API client (List, Get, Upload, Content) with tests
- `pkg/embeddings`: Embeddings API client
- `pkg/vectorstore`: Vector Stores API client (Create, Search placeholder)

**Computer Use**:
- `pkg/computer`: Sync and async interfaces (Screenshot, Click, Type, Scroll, etc.)
- `pkg/computer/mock`: MockComputer and AsyncMockComputer for testing
- `tools/computer_tool.go`: Computer tool for agents with screenshot/click/type actions

**Diff & Patch**:
- `pkg/diff`: Parser and applicator for unified diffs
- `tools/apply_patch_tool.go`: Apply-patch tool for agent-driven file edits

**Session & Plugins**:
- `session/compaction.go`: Message compaction (DefaultCompactor) for context window management
- `plugins/redis`: Redis session backend with tests

**Tools**:
- `tools/file_tool.go`: list_files, get_file, get_file_content tools
- `tools/rag_tool.go`: RAG tool for retrieval-augmented generation

**Structure**:
- Core packages moved under `pkg/` (computer, diff, embeddings, files, mcp, vectorstore)
- Examples renumbered to 04–22 (e.g. 03b_advanced_handoffs → 04_advanced_handoffs)

### Changed
- Examples: directory renumbering (04–22) and `examples/README.md` updates
- Tools and internal code updated to use new `pkg/` imports

### Fixed
- **Linting**: Resolved all golangci-lint issues (errcheck, gocritic, revive)
  - errcheck: `rc.Close()` return handled in files_test and file_tool
  - gocritic: if-else chain rewritten as switch in pkg/diff/parser
  - revive: package/exported comments and unused-parameter renames across pkg/

## [v0.4.0] - 2026-01-23

### Added

**Parallel Tools Integration**:
- **Three-State API Parameter**: Full parity with Python SDK's `parallel_tool_calls`
  - `true`: Explicitly enable parallel tool calls (model can request multiple tools)
  - `false`: Restrict to one tool call per turn
  - `nil`: Use provider default (typically parallel)
- **Enhanced ConfigMerger**: New `GetParallelToolCallsPtr()` method for proper API parameter transmission
- **Goroutine-Based Execution**: Client-side parallel tool execution using goroutines
  - Semaphore pattern for concurrency limiting via `MaxToolConcurrency`
  - Order preservation despite async execution
  - Error isolation (one tool's error doesn't block others)
- **Comprehensive Test Suite**: 5 new tests in `internal/runner/toolhandler_parallel_test.go`
  - `TestParallelToolExecution`: Verifies concurrent execution (~100ms for 2 tools)
  - `TestSequentialToolExecution`: Verifies sequential mode (~100ms total)
  - `TestConcurrencyLimiting`: Verifies semaphore limits concurrent execution
  - `TestOrderPreservation`: Verifies results maintain tool call order
  - `TestParallelErrorHandling`: Verifies error isolation in parallel mode
- **Example**: `examples/21_parallel_tools/` demonstrating:
  - Parallel execution (default, ~2s for 3 tools)
  - Sequential execution (~6s for 3 tools)
  - Limited concurrency (max 2 tools, ~4s)
  - Performance comparison showing 3x speedup
  - Works with or without OpenAI API key (demo mode)
- **Documentation**: Comprehensive "Parallel Tool Execution" section in `docs/tools.md`
  - Configuration examples (agent-level and runtime override)
  - Performance comparison table
  - API integration details
  - Best practices and use cases

### Changed
- **API Parameter Logic**: Enhanced `internal/runner/preparation.go` to support three-state behavior
- **Runner Integration**: Updated `runner.go` to use `GetParallelToolCallsPtr()` for API requests

### Fixed
- All linting issues resolved (0 issues)
- Race detection clean for parallel tools implementation

## [v0.3.5] - 2026-01-22

### Added
- **Handoff Parity**:
  - New `handoff` package with functional options pattern.
  - Input filtering support via `WithInputFilter`.
  - History nesting (summarization) support via `WithHistoryNesting`.
  - Dynamic enablement via `WithEnabledPredicate`.
  - Type-safe `isHandoffTool` marker.

### Removed
- **Error Handling**: Removed `retry.go`, `retry_test.go` and `RetryWithBackoff` functionality.
- **Example**: Removed `examples/14_error_handling`.

## [v0.3.0] - 2026-01-20

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
- **Session Utilities**:
  - `WithCompression()`: Transparent GZIP compression for session data
  - `WithEncryption()`: AES-GCM encryption for secure storage

### Changed
- Tool callbacks can now return `Content` objects for rich responses
- Session backends now auto-register via `init()` functions
- **Docs**: Comprehensive updates to `docs/guardrails.md`, `docs/sessions/index.md`, and README.

### Fixed
- **Security**: Resolved potential integer overflows in `runner` package (CWE-190).
- **Security**: Added strict validation for session encryption keys.
- **Stability**: Fixed potential race conditions in parallel guardrail execution.
- **Linting**: Resolved all `goconst`, `revive`, and `misspell` issues.

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

- [v0.5.0](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.5.0)
- [v0.4.0](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.4.0)
- [v0.3.5](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.3.5)
- [v0.3.0](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.3.0)
- [v0.2.5](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.5)
- [v0.2.3](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.3)
- [v0.2.2](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.2)
- [v0.2.1](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.1)
- [v0.2.0](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.2.0)
- [v0.1.0](https://github.com/MitulShah1/openai-agents-go/releases/tag/v0.1.0)

---

## Future Releases

See [ROADMAP.md](./ROADMAP.md) for planned features:
- v0.4.0: Tracing (OpenTelemetry), Metrics & Streaming support
- v1.0.0: Stable release with API guarantees
- v1.1.0+: Advanced integrations (Batch API, Realtime API, RAG, MCP)

