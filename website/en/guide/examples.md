# Examples

This page shows practical evaluation configurations for common use cases. Each example is self-contained and can be adapted to your needs.

## Simple: Command Agent

The simplest way to get started is with a `command` agent that runs a shell command and checks its output. This is useful for testing grader configurations without needing API keys.

**eval.yaml:**

```yaml
name: "simple-eval"
description: "Simple evaluation example using a command agent"

agent:
  type: command
  config:
    command: echo
    args: ["Paris"]

defaults:
  trials_per_task: 3
  graders:
    - type: exact_match
      config:
        ignore_case: true
        ignore_whitespace: true

execution:
  concurrency: 2
  timeout: 30s

task_files:
  - tasks/*.yaml

output:
  format: all
  dir: ./results
```

**tasks/geography.yaml:**

```yaml
- id: capital-france
  name: "Capital of France"
  input:
    prompt: "What is the capital of France?"
  expected:
    text: "Paris"
```

The `command` agent runs `echo Paris` for every task, ignoring the prompt. The `exact_match` grader compares the output against the expected text with case and whitespace normalization enabled. This configuration is deterministic -- all 3 trials will pass every time.

## Coding Agent: OpenAI Evaluation

This example evaluates an OpenAI model on coding tasks with rate limiting and multiple trials.

**eval.yaml:**

```yaml
name: "coding-agent-eval"
description: "Evaluation suite for a coding agent"

agent:
  type: openai
  config:
    model: gpt-4
    api_key: ${OPENAI_API_KEY}
    base_url: https://api.openai.com/v1
    temperature: 0.0

defaults:
  trials_per_task: 3
  pass_threshold: 0.5

execution:
  concurrency: 4
  rate_limit_rps: 5
  timeout: 120s

task_files:
  - tasks/*.yaml

output:
  format: all
  dir: ./results
```

Key configuration details:

- **`temperature: 0.0`** -- Minimizes randomness for more reproducible results. Even at temperature 0, some non-determinism remains in LLM APIs.
- **`rate_limit_rps: 5`** -- Limits API calls to 5 per second to stay within rate limits. Adjust based on your API tier.
- **`concurrency: 4`** -- Runs 4 trials in parallel for faster evaluation. Combined with rate limiting, this prevents API throttling.
- **`timeout: 120s`** -- Allows up to 2 minutes per trial, appropriate for complex coding tasks.
- **`${OPENAI_API_KEY}`** -- Environment variable expansion. Set this in your shell before running.

## Multi-Grader: Combined Scoring

Tasks can override the suite-level grader defaults with their own grader configurations. This example uses multiple graders with different weights.

**tasks/chemistry.yaml:**

```yaml
- id: water-formula
  name: "Chemical Formula"
  input:
    prompt: "What is the chemical formula for water?"
  expected:
    text: "H2O"
  graders:
    - type: exact_match
      weight: 2.0
    - type: contains
      config:
        keywords: ["H", "O"]
    - type: constraint
      config:
        checks:
          - max_words: 10
```

This task applies three graders:

1. **`exact_match`** (weight 2.0) -- Checks if the output is exactly "H2O". This grader has double weight, making it contribute more to the final score.
2. **`contains`** -- Verifies the output contains both "H" and "O" as substrings.
3. **`constraint`** -- Ensures the response is concise (at most 10 words).

The trial's final score is the weighted average: `(exact_match * 2.0 + contains * 1.0 + constraint * 1.0) / 4.0`. However, the trial only passes if **all three** graders pass -- a correct formula that exceeds 10 words would still fail.

## LLM Grader: LLM-as-Judge

For subjective or complex evaluation criteria, the `llm` grader uses a language model to judge agent outputs against a rubric.

**eval.yaml:**

```yaml
name: "llm-graded-eval"
description: "Evaluation with LLM-based grading"

agent:
  type: openai
  config:
    model: gpt-4
    api_key: ${OPENAI_API_KEY}
    temperature: 0.7

defaults:
  trials_per_task: 2
  pass_threshold: 0.6
  graders:
    - type: llm
      config:
        provider: openai
        api_key: ${OPENAI_API_KEY}
        model: gpt-4
        rubric: |
          Evaluate the response based on:
          1. Correctness - Is the information factually accurate?
          2. Relevance - Does the response address the question directly?
          3. Completeness - Does the response cover all key aspects?

          Score 1.0 if all three criteria are met.
          Score 0.5 if the response is partially correct or incomplete.
          Score 0.0 if the response is incorrect or irrelevant.

task_files:
  - tasks/*.yaml

output:
  format: all
  dir: ./results
```

**tasks/knowledge.yaml:**

```yaml
- id: explain-photosynthesis
  name: "Explain Photosynthesis"
  input:
    prompt: "Explain the process of photosynthesis in 2-3 sentences."
  expected:
    text: "Photosynthesis is the process by which plants convert sunlight, water, and carbon dioxide into glucose and oxygen."
```

The LLM grader sends the agent's output, the expected output, and the rubric to the judge model. The judge returns a score and reasoning. This approach handles cases where exact matching is too strict -- paraphrased but correct answers can still receive high scores.

Key considerations for LLM grading:

- **Cost** -- Each grading call is an additional LLM API call. With many trials, this adds up.
- **Judge model** -- Using a stronger model (e.g., GPT-4) as a judge tends to produce more reliable grades than weaker models.
- **Rubric design** -- Clear, specific rubrics with concrete scoring criteria produce more consistent grades than vague instructions.
- **`base_url`** -- You can set `base_url` in the grader config to point to an OpenAI-compatible API endpoint.

## Next Steps

- Read [Core Concepts](/en/guide/concepts) to understand the scoring and metrics model
- See [Advanced Usage](/en/guide/advanced) for CI/CD integration, caching, and custom extensions
