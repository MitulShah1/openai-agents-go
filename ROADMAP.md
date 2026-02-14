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
Week 15-16│ v0.6.0 - Tool Approvals (Runner Integration) ✅
Week 17-18│ v0.7.0 - Model Abstraction & Prompts API ✅
Week 19-20│ v1.0.0 - Stable Release 🚀
Future    │ v1.1.0+ - Advanced Integrations
```

---

## Version Roadmap

### Completed Releases ✅

For detailed release notes, see [CHANGELOG.md](./CHANGELOG.md).

#### v0.6.0 - Tool Approvals Runner Integration (2026-02-09) ✅
- **Tool Approvals**: Full human-in-the-loop approval workflow integrated into all runner paths
  - `Run()`, `Stream()`, and `StreamWithResult()` check approvals before tool execution
  - Pause/resume via `ToolApprovalRequiredError` + `RunState` + `Runner.Resume()`
  - Inline approval via `WithApprovalHandler()` run option
  - Streaming support with `StreamEventApprovalRequired` and `ApprovalRequiredEvent`
  - Parallel batch safety (no partial execution)

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

#### v0.7.0 - Model Abstraction & Prompts API (2026-02-14) ✅
- **Model Provider Abstraction**: Pluggable LLM backends via `models.Model` and `models.ModelProvider` interfaces
  - `OpenAIProvider` built-in implementation
  - `MultiProvider` for prefix-based routing to multiple providers
  - `ModelSettings` with `Resolve()` merge logic
  - 3-tier resolution: Agent.ModelProvider → Runner.ModelProvider → Runner.Client
  - `NewRunnerWithProvider()` constructor for explicit control
- **Prompts API**: Dynamic prompt configuration via `Agent.Prompt` field
  - `prompts.Prompt` for static prompts
  - `prompts.DynamicPromptFunc` for runtime-resolved prompts
  - Runner integration in all execution paths
- **Examples**: `24_prompts_demo/`, `25_multi_provider/`
- **Documentation**: `docs/models.md`, `docs/prompts.md`

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

**Last Updated**: 2026-02-14  
**Current Version**: v0.7.0
