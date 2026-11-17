# System Architecture: agentdiff

## Overview

agentdiff evaluates autonomous coding and research agents by comparing their operational trajectories on standardized software engineering tasks. Rather than relying solely on end-state unit test pass rates, agentdiff analyzes intermediate decisions: tool choices, execution latency, file mutations, and AST semantic deltas.

```
       +-------------------------------------------------------------+
       |                  Task Suite Manifest                        |
       |  (Target Repository, Initial Commit, Verification Tests)    |
       +-------------------------------------------------------------+
                                      |
                         Concurrent Task Scheduler
                                      v
       +-------------------------------------------------------------+
       |                Workspace Isolation Manager                  |
       |  Creates two independent Git worktrees / sandbox roots:     |
       |    - Workspace A: Candidate Agent (Model A / Prompt A)      |
       |    - Workspace B: Baseline Agent (Model B / Prompt B)       |
       +-------------------------------------------------------------+
                                      |
                         Parallel Agent Execution
                                      v
       +-------------------------------------------------------------+
       |                Trajectory Stream Recorders                  |
       |  Captures chronological events per agent:                   |
       |    - Thought text & reasoning chains                        |
       |    - Tool invocations with arguments & returned stdout      |
       |    - Shell command executions with exit codes               |
       |    - File modifications & git diff patches                  |
       |    - Final verification test status and execution time      |
       +-------------------------------------------------------------+
                                      |
                         Trajectory Analysis Engine
                                      v
       +-------------------------------------------------------------+
       |  1. AST Semantic Diff: parses modified source files,        |
       |     calculates Surgical Precision Score (S_surgical).       |
       |  2. Sequence Alignment: computes Trajectory Edit Distance   |
       |     (D_traj) and isolates points of strategy divergence.    |
       |  3. Anti-Pattern Engine: flags doom loops, redundant reads, |
       |     and tool invocation bloat (B_tools).                    |
       +-------------------------------------------------------------+
                                      |
                                      v
       +-------------------------------------------------------------+
       |                     Outputs & Deliverables                  |
       |  - Embedded React 19 Trajectory Studio (localhost:7676)     |
       |  - CLI Markdown Scorecard & GitHub Step Summary             |
       |  - Machine-readable evaluation dataset (.trajectory.json)   |
       +-------------------------------------------------------------+
```

## Component Breakdown

### 1. Workspace Manager (`internal/runner/workspace.go`)
- Uses Git worktrees (`git worktree add -b ...`) or isolated temporary working directories to clone the benchmark task repository.
- Guarantees that candidate and baseline models execute against identical initial states without file collisions.
- Cleans up worktrees and temporary branches upon benchmark completion.

### 2. Agent Execution Harness (`internal/runner/`)
- `built_in_agent.go`: A native agent loop that executes against standard LLM providers (OpenAI, Anthropic, Gemini, Ollama, custom HTTP). Implements tools: `read_file`, `write_file`, `edit_file`, `list_dir`, `exec_bash`, `run_tests`, `git_diff`.
- `mock_agent.go`: A deterministic agent simulator executing pre-recorded or deterministic behavior profiles for hermetic testing and verification without external network access.
- `executor.go`: An external process runner capable of driving external coding agent CLIs (e.g. Aider, Claude Code) and capturing standard streams.

### 3. Trajectory Serializer & Alignment Engine (`internal/trajectory/`)
- Encodes full execution traces into JSON or JSONL format.
- Employs Needleman-Wunsch global sequence alignment and Levenshtein distance on tool invocation tokens.
- Identifies exact turns where the candidate model diverged from the baseline path.

### 4. AST Semantic Diff Engine (`internal/ast/`)
- Parses modified source files in Go (`go/parser`, `go/ast`), Python, and TypeScript.
- Differentiates between functional syntax tree modifications (statement and declaration mutations) and cosmetic changes (whitespace reformatting, comments).
- Computes the Surgical Precision Score:

$$S_{\text{surgical}} = \frac{N_{\text{ast}}}{L_{\text{diff}} + 1}$$

### 5. Metrics & Anti-Pattern Detection (`internal/metrics/`)
- Detects redundant file reads (reading the identical file multiple times without intervening modifications).
- Detects command doom loops (executing the same failing terminal command three or more times without code changes).
- Calculates Tool Invocation Efficiency Ratio and Tool Bloat Factor.
- Computes aggregate pass/fail deltas, token cost variance, and latency overhead.

### 6. Trajectory Studio Server (`internal/server/`)
- Go HTTP server with embedded React 19 static assets.
- Serves REST APIs for runs, trajectories, and comparison metrics.
- Provides a Server-Sent Events (SSE) stream for real-time benchmark monitoring.
- Targets a memory footprint under 20MB resident set size.
