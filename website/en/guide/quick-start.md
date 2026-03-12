# Quick Start

This guide walks you through installing AgentEval, creating your first evaluation, and viewing results.

## Installation

### One-line Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/wallezhang/agent-eval/main/install.sh | bash
```

This automatically detects your OS and architecture, downloads the latest binary, verifies the checksum, and installs to `/usr/local/bin`. You can customize the install directory with `INSTALL_DIR`:

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/wallezhang/agent-eval/main/install.sh | bash
```

### Using `go install`

```bash
go install github.com/wallezhang/agent-eval@latest
```

### Download Binary

Download pre-built binaries from the [GitHub Releases](https://github.com/wallezhang/agent-eval/releases) page. Binaries are available for Linux, macOS, and Windows across multiple architectures.

## Initialize a Project

```bash
agent-eval init my-eval
cd my-eval
```

This creates:

- `eval.yaml` -- evaluation configuration
- `tasks/` -- directory for task definitions
- `tasks/sample.yaml` -- a sample task file
- `results/` -- output directory

## Write Your First Evaluation

Edit `eval.yaml`:

```yaml
name: "my-first-eval"
description: "A simple evaluation"

agent:
  type: command
  config:
    command: echo
    args: ["Hello, World!"]

defaults:
  trials_per_task: 3
  graders:
    - type: exact_match
      config:
        ignore_case: true

execution:
  concurrency: 2
  timeout: 30s

task_files:
  - tasks/*.yaml

output:
  format: all
  dir: ./results
```

Edit `tasks/sample.yaml`:

```yaml
- id: greeting
  name: "Greeting Test"
  input:
    prompt: "Say hello"
  expected:
    text: "Hello, World!"
```

## Run the Evaluation

```bash
agent-eval run -c eval.yaml
```

The runner will execute 3 trials for the greeting task, grade each one with the exact_match grader, and print a summary table to the terminal.

## View Results

Results are displayed as a table in the terminal. You can also find detailed reports in the `results/` directory:

- `results/summary.json` -- machine-readable summary
- `results/report.html` -- interactive HTML report

## List Past Runs

```bash
agent-eval list
```

This shows all previous evaluation runs stored in the local SQLite database, including timestamps, pass rates, and run IDs.

## Compare Two Runs

```bash
agent-eval compare <runA> <runB>
```

This generates a diff report highlighting changes in pass rates, scores, and latency between two runs. Useful for tracking regressions or improvements after configuration changes.

## Next Steps

- Read [Core Concepts](/en/guide/concepts) to understand the evaluation model
- Explore [Examples](/en/guide/examples) for real-world configurations
- See the [Advanced Usage](/en/guide/advanced) guide for CI/CD integration, caching, and custom extensions
