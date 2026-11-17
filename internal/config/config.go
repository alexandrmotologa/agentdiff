package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultGates provides conservative production-ready regression bounds.
func DefaultGates() RegressionGates {
	return RegressionGates{
		MaxToolBloatFactor:      1.25,
		MinSurgicalScore:        0.40,
		MaxSurgicalScoreDrop:    0.15,
		MaxLatencyIncreaseRatio: 1.30,
		MaxCosmeticChurnRatio:   0.45,
		RequirePassRateParity:   true,
		FailOnDoomLoop:          true,
	}
}

// LoadConfig reads and parses a benchmark configuration file (YAML or JSON).
func LoadConfig(path string) (*BenchmarkConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	cfg := &BenchmarkConfig{
		Concurrency:    2,
		DefaultTimeout: 180,
		Gates:          DefaultGates(),
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("unmarshaling json config: %w", err)
		}
	default:
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("unmarshaling yaml config: %w", err)
		}
	}

	// Resolve environment variables for API keys
	resolveAPIKey(&cfg.Candidate)
	resolveAPIKey(&cfg.Baseline)

	// Apply sensible defaults
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 2
	}
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = 180
	}
	if cfg.Candidate.MaxTurns <= 0 {
		cfg.Candidate.MaxTurns = 25
	}
	if cfg.Baseline.MaxTurns <= 0 {
		cfg.Baseline.MaxTurns = 25
	}

	// If TasksDir is provided, scan and load individual task files
	if cfg.TasksDir != "" {
		tasks, err := DiscoverTasks(cfg.TasksDir, cfg.DefaultTimeout)
		if err != nil {
			return nil, fmt.Errorf("discovering tasks from %q: %w", cfg.TasksDir, err)
		}
		cfg.Tasks = append(cfg.Tasks, tasks...)
	}

	// Validate configuration
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid benchmark config: %w", err)
	}

	return cfg, nil
}

func resolveAPIKey(m *ModelConfig) {
	if m.APIKey == "" && m.APIKeyEnv != "" {
		m.APIKey = os.Getenv(m.APIKeyEnv)
	}
}

// ValidateConfig verifies that required fields are set and properly configured.
func ValidateConfig(cfg *BenchmarkConfig) error {
	if cfg.Candidate.Model == "" {
		return fmt.Errorf("candidate model must be specified")
	}
	if cfg.Baseline.Model == "" {
		return fmt.Errorf("baseline model must be specified")
	}
	if len(cfg.Tasks) == 0 {
		return fmt.Errorf("at least one task must be defined or discovered in tasks_dir")
	}
	for i, task := range cfg.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task #%d is missing an id", i+1)
		}
		if task.TestCommand == "" {
			return fmt.Errorf("task %q is missing a test_command", task.ID)
		}
	}
	return nil
}

// DiscoverTasks scans a directory for task definitions (task.yaml or subdirectories).
func DiscoverTasks(dir string, defaultTimeout int) ([]TaskDefinition, error) {
	var tasks []TaskDefinition

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskYamlPath := filepath.Join(dir, entry.Name(), "task.yaml")
		if _, err := os.Stat(taskYamlPath); err == nil {
			data, err := os.ReadFile(taskYamlPath)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", taskYamlPath, err)
			}
			var t TaskDefinition
			if err := yaml.Unmarshal(data, &t); err != nil {
				return nil, fmt.Errorf("parsing %s: %w", taskYamlPath, err)
			}
			if t.ID == "" {
				t.ID = entry.Name()
			}
			if t.WorkingDir == "" {
				t.WorkingDir = filepath.Join(dir, entry.Name())
			} else if !filepath.IsAbs(t.WorkingDir) {
				t.WorkingDir = filepath.Join(dir, entry.Name(), t.WorkingDir)
			}
			if t.TimeoutSeconds <= 0 {
				t.TimeoutSeconds = defaultTimeout
			}
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}
