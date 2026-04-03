# Web UI

AgentEval includes a browser-based Web UI for managing projects, editing configurations, running evaluations with real-time progress, and viewing results. The frontend is embedded into the single binary via `go:embed` — no separate process or install is required.

## Starting the Server

```bash
agent-eval server
```

Open [http://localhost:8080](http://localhost:8080) in your browser. The server auto-discovers projects registered in `~/.agent-eval/projects.json`.

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-p, --port` | `8080` | Server listen port |
| `--home` | `~/.agent-eval` | Home directory for project registry |

```bash
# Custom port
agent-eval server -p 3000

# Custom home directory
agent-eval server --home /data/agent-eval
```

## Adding a Project

On first launch, the Web UI has no projects. Click **"+ Add Project"** in the sidebar project switcher and provide:

- **Project Path** — absolute path to an existing `agent-eval init` directory (e.g., `/home/user/my-eval`)
- **Project Name** — auto-filled from the directory name, can be customized

The project is registered in `~/.agent-eval/projects.json`. You can add multiple projects and switch between them from the sidebar dropdown.

## Pages

### Dashboard

Overview of the selected project:

- **Summary cards** — total runs, configs count, average pass rate, active runs
- **Recent runs table** — clickable rows to view detailed results, with pass rate, duration, and date

### Configurations

YAML config editor with a file tree browser:

- **File tree** (left panel) — browse, create, and organize `.yaml` config files and folders
- **Editor** (center) — CodeMirror 6 with YAML syntax highlighting, real-time validation as you type
- **Quick Insert** (right panel) — one-click templates for agent, task, and grader blocks
- **Reference** (right panel) — available agent and grader types for quick lookup

Validation runs automatically on every edit and on file switch. Errors display inline above the editor.

### Runs

Start and manage evaluation runs:

- **New Run** — select a config file and start an evaluation. You are redirected to the live run view.
- **Active runs** — cards with animated progress indicator, click to view live SSE stream
- **Run history** — table with suite name, agent type, pass rate (with mini progress bar), duration, and relative timestamps
- **Compare** — check any 2 runs, then click Compare to see a side-by-side diff with charts

### Run Detail (Live)

Real-time view of an in-progress evaluation:

- **Progress bar** — animated gradient fill showing completion percentage
- **Status badges** — pass/fail/error counts update in real-time via SSE
- **Log terminal** — scrolling terminal-style log with color-coded lines (green for pass, red for errors, blue for start events)
- **Cancel** — stop a running evaluation at any time

### Results

Detailed breakdown of a completed run:

- **Summary cards** — pass rate, average score, total trials, estimated cost
- **Task results** — expandable rows per task showing pass/fail/error counts, average score, and latency percentiles (P50/P90)
- **Trial details** — per-trial grades, scores, agent output, metadata, and transcript

### Compare

Side-by-side comparison of two runs:

- **Run meta cards** — orange-accented (Run A) and indigo-accented (Run B) summary cards
- **Bar chart** — ECharts visualization comparing pass rate, avg score, pass@k, and pass^k
- **Metrics table** — numeric comparison with directional arrows (↑ improved, ↓ regressed)
- **Per-task drill-down** — expandable rows showing trial-level diffs, filterable by status (improved/regressed/unchanged)

### Settings

Project information display:

- Project name, path, and database path
