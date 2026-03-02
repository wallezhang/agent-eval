# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build                          # Build binary with version injection
make test                           # go test ./... -v
make vet                            # go vet ./...
make lint                           # go vet + golangci-lint
go test ./pkg/model/ -v             # Run tests for a single package
go test ./pkg/grader/ -run TestExactMatch -v  # Run a single test
make run                            # Run examples/simple/eval.yaml end-to-end
make license-check                  # Check all files have Apache-2.0 license header
make license-fix                    # Add missing Apache-2.0 license headers
```

## Architecture

This is a YAML-config-driven CLI tool for evaluating AI agents. Module: `github.com/wallezhang/agent-eval`.

### Execution Flow

```
CLI (cmd/run.go)
  → config.Load()          parse YAML, expand ${ENV_VARS}, glob-load task_files, apply defaults, validate
  → engine.New(suite).Execute(ctx)
      → agent.Create()     look up factory in registry by type name, instantiate
      → resolveGraders()   same registry pattern for each task's grader refs
      → buildWorkItems()   cartesian product: task × trials_per_task
      → scheduler.Run()    errgroup with SetLimit(concurrency) + rate.Limiter(rps)
          → runner.Run()   per work-item: agent.Execute → grader.Grade × N → weighted score
      → aggregateResults() group trials by task, compute pass@k / pass^k
  → store.SaveRun()        persist to SQLite
  → report.GenerateAll()   table (stdout) + JSON file + HTML file
```

### Registry Pattern (agent + grader + llm)

All three extension points use the same pattern: a package-level `map[string]Factory` populated via `init()` in each implementation file. New types are added by creating a file with `init() { Register("name", factory) }` — no central switch statement.

- `agent.Create(typeName, config)` → `Agent` interface (`Execute` + `Close`)
- `grader.Create(typeName, config)` → `Grader` interface (`Grade` + `Type`)
- `llm.Create(providerName, config)` → `Client` interface (`Complete` + `Close`)

The `llm` package is only used by the `llm` and `pairwise` graders, not by agents (agents call APIs directly).

### Scoring Model

A trial's final score is a **weighted average** across all grader results, but pass/fail is an **AND** — all graders must pass for the trial to pass. Individual grader errors don't abort the trial; they record score=0 and continue.

### Config Defaults Cascade

`EvalSuite.Defaults.Graders` and `TrialsPerTask` are applied to any task that doesn't override them. Grader weights default to 1.0. Config validation runs after defaults are applied.

### Storage

SQLite via `modernc.org/sqlite` (pure Go, no CGO). Task results and summary are stored as JSON text columns. The `compare` command supports ID prefix matching when looking up runs.

### HTML Reports

The template lives at `pkg/report/templates/report.html.tmpl` and is embedded via `//go:embed` in `pkg/report/html.go`. Custom template funcs: `pct`, `score`, `shortID`, `statusClass`.

### Key Design Decisions

- Runner returns `(*Trial, nil)` even on agent/grading errors — errors are recorded in the Trial struct, not propagated up. This keeps the scheduler simple.
- The `command` agent passes task prompt via stdin; the `command` grader passes a JSON payload via stdin and parses structured JSON from stdout (falling back to exit code).
- `pass@k` uses log-space arithmetic to avoid overflow with large n.
