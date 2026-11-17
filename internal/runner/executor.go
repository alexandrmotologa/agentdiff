package runner

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/config"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// ExecuteAgent delegates task execution to the configured provider runner.
func ExecuteAgent(ws *Workspace, task config.TaskDefinition, modelCfg config.ModelConfig) (*trajectory.Trajectory, error) {
	provider := strings.ToLower(modelCfg.Provider)
	if provider == "" {
		provider = "mock"
	}

	switch provider {
	case "mock":
		return RunMockAgent(ws, task, modelCfg)
	case "openai", "anthropic", "ollama":
		return RunBuiltInAgent(ws, task, modelCfg)
	case "cli":
		return runCLIAgent(ws, task, modelCfg)
	default:
		return nil, fmt.Errorf("unsupported agent provider: %s", modelCfg.Provider)
	}
}

func runCLIAgent(ws *Workspace, task config.TaskDefinition, modelCfg config.ModelConfig) (*trajectory.Trajectory, error) {
	rec := trajectory.NewRecorder(task.ID, modelCfg.ID, modelCfg.Model, modelCfg.SystemPrompt)
	rec.StartTurn(fmt.Sprintf("Executing external agent CLI: %s", modelCfg.CLICommand))

	fullCmd := modelCfg.CLICommand
	if len(modelCfg.CLIArgs) > 0 {
		fullCmd += " " + strings.Join(modelCfg.CLIArgs, " ")
	}

	stdout, stderr, exitCode, durMs, _ := ws.ExecCommand(fullCmd, 180)
	rec.RecordTerminalExecution(trajectory.TerminalExecution{
		Command:    fullCmd,
		ExitCode:   exitCode,
		Stdout:     stdout,
		Stderr:     stderr,
		DurationMs: durMs,
	})
	rec.EndTurn()

	// Verification
	testStdout, _, testExit, _, _ := ws.ExecCommand(task.TestCommand, 180)
	verified := (testExit == 0)
	status := "COMPLETED"
	if !verified {
		status = "FAILED"
	}

	patch, mutations, astMetrics, _ := ws.ComputeDiff()
	rec.SetFinalPatch(patch)
	for _, m := range mutations {
		rec.RecordFileMutation(m)
	}
	rec.SetASTMetrics(astMetrics)

	return rec.Complete(status, verified, testStdout), nil
}
