# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build                          # Build binary with version injection
make build-frontend                 # Build Vue 3 frontend (npm install + npm run build)
make build-web                      # Build frontend + Go binary (full web build)
make test                           # go test ./... -v
make vet                            # go vet ./...
make lint                           # go vet + golangci-lint
go test ./pkg/model/ -v             # Run tests for a single package
go test ./pkg/grader/ -run TestExactMatch -v  # Run a single test
go test ./web/server/ -v            # Run web server tests
make run                            # Run examples/simple/eval.yaml end-to-end
make license-check                  # Check all files have Apache-2.0 license header
make license-fix                    # Add missing Apache-2.0 license headers
```

## Conventions

- All `.go` files must have the Apache-2.0 license header. Run `make license-check` to verify, `make license-fix` to add missing headers.
- Error handling: agent/grading errors are recorded in the `Trial` struct (score=0), not propagated up. Runner always returns `(*Trial, nil)`.
- Hook and cache failures are non-fatal — log a warning and continue.

## Architecture

This is a YAML-config-driven CLI tool for evaluating AI agents. Module: `github.com/wallezhang/agent-eval`.

### Package Overview

| Package | Role |
|---------|------|
| `cmd/` | CLI commands (Cobra): run, list, compare, init, server |
| `pkg/config/` | YAML loading, `${ENV}` expansion, glob task files, defaults cascade, validation |
| `pkg/model/` | Domain types (EvalSuite, Task, Trial, GradeResult) + pass@k/pass^k computation |
| `pkg/agent/` | Agent interface + adapters (openai, anthropic, http, command) |
| `pkg/grader/` | Grader interface + implementations (exact_match, contains, regex, json_match, command, llm, pairwise, constraint) |
| `pkg/engine/` | Orchestration: Engine, Scheduler (concurrent errgroup + rate limiter), Runner (per-trial execution with retry), Hooks |
| `pkg/cache/` | File-based response caching, wraps Agent interface |
| `pkg/storage/` | SQLite persistence (pure Go via `modernc.org/sqlite`, no CGO) + checkpoint store for resume |
| `pkg/report/` | Report generation: table (stdout), JSON, HTML (embedded template via `//go:embed`), diff |
| `pkg/llm/` | LLM client, used only by `llm` and `pairwise` graders (agents call APIs directly) |
| `web/server/` | Web UI backend: Chi router, REST API handlers, SSE streaming, RunManager, Service layer |
| `web/frontend/` | Web UI frontend: Vue 3 SPA (TypeScript, Naive UI, CodeMirror 6, Pinia) |
| `web/embed.go` | `//go:embed` for embedding frontend dist into Go binary |

### Execution Flow

```
CLI (cmd/run.go)
  → config.Load()           → engine.New(suite).Execute(ctx)
      → agent.Create()      → [cache.Wrap()]  → resolveGraders()
      → [filterTasksByTags()] → buildWorkItems() → [loadCheckpoint()]
      → [hooks.BeforeRun()]
      → scheduler.Run()     → runner.Run() per work item
      → [hooks.AfterRun()]  → aggregateResults()
  → store.SaveRun()         → report.GenerateAll()
```

Steps in `[]` are optional (enabled by config flags).

### Registry Pattern (agent + grader + llm)

All three extension points use the same pattern: a package-level `map[string]Factory` populated via `init()` in each implementation file. New types are added by creating a file with `init() { Register("name", factory) }` — no central switch statement.

- `agent.Create(typeName, config)` → `Agent` interface (`Execute` + `Close`)
- `grader.Create(typeName, config)` → `Grader` interface (`Grade` + `Type`)
- `llm.Create(providerName, config)` → `Client` interface (`Complete` + `Close`)

### Adding a New Agent or Grader

1. Create a new file in `pkg/agent/` or `pkg/grader/` (e.g., `pkg/grader/my_grader.go`)
2. Implement the interface (`Agent` or `Grader`)
3. Add `func init() { Register("my_grader", newMyGrader) }` — the factory function receives `map[string]any` config
4. No other files need modification — the registry pattern handles discovery

### Key Invariants

These are non-obvious rules that must be preserved across changes:

- **Scoring**: A trial's final score is a weighted average across grader results, but pass/fail is an AND — all graders must pass for the trial to pass.
- **Error propagation**: Runner returns `(*Trial, nil)` even on errors. Errors go into `Trial.Error`, keeping the scheduler simple.
- **Retry scope**: Exponential backoff retries only agent execution errors, never grading failures.
- **Config defaults**: `EvalSuite.Defaults.Graders` and `TrialsPerTask` are applied to tasks that don't override them. Validation runs after defaults are applied.
- **Latency tracking**: `AgentDurationMS` measures agent execution time only (excludes grading). Latency percentiles use this field when available.
- **Token extraction**: Usage data is extracted from `AgentOutput.Metadata["usage"]`, supporting both OpenAI field names (`input_tokens`/`output_tokens`) and Anthropic field names (`prompt_tokens`/`completion_tokens`).
- **pass@k**: Uses log-space arithmetic to avoid overflow with large n.
- **Stdin protocols**: `command` agent passes prompt via stdin. `command` grader passes JSON payload via stdin (`task_id`, `agent_output`, `expected`).

### Web UI Architecture

The `server` subcommand starts a web UI (single binary, frontend embedded via `go:embed`).

**Backend** (`web/server/`):
- Chi router with REST APIs scoped under `/api/projects/{name}/...`
- `ProjectRegistry` manages `~/.agent-eval/projects.json` (multi-project support)
- `Service` bridges HTTP handlers to `pkg/*` (config, engine, storage)
- `RunManager` tracks concurrent evaluation runs, each with its own context and SSE event channel
- SSE (`text/event-stream`) for real-time run progress
- Static file serving with SPA fallback for client-side routing

**Frontend** (`web/frontend/`):
- Vue 3 + TypeScript + Vite, Naive UI components, Pinia state management
- CodeMirror 6 for YAML config editing
- Pages: Dashboard, Configurations (CRUD + editor), Runs (list + SSE detail), Results, Settings

**Key patterns**:
- Filename sanitization on all config CRUD endpoints (path traversal protection)
- RunManager uses `sync.Once` + `done` channel to prevent send-on-closed-channel races
- SQLite WAL mode supports concurrent writes from multiple active runs
