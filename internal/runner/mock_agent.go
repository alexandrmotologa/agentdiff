package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/alexandrmotologa/agentdiff/internal/config"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// RunMockAgent executes a deterministic simulated agent run on the workspace.
func RunMockAgent(ws *Workspace, task config.TaskDefinition, modelCfg config.ModelConfig) (*trajectory.Trajectory, error) {
	rec := trajectory.NewRecorder(task.ID, modelCfg.ID, modelCfg.Model, modelCfg.SystemPrompt)

	mode := strings.ToLower(modelCfg.Model)
	if mode == "" {
		mode = "mock-baseline"
	}

	targetFile := "main.go"
	if len(task.TargetFiles) > 0 {
		targetFile = task.TargetFiles[0]
	}

	switch {
	case strings.Contains(mode, "bloat") || strings.Contains(mode, "v2-regressed"):
		return runBloatedScenario(rec, ws, task, targetFile)
	case strings.Contains(mode, "churn") || strings.Contains(mode, "formatting"):
		return runChurnScenario(rec, ws, task, targetFile)
	case strings.Contains(mode, "doom") || strings.Contains(mode, "fail"):
		return runDoomLoopScenario(rec, ws, task, targetFile)
	default:
		return runEfficientScenario(rec, ws, task, targetFile)
	}
}

// runEfficientScenario: reads target, applies surgical fix, runs test, succeeds.
func runEfficientScenario(rec *trajectory.Recorder, ws *Workspace, task config.TaskDefinition, targetFile string) (*trajectory.Trajectory, error) {
	// Turn 1: Inspect target file
	rec.StartTurn(fmt.Sprintf("Reading %s to inspect the implementation.", targetFile))
	content, err := ws.ReadFile(targetFile)
	call1 := trajectory.ToolCall{
		ToolName:         "read_file",
		Arguments:        map[string]interface{}{"path": targetFile},
		NormalizedTarget: targetFile,
		DurationMs:       15,
	}
	if err != nil {
		call1.Status = "ERROR"
		call1.Output = err.Error()
	} else {
		call1.Status = "SUCCESS"
		call1.Output = content
	}
	rec.RecordToolCall(call1)
	rec.EndTurn()

	// Turn 2: Apply surgical fix
	rec.StartTurn(fmt.Sprintf("Applying surgical fix to %s.", targetFile))
	applyMockTaskFix(ws, task, false)
	call2 := trajectory.ToolCall{
		ToolName:         "edit_file",
		Arguments:        map[string]interface{}{"path": targetFile},
		NormalizedTarget: targetFile,
		Status:           "SUCCESS",
		DurationMs:       28,
		Output:           fmt.Sprintf("Updated %s with targeted fix.", targetFile),
	}
	rec.RecordToolCall(call2)
	rec.EndTurn()

	// Turn 3: Verify with test command
	rec.StartTurn(fmt.Sprintf("Running verification test: %s", task.TestCommand))
	stdout, stderr, exitCode, durMs, _ := ws.ExecCommand(task.TestCommand, time.Duration(task.TimeoutSeconds)*time.Second)
	testCall := trajectory.ToolCall{
		ToolName:         "run_tests",
		Arguments:        map[string]interface{}{"command": task.TestCommand},
		NormalizedTarget: task.TestCommand,
		DurationMs:       durMs,
	}
	if exitCode == 0 {
		testCall.Status = "SUCCESS"
		testCall.Output = stdout
	} else {
		testCall.Status = "ERROR"
		testCall.Output = stdout + "\n" + stderr
	}
	rec.RecordToolCall(testCall)
	rec.RecordTerminalExecution(trajectory.TerminalExecution{
		Command:    task.TestCommand,
		ExitCode:   exitCode,
		Stdout:     stdout,
		Stderr:     stderr,
		DurationMs: durMs,
	})
	rec.EndTurn()

	patch, mutations, astMetrics, _ := ws.ComputeDiff()
	rec.SetFinalPatch(patch)
	for _, m := range mutations {
		rec.RecordFileMutation(m)
	}
	rec.SetASTMetrics(astMetrics)
	rec.SetTokenUsage(trajectory.TokenUsage{
		PromptTokens:     1250,
		CompletionTokens: 320,
		TotalTokens:      1570,
		EstimatedCostUSD: 0.005,
	})

	verified := (exitCode == 0)
	status := "COMPLETED"
	if !verified {
		status = "FAILED"
	}
	return rec.Complete(status, verified, stdout), nil
}

