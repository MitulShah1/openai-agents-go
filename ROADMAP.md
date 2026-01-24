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
Week 10-12│ v0.4.0 - Streaming & Performance ✅
Week 13-14│ v0.5.0 - Feature Parity & Extensions
Week 15-16│ v1.0.0 - Stable Release
Future    │ v1.1.0+ - Advanced Integrations
```

---

## Version Roadmap

### Completed Releases ✅

For detailed release notes, see [CHANGELOG.md](./CHANGELOG.md).

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

### v0.5.0 - Feature Parity & Extensions 🚀
**Timeline**: Q1 2026
**Status**: Planned

#### Goals
Achieve feature parity with the Python SDK by implementing key missing features:

- **MCP Support**: Model Context Protocol integration for external tool/data connections
  - MCP server integration
  - Tool approval functions
  - Hosted MCP tool support
  
- **Computer Use Interface**: Browser and desktop automation capabilities
  - Computer abstraction (sync and async)
  - Operations: screenshot, click, scroll, type, keypress, drag, move
  - Support for mac, windows, ubuntu, and browser environments
  
- **Diff Application Logic**: Structured code change application
  - V4A diff parser
  - Apply patch editor and tool
  - Support for create and update modes
  
- **Enhanced Session Backends**: Advanced conversation management
  - Message compaction for long conversations
  - Redis session backend for distributed deployments
  - OpenAI Responses compaction-aware sessions

- **Embeddings & Vector Stores**: RAG capabilities via openai-go SDK
  - Embeddings API integration for semantic search
  - Vector Stores for managed RAG
  - File chunking strategies
  - Semantic search and knowledge retrieval

- **Files API**: File management capabilities
  - File upload and management
  - Document processing for agents
  - File attachments for tools
  - Content retrieval
  
- **Extensions Framework**: Experimental features and utilities
  - Handoff filters and prompts
  - Visualization tools
  - Experimental integrations

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

**Last Updated**: 2026-01-24
**Current Version**: v0.4.0
**Next Focus**: v0.5.0 - Feature Parity & Extensions
**SDK Version**: openai-go/v3 v3.16.0
