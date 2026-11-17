package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/agentdiff/internal/config"
)

func TestSuiteExecution(t *testing.T) {
	// Create mock task directory with a Go task
	tmpBase := t.TempDir()
	taskDir := filepath.Join(tmpBase, "task_dummy")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	taskYaml := `id: "task_dummy"
title: "Dummy Task"
language: "go"
target_files: ["main.go"]
test_command: "go version"
timeout_seconds: 30
`
	_ = os.WriteFile(filepath.Join(taskDir, "task.yaml"), []byte(taskYaml), 0644)
	_ = os.WriteFile(filepath.Join(taskDir, "main.go"), []byte("package main\n"), 0644)

	tasks, err := config.DiscoverTasks(tmpBase, 30)
	if err != nil {
		t.Fatalf("DiscoverTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 discovered task, got %d", len(tasks))
	}

	cfg := &config.BenchmarkConfig{
		Name: "Test Suite",
		Candidate: config.ModelConfig{
			ID:       "candidate",
			Provider: "mock",
			Model:    "mock-v2-regressed",
		},
		Baseline: config.ModelConfig{
			ID:       "baseline",
			Provider: "mock",
			Model:    "mock-efficient",
		},
		Gates:       config.DefaultGates(),
		Concurrency: 1,
		Tasks:       tasks,
	}

	outDir := filepath.Join(tmpBase, "results")
	results, err := RunBenchmarkSuite(cfg, outDir, nil)
	if err != nil {
		t.Fatalf("RunBenchmarkSuite failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	res := results[0]
	if res.TaskID != "task_dummy" {
		t.Errorf("expected task_dummy, got %s", res.TaskID)
	}
	if res.ToolBloatFactor <= 1.0 {
		t.Errorf("expected candidate bloat factor > 1.0, got %f", res.ToolBloatFactor)
	}

	// Verify generated trace files
	candTrace := filepath.Join(outDir, "task_dummy", "candidate.trajectory.json")
	if _, err := os.Stat(candTrace); err != nil {
		t.Errorf("expected candidate trace file to exist: %v", err)
	}
}
