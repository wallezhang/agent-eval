# Configuration Reference

Full YAML configuration reference for AgentEval.

## Complete Schema

```yaml
# Suite name (required)
name: "my-eval"
# Description (optional)
description: "Evaluation suite description"

# Agent configuration (required)
agent:
  type: openai | anthropic | http | command
  config:
    # Type-specific config (see Agent Types below)

# Default settings applied to all tasks (optional)
defaults:
  trials_per_task: 1          # Number of trials per task
  pass_threshold: 0.5         # Minimum score to pass
  graders:                    # Default graders for tasks
    - type: exact_match
      weight: 1.0
      config: {}

# Execution settings (optional)
execution:
  concurrency: 1              # Max parallel trials
  rate_limit_rps: 0           # Requests per second (0 = unlimited)
  timeout: "60s"              # Per-trial timeout
  retry:
    max_retries: 0            # Max retry attempts for agent errors
    initial_delay: "1s"       # Initial backoff delay
    max_delay: "30s"          # Maximum backoff delay

# Task file globs (at least one required if no inline tasks)
task_files:
  - tasks/*.yaml

# Inline tasks (alternative to task_files)
tasks:
  - id: task-1
    name: "Task Name"
    input:
      prompt: "Your prompt"

# Output configuration (optional)
output:
  format: table | json | html | all   # Default: table
  dir: ./results                       # Output directory

# Response cache (optional)
cache:
  enabled: false
  dir: .cache/agent-eval

# Lifecycle hooks (optional)
hooks:
  before_run: ""              # Shell command before evaluation
  after_run: ""               # Shell command after evaluation
```

## Agent Types

### openai

```yaml
agent:
  type: openai
  config:
    api_key: ${OPENAI_API_KEY}    # Required
    model: gpt-4                   # Default: gpt-4
    base_url: https://api.openai.com/v1  # Optional, for compatible APIs
    temperature: 0.0               # Default: 0.0
```

Compatible with any OpenAI-compatible API (e.g., vLLM, Ollama, Azure OpenAI) by setting `base_url`.

### anthropic

```yaml
agent:
  type: anthropic
  config:
    api_key: ${ANTHROPIC_API_KEY}  # Required
    model: claude-sonnet-4-20250514  # Default
    base_url: https://api.anthropic.com  # Optional
    temperature: 0.0               # Optional
    max_tokens: 4096               # Default: 4096
```

### http

```yaml
agent:
  type: http
  config:
    url: https://my-api.example.com/evaluate  # Required
    method: POST                              # Default: POST
    headers:                                  # Optional
      Authorization: "Bearer ${API_TOKEN}"
    body_template: ""                         # Optional custom body template
    response_path: "text"                     # JSON path in response, default: "text"
```

The HTTP agent sends a JSON body with `prompt`, `system`, `messages`, and `params` fields. Response text is extracted from the JSON path specified by `response_path`, falling back to raw response text.

### command

```yaml
agent:
  type: command
  config:
    command: python                  # Required
    args: ["eval_script.py"]         # Optional
    working_dir: ./scripts           # Optional
    timeout: "60s"                   # Default: 60s
    env:                             # Optional environment variables
      MY_VAR: "value"
```

The prompt is passed via stdin by default. If any argument contains <code v-pre>{{.Prompt}}</code>, the prompt is substituted into args instead.

## Grader Types

### exact_match

```yaml
- type: exact_match
  weight: 1.0
  config:
    ignore_case: false         # Case-insensitive comparison
    ignore_whitespace: false   # Ignore leading/trailing whitespace
```

Compares agent output text against `expected.text`. Score: 1.0 if match, 0.0 otherwise.

### contains

```yaml
- type: contains
  weight: 1.0
  config:
    ignore_case: false
    keywords: ["keyword1", "keyword2"]  # Additional required keywords
```

Checks that `expected.text` and all `keywords` are present in agent output. Score: matched / total.

### regex

