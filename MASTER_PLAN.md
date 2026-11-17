# agentdiff: Architecture and Engineering Master Plan

Comparative Benchmarking and Trajectory Regression Platform for AI Coding Agents.
Target: Single-binary Go evaluation engine with embedded React 19 Trajectory Studio and AST diff analyzer.

---

## 1. Problem Statement and Motivation

### 1.1 The Silent Regressions in Autonomous Agent Upgrades
Engineering organizations frequently deploy autonomous coding and research agents based on foundation models. When upgrading foundation models (for example, switching between model checkpoints) or modifying system prompts, traditional static benchmarks (HumanEval, SWE-bench) fail to capture operational agent regressions:
1. Tool Invocation Bloat: A new model may pass tests but require 3x more tool calls (redundant directory listings, repeated file reads), tripling runtime latency and token costs.
2. AST Diff Churn: Rather than applying targeted 3-line surgical changes, drifting models frequently reformat, re-indent, or rewrite entire files, corrupting version control history.
3. Degraded Error Recovery: When a test fails, regressed models often enter repetitive retry cycles instead of diagnosing error outputs.
4. Lack of Trajectory Observability: Upgrades are often evaluated through subjective manual checks rather than reproducible, turn-by-turn trajectory metrics.

### 1.2 The agentdiff Solution
agentdiff is an automated comparative evaluation and trajectory regression platform written in Go.
Key capabilities:
- Parallel Sandboxed Benchmark Execution: Runs tasks concurrently in isolated Git worktrees.
- Full Trajectory Serialization: Captures thoughts, tool invocations, bash executions, and file diffs in structured JSON/JSONL.
- AST Semantic Diff Analysis: Measures the Surgical Precision of code edits across Go, Python, and TypeScript.
- Trajectory Edit Distance: Quantifies decision path divergence using sequence alignment.
- Anti-Pattern Detection: Flags doom loops, redundant reads, context bloat, and hallucinated tool calls.
- Automated CI Quality Gate: Blocks pull requests if surgical precision degrades or tool bloat exceeds configured limits.
- Embedded React 19 Trajectory Studio: Provides side-by-side synchronized playback, flamegraphs, and diff inspections from a single Go binary.

---

## 2. System Architecture

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
       |  +-------------------------------------------------------+  |
       |                               |                             |
       +-------------------------------|-----------------------------+
                                       |
                                       v
       +-------------------------------------------------------------+
       |                 Output and Decision Interfaces              |
       |  1. Synchronized React 19 Playback Studio (Local HTTP :7676)|
       |  2. Executive Regression Scorecard (Markdown and CI Gate)   |
       |  3. Structured Evaluation Dataset (agent_eval.jsonl)        |
       +-------------------------------------------------------------+
