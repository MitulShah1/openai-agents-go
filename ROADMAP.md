# OpenAI Agents Go SDK - Development Roadmap

> A comprehensive plan to build a robust, production-ready Go SDK for OpenAI Agents with feature parity to the official [Python SDK](https://github.com/openai/openai-agents-python).

---

## Vision

Build a Go SDK that provides:
- **Zero-dependency core** for easy adoption
- **Production-ready features** (guardrails, sessions, tracing)
- **Idiomatic Go** patterns and best practices
- **Feature parity** with Python SDK
- **Excellent documentation** and examples

---

## Timeline Overview

```
Week 1-2  │ v0.1.0 - Core Foundation ✅
Week 3-4  │ v0.2.0 - Guardrails & Sessions ✅
Week 4-5  │ v0.2.3 - Enhanced Guardrails & Error Handling ✅
Week 6-9  │ v0.3.0 - DB Backends, Tracing & Composition ✅
Week 13-14│ v0.5.2 - P0 Bug Fixes ✅
Week 15-16│ v0.6.0 - Python SDK Parity 🚀
Week 17-18│ v1.0.0 - Stable Release
Future    │ v1.1.0+ - Advanced Integrations
```

---

## Version Roadmap

### Completed Releases ✅

For detailed release notes, see [CHANGELOG.md](./CHANGELOG.md).

#### v0.5.2 - P0 Bug Fixes (2026-02-07) ✅
- **Streaming CPU Fix**: Removed spin loop in streaming result iterator
- **History Fix**: Prevented exponential message duplication in sessions
- **Tracing Fix**: Explicitly ended spans in loops to avoid memory leak
- **Concurrency Fix**: Resolved map data race in parallel tool calls
- **Config Fix**: Correctly respect `MaxToolConcurrency` in all paths
- **Streaming Parity**: Fixed default parallel tool calls in streaming
- **Guardrail Fix**: Proper text extraction for input guardrails

#### v0.5.1 - Cleanup & Restructure (2026-02-06) ✅
- **Package Cleanup**: Removed bloat outside agent SDK scope
  - Removed `pkg/embeddings/`, `pkg/files/`, `pkg/vectorstore/`
  - Users should use `openai-go` directly for these features
- **Package Restructure**: Better API design matching Python SDK
  - Moved `pkg/mcp/` → `mcp/`, `pkg/computer/` → `computer/`, `pkg/diff/` → `diff/`
  - Top-level packages for core features
- **Plugin Simplification**: Build tag approach for optional backends
  - Merged `plugins/postgres/` and `plugins/redis/` into `session/` package
  - PostgreSQL backend: `session.NewPostgresSession()` (build tag: `-tags postgres`)
  - Redis backend: `session.NewRedisSession()` (build tag: `-tags redis`)

#### v0.5.0 - MCP, Computer Use & Extensions (2026-02-05) ✅
- **MCP Support**: Model Context Protocol integration ✅
- **Computer Use Interface**: Browser/desktop automation ✅
- **Diff Application**: Structured code changes ✅
- **Session Enhancements**: Message compaction, Redis backend ✅

#### v0.4.0 - Tracing, Streaming & Performance (2026-01-23) ✅
- **Parallel Tools**: Idiomatic goroutine-based parallel execution
- **Tracing**: OpenTelemetry support for agents, tools, and sessions
- **Streaming**: Token-by-token and object-based streaming
- **Plugins**: Redis and PostgreSQL backends

#### v0.3.5 - Handoff Parity & Idiomatic Go (2026-01-22) ✅
- **Handoff Parity**: Full feature parity with Python SDK
  - Input filtering, history nesting, dynamic enablement
- **Refactoring**: 
  - Implementation of Idiomatic Go patterns
  - Improved type safety and error handling
  - Removal of legacy retry logic in favor of robust error types

#### v0.3.0 - Database Backends, Tracing & Composition (2026-01-20) ✅
- **Multimodal Tool Outputs**: Rich responses (text, images, files)
- **Guardrail Composition**: Chain builder with seq/parallel execution
- **Database Sessions**: SQLite backend with connection pooling
- **Metrics Collection**: Guardrail telemetry (P95/P99)

#### v0.2.3 - Enhanced Guardrails & Error Handling (2026-01-19) ✅
- 9+ production-ready guardrails (PII, Moderation, Rate Limiting, Profanity, Secrets, Prompt Injection, etc.)
- Production-grade error handling with 4 retry strategies
- 98.1% test coverage on guardrails

#### v0.2.2 - Conversations API (2026-01-16) ✅
- Cloud-based session persistence via OpenAI Conversations API

#### v0.2.1 - Moderation Guardrail (2026-01-13) ✅
- OpenAI Moderation API integration (13 categories)

#### v0.2.0 - Guardrails & Sessions (2026-01-16) ✅
- Guardrail framework with PII, URL filtering, custom regex
- Session framework with in-memory, file-based, and cloud backends

#### v0.1.0 - Core Foundation (2026-01-12) ✅
- Agent, Runner, Tool abstractions
- Structured outputs with JSON schema builder
- Multi-agent workflows and handoffs

---



---

### v0.6.0 - Python SDK Parity 🚀
**Timeline**: Q1 2026  
**Status**: In Progress (2/6 features remaining)

#### Completed in v0.5.x ✅
- ✅ **MCP Support**: Model Context Protocol integration
- ✅ **Computer Use Interface**: Browser and desktop automation
- ✅ **Diff Application Logic**: Structured code change application
- ✅ **Enhanced Session Backends**: Compaction and distributed sessions
- ✅ **Package Structure**: Cleaned and reorganized for better API

#### Completed in v0.6.0 (in progress) ✅
- ✅ **Tool Approvals**: Full human-in-the-loop approval workflow
  - Runner checks `NeedsApproval`/`ApprovalFunc` before tool execution
  - Pause/resume via `ToolApprovalRequiredError` + `RunState` + `Runner.Resume()`
  - Inline approval via `WithApprovalHandler()` run option
  - Streaming support with `StreamEventApprovalRequired`
  - Parallel batch safety (no partial execution)

#### Remaining Work 📋

**Critical Features** (Python SDK parity gaps):

1. **Prompts API** - Dynamic prompt configuration
   - `prompts.Prompt` type for OpenAI Prompts API
   - `DynamicPromptFunc` for runtime prompt generation
   - Integration with `Agent` and `Runner`
   - Estimated: 3-4 hours

2. **Model Abstraction** - Multi-provider support (deferred to v0.7.0)
See [CHANGELOG.md](file:///home/mitul/project/openai-agents-go/CHANGELOG.md) for detailed migration guides.

---

## v0.6.0 (Planned - Q1 2026)

**Status**: In Planning  
**Focus**: Complete Python SDK feature parity with full integration

### Goals

1. **Complete Prompts API Integration**
   - [ ] Add `Prompt` field to `Agent` struct
   - [ ] Implement dynamic prompt resolution in runner
   - [ ] Support OpenAI Prompts API parameters
   - [ ] Add prompt template examples
   - [ ] Add comprehensive tests

2. **Complete Tool Approvals Integration** ✅
   - [x] Update `Runner` to detect approval requirements
   - [x] Implement approval workflow interruption (`ToolApprovalRequiredError` + `RunState`)
   - [x] Add approval callback (`WithApprovalHandler()` run option)
   - [x] Add `Runner.Resume()` for pause/resume workflow
   - [x] Create approval workflow examples (`examples/23_tool_approvals/`)
   - [x] Add comprehensive approval tests (15 test cases)

3. **Implement Model Provider Abstraction**
   - [ ] Create `models.Provider` interface
   - [ ] Implement `models.OpenAIProvider`
   - [ ] Implement `models.AnthropicProvider`
   - [ ] Implement `models.GoogleProvider`
   - [ ] Add provider configuration utilities
   - [ ] Update `Agent` to use `Provider` interface
   - [ ] Create multi-provider examples
   - [ ] Add provider integration tests

4. **Documentation & Examples**
   - [ ] Create `examples/XX_prompts_demo/`
   - [x] Create `examples/23_tool_approvals/`
   - [ ] Create `examples/XX_multi_provider/`
   - [ ] Update README.md with new features
   - [ ] Create feature-specific documentation
   - [ ] Add migration guide for v0.5.0 → v0.6.0

### Success Criteria

- ✅ All foundation APIs (v0.5.1) fully integrated with agent runtime
- ✅ Support for 3+ LLM providers (OpenAI, Anthropic, Google)
- ✅ Production-ready approval workflows
- ✅ Comprehensive examples for all new features
- ✅ 100% test coverage on new features
- ✅ Python SDK feature parity achieved

### Timeline

- Week 1-2: Prompts API integration
- Week 3-4: Tool Approvals integration
- Week 5-6: Model Provider abstraction
- Week 7-8: Examples, docs, and testing

---

### v1.0.0 - Stable Release 🎯
**Timeline**: Q2 2026
**Status**: Planned

#### Goals
- ✅ API stability guarantees
- ✅ 90%+ test coverage
- ✅ Performance benchmarks published
- ✅ Migration guides for all versions

---

### v1.1.0+ - Advanced Integrations 🔮
**Timeline**: Post-v1.0
**Status**: Future Planning

#### Planned Features

**Audio & Voice APIs**:
- Audio Transcriptions (speech-to-text)
- Audio Translations (multilingual audio)
- Audio Speech (text-to-speech)
- Voice agent capabilities
- Realtime API (WebSocket-based voice/text sessions)

**Batch Processing**:
- Batch API for cost-effective bulk operations (50% cost savings)
- Asynchronous agent processing
- Bulk evaluations and testing

**Images & Video**:
- DALL-E image generation tool
- Image editing and variations
- Video processing API (when stable)

**Advanced Integrations**:
- Beta Assistants API (alternative agent backend)
- Fine-Tuning API (custom agent models)
- Webhooks (event-driven architectures)
- Containers API (sandboxed code execution)

**Other Features**:
- Advanced MCP integrations
- Graders API (agent evaluation)

---

## Development Guidelines

### Code Quality Standards
- Test coverage: >85%
- All public APIs must have godoc comments
- Follow Go best practices and idioms
- Use semantic versioning strictly
- Comprehensive error handling

### Testing Strategy
- Unit tests for all core functionality
- Integration tests with real API (opt-in)
- Benchmark tests for performance tracking
- Fuzz tests for critical parsers

### Documentation Requirements
- API documentation (godoc)
- Migration guides for breaking changes
- Examples for all major features
- Architecture decision records (ADRs)
- Troubleshooting guides

---

## Contributing

We welcome contributions! Please:
1. Check open issues for tasks
2. Read [CONTRIBUTING.md](./CONTRIBUTING.md) (coming in v0.3.0)
3. Submit PRs with tests and documentation
4. Follow Go best practices
5. Ensure `make check` passes

---

## Questions?

- **Why start with zero dependencies?** Easy adoption, no build complexity, works everywhere Go works
- **Why not implement Batch/Realtime/Voice in v1.0?** Focus on stable core first, advanced features follow based on demand
- **Can I use SQLite from day one?** Use file-based sessions (zero deps) now, or wait for v0.3.0 for full database support
- **Will there be breaking changes?** Minimal after v1.0. We use semantic versioning and provide migration guides
- **How does this compare to Python SDK?** We aim for feature parity with unique Go advantages (type safety, performance, zero deps)

---

## Success Metrics

### Technical Metrics
- **Test Coverage**: >85% across all packages
- **Performance**: <100ms streaming latency, <1s tool execution
- **Reliability**: <0.1% error rate in production scenarios
- **API Stability**: Zero breaking changes after v1.0

### Community Metrics
- **Adoption**: 500+ GitHub stars by v1.0
- **Engagement**: <48h issue response time
- **Quality**: >95% example success rate
- **Documentation**: <5% documentation-related issues

---

**Last Updated**: 2026-02-08
**Current Version**: ## v0.5.2 (2026-02-07) ✅ COMPLETED

**Status**: Released  
**Focus**: Library cleanup, package restructuring, and Python SDK parity foundations

### Goals Achieved

1. ✅ **Remove Bloat** (~1,800 LOC removed)
   - Deleted `pkg/embeddings/`, `pkg/files/`, `pkg/vectorstore/` (out of SDK scope)
   - Simplified plugin architecture (merged into `session/` with build tags)
   - Removed dependent tools (`rag_tool.go`, `file_tool.go`)

2. ✅ **Package Restructuring** (Better API design)
   - Moved `pkg/mcp/` → `mcp/` (top level)
   - Moved `pkg/computer/` → `computer/` (top level)
   - Moved `pkg/diff/` → `diff/` (top level)
   - Updated 7 files with new import paths

3. ✅ **Python SDK Parity Foundations**
   - Created `prompts` package with `Prompt` and `DynamicPromptFunc` types
   - Implemented Tool Approvals (`NeedsApproval`, `ApprovalFunc`, `RequiresApproval()`)
   - Added approval types (`ApprovalRequest`, `ApprovalResponse`, `ApprovalHandler`)

### Breaking Changes

**Removed packages** (use alternatives):
- `pkg/embeddings/` → Use `github.com/openai/openai-go/v3` directly
- `pkg/files/` → Use `github.com/openai/openai-go/v3` directly
- `pkg/vectorstore/` → Use dedicated vector DB libraries

**Package moves** (update imports):
```go
// Old
import "github.com/MitulShah1/openai-agents-go/pkg/mcp"
import "github.com/MitulShah1/openai-agents-go/plugins/redis"

// New
import "github.com/MitulShah1/openai-agents-go/mcp"
import "github.com/MitulShah1/openai-agents-go/session"
store := session.NewRedisStore(...) // Build with: go build -tags redis
```

### Migration (Prompts API, Tool Approvals, Model Abstraction)
**SDK Version**: openai-go/v3 v3.16.0
