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
Week 1-2  │ v0.1.0 - Core Foundation
Week 3-4  │ v0.2.0 - Guardrails & Sessions  
Week 5-6  │ v0.3.0 - Tracing & Observability
Week 7-8  │ v0.4.0 - Advanced Features
Week 9-10 │ v1.0.0 - Stable Release
Future    │ v1.1.0+ - Advanced Integrations
```

---

## Version Roadmap

### v0.1.0 - Core Foundation 🏗️ ✅ COMPLETE

**Timeline**: Week 1-2  
**Status**: ✅ Complete  
**Dependencies**: Zero external dependencies

#### Features
- ✅ Enhanced Agent configuration (temperature, max_tokens, lifecycle hooks)
- ✅ RunConfig for execution control (max_turns, timeout, debug mode)
- ✅ Custom error types for better error handling
- ✅ Usage tracking (tokens, costs)
- ✅ Execution steps recording
- ✅ Function schema generation from Go structs
- ✅ Improved tool execution with context
- ✅ **Structured Outputs** - JSON schema builder with fluent API
  - Complete JSON schema support (objects, arrays, primitives, validation)
  - OpenAI Structured Outputs integration
  - Response format control (text vs JSON schema)
  - Modular `internal/jsonschema` package

#### Use Cases
- Simple chatbots
- Basic tool-calling agents
- Single-turn interactions
- Development and testing

---

### v0.2.0 - Guardrails & Sessions 🛡️ ✅ SUBSTANTIALLY COMPLETE

**Timeline**: Week 3-4  
**Status**: ✅ Substantially Complete (Core features done, examples/docs pending)  
**Dependencies**: Zero external dependencies

#### Features

**Guardrails** (Input/Output Validation):
- ✅ Guardrail framework with pluggable validators
- ✅ OpenAI Moderation API integration (✅ v0.2.1)
- ✅ PII detection (email, phone, SSN, credit card)
- ✅ URL filtering (blocklist/allowlist)
- ✅ Custom regex-based validation
- ✅ Tripwire support (halt on failure)

**Sessions** (Conversation Persistence):
- ✅ Session interface for pluggable backends
- ✅ In-Memory session (thread-safe, zero deps)
- ✅ File-Based session (JSON files with atomic writes, zero deps)
- ✅ OpenAI Conversations API session (cloud-based persistence with mock implementation)
- ✅ Automatic history management (load/save integrated in Runner)

**What's Complete**:
- 32 new tests passing (20 guardrail + 12 session)
- Zero new dependencies added
- Runner.Run signature updated (breaking change: added session parameters)
- All existing tests and examples fixed
- ✅ v0.2.1: OpenAI Moderation API guardrail with 13 categories
- ✅ v0.2.1: Examples 08, 09, 10 demonstrating production usage
- ✅ v0.2.1: Comprehensive documentation and godoc

**Completed** (v0.2.1 released):
- ✅ OpenAI Moderation API guardrail
- ✅ All examples (08, 09, 10)
- ✅ Documentation updates (README, ROADMAP, godoc)

#### Use Cases
- Multi-turn conversations
- Chatbots with memory
- Content moderation
- Compliance checks (PII protection)
- Production deployments

---

### v0.3.0 - Enhanced Guardrails, Database Session Backends & Tracing 🛡️��🔍

**Timeline**: Week 5-6  
**Status**: Planned  
**Dependencies**: Optional session backends (SQLite, Redis, PostgreSQL drivers)

#### Features

**Enhanced Guardrails** (Advanced Security & Validation) - NEW:
- ⏳ **Content Length Guardrail**: Min/max character/word/line limits
- ⏳ **Rate Limiting Guardrail**: Token bucket algorithm for API protection
  - Per-user, per-IP, or global limits
  - Configurable time windows and burst sizes
  - Thread-safe concurrent execution
- ⏳ **Profanity Detection**: Pattern-based toxicity detection
  - Comprehensive profanity word lists
  - Leetspeak normalization
  - Configurable severity levels
- ⏳ **Prompt Injection Detection**: LLM security guardrail
  - Detect instruction override attempts
  - Role manipulation prevention
  - Delimiter and encoding attack detection
- ⏳ **Secrets Detection**: Credential leakage prevention
  - API keys (AWS, OpenAI, GitHub, etc.)
  - Private keys (RSA, SSH, PGP)
  - High-entropy string detection
  - JWT token patterns

**Guardrail Composition** (Advanced Features) - NEW:
- ⏳ **Chaining Support**: Combine multiple guardrails
  - Sequential mode (short-circuit on failure)
  - Parallel mode (collect all results)
  - StopOnFirst mode (early exit on pass)
- ⏳ **Async Validation**: Timeout and cancellation support
  - Context-aware execution
  - Graceful degradation
- ⏳ **Basic Metrics Collection**: Guardrail telemetry
  - Success/failure counters
  - Tripwire statistics
  - Average latency tracking

**Session Backends** (Production-Ready Persistence):
- ⏳ SQLite session backend (file-based database)
  - Pure Go implementation (`modernc.org/sqlite`) or CGO (`github.com/mattn/go-sqlite3`)
  - SQL schema with indexes for performance
  - Connection pooling support
  - Migration system
- ⏳ Redis session backend (distributed/scalable)
  - `github.com/redis/go-redis/v9` integration
  - Connection pooling and retry logic
  - TTL/expiry support for auto-cleanup
  - Clustering support
- ⏳ PostgreSQL session backend (enterprise-grade)
  - `github.com/lib/pq` or `github.com/jackc/pgx` integration
  - JSONB column type for message storage
  - Full-text search capability
  - Partitioning support for scale
- ⏳ Session options and utilities:
  - Pagination/limit support for large conversations
  - Compression (gzip) for storage efficiency
  - Encryption wrapper for sensitive data
  - Session migration tools between backends

**Tracing** (Observability):
- ⏳ Basic tracing framework with spans
- ⏳ Console trace processor (stdout)
- ⏳ OpenTelemetry integration (optional)
- ⏳ Automatic tracing of operations (LLM, tools, sessions, guardrails)

#### Use Cases
- LLM security (prompt injection prevention)
- Credential protection (secrets detection)
- Content moderation (profanity, PII)
- API rate limiting and resource protection
- Complex validation workflows (chaining)
- Production deployments with database persistence
- High-scale distributed systems (Redis)
- Enterprise applications (PostgreSQL)
- Multi-server/containerized environments
- Long-term conversation storage and analytics
- GDPR compliance with encryption
- Observability and debugging with tracing

---

### v0.4.0 - Streaming & Advanced Features 🚀

**Timeline**: Week 7-8  
**Status**: Planned  
**Dependencies**: Optional SQLite driver (`modernc.org/sqlite`)

#### Features

**Streaming Support**:
- ⏳ Token-by-token response streaming
- ⏳ Server-Sent Events (SSE) support
- ⏳ Stream cancellation and error handling

**Advanced Features**:
- ⏳ Parallel tool execution
- ⏳ Advanced handoff patterns:
  - Conditional handoffs
  - Parallel agent execution
  - Sequential agent chains
- ⏳ 10+ comprehensive examples
- ⏳ Full documentation
- ⏳ Performance benchmarks

**Additional Guardrails** (Optional):
- ⏳ Sentiment analysis guardrail
- ⏳ Language detection guardrail
- ⏳ SQL injection detection
- ⏳ Data Loss Prevention (DLP) with comprehensive patterns
- ⏳ Conditional guardrails (context-based)
- ⏳ Validation caching (LRU with TTL)

#### Use Cases
- High-performance real-time agents
- Complex multi-agent workflows
- Production applications at scale
- Advanced content moderation

---

### v1.0.0 - Stable Release 🎯

**Timeline**: Week 9-10  
**Status**: Planned  
**Dependencies**: Same as v0.4.0

#### Goals
- ✅ API stability guarantees
- ✅ 90%+ test coverage
- ✅ Performance benchmarks published
- ✅ Migration guides for all versions
- ✅ Comprehensive documentation
- ✅ Production-ready examples
- ✅ Community feedback incorporated

#### Deliverables
- Stable v1.0.0 release
- Full API documentation
- Migration guides
- Performance benchmarks
- Security audit (if applicable)
- Production deployment guide

---

### v1.1.0+ - Advanced Integrations 🔮

**Timeline**: Post-v1.0 (Based on community demand)  
**Status**: Future

#### Planned Features
- 🔮 **MCP (Model Context Protocol)**: Integration with MCP servers for dynamic tool discovery
- 🔮 **Realtime API**: WebSocket-based real-time agent interactions
- 🔮 **Voice Support**: Voice input/output for agents
- 🔮 **MySQL Session Backend**: For users preferring MySQL over PostgreSQL
- 🔮 **Batch API Support**: Process multiple requests in batch mode
- 🔮 **Fine-tuned Model Support**: Integration with custom fine-tuned models
- 🔮 **AI-Powered Guardrails**: Hallucination detection, NSFW detection, jailbreak detection (requires external APIs)
- 🔮 **Advanced Metrics**: Prometheus exporter, OpenTelemetry full integration
- 🔮 **Multi-modal Support**: Image, audio, video processing

---

## Contributing

We welcome contributions! Please:
1. Check open issues for tasks
2. Read CONTRIBUTING.md (coming in v0.3.0)
3. Submit PRs with tests
4. Follow Go best practices

---

## Questions?

- **Why start with zero dependencies?** Easy adoption, no build complexity, works everywhere Go works
- **Why not implement MCP/Realtime/Voice in v1.0?** Focus on stable core first, advanced features can follow
- **Can I use SQLite from day one?** No, but you can use file-based sessions (zero deps) or wait for v0.3.0
- **Will there be breaking changes?** Minimal. We'll use semantic versioning and provide migration guides

---

**Last Updated**: 2026-01-16  
**Current Focus**: v0.2.2 - Conversations API Session Backend (Mock Implementation Ready)


