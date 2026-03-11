# agent-eval

[![Go Reference](https://pkg.go.dev/badge/github.com/wallezhang/agent-eval.svg)](https://pkg.go.dev/github.com/wallezhang/agent-eval)
[![Go Report Card](https://goreportcard.com/badge/github.com/wallezhang/agent-eval)](https://goreportcard.com/report/github.com/wallezhang/agent-eval)
[![License](https://img.shields.io/github/license/wallezhang/agent-eval)](LICENSE)

A general-purpose AI agent evaluation framework. YAML-config-driven, supporting multiple agent types and grading strategies, with SQLite persistence and table/JSON/HTML report output.

Methodology inspired by Anthropic's [Demystifying Evals for AI Agents](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents).

[中文文档](README_zh.md)

## Core Highlights

- **Reliability Metrics (pass@k / pass^k)** — Implements Anthropic's methodology for measuring both agent capability ceiling and reliability across repeated trials
- **8 Built-in Graders** — exact_match, contains, regex, json_match, command, llm, pairwise, constraint — covering deterministic checks, LLM rubric grading, A/B comparison, and compliance rules
- **Token/Cost Tracking** — Automatically extracts token usage from agent metadata and estimates cost per task and run
- **Latency Percentiles (P50/P90/P99)** — Measures agent response time distribution for SLA assessment
- **Response Caching** — File-based caching avoids redundant API calls when iterating on grader logic
- **CI/CD Integration** — `--fail-under` flag enables pass rate gating in automated pipelines, with machine-readable `summary.json` output
- **Tag-based Filtering** — Run subsets of tasks by tag (e.g., `--tags safety` or `--exclude-tags slow`)
- **Retry with Backoff** — Configurable exponential backoff for transient API failures
- **Checkpoint Resume** — Resume interrupted evaluations from where they left off with `--resume`
- **Lifecycle Hooks** — Execute shell commands before/after runs, tasks, or trials for custom logic
- **4 Agent Adapters** — OpenAI, Anthropic, HTTP, Command — plus a registry for custom agents
- **4 Report Formats** — Table (stdout), JSON, HTML, Diff comparison

## Installation

```bash
go install github.com/wallezhang/agent-eval@latest
```

Or build from source:

```bash
git clone https://github.com/wallezhang/agent-eval.git
cd agent-eval
make build
```

## Quick Start

### 1. Initialize a Project

```bash
agent-eval init my-eval
cd my-eval
```

Generated structure:

```
my-eval/
├── eval.yaml          # Evaluation config
├── tasks/
│   └── sample.yaml    # Sample tasks
└── results/           # Report output directory
```

### 2. Edit Configuration

`eval.yaml`:

```yaml
name: "my-eval"
description: "Agent evaluation suite"

agent:
  type: openai
  config:
    model: gpt-4
    api_key: ${OPENAI_API_KEY}
    temperature: 0.0
  cost_per_input_token: 0.00003     # Optional, for cost estimation
  cost_per_output_token: 0.00006

defaults:
  trials_per_task: 3
  graders:
    - type: exact_match
      config:
        ignore_case: true

execution:
  concurrency: 4
  rate_limit_rps: 10
  timeout: 60s
  max_retries: 2          # Retry on transient errors (default 0)
  retry_delay: 1s         # Initial retry delay (default 1s)

cache:                      # Optional, response caching
  enabled: true
  dir: .cache/
  ttl: 24h

task_files:
  - tasks/*.yaml

output:
  format: all       # table | json | html | all
  dir: ./results
```

`tasks/sample.yaml`:

```yaml
- id: capital-of-france
  name: "Capital of France"
  tags: [geography]
  input:
    prompt: "What is the capital of France? Answer with just the city name."
  expected:
    text: "Paris"
```

### 3. Run Evaluation

```bash
agent-eval run -c eval.yaml
```

Sample output:

```
=== Evaluation Report: my-eval ===
Agent: openai | Run ID: a1b2c3d4
Duration: 3250ms

TASK              PASS  FAIL  ERR  AVG SCORE  PASS@K  PASS^K  P50ms  P90ms  P99ms
----              ----  ----  ---  ---------  ------  ------  -----  -----  -----
Capital of France 3     0     0    1.000      1.000   1.000   980    1050   1080

--- Summary ---
Tasks: 1 | Trials: 3 (passed: 3, failed: 0, error: 0)
Overall Pass Rate: 100.0% | Avg Score: 1.000
Avg pass@k: 1.000 | Avg pass^k: 1.000

--- Token Usage ---
Total Input Tokens: 150 | Total Output Tokens: 18 | Total: 168
Estimated Cost: $0.0056
```

### 4. CI/CD Usage

```bash
# Fail the pipeline if pass rate drops below 80%
agent-eval run -c eval.yaml --fail-under 0.8

# Run only safety-tagged tasks
agent-eval run -c eval.yaml --tags safety

# Resume an interrupted run
agent-eval run -c eval.yaml --resume <run-id>
```

A `summary.json` file is always written alongside reports for machine consumption.

## CLI Commands

| Command | Description |
|---------|-------------|
| `agent-eval run -c <config>` | Run an evaluation suite |
| `agent-eval list [--db path]` | List historical runs |
| `agent-eval compare <runA> <runB>` | Compare two runs |
| `agent-eval init [directory]` | Initialize an evaluation project |

### run

```bash
agent-eval run -c eval.yaml [flags]
```

| Flag | Description | Default |
|------|-------------|---------|
| `-c, --config` | Config file path | `eval.yaml` |
| `--db` | SQLite database path | `<output_dir>/agent-eval.db` |
| `--verbose` | Enable verbose logging | `false` |
| `--fail-under` | Minimum pass rate (0.0-1.0), exit code 1 if below | `0` (disabled) |
| `--tags` | Only run tasks matching these tags (comma-separated) | |
| `--exclude-tags` | Exclude tasks matching these tags (comma-separated) | |
| `--no-cache` | Bypass response cache | `false` |
| `--resume` | Resume a previous run by run ID | |

### list

```bash
agent-eval list [--db results/agent-eval.db]
```

```
ID        SUITE        AGENT    TASKS  PASS RATE  DURATION  DATE
eefc2b36  simple-eval  command  3      33.3%      12ms      2026-02-27 23:31
```

### compare

```bash
agent-eval compare eefc2b36 b3c4d5e6
```

Supports ID prefix matching. Outputs per-task score comparison with regression/improvement annotations.

### init

```bash
agent-eval init my-project
```

Generates `eval.yaml` and `tasks/sample.yaml` templates in the specified directory.

## Configuration Reference

### Full YAML Structure

```yaml
name: "suite-name"                    # Required, suite name
description: "..."                    # Optional, description

agent:                                # Required, agent under test
  type: openai                        # Agent type
  config:                             # Type-specific config
    model: gpt-4
    api_key: ${OPENAI_API_KEY}        # Supports env var expansion
  cost_per_input_token: 0.00003       # Optional, for cost estimation
  cost_per_output_token: 0.00006      # Optional, for cost estimation

defaults:                             # Optional, global defaults
  trials_per_task: 3                  # Trials per task (default 1)
  pass_threshold: 0.5                 # Pass threshold (default 0.5)
  graders:                            # Default graders
    - type: exact_match
      weight: 1.0                     # Weight (default 1.0)
      config: {}

execution:                            # Optional, execution control
  concurrency: 4                      # Concurrency (default 1)
  rate_limit_rps: 10                  # Requests per second limit (0=unlimited)
  timeout: 60s                        # Per-trial timeout
  max_retries: 2                      # Retries on transient errors (default 0)
  retry_delay: 1s                     # Initial retry delay, doubles each attempt (default 1s)

cache:                                # Optional, response caching
  enabled: true                       # Enable/disable cache (default false)
  dir: .cache/                        # Cache directory (default .cache/)
  ttl: 24h                            # Cache entry TTL (default 24h)

hooks:                                # Optional, lifecycle hooks (shell commands)
  before_run: "echo 'Starting run'"
  after_run: "curl -X POST https://slack.example.com/webhook -d @-"
  before_task: ""
  after_task: ""
  before_trial: ""
  after_trial: ""

task_files:                           # Optional, external task files (glob supported)
  - tasks/*.yaml

tasks:                                # Optional, inline task definitions
  - id: task-1
    name: "Task name"
    tags: [tag1, tag2]
    trials_per_task: 5                # Can override default
    step_limit: 10                    # Optional, expected max steps for efficiency tracking
    input:
      prompt: "..."
      system: "..."                   # Optional, system prompt
      messages:                       # Optional, multi-turn conversation
        - role: user
          content: "..."
    expected:
      text: "Expected text"
      fields:                         # JSON field matching
        name: "value"
    graders:                          # Can override default graders
      - type: llm
        weight: 0.5
        config:
          rubric: "Grading criteria"

output:                               # Optional, report config
  format: all                         # table | json | html | all
  dir: ./results                      # Output directory
```

### Agent Types

| Type | Description | Required Config |
|------|-------------|-----------------|
| `openai` | OpenAI Chat Completions API | `api_key` |
| `anthropic` | Anthropic Messages API | `api_key` |
| `http` | Generic HTTP API | `url` |
| `command` | External command (stdin/stdout) | `command` |

**openai**

```yaml
agent:
  type: openai
  config:
    api_key: ${OPENAI_API_KEY}
    base_url: https://api.openai.com/v1   # Optional
    model: gpt-4                           # Optional, default gpt-4
    temperature: 0.0                       # Optional, default 0.0
  cost_per_input_token: 0.00003            # Optional, for cost reports
  cost_per_output_token: 0.00006
```

**anthropic**

```yaml
agent:
  type: anthropic
  config:
    api_key: ${ANTHROPIC_API_KEY}
    base_url: https://api.anthropic.com    # Optional
    model: claude-sonnet-4-20250514       # Optional
    temperature: 0.0
    max_tokens: 4096
  cost_per_input_token: 0.000003
  cost_per_output_token: 0.000015
```

**http**

```yaml
agent:
  type: http
  config:
    url: http://localhost:8080/api/chat
    method: POST                           # Optional, default POST
    headers:
      Authorization: "Bearer ${TOKEN}"
    response_path: text                    # Optional, extract field from JSON response
```

**command**

```yaml
agent:
  type: command
  config:
    command: python
    args: ["-m", "my_agent"]
    working_dir: /path/to/project      # Optional, working directory for command execution
    timeout: 120s
    env:
      MODEL_PATH: /path/to/model
```

The task `prompt` is passed via stdin by default, with stdout as agent output. If `args` contains `{{.Prompt}}`, the prompt is substituted into the arguments instead of being passed via stdin:

```yaml
agent:
  type: command
  config:
    command: echo
    args: ["{{.Prompt}}"]
```

### Grader Types

| Type | Description | Use Case |
|------|-------------|----------|
| `exact_match` | Exact string matching | Q&A with definite answers |
| `contains` | Keyword presence check | Semi-structured output |
| `regex` | Regular expression matching | Formatted output validation |
| `json_match` | JSON field value matching | API response validation |
| `command` | External command grading | Coding agents (unit tests) |
| `llm` | LLM grading + rubric | Open-ended output evaluation |
| `pairwise` | A/B pairwise comparison | Model comparison |
| `constraint` | Compliance/policy checks | Safety, PII, word limits |

**exact_match**

```yaml
graders:
  - type: exact_match
    config:
      ignore_case: true        # Case insensitive
      ignore_whitespace: true  # Trim whitespace
```

**contains**

```yaml
graders:
  - type: contains
    config:
      ignore_case: true
      keywords: ["keyword1", "keyword2"]  # All keywords must match to pass
```

**regex**

```yaml
graders:
  - type: regex
    config:
      pattern: "^\\d{3}-\\d{4}$"
```

**json_match**

```yaml
graders:
  - type: json_match
    config:
      ignore_case: true
# Expected field values should be defined in the task's expected.fields
```

**command**

```yaml
graders:
  - type: command
    config:
      command: python
      args: ["-m", "pytest", "tests/"]
      timeout: 60s
```

The grading command receives JSON via stdin (containing `task_id`, `agent_output`, `expected`) and can return JSON via stdout (containing `score`, `pass`, `reason`), or simply use exit code 0/non-zero to indicate pass/fail.

**llm**

```yaml
graders:
  - type: llm
    weight: 0.5
    config:
      provider: openai           # openai | anthropic
      api_key: ${OPENAI_API_KEY}
      base_url: https://api.openai.com/v1  # Optional, for OpenAI-compatible services
      model: gpt-4
      rubric: |
        Evaluation criteria:
        1. Answer accuracy
        2. Conciseness
```

**pairwise**

```yaml
graders:
  - type: pairwise
    config:
      provider: openai
      api_key: ${OPENAI_API_KEY}
      base_url: https://api.openai.com/v1  # Optional, for OpenAI-compatible services
      criteria: "Which answer is more accurate and complete?"
      reference: "Reference answer text"     # Optional, defaults to expected.text
```

**constraint**

The constraint grader checks agent output against a set of compliance rules. All checks must pass for the trial to pass; the score reflects the fraction of checks passed.

```yaml
graders:
  - type: constraint
    config:
      checks:
        - name: "no_pii"
          pattern: '(?i)(ssn|credit card|social security)'
          must_not_match: true           # Fail if pattern is found
        - name: "has_disclaimer"
          pattern: "(?i)disclaimer"
          must_match: true               # Fail if pattern is NOT found
        - name: "word_limit"
          max_words: 500                 # Fail if output exceeds word count
        - name: "min_length"
          min_words: 10                  # Fail if output is too short
```

### Weighted Composite Scoring

Multiple graders can be configured per task with `weight`:

```yaml
graders:
  - type: exact_match
    weight: 2.0                    # Weight 2.0
    config: { ignore_case: true }
  - type: llm
    weight: 0.5                    # Weight 0.5
    config: { rubric: "..." }
```

Final score = weighted average. Pass condition: all graders must pass.

### Response Caching

When `cache.enabled` is true, agent responses are cached to disk based on a hash of (agent type, agent config, task input). This is useful when iterating on grader logic — you can re-run evaluations without re-calling the agent API.

```yaml
cache:
  enabled: true
  dir: .cache/       # Cache directory
  ttl: 24h           # Time-to-live for cache entries
```

Use `--no-cache` to bypass the cache for a single run.

### Lifecycle Hooks

Hooks let you run shell commands at key points during evaluation. Each hook receives a JSON context object on stdin with relevant data (run info, task info, trial info depending on the hook type).

```yaml
hooks:
  before_run: "echo 'Starting evaluation'"
  after_run: "python scripts/notify.py"
  before_trial: ""
  after_trial: "python scripts/log_trial.py"
```

### Retry Configuration

Transient API errors (rate limits, network timeouts) are automatically retried with exponential backoff when `max_retries > 0`. Only agent execution errors trigger retries — grading failures do not.

```yaml
execution:
  max_retries: 3        # Max retry attempts per trial
  retry_delay: 1s       # Initial delay, doubles each attempt (1s → 2s → 4s)
```

## Key Metrics

### pass@k

Probability of at least 1 pass (optimistic metric, measures agent capability ceiling):

```
pass@k = 1 - C(n-c, k) / C(n, k)
```

### pass^k

Probability of all k passes (strict metric, measures agent reliability):

```
pass^k = C(c, k) / C(n, k)
```

Where `n` = total trials, `c` = pass count, `k` = sample size.

### Latency Percentiles

Each task result includes P50, P90, and P99 latency percentiles computed from agent execution times (excluding grading time). This helps assess whether the agent meets latency SLAs.

### Token Usage & Cost

Token counts (input/output) are automatically extracted from agent metadata when available (OpenAI and Anthropic agents report this). When `cost_per_input_token` and `cost_per_output_token` are configured, estimated cost is computed per task and for the entire run.

### Step Count

Agent step counts are extracted from metadata or transcripts. Combined with the optional `step_limit` task field, this measures agent efficiency — how many steps the agent takes relative to expectations.

## Report Formats

### Table (stdout)

Terminal output with per-task pass/fail/error counts, average score, pass@k, pass^k, P50/P90/P99 latency, and token usage summary.

### JSON

Full structured data with all trial details, grading results, and metadata. Output to `results/<suite>-<id>.json`.

### HTML

Styled visual report with summary cards (pass rate, avg score, token usage, cost), task result tables with latency and usage columns, and trial detail tables with step counts. Output to `results/<suite>-<id>.html`.

### Summary JSON

Machine-readable `summary.json` is always written to the output directory, containing the run summary for CI/CD integration.

## Project Structure

```
agent-eval/
├── main.go
├── go.mod
├── Makefile
├── cmd/                          # CLI commands
│   ├── root.go
│   ├── run.go
│   ├── list.go
│   ├── compare.go
│   └── init.go
├── pkg/
│   ├── model/                    # Domain models + metric computation
│   ├── config/                   # YAML config loading & validation
│   ├── agent/                    # Agent interface & adapters
│   ├── grader/                   # Grader interface & implementations
│   ├── engine/                   # Evaluation engine (concurrent scheduling)
│   ├── cache/                    # Response caching
│   ├── storage/                  # Result persistence (SQLite / in-memory)
│   ├── report/                   # Report generation (table / JSON / HTML / diff)
│   └── llm/                      # LLM client (used by graders)
├── examples/
│   ├── simple/                   # Command agent example
│   ├── multi-grader/             # Multi-grader example
│   └── coding-agent/             # Coding agent example
└── templates/
    └── report.html.tmpl
```

## Extending

Agents and graders are extended via factory registries. Register in `init()`:

```go
// Register a custom agent
agent.Register("my-agent", func(config map[string]any) (agent.Agent, error) {
    return &MyAgent{config: config}, nil
})

// Register a custom grader
grader.Register("my-grader", func(config map[string]any) (grader.Grader, error) {
    return &MyGrader{config: config}, nil
})
```

Agent interface:

```go
type Agent interface {
    Execute(ctx context.Context, input model.TaskInput) (*model.AgentOutput, error)
    Close() error
}
```

Grader interface:

```go
type Grader interface {
    Grade(ctx context.Context, input GradeInput) (*model.GradeResult, error)
    Type() string
}
```

## Development

```bash
make build        # Build binary
make test         # Run tests
make vet          # Static analysis
make lint         # golangci-lint (requires installation)
make clean        # Clean build artifacts
```

Run examples:

```bash
make run
# Equivalent to: go run . run -c examples/simple/eval.yaml
```
