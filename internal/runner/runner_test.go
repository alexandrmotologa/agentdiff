package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/agentdiff/internal/config"
)

func TestWorkspaceOperations(t *testing.T) {
	tmpDir := t.TempDir()
	sampleFile := filepath.Join(tmpDir, "counter.go")
	initialCode := `package main

type Counter struct {
	count int
}

func (c *Counter) Inc() {
	c.count++
}
`
	if err := os.WriteFile(sampleFile, []byte(initialCode), 0644); err != nil {
		t.Fatalf("writing initial code: %v", err)
	}

	ws, err := NewWorkspace("test_task", "test_run", tmpDir)
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Read
	data, err := ws.ReadFile("counter.go")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if data != initialCode {
		t.Errorf("content mismatch")
	}

	// Edit
	err = ws.EditFile("counter.go", "c.count++", "c.count += 2")
	if err != nil {
		t.Fatalf("EditFile failed: %v", err)
	}

	// Diff
	patch, mutations, astMetrics, err := ws.ComputeDiff()
	if err != nil {
		t.Fatalf("ComputeDiff failed: %v", err)
	}
	if len(mutations) != 1 {
		t.Fatalf("expected 1 mutation, got %d", len(mutations))
	}
	if astMetrics.TotalLinesDiff == 0 {
		t.Errorf("expected non-zero lines diff in AST metrics")
	}
	if patch == "" {
		t.Errorf("expected non-empty patch")
	}
}

func TestMockAgentExecution(t *testing.T) {
	tmpDir := t.TempDir()
	codeFile := filepath.Join(tmpDir, "counter.go")
	initialCode := `package main

import "sync"

type Counter struct {
	mu sync.Mutex
	count int
}

func (c *Counter) Inc() {
	c.count++
}
`
	_ = os.WriteFile(codeFile, []byte(initialCode), 0644)

	ws, err := NewWorkspace("test_task", "mock_run", tmpDir)
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	task := config.TaskDefinition{
		ID:             "task_01_go_race",
		Title:          "Fix Data Race",
		Language:       "go",
		WorkingDir:     tmpDir,
		TargetFiles:    []string{"counter.go"},
		TestCommand:    "go version", // simple command guaranteed to pass
		TimeoutSeconds: 30,
	}

	modelCfg := config.ModelConfig{
		ID:       "test-agent",
		Provider: "mock",
		Model:    "mock-efficient",
	}

	traj, err := RunMockAgent(ws, task, modelCfg)
	if err != nil {
		t.Fatalf("RunMockAgent failed: %v", err)
	}
	if !traj.Verified {
		t.Errorf("expected mock efficient agent to verify successfully")
	}
	if len(traj.Turns) != 3 {
		t.Errorf("expected 3 turns for efficient mock agent, got %d", len(traj.Turns))
	}
}
