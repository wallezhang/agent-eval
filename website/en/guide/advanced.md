# Advanced Usage

This guide covers CI/CD integration, caching, checkpoints, tag filtering, hooks, and extending AgentEval with custom agents and graders.

## CI/CD Integration

Use the `--fail-under` flag to enforce a minimum pass rate in your CI pipeline. The command exits with a non-zero status code if the overall pass rate falls below the threshold.

### GitHub Actions Example

```yaml
name: Agent Evaluation
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  eval:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go install github.com/wallezhang/agent-eval@latest
      - run: agent-eval run -c eval.yaml --fail-under 0.8
        env:
          OPENAI_API_KEY: $&#123;&#123; secrets.OPENAI_API_KEY &#125;&#125;
```

The `--fail-under 0.8` flag means the pipeline fails if the overall pass rate drops below 80%. This acts as a quality gate -- preventing merges that degrade agent performance.

You can combine this with JSON output for further processing:

```bash
agent-eval run -c eval.yaml --fail-under 0.8
# Results are written to the output directory as summary.json
```

## Response Caching

Enable file-based caching to avoid redundant API calls during development and debugging.

```yaml
cache:
  enabled: true
  dir: .cache/agent-eval
```

The cache is content-addressed: identical prompts to the same agent configuration return cached responses. This is particularly useful when:

- Iterating on grader configurations without re-running the agent
- Debugging evaluation logic with consistent inputs
- Reducing API costs during development

To bypass the cache for a fresh run:

```bash
agent-eval run -c eval.yaml --no-cache
```

::: warning
Cache failures are non-fatal. If the cache directory is inaccessible or a cache entry is corrupted, AgentEval logs a warning and proceeds with a live API call.
:::

## Checkpoint Resume

Long-running evaluations can be resumed from where they left off if interrupted.

```bash
# First run gets interrupted (Ctrl+C, timeout, crash)
agent-eval run -c eval.yaml

# Resume from where it left off using the run ID
agent-eval run -c eval.yaml --resume <run-id>
```

The checkpoint system stores completed trial results in a local checkpoint store. When resuming, already-completed trials are loaded from the checkpoint and only remaining work items are executed. This avoids re-running expensive API calls.

You can find the run ID from the output of the initial run or by listing past runs:

```bash
agent-eval list
```

## Tag-based Filtering

Tags allow you to run subsets of tasks without modifying configuration files.

### Defining Tags

Add tags to individual tasks in your task YAML files:

```yaml
- id: math-addition
  name: "Simple Addition"
  tags: [math, easy]
  input:
    prompt: "What is 2 + 2?"
  expected:
    text: "4"

- id: code-generation
  name: "Generate Fibonacci"
  tags: [coding, hard]
  input:
    prompt: "Write a Fibonacci function in Python"
  expected:
    text: "def fibonacci(n):"
```

### Filtering at Runtime

```bash
# Only run tasks tagged "math"
agent-eval run -c eval.yaml --tags math

# Only run tasks tagged both "coding" and "easy"
agent-eval run -c eval.yaml --tags coding,easy

# Exclude tasks tagged "hard"
agent-eval run -c eval.yaml --exclude-tags hard
```

Tag filtering is applied after loading all task files, so you can maintain a comprehensive task set and select relevant subsets for different evaluation scenarios.

## Lifecycle Hooks

Hooks run shell commands before and after the evaluation run. They are useful for notifications, environment setup, or post-processing.

```yaml
hooks:
  before_run: "echo 'Starting evaluation at $(date)'"
  after_run: "curl -X POST https://slack.webhook/... -d '{\"text\": \"Eval complete\"}'"
```

Use cases for hooks:

- **Notifications** -- Send Slack or email alerts when evaluations complete
- **Environment setup** -- Start local services or prepare test data before the run
- **Post-processing** -- Upload results to a dashboard or trigger downstream pipelines

::: warning
Hook failures are non-fatal. If a hook command exits with an error, AgentEval logs a warning and continues execution. This prevents a broken notification script from blocking evaluations.
:::

## Custom Agent Development

AgentEval uses a registry pattern that makes it straightforward to add new agent types without modifying existing code.

### Steps

1. Create a new file in `pkg/agent/` (e.g., `pkg/agent/my_agent.go`)
2. Implement the `Agent` interface:

```go
type Agent interface {
    Execute(ctx context.Context, input TaskInput) (*AgentOutput, error)
    Close() error
}
```

3. Register the agent in an `init()` function:

```go
func init() {
    Register("my_agent", func(config map[string]any) (Agent, error) {
        // Parse config and return your agent implementation
        return &myAgent{}, nil
    })
}
```

No other files need modification. The registry pattern automatically discovers and makes your agent available as `type: my_agent` in configuration files.

## Custom Grader Development

Custom graders follow the same registry pattern as agents.

### Steps

1. Create a new file in `pkg/grader/` (e.g., `pkg/grader/my_grader.go`)
2. Implement the `Grader` interface:

```go
type Grader interface {
    Grade(ctx context.Context, task Task, output AgentOutput) (*GradeResult, error)
    Type() string
}
```

3. Register the grader:

```go
func init() {
    Register("my_grader", func(config map[string]any) (Grader, error) {
        return &myGrader{}, nil
    })
}
```

The grader is then available as `type: my_grader` in task and suite configurations.

## Token Usage and Cost Estimation

AgentEval automatically extracts token usage data from agent responses. This data is included in reports and can be used for cost estimation and latency analysis.

### How It Works

Token data is extracted from the `usage` field in agent output metadata. AgentEval supports both naming conventions:

| Provider | Input field | Output field |
|----------|-------------|--------------|
| OpenAI | `prompt_tokens` | `completion_tokens` |
| Anthropic | `input_tokens` | `output_tokens` |

### Reported Metrics

The generated reports include:

- **Total tokens** -- Sum of input and output tokens across all trials
- **Latency percentiles** -- P50, P90, and P99 of agent execution time (excludes grading time)
- **Per-trial breakdown** -- Token counts and latency for each individual trial

Latency percentiles are computed from the `AgentDurationMS` field, which measures only the time spent executing the agent -- grading time is excluded. This gives an accurate picture of agent performance for SLA assessment.

## Next Steps

- Review [Core Concepts](/en/guide/concepts) for the evaluation model and scoring logic
- See [Examples](/en/guide/examples) for ready-to-use configurations
- Return to the [Quick Start](/en/guide/quick-start) for initial setup
