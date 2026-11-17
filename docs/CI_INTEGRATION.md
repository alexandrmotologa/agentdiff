# CI/CD Integration Guide: agentdiff

## Overview
agentdiff operates as an automated regression gate in continuous integration pipelines. When modifying foundation model configurations, prompt templates, or tool definitions, agentdiff executes benchmark tasks and evaluates regressions against defined numerical thresholds.

---

## 1. CLI Commands for Automation

### Benchmark Execution
```bash
agentdiff bench \
  --config agentdiff.yaml \
  --output-dir ./eval_results \
  --fail-on-regression \
  --github-summary
```

### Direct Trajectory Comparison
```bash
agentdiff compare \
  --candidate ./eval_results/candidate.trajectory.json \
  --baseline ./eval_results/baseline.trajectory.json \
  --thresholds ./thresholds.yaml \
  --markdown-report ./eval_summary.md
```

---

## 2. Exit Codes

| Exit Code | Meaning | Action in CI |
|-----------|---------|--------------|
| `0` | All quality gates passed without regressions. | Pipeline succeeds. |
| `1` | Trajectory regression detected (e.g. tool bloat, surgical precision loss, or pass rate drop). | Pipeline fails. |
| `2` | Configuration error or invalid task definition. | Pipeline fails immediately. |
| `3` | Internal execution timeout or sandbox failure. | Pipeline fails. |

---

## 3. GitHub Actions Pipeline Example

Create `.github/workflows/agent-eval.yml`:

```yaml
name: Agent Trajectory Regression Evaluation

on:
  pull_request:
    paths:
      - 'prompts/**'
      - 'agent/**'
      - 'models.yaml'

jobs:
  evaluate-agent-drift:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Build agentdiff CLI
        run: go build -o /usr/local/bin/agentdiff ./cmd/agentdiff

      - name: Run agentdiff Benchmark Suite
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          agentdiff bench \
            --config examples/agentdiff.yaml \
            --output-dir ./results \
            --fail-on-regression \
            --github-summary

      - name: Upload Evaluation Artifacts
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: trajectory-traces
          path: ./results/*.trajectory.json
```

---

## 4. Markdown Step Summary Output

When `--github-summary` is specified, agentdiff writes directly to `$GITHUB_STEP_SUMMARY`:

```markdown
### agentdiff Regression Evaluation Summary

| Metric | Baseline | Candidate | Delta | Status |
|--------|----------|-----------|-------|--------|
| Pass Rate | 100.0% | 100.0% | +0.0% | PASS |
| Avg Duration | 14.2s | 11.8s | -2.4s | FASTER |
| Tool Invocations | 12 | 19 | +58.3% | REGRESSION |
| AST Surgical Score | 0.88 | 0.42 | -0.46 | REGRESSION |
| Redundant Reads | 0 | 4 | +4 | WARNING |

**Verdict:** REGRESSION DETECTED. Candidate model exhibited significant tool bloat (+58.3%) and reduced AST surgical precision.
```
