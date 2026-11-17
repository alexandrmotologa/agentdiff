package trajectory

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRecorderAndSerializer(t *testing.T) {
	rec := NewRecorder("task_01", "agent_cand", "test-model", "System prompt here")
	t1 := rec.StartTurn("Inspecting files")
	if t1 != 1 {
		t.Fatalf("expected turn 1, got %d", t1)
	}

	rec.RecordToolCall(ToolCall{
		ToolName:         "list_dir",
		Arguments:        map[string]interface{}{"path": "."},
		NormalizedTarget: ".",
		Status:           "SUCCESS",
		DurationMs:       10,
		Output:           "main.go",
	})
	rec.EndTurn()

	rec.StartTurn("Editing code")
	rec.RecordToolCall(ToolCall{
		ToolName:         "edit_file",
		Arguments:        map[string]interface{}{"path": "main.go"},
		NormalizedTarget: "main.go",
		Status:           "SUCCESS",
		DurationMs:       25,
		Output:           "Saved",
	})
	rec.RecordFileMutation(FileMutation{
		Path:         "main.go",
		ChangeType:   "MODIFIED",
		LinesAdded:   2,
		LinesRemoved: 1,
		Diff:         "@@ -1,1 +1,2 @@",
	})
	rec.EndTurn()

	traj := rec.Complete("COMPLETED", true, "All tests passed")
	if traj.Status != "COMPLETED" || !traj.Verified {
		t.Errorf("unexpected status or verified flag: %s, %v", traj.Status, traj.Verified)
	}
	if len(traj.Turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(traj.Turns))
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "trace.json")
	if err := SaveTrajectory(filePath, traj); err != nil {
		t.Fatalf("SaveTrajectory failed: %v", err)
	}

	loaded, err := LoadTrajectory(filePath)
	if err != nil {
		t.Fatalf("LoadTrajectory failed: %v", err)
	}
	if loaded.TaskID != "task_01" || len(loaded.Turns) != 2 {
		t.Errorf("loaded trajectory mismatch")
	}
}

func TestAlignmentAndEditDistance(t *testing.T) {
	now := time.Now()

	// Candidate: list_dir -> read_file:a.go -> edit_file:a.go -> run_tests
	cand := &Trajectory{
		TaskID: "task_test",
		Turns: []Turn{
			{
				TurnIndex: 1,
				Timestamp: now,
				ToolCalls: []ToolCall{
					{ToolName: "list_dir", NormalizedTarget: "."},
					{ToolName: "read_file", NormalizedTarget: "a.go"},
				},
			},
			{
				TurnIndex: 2,
				Timestamp: now,
				ToolCalls: []ToolCall{
					{ToolName: "edit_file", NormalizedTarget: "a.go"},
					{ToolName: "run_tests", NormalizedTarget: "go test"},
				},
			},
		},
	}

	// Baseline: list_dir -> edit_file:a.go -> run_tests (missing read_file:a.go)
	base := &Trajectory{
		TaskID: "task_test",
		Turns: []Turn{
			{
				TurnIndex: 1,
				Timestamp: now,
				ToolCalls: []ToolCall{
					{ToolName: "list_dir", NormalizedTarget: "."},
					{ToolName: "edit_file", NormalizedTarget: "a.go"},
				},
			},
			{
				TurnIndex: 2,
				Timestamp: now,
				ToolCalls: []ToolCall{
					{ToolName: "run_tests", NormalizedTarget: "go test"},
				},
			},
		},
	}

	editDist, normDist, divTurn, steps := AlignTrajectories(cand, base)

	// One insertion in cand (read_file:a.go)
	if editDist != 1 {
		t.Errorf("expected edit distance 1, got %d", editDist)
	}
	if normDist != 0.25 {
		t.Errorf("expected normalized distance 0.25 (1/4), got %f", normDist)
	}
	if divTurn != 1 {
		t.Errorf("expected divergence turn 1, got %d", divTurn)
	}
	if len(steps) != 4 {
		t.Fatalf("expected 4 alignment steps, got %d", len(steps))
	}
}