// runBloatedScenario: redundant reads, directory listings, extra calls, passes.
func runBloatedScenario(rec *trajectory.Recorder, ws *Workspace, task config.TaskDefinition, targetFile string) (*trajectory.Trajectory, error) {
	// Turn 1: List directory
	rec.StartTurn("Listing repository files.")
	entries, _ := ws.ListDir(".")
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "list_dir",
		Arguments:        map[string]interface{}{"path": "."},
		NormalizedTarget: ".",
		Status:           "SUCCESS",
		DurationMs:       10,
		Output:           strings.Join(entries, "\n"),
	})
	rec.EndTurn()

	// Turn 2: Read target file
	rec.StartTurn(fmt.Sprintf("Reading %s first time.", targetFile))
	content, _ := ws.ReadFile(targetFile)
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "read_file",
		Arguments:        map[string]interface{}{"path": targetFile},
		NormalizedTarget: targetFile,
		Status:           "SUCCESS",
		DurationMs:       15,
		Output:           content,
	})
	rec.EndTurn()

	// Turn 3: Redundant read of same target file
	rec.StartTurn("Re-reading same file to double check lines.")
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "read_file",
		Arguments:        map[string]interface{}{"path": targetFile},
		NormalizedTarget: targetFile,
		Status:           "SUCCESS",
		DurationMs:       14,
		Output:           content,
	})
	rec.EndTurn()

	// Turn 4: Another redundant read
	rec.StartTurn("Checking syntax one more time.")
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "read_file",
		Arguments:        map[string]interface{}{"path": targetFile},
		NormalizedTarget: targetFile,
		Status:           "SUCCESS",
		DurationMs:       12,
		Output:           content,
	})
	rec.EndTurn()

	// Turn 5: Apply fix
	rec.StartTurn("Applying fix now.")
	applyMockTaskFix(ws, task, false)
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "edit_file",
		Arguments:        map[string]interface{}{"path": targetFile},
		NormalizedTarget: targetFile,
		Status:           "SUCCESS",
		DurationMs:       32,
		Output:           "Edited file.",
	})
	rec.EndTurn()

	// Turn 6: Run tests
	rec.StartTurn("Running verification test.")
	stdout, stderr, exitCode, durMs, _ := ws.ExecCommand(task.TestCommand, time.Duration(task.TimeoutSeconds)*time.Second)
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "run_tests",
		Arguments:        map[string]interface{}{"command": task.TestCommand},
		NormalizedTarget: task.TestCommand,
		Status:           "SUCCESS",
		DurationMs:       durMs,
		Output:           stdout,
	})
	rec.RecordTerminalExecution(trajectory.TerminalExecution{
		Command:    task.TestCommand,
		ExitCode:   exitCode,
		Stdout:     stdout,
		Stderr:     stderr,
		DurationMs: durMs,
	})
	rec.EndTurn()

	patch, mutations, astMetrics, _ := ws.ComputeDiff()
	rec.SetFinalPatch(patch)
	for _, m := range mutations {
		rec.RecordFileMutation(m)
	}
	rec.SetASTMetrics(astMetrics)
	rec.SetTokenUsage(trajectory.TokenUsage{
		PromptTokens:     4800,
		CompletionTokens: 1100,
		TotalTokens:      5900,
		EstimatedCostUSD: 0.019,
	})

	verified := (exitCode == 0)
	status := "COMPLETED"
	if !verified {
		status = "FAILED"
	}
	return rec.Complete(status, verified, stdout), nil
}

// runChurnScenario: applies cosmetic comments and reformatting, reducing surgical score.
func runChurnScenario(rec *trajectory.Recorder, ws *Workspace, task config.TaskDefinition, targetFile string) (*trajectory.Trajectory, error) {
	rec.StartTurn("Reading target file.")
	content, _ := ws.ReadFile(targetFile)
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "read_file",
		Arguments:        map[string]interface{}{"path": targetFile},
		NormalizedTarget: targetFile,
		Status:           "SUCCESS",
		DurationMs:       15,
		Output:           content,
	})
	rec.EndTurn()

	rec.StartTurn("Rewriting file with full reformatting and extra comments.")
	applyMockTaskFix(ws, task, true) // churn mode
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "write_file",
		Arguments:        map[string]interface{}{"path": targetFile},
		NormalizedTarget: targetFile,
		Status:           "SUCCESS",
		DurationMs:       45,
		Output:           "Rewrote entire file.",
	})
	rec.EndTurn()

	rec.StartTurn("Running verification test.")
	stdout, stderr, exitCode, durMs, _ := ws.ExecCommand(task.TestCommand, time.Duration(task.TimeoutSeconds)*time.Second)
	rec.RecordToolCall(trajectory.ToolCall{
		ToolName:         "run_tests",
		Arguments:        map[string]interface{}{"command": task.TestCommand},
		NormalizedTarget: task.TestCommand,
		Status:           "SUCCESS",
		DurationMs:       durMs,
		Output:           stdout,
	})
	rec.RecordTerminalExecution(trajectory.TerminalExecution{
		Command:    task.TestCommand,
		ExitCode:   exitCode,
		Stdout:     stdout,
		Stderr:     stderr,
		DurationMs: durMs,
	})
	rec.EndTurn()

	patch, mutations, astMetrics, _ := ws.ComputeDiff()
	rec.SetFinalPatch(patch)
	for _, m := range mutations {
		rec.RecordFileMutation(m)
	}
	rec.SetASTMetrics(astMetrics)
	rec.SetTokenUsage(trajectory.TokenUsage{
		PromptTokens:     2400,
		CompletionTokens: 850,
		TotalTokens:      3250,
		EstimatedCostUSD: 0.011,
	})

	verified := (exitCode == 0)
	status := "COMPLETED"
	if !verified {
		status = "FAILED"
	}
	return rec.Complete(status, verified, stdout), nil
}