```yaml
- type: regex
  weight: 1.0
  config:
    pattern: "\\d{3}-\\d{4}"    # Required, Go regex syntax
```

Tests pattern against agent output text. Score: 1.0 if match, 0.0 otherwise.

### json_match

```yaml
- type: json_match
  weight: 1.0
  config:
    ignore_case: false
```

Parses agent output as JSON and compares against `expected.fields`. Score: matched / total fields.

Task definition with json_match:
```yaml
- id: json-test
  input:
    prompt: "Return user info as JSON"
  expected:
    fields:
      name: "Alice"
      age: 30
  graders:
    - type: json_match
```

### command

```yaml
- type: command
  weight: 1.0
  config:
    command: python               # Required
    args: ["grade_script.py"]
    timeout: "60s"                # Default: 60s
```

Receives JSON via stdin: `{"task_id": "...", "agent_output": "...", "expected": {...}}`. Exit code 0 = pass, non-zero = fail. Optionally returns JSON via stdout: `{"score": 0.8, "pass": true, "reason": "..."}`.

### llm

```yaml
- type: llm
  weight: 1.0
  config:
    provider: openai              # Default: openai
    api_key: ${OPENAI_API_KEY}
    base_url: ""                  # Optional, for compatible APIs
    model: gpt-4
    rubric: |                     # Required
      Evaluate the response for:
      1. Correctness
      2. Completeness
```

Uses an LLM to grade the agent's output against a rubric. The LLM response is parsed for: `SCORE: <0.0-1.0>`, `PASS: <true/false>`, `REASON: <text>`.

### pairwise

```yaml
- type: pairwise
  weight: 1.0
  config:
    provider: openai
    api_key: ${OPENAI_API_KEY}
    base_url: ""                  # Optional
    model: gpt-4
    criteria: ""                  # Optional, default compares accuracy/completeness/helpfulness
    reference: ""                 # Optional, uses expected.text if not set
```

Compares agent output against a reference using LLM. Parses: `VERDICT: <A_BETTER|B_BETTER|TIE>`, `REASON: <text>`. Scores: B_BETTER=1.0, TIE=0.5, A_BETTER=0.0.

### constraint

```yaml
- type: constraint
  weight: 1.0
  config:
    checks:
      - pattern: "\\d+"
        must_match: true
      - pattern: "error"
        must_not_match: true
      - max_words: 100
      - min_words: 10
```

Applies structural constraints to agent output. ALL checks must pass. Score: passed / total checks.

## Task File Format

Task files are YAML arrays:

```yaml
- id: task-1                       # Required, unique
  name: "Task Name"                # Optional
  tags: [math, easy]               # Optional, for filtering
  input:
    prompt: "Your prompt here"     # Required (or messages)
    system: "System prompt"        # Optional
    messages:                      # Alternative to prompt
      - role: user
        content: "Hello"
      - role: assistant
        content: "Hi there"
    params:                        # Optional key-value params
      language: "python"
  expected:                        # Optional
    text: "Expected output"
    fields:                        # For json_match
      key: "value"
  graders:                         # Optional, overrides defaults
    - type: exact_match
  trials_per_task: 5               # Optional, overrides default
  step_limit: 10                   # Optional, expected max steps
```

## Environment Variable Expansion

Use `${ENV_VAR}` syntax in any string value:

```yaml
agent:
  type: openai
  config:
    api_key: ${OPENAI_API_KEY}
```

## Defaults Cascade

Suite-level defaults are applied to tasks that don't override them:

```yaml
defaults:
  trials_per_task: 3
  pass_threshold: 0.7
  graders:
    - type: exact_match

tasks:
  - id: task-1
    input:
      prompt: "..."
    # Inherits: trials_per_task=3, pass_threshold=0.7, graders=[exact_match]

  - id: task-2
    input:
      prompt: "..."
    trials_per_task: 5              # Overrides default
    graders:                        # Overrides default graders
      - type: contains
```
