# agentdiff

Comparative benchmarking and trajectory regression platform for AI coding agents.

agentdiff runs multi-step coding benchmarks across competing model versions or system prompt configurations in isolated Git worktrees. It serializes agent turns, tool invocations, shell outputs, and file diffs. It then calculates Trajectory Edit Distance and AST Surgical Precision to pinpoint where an updated model becomes slower, wasteful, or prone to cosmetic file churn.

## Key Features

- **Isolated Workspace Execution**: Concurrently runs candidate and baseline models in separate Git worktrees under identical starting conditions.
- **Trajectory Serialization**: Records thoughts, tool calls, bash arguments, outputs, and diffs in structured JSON format.
- **AST Semantic Diff Engine**: Computes AST node modifications across Go, Python, and TypeScript, measuring surgical edit precision versus formatting churn.
- **Sequence Alignment & Edit Distance**: Calculates Levenshtein distance on tool sequences and detects the exact turn of strategy divergence.
- **Anti-Pattern Diagnostics**: Detects command doom loops, redundant file reads, and tool bloat.
- **Divergence Root-Cause Explainer**: Automatically analyzes the separation turn and recommends prompt steering adjustments.
- **Token Economics & Scaling Forecaster**: Projects monthly monetary budget deltas at 1k, 10k, and 50k task volumes.
- **Multi-Model Leaderboard Matrix**: Evaluates and ranks models across pass rates, AST surgical precision, and cost per solved task.
- **Self-Contained HTML Report Exporter**: Generates standalone single-file interactive reports that run offline in any browser without a server.
- **Interactive Terminal TUI Replayer**: Step through turns, inspect reasoning, and review diffs directly from the terminal.
- **Automated CI Quality Gate**: Fails CI runs when surgical precision degrades or tool bloat exceeds configured limits.
- **Low Memory Footprint**: Compiled as a single static Go binary with memory usage below 20MB RAM (~11.6MB working set).

## System Architecture

```
       +-------------------------------------------------------------+
       |               Agent Task Benchmark Suite                    |
       |       (Multi-Step Software Tasks in Go, Python, TS)         |
       +-------------------------------------------------------------+
                                      |
                         Concurrent Benchmark Dispatcher
                                      v
       +-------------------------------------------------------------+
       |                     agentdiff Core Engine                   |
       |                   (High-Performance Go CLI)                |
       |                                                             |
       |  +-------------------------------------------------------+  |
       |  |          Parallel Sandboxed Workspace Runner          |  |
       |  |   [Workspace A: Model Candidate]  [Workspace B: Base] |  |
       |  +-------------------------------------------------------+  |
       |                               |                             |
       |                               v                             |
       |  +-------------------------------------------------------+  |
       |  |            Deterministic Trajectory Recorder          |  |
       |  |   Serializes: Turns, Tool Calls, Bash Runs, File Diffs|  |
       |  +-------------------------------------------------------+  |
       |                               |                             |
       |                               v                             |
       |  +-------------------------------------------------------+  |
       |  |           Trajectory Variance & AST Diff Engine       |  |
       |  |  - AST Surgical Edit Precision Score                 |  |
       |  |  - Tool Call Bloat Factor & Redundancy Ratio          |  |
       |  |  - Trajectory Edit Distance (Sequence Alignment)      |  |
       |  |  - Anti-Pattern and Doom Loop Detector                |  |
       |  |  - Token Economics & Scale Impact Forecaster          |  |
       |  +-------------------------------------------------------+  |
       |                               |                             |
       +-------------------------------|-----------------------------+
                                       |
                                       v
       +-------------------------------------------------------------+
       |                 Output and Decision Interfaces              |
       |  1. Synchronized React 19 Playback Studio (Local HTTP :7676)|
       |  2. Standalone Self-Contained HTML Report (report.html)     |
       |  3. Terminal Interactive TUI Replayer (agentdiff replay -i) |
       |  4. Executive Regression Scorecard (Markdown & CI Gate)     |
       |  5. Multi-Model Benchmark Leaderboard & Frontier Matrix     |
       +-------------------------------------------------------------+
```

## Quick Start

### Installation

```bash
git clone https://github.com/alexandrmotologa/agentdiff.git
cd agentdiff
go build -o bin/agentdiff ./cmd/agentdiff
```

### Running a Benchmark

```bash
# Execute comparative benchmark on configured tasks
./bin/agentdiff bench --config examples/agentdiff.yaml --output-dir ./eval_results
```

### Comparing Two Trajectories

```bash
# Compare candidate vs baseline trajectory traces
./bin/agentdiff compare \
  --candidate ./eval_results/task_01_go_race_condition/candidate.trajectory.json \
  --baseline ./eval_results/task_01_go_race_condition/baseline.trajectory.json
```

### Interactive Terminal Replayer

```bash
# Step through trajectory turns interactively with keyboard navigation
./bin/agentdiff replay --trace ./eval_results/task_01_go_race_condition/candidate.trajectory.json -i
```

### Exporting Standalone Interactive HTML Report

```bash
# Generate single-file report with embedded React studio and traces (runs offline)
./bin/agentdiff export --traces-dir ./eval_results --output ./eval_results/report.html
```

### Starting the Web Studio

```bash
# Start the local playback server
./bin/agentdiff studio --port 7676 --traces-dir ./eval_results
```

Open `http://localhost:7676` in your browser.

## Core Metrics

### 1. Tool Invocation Efficiency Ratio
Measures the proportion of essential tool calls contributing directly to resolving the task:
$$\eta_{\text{tools}} = \frac{|\mathcal{T}_{\text{essential}}|}{|\mathcal{T}|}$$

### 2. AST Surgical Precision Score
Measures semantic changes against overall line diff churn:
$$S_{\text{surgical}} = \frac{N_{\text{ast}}}{L_{\text{diff}} + 1}$$

### 3. Trajectory Edit Distance
Calculates Levenshtein distance on tool sequences to evaluate strategy divergence:
$$D_{\text{traj}} = \text{Levenshtein}(\text{Tokens}(\mathcal{T}_A), \text{Tokens}(\mathcal{T}_B))$$

### 4. Token Economics & Scale Impact
Projects monthly cost delta at 1,000, 10,000, and 50,000 tasks/month based on actual input/output token counts and model pricing tables.

## Repository Layout

- `cmd/agentdiff`: CLI entry point (`bench`, `compare`, `replay`, `studio`, `export`, `eval`).
- `internal/config`: YAML configuration parser and task schema.
- `internal/runner`: Git worktree manager, native agent harness, and mock runner.
- `internal/trajectory`: Serialization, recorder, and sequence alignment algorithms.
- `internal/ast`: Multi-language AST semantic diff parser and surgical precision engine.
- `internal/metrics`: Efficiency calculations, cost projections, divergence diagnosis, and leaderboard.
- `internal/server`: HTTP server with embedded React 19 studio, SSE streaming, and HTML exporter.
- `web`: React 19 + TypeScript single-page application with Leaderboard matrix and AST hierarchy.
- `examples`: Sample benchmark tasks in Go, Python, TypeScript and configuration templates.

## License

MIT