// runDoomLoopScenario: repeats failing command multiple times without code edit.
func runDoomLoopScenario(rec *trajectory.Recorder, ws *Workspace, task config.TaskDefinition, targetFile string) (*trajectory.Trajectory, error) {
	for i := 1; i <= 3; i++ {
		rec.StartTurn(fmt.Sprintf("Attempt %d running test command without modifying code.", i))
		stdout, stderr, exitCode, durMs, _ := ws.ExecCommand(task.TestCommand, time.Duration(task.TimeoutSeconds)*time.Second)
		rec.RecordToolCall(trajectory.ToolCall{
			ToolName:         "run_tests",
			Arguments:        map[string]interface{}{"command": task.TestCommand},
			NormalizedTarget: task.TestCommand,
			Status:           "ERROR",
			DurationMs:       durMs,
			Output:           stderr,
		})
		rec.RecordTerminalExecution(trajectory.TerminalExecution{
			Command:    task.TestCommand,
			ExitCode:   exitCode,
			Stdout:     stdout,
			Stderr:     stderr,
			DurationMs: durMs,
		})
		rec.EndTurn()
	}

	rec.SetTokenUsage(trajectory.TokenUsage{
		PromptTokens:     3500,
		CompletionTokens: 600,
		TotalTokens:      4100,
		EstimatedCostUSD: 0.012,
	})
	return rec.Complete("FAILED", false, "Test command repeatedly failed in doom loop"), nil
}

// applyMockTaskFix resolves known bugs in the benchmark tasks.
func applyMockTaskFix(ws *Workspace, task config.TaskDefinition, churn bool) {
	targetFile := "main.go"
	if len(task.TargetFiles) > 0 {
		targetFile = task.TargetFiles[0]
	}

	content, err := ws.ReadFile(targetFile)
	if err != nil {
		return
	}

	// Task 1: Go race condition
	if strings.Contains(task.ID, "go") || strings.Contains(targetFile, "counter.go") || strings.Contains(content, "c.count++") {
		if churn {
			// Add heavy cosmetic comments and blank lines
			churnText := "// Counter holds synchronized counter state\n// Mutex added for concurrency protection\n" +
				"// Thread safety ensured via sync.Mutex\n\n\n\n\tc.mu.Lock()\n\tc.count++\n\tc.mu.Unlock()\n\n\n"
			_ = ws.EditFile(targetFile, "\tc.count++", churnText)
		} else {
			_ = ws.EditFile(targetFile, "\tc.count++", "\tc.mu.Lock()\n\tc.count++\n\tc.mu.Unlock()")
		}
		return
	}

	// Task 2: Python data pipeline
	if strings.Contains(task.ID, "python") || strings.Contains(targetFile, "pipeline.py") || strings.Contains(content, "record['amount']") {
		if churn {
			churnPy := "# Safe extraction of amount\n        # Checks if record is valid\n        # Handles None values\n\n\n        if record and 'amount' in record and record['amount'] is not None:\n            total += float(record['amount'])\n\n"
			_ = ws.EditFile(targetFile, "        total += float(record['amount'])", churnPy)
		} else {
			_ = ws.EditFile(targetFile, "        total += float(record['amount'])", "        if record and 'amount' in record and record['amount'] is not None:\n            total += float(record['amount'])")
		}
		return
	}

	// Task 3: TypeScript null safety
	if strings.Contains(task.ID, "typescript") || strings.Contains(targetFile, "user.ts") || strings.Contains(content, "user.profile.age") {
		if churn {
			churnTS := "  // Retrieve user age safely\n  // Protect against null profile\n\n\n  if (!user || !user.profile) {\n    return 0;\n  }\n  return user.profile.age;\n\n"
			_ = ws.EditFile(targetFile, "  return user.profile.age;", churnTS)
		} else {
			_ = ws.EditFile(targetFile, "  return user.profile.age;", "  if (!user || !user.profile) {\n    return 0;\n  }\n  return user.profile.age;")
		}
		return
	}
}
