# Core Concepts

This page explains the foundational ideas behind AgentEval and how evaluations are structured.

## Why Evaluate AI Agents?

AI agents are inherently non-deterministic. The same prompt can produce different outputs across runs, and quality can vary significantly depending on model parameters, prompt phrasing, and external API behavior. Manual spot-checking does not scale and misses reliability problems that only surface over many runs.

Systematic evaluation addresses these challenges:

- **Quantify reliability** -- measure how often an agent produces correct results, not just whether it can
- **Track regressions** -- detect quality drops when prompts, models, or configurations change
- **Compare configurations** -- make data-driven decisions about model selection, temperature settings, and prompt design
- **Enforce quality gates** -- block deployments that fall below minimum pass rates

## Key Terminology

### Suite

An evaluation suite is a complete configuration that defines which agent to test, what tasks to run, and how to grade results. A suite is defined in a single YAML file (typically `eval.yaml`) and references one or more task files.

### Task

A task represents a single evaluation case with an input prompt and expected output. Tasks are defined in YAML files and referenced from the suite via glob patterns. Each task has a unique ID, a prompt to send to the agent, and expected output for grading.

### Trial

Each task is run multiple times (configurable via `trials_per_task`). Each run is called a trial. Multiple trials enable statistical reliability measurements -- a single pass tells you the agent *can* answer correctly, but multiple trials tell you how *often* it does.

### Grader

A grader evaluates an agent's output against expected results. Each task can have one or more graders, each producing a score (0.0-1.0) and a pass/fail status. The 8 built-in graders are:

| Grader | Purpose |
|--------|---------|
| `exact_match` | Exact string comparison (with optional case/whitespace normalization) |
| `contains` | Checks for required keywords in the output |
| `regex` | Matches output against a regular expression |
| `json_match` | Compares JSON structures with configurable path matching |
| `command` | Runs an external command to grade output |
| `llm` | Uses an LLM as a judge with a rubric |
| `pairwise` | Compares two outputs using an LLM to pick the better one |
| `constraint` | Validates structural constraints (word count, length, format) |

## Evaluation Flow

The evaluation follows a well-defined pipeline:

1. **Load configuration** -- Parse the YAML suite file, expand environment variables, apply defaults
2. **Create agent** -- Instantiate the agent adapter (OpenAI, Anthropic, HTTP, or Command)
3. **Build work items** -- Combine tasks and trial counts into a flat list of work items
4. **Load checkpoint** (optional) -- Skip already-completed trials if resuming an interrupted run
5. **Run trials concurrently** -- Execute work items using a concurrent scheduler with rate limiting
6. **Grade each trial** -- Apply all configured graders to each trial's output
7. **Aggregate results** -- Compute per-task and suite-level statistics (pass rates, scores, pass@k, pass^k)
8. **Generate reports** -- Output results as terminal table, JSON, and/or HTML

## Scoring Logic

The scoring system uses a multi-grader model with AND-based pass logic:

- Each grader produces a **score** (0.0 to 1.0) and a **pass/fail** status
- A trial's final score is the **weighted average** across all graders
- A trial **passes** only if **ALL** graders pass (AND logic)
- This means a single failing grader causes the whole trial to fail, regardless of scores

This design ensures that evaluation criteria are strictly enforced. For example, if a response is factually correct but exceeds a word limit, it still fails.

## pass@k and pass^k

These two complementary metrics capture different aspects of agent performance.

### pass@k (Capability Ceiling)

The probability of getting at least one correct answer in k attempts. This is computed using the formula:

```
pass@k = 1 - C(n-c, k) / C(n, k)
```

where `n` is the total number of trials and `c` is the number of passing trials. This measures what the agent *can* do at its best -- given enough tries, how likely is it to succeed at least once?

### pass^k (Reliability)

The probability of getting all k attempts correct. This is computed using the formula:

```
pass^k = C(c, k) / C(n, k)
```

This measures how dependably the agent performs. High pass^k indicates consistent, production-ready behavior.

### Example

If 7 out of 10 trials pass:

- **pass@1 = 0.70** -- 70% chance a single attempt succeeds
- **pass@3 = 0.99** -- 99% chance at least one of three attempts succeeds
- **pass^3 = 0.29** -- 29% chance all three attempts succeed

The gap between pass@k and pass^k reveals reliability issues. A high pass@3 combined with a low pass^3 means the agent is capable but inconsistent -- it usually gets it right, but not always. For production systems that need every request to succeed, pass^k is the metric that matters.

::: tip
AgentEval uses log-space arithmetic for pass@k and pass^k computation to avoid numerical overflow when working with large trial counts.
:::

## Next Steps

- Follow the [Quick Start](/en/guide/quick-start) to run your first evaluation
- See [Examples](/en/guide/examples) for practical evaluation configurations
- Read [Advanced Usage](/en/guide/advanced) for CI/CD integration and custom extensions
