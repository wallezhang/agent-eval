# CLI Reference

## agent-eval run

Run an evaluation suite.

```bash
agent-eval run [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-c, --config` | string | `eval.yaml` | Path to evaluation config file |
| `--db` | string | | SQLite database path |
| `--verbose` | bool | `false` | Enable verbose logging |
| `--fail-under` | float | `0.0` | Minimum pass rate (0.0-1.0). Exit code 1 if below threshold |
| `--tags` | string | | Only run tasks matching these tags (comma-separated) |
| `--exclude-tags` | string | | Exclude tasks matching these tags (comma-separated) |
| `--no-cache` | bool | `false` | Bypass response cache for this run |
| `--resume` | string | | Resume a previous run by run ID |

### Examples

```bash
# Run with default config
agent-eval run

# Run with specific config
agent-eval run -c my-eval.yaml

# CI mode: fail if pass rate below 80%
agent-eval run -c eval.yaml --fail-under 0.8

# Run only math tasks
agent-eval run --tags math

# Resume interrupted run
agent-eval run --resume abc123
```

## agent-eval list

List historical evaluation runs.

```bash
agent-eval list [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--db` | string | `./results/agent-eval.db` | Path to SQLite database |

### Output

Displays a table with columns:
- **ID** -- Run identifier
- **SUITE** -- Suite name
- **AGENT** -- Agent type
- **TASKS** -- Number of tasks
- **PASS RATE** -- Overall pass rate
- **DURATION** -- Total run duration
- **DATE** -- Run timestamp

### Example

```bash
agent-eval list
agent-eval list --db ./my-results/agent-eval.db
```

## agent-eval compare

Compare two evaluation runs side by side.

```bash
agent-eval compare <runA> <runB> [flags]
```

### Arguments

- `runA` -- First run ID (supports prefix matching)
- `runB` -- Second run ID (supports prefix matching)

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--db` | string | `./results/agent-eval.db` | Path to SQLite database |

### Example

```bash
agent-eval compare abc123 def456
# Prefix matching
agent-eval compare abc def
```

## agent-eval init

Initialize a new evaluation project.

```bash
agent-eval init [directory]
```

### Arguments

- `directory` -- Target directory name (optional, defaults to current directory)

### Created Files

```
<directory>/
  eval.yaml          # Evaluation configuration template
  tasks/
    sample.yaml      # Sample task file
  results/           # Output directory
```

### Example

```bash
agent-eval init my-eval
cd my-eval
# Edit eval.yaml and tasks/sample.yaml, then:
agent-eval run
```

## agent-eval server

Start the Web UI server. The frontend is embedded in the binary — no separate install needed.

```bash
agent-eval server [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-p, --port` | int | `8080` | Server listen port |
| `--home` | string | `~/.agent-eval` | Home directory for project registry |

### Examples

```bash
# Start with defaults (port 8080)
agent-eval server

# Custom port
agent-eval server -p 3000

# Custom home directory
agent-eval server --home /data/agent-eval
```

Open [http://localhost:8080](http://localhost:8080) in your browser to manage projects, edit configs, run evaluations with real-time progress, and view results. See the [Web UI guide](/en/guide/web-ui) for details on each page.