```

---

## 3. Mathematical Formulation: Trajectory Variance Metrics

### 3.1 Tool Invocation Efficiency Ratio (eta_tools)
Let T = (t_1, t_2, ..., t_K) represent the ordered sequence of tool invocations emitted by the agent. Let T_essential subseteq T denote the minimal subsequence of tool calls contributing directly to resolving the task:

$$\eta_{\text{tools}} = \frac{|T_{\text{essential}}|}{|T|} \in (0, 1]$$

A model update where Delta eta = eta_candidate - eta_baseline < -0.15 flags a tool bloat regression.

### 3.2 Tool Call Bloat Factor (B_tools)
Given total tool calls K_candidate and K_baseline on the identical task:

$$B_{\text{tools}} = \frac{K_{\text{candidate}}}{K_{\text{baseline}}}$$

A bloat factor B_tools > 1.25 indicates that the candidate model took over 25% more tool calls.

### 3.3 AST Surgical Precision Score (S_surgical)
Given total lines modified L_diff and modified AST declaration or statement nodes N_ast:

$$S_{\text{surgical}} = \frac{N_{\text{ast}}}{L_{\text{diff}} + 1}$$

Values close to 1.0 or higher indicate surgical changes. Low values indicate mass reformatting, cosmetic comment churn, or broad file rewrites.

### 3.4 Trajectory Edit Distance (D_traj)
Aligns tool call sequences T_A and T_B using sequence alignment:

$$D_{\text{traj}}(T_A, T_B) = \text{Levenshtein}(\text{Tokens}(T_A), \text{Tokens}(T_B))$$

Also detects the first point of divergence turn index where decisions parted.

---

## 4. Repository Structure

```
agentdiff/
├── cmd/
│   └── agentdiff/
│       └── main.go                 # CLI entry point: bench, compare, replay, studio, eval
├── internal/
│   ├── config/
│   │   ├── config.go               # Configuration parser for models, tasks, gates
│   │   └── schema.go               # Types for benchmarks, models, gates, metrics
│   ├── runner/
│   │   ├── workspace.go            # Sandboxed git worktrees and temporary environments
│   │   ├── executor.go             # Autonomous ReAct agent runner and external CLI harness
│   │   ├── built_in_agent.go       # Native agent loop with tool sandboxing
│   │   ├── mock_agent.go           # Deterministic replay and benchmark simulator
│   │   └── pool.go                 # Parallel worker pool with concurrency control
│   ├── trajectory/
│   │   ├── types.go                # Trajectory, Turn, Step, ToolCall data models
│   │   ├── recorder.go             # Thread-safe event recorder
│   │   ├── serializer.go           # JSON/JSONL reader, writer, and validator
│   │   └── alignment.go            # Needleman-Wunsch and Levenshtein alignment
│   ├── ast/
│   │   ├── parser.go               # Go, Python, and TypeScript structural parsers
│   │   ├── diff.go                 # Semantic AST node comparison
│   │   └── surgical.go             # Surgical precision calculation and churn penalty
│   ├── metrics/
│   │   ├── efficiency.go           # Tool bloat, redundant reads, doom loop detection
│   │   ├── scorecard.go            # Aggregated scorecards and regression verdicts
│   │   └── report.go               # Terminal tables, Markdown, and JSON exporters
│   └── server/
│       ├── api.go                  # REST endpoints for runs, tasks, comparisons
│       ├── sse.go                  # Server-Sent Events for real-time benchmark streaming
│       └── embed.go                # Embedded React 19 studio assets (go:embed)
├── web/
│   ├── src/
│   │   ├── App.tsx                 # React 19 root application
│   │   ├── components/
│   │   │   ├── SynchronizedPlayer.tsx# Side-by-side synchronized turn playback
│   │   │   ├── AstDiffViewer.tsx   # AST-annotated semantic diff viewer
│   │   │   ├── ToolCallFlamegraph.tsx# Timeline Gantt and latency distribution
│   │   │   ├── ScorecardHero.tsx   # Top KPI metrics and regression delta cards
│   │   │   ├── AntiPatternAlerts.tsx# Visual flags for doom loops and bloat
│   │   │   └── TaskSelector.tsx    # Filterable task and run picker
│   │   └── pages/
│   │       └── Dashboard.tsx       # Main workbench view
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── examples/
│   ├── agentdiff.yaml              # Reference configuration
│   └── benchmark_tasks/            # Real multi-step engineering tasks
│       ├── task_01_go_race_condition/
│       ├── task_02_python_data_pipeline/
│       └── task_03_typescript_null_safety/
├── docs/
│   ├── ARCHITECTURE.md             # System design and execution lifecycle
│   ├── METRICS_FORMULATION.md      # Detailed mathematical foundations
│   ├── CI_INTEGRATION.md           # GitHub Actions and quality gates
│   └── TRAJECTORY_SPEC.md          # JSON schema for trajectory serialization
├── .github/
│   └── workflows/
│       └── ci.yml                  # Continuous integration testing pipeline
├── Makefile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 5. Implementation Roadmap and Tasks

### Task 1: Repository Initialization and Base Documentation
- Initialize Git repository and create public GitHub repository.
- Author humanized documentation: README.md, docs/ARCHITECTURE.md, docs/METRICS_FORMULATION.md, docs/CI_INTEGRATION.md, docs/TRAJECTORY_SPEC.md.
- Set up root Makefile, .gitignore, and LICENSE.

### Task 2: Go Core Engine and Configuration
- Initialize Go module `github.com/alexandrmotologa/agentdiff`.
- Implement `internal/config` to parse YAML configurations, task manifests, candidate/baseline specifications, and CI regression thresholds.
- Add unit tests for configuration parsing and validation.

