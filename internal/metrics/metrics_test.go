package metrics

import (
	"testing"
	"time"

	"github.com/alexandrmotologa/agentdiff/internal/config"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

func TestMetricsAntiPatternsAndComparison(t *testing.T) {
	now := time.Now()

	// Candidate with redundant reads and bloat
	cand := &trajectory.Trajectory{
		TaskID:     "task_01",
		Model:      "model-cand",
		Status:     "COMPLETED",
		Verified:   true,
		DurationMs: 20000,
		ASTMetrics: trajectory.ASTMetrics{
			TotalLinesDiff:   10,
			ASTNodesModified: 3,
			SurgicalScore:    0.30,
		},
		Turns: []trajectory.Turn{
			{
				TurnIndex: 1,
				Timestamp: now,
				ToolCalls: []trajectory.ToolCall{
					{ToolName: "read_file", NormalizedTarget: "worker.go"},
				},
			},
			{
				TurnIndex: 2,
				Timestamp: now,
				ToolCalls: []trajectory.ToolCall{
					{ToolName: "read_file", NormalizedTarget: "worker.go"}, // redundant read
				},
			},
			{
				TurnIndex: 3,
				Timestamp: now,
				ToolCalls: []trajectory.ToolCall{
					{ToolName: "read_file", NormalizedTarget: "worker.go"}, // redundant read
				},
			},
			{
				TurnIndex: 4,
				Timestamp: now,
				TerminalExecutions: []trajectory.TerminalExecution{
					{Command: "go test", ExitCode: 1},
				},
			},
			{
				TurnIndex: 5,
				Timestamp: now,
				TerminalExecutions: []trajectory.TerminalExecution{
					{Command: "go test", ExitCode: 1},
				},
			},
			{
				TurnIndex: 6,
				Timestamp: now,
				TerminalExecutions: []trajectory.TerminalExecution{
					{Command: "go test", ExitCode: 1}, // doom loop
				},
			},
		},
	}

	// Baseline without bloat
	base := &trajectory.Trajectory{
		TaskID:     "task_01",
		Model:      "model-base",
		Status:     "COMPLETED",
		Verified:   true,
		DurationMs: 10000,
		ASTMetrics: trajectory.ASTMetrics{
			TotalLinesDiff:   4,
			ASTNodesModified: 3,
			SurgicalScore:    0.75,
		},
		Turns: []trajectory.Turn{
			{
				TurnIndex: 1,
				Timestamp: now,
				ToolCalls: []trajectory.ToolCall{
					{ToolName: "read_file", NormalizedTarget: "worker.go"},
				},
			},
			{
				TurnIndex: 2,
				Timestamp: now,
				ToolCalls: []trajectory.ToolCall{
					{ToolName: "edit_file", NormalizedTarget: "worker.go"},
				},
				FileMutations: []trajectory.FileMutation{
					{Path: "worker.go", LinesAdded: 2},
				},
			},
		},
	}

	gates := config.DefaultGates()
	comp := CompareTrajectories(cand, base, []string{"worker.go"}, gates)

	if comp.Verdict != "REGRESSION" {
		t.Errorf("expected VERDICT REGRESSION, got %s", comp.Verdict)
	}
	if len(comp.Regressions) == 0 {
		t.Errorf("expected detected regressions")
	}

	terminalRep := FormatTerminalReport(comp)
	if terminalRep == "" {
		t.Errorf("expected non-empty terminal report")
	}

	mdRep := FormatMarkdownSummary(comp)
	if mdRep == "" {
		t.Errorf("expected non-empty markdown summary")
	}
}
