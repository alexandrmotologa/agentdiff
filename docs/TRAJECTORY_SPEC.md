# Trajectory Trace Specification (`.trajectory.json`)

## Overview
A `.trajectory.json` file contains the complete, serial execution record of an autonomous coding agent solving a software engineering task.

---

## 1. Top-Level Schema

```json
{
  "$schema": "https://raw.githubusercontent.com/alexandrmotologa/agentdiff/main/docs/trajectory_schema.json",
  "version": "1.0.0",
  "task_id": "task_01_go_race_condition",
  "agent_id": "claude-3-7-sonnet-candidate",
  "model": "claude-3-7-sonnet-20250219",
  "system_prompt_hash": "sha256:4a8b79f610...",
  "started_at": "2026-03-30T10:15:00Z",
  "completed_at": "2026-03-30T10:15:32Z",
  "duration_ms": 32150,
  "status": "COMPLETED",
  "verified": true,
  "turns": [],
  "final_patch": "diff --git a/worker.go b/worker.go...",
  "ast_metrics": {
    "total_lines_diff": 4,
    "ast_nodes_modified": 3,
    "surgical_score": 0.60,
    "cosmetic_churn_ratio": 0.20
  },
  "token_usage": {
    "prompt_tokens": 14250,
    "completion_tokens": 1820,
    "total_tokens": 16070,
    "estimated_cost_usd": 0.054
  },
  "anti_patterns": []
}
```

---

## 2. Turn Schema

A turn represents one cycle of reasoning, tool invocation, and environment response.

```json
{
  "turn_index": 1,
  "timestamp": "2026-03-30T10:15:02Z",
  "duration_ms": 4200,
  "thought": "Let's first list files in the repository to locate the concurrency issue.",
  "tool_calls": [
    {
      "call_id": "call_01",
      "tool_name": "list_dir",
      "arguments": {
        "path": "."
      },
      "status": "SUCCESS",
      "duration_ms": 12,
      "output": "main.go\nworker.go\nworker_test.go\ngo.mod\n"
    }
  ],
  "terminal_executions": [],
  "file_mutations": []
}
```

---

## 3. Tool Call Schema

```json
{
  "call_id": "call_03",
  "tool_name": "edit_file",
  "arguments": {
    "path": "worker.go",
    "target": "c.count++",
    "replacement": "c.mu.Lock()\nc.count++\nc.mu.Unlock()"
  },
  "status": "SUCCESS",
  "duration_ms": 45,
  "output": "Successfully replaced 1 occurrence in worker.go"
}
```

---

## 4. File Mutation Schema

```json
{
  "path": "worker.go",
  "change_type": "MODIFIED",
  "lines_added": 3,
  "lines_removed": 1,
  "diff": "@@ -24,1 +24,3 @@\n- c.count++\n+ c.mu.Lock()\n+ c.count++\n+ c.mu.Unlock()",
  "ast_nodes_affected": [
    {
      "kind": "CallExpr",
      "identifier": "c.mu.Lock",
      "line_start": 24,
      "line_end": 24
    },
    {
      "kind": "CallExpr",
      "identifier": "c.mu.Unlock",
      "line_start": 26,
      "line_end": 26
    }
  ]
}
```

---

## 5. Anti-Pattern Schema

```json
{
  "kind": "REDUNDANT_READ",
  "severity": "WARNING",
  "turn_index": 4,
  "tool_name": "read_file",
  "details": "File 'worker.go' read again without any intermediate modification."
}
```