### Task 3: Workspace Isolation and Agent Runner Harness
- Implement `internal/runner/workspace.go` using isolated Git worktrees and temporary directories.
- Implement `internal/runner/built_in_agent.go`: Native agent loop supporting tools (`read_file`, `write_file`, `edit_file`, `list_dir`, `exec_bash`, `run_tests`).
- Implement `internal/runner/mock_agent.go`: Deterministic agent simulator for hermetic benchmarks and offline testing.
- Implement `internal/runner/executor.go` and `internal/runner/pool.go` for concurrent worker execution.
- Add unit tests for workspace lifecycle and agent harnesses.

### Task 4: Trajectory Serialization, Sequence Alignment, and Distance
- Define rich trajectory data structures in `internal/trajectory/types.go`.
- Implement `internal/trajectory/recorder.go` for thread-safe streaming capture.
- Implement `internal/trajectory/serializer.go` for JSON and JSONL format support.
- Implement `internal/trajectory/alignment.go` with Needleman-Wunsch sequence alignment, Levenshtein distance, and divergence point identification.
- Add unit tests for alignment algorithms.

### Task 5: Multi-Language AST Semantic Diff Engine
- Implement `internal/ast/parser.go` supporting Go (`go/parser`, `go/ast`), Python, and TypeScript syntax structures.
- Implement `internal/ast/diff.go` comparing AST node modifications versus line-level text churn.
- Implement `internal/ast/surgical.go` computing Surgical Precision ($S_{\text{surgical}}$) and formatting penalties.
- Add unit tests across Go, Python, and TypeScript test cases.

### Task 6: Metrics Aggregator, Anti-Pattern Detector, and CI Reporting
- Implement `internal/metrics/efficiency.go` for tool bloat, redundant file read detection, and doom loop detection.
- Implement `internal/metrics/scorecard.go` aggregating pass/fail deltas, token cost, latency, and regression flags.
- Implement `internal/metrics/report.go` generating formatted terminal tables, GitHub Action Markdown step summaries, and JSON reports.
- Add unit tests for metrics calculations.

### Task 7: CLI Interface Implementation
- Implement `cmd/agentdiff/main.go` with subcommands:
  - `bench`: execute comparative benchmark runs across tasks.
  - `compare`: inspect and diff two `.trajectory.json` files directly.
  - `replay`: turn-by-turn CLI trajectory playback.
  - `studio`: start the embedded web server on `:7676`.
  - `eval`: evaluate a single trajectory against quality gates.

### Task 8: Web Studio Frontend (React 19 + TypeScript + Tailwind)
- Set up Vite + React 19 + TypeScript in `web/`.
- Implement interactive components:
  - `ScorecardHero`: KPI metrics and delta indicators.
  - `TaskSelector`: Filterable task and benchmark runs.
  - `SynchronizedPlayer`: Turn-by-turn side-by-side synchronized player with scrubbing.
  - `AstDiffViewer`: Visual semantic diff with AST node highlights.
  - `ToolCallFlamegraph`: Gantt chart of tool invocations and execution durations.
  - `AntiPatternAlerts`: Badges and diagnostics for detected inefficiencies.
- Build production assets into `web/dist`.
- Implement `internal/server/embed.go` using `//go:embed` to serve the bundle directly from the Go binary.
- Implement `internal/server/api.go` and `internal/server/sse.go` for REST and live streaming.

### Task 9: Real Benchmark Tasks Suite and Verification
- Construct realistic multi-step benchmark tasks in `examples/benchmark_tasks`:
  - `task_01_go_race_condition`: Concurrency data race in Go channel worker.
  - `task_02_python_data_pipeline`: Null pointer and boundary error in Python processor.
  - `task_03_typescript_null_safety`: Type guard and missing property handling in TypeScript.
- Verify end-to-end execution of benchmarks, comparisons, and web studio.
- Validate binary memory footprint stays below 20MB RAM.

### Task 10: CI/CD Pipeline and Final Release Verification
- Set up `.github/workflows/ci.yml` running Go tests, linting, frontend build, and verification.
- Review all documentation to ensure zero placeholders and strict compliance with humanizer guidelines.
