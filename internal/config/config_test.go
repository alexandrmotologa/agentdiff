package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigYAML(t *testing.T) {
	content := `
name: "Regression Test Suite"
candidate:
  id: "candidate-v2"
  provider: "mock"
  model: "test-model-candidate"
baseline:
  id: "baseline-v1"
  provider: "mock"
  model: "test-model-baseline"
tasks:
  - id: "task_01"
    title: "Fix Race Condition"
    language: "go"
    test_command: "go test -race ./..."
    timeout_seconds: 60
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "agentdiff.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Name != "Regression Test Suite" {
		t.Errorf("expected name 'Regression Test Suite', got %q", cfg.Name)
	}
	if cfg.Candidate.Model != "test-model-candidate" {
		t.Errorf("expected candidate model 'test-model-candidate', got %q", cfg.Candidate.Model)
	}
	if cfg.Baseline.Model != "test-model-baseline" {
		t.Errorf("expected baseline model 'test-model-baseline', got %q", cfg.Baseline.Model)
	}
	if len(cfg.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(cfg.Tasks))
	}
	if cfg.Tasks[0].ID != "task_01" {
		t.Errorf("expected task id 'task_01', got %q", cfg.Tasks[0].ID)
	}
	if cfg.Gates.MaxToolBloatFactor != 1.25 {
		t.Errorf("expected default MaxToolBloatFactor 1.25, got %v", cfg.Gates.MaxToolBloatFactor)
	}
}
