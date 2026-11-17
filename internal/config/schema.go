package config

// ModelConfig defines an agent model runner configuration.
type ModelConfig struct {
	ID           string   `yaml:"id" json:"id"`
	Provider     string   `yaml:"provider" json:"provider"` // "mock", "openai", "anthropic", "ollama", "cli"
	Model        string   `yaml:"model" json:"model"`
	Endpoint     string   `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	APIKeyEnv    string   `yaml:"api_key_env,omitempty" json:"api_key_env,omitempty"`
	APIKey       string   `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	SystemPrompt string   `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	MaxTurns     int      `yaml:"max_turns,omitempty" json:"max_turns,omitempty"`
	Temperature  *float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	CLICommand   string   `yaml:"cli_command,omitempty" json:"cli_command,omitempty"`
	CLIArgs      []string `yaml:"cli_args,omitempty" json:"cli_args,omitempty"`
}

// RegressionGates defines tolerance thresholds for candidate vs baseline.
type RegressionGates struct {
	MaxToolBloatFactor      float64 `yaml:"max_tool_bloat_factor" json:"max_tool_bloat_factor"`             // e.g. 1.25 (+25%)
	MinSurgicalScore        float64 `yaml:"min_surgical_score" json:"min_surgical_score"`                   // e.g. 0.40
	MaxSurgicalScoreDrop    float64 `yaml:"max_surgical_score_drop" json:"max_surgical_score_drop"`         // e.g. 0.15
	MaxLatencyIncreaseRatio float64 `yaml:"max_latency_increase_ratio" json:"max_latency_increase_ratio"`   // e.g. 1.30 (+30%)
	MaxCosmeticChurnRatio   float64 `yaml:"max_cosmetic_churn_ratio" json:"max_cosmetic_churn_ratio"`       // e.g. 0.40
	RequirePassRateParity   bool    `yaml:"require_pass_rate_parity" json:"require_pass_rate_parity"`       // must pass if baseline passes
	FailOnDoomLoop          bool    `yaml:"fail_on_doom_loop" json:"fail_on_doom_loop"`                     // flag as regression if doom loop occurs
}

// TaskDefinition defines a benchmark problem to be solved by agents.
type TaskDefinition struct {
	ID             string   `yaml:"id" json:"id"`
	Title          string   `yaml:"title" json:"title"`
	Description    string   `yaml:"description" json:"description"`
	Language       string   `yaml:"language" json:"language"` // "go", "python", "typescript"
	WorkingDir     string   `yaml:"working_dir" json:"working_dir"` // source path of the task template
	TargetFiles    []string `yaml:"target_files" json:"target_files"`
	TestCommand    string   `yaml:"test_command" json:"test_command"`
	TimeoutSeconds int      `yaml:"timeout_seconds" json:"timeout_seconds"`
	ExpectedOutput string   `yaml:"expected_output,omitempty" json:"expected_output,omitempty"`
}

// BenchmarkConfig represents the root configuration file (agentdiff.yaml).
type BenchmarkConfig struct {
	Name            string           `yaml:"name" json:"name"`
	Description     string           `yaml:"description,omitempty" json:"description,omitempty"`
	Candidate       ModelConfig      `yaml:"candidate" json:"candidate"`
	Baseline        ModelConfig      `yaml:"baseline" json:"baseline"`
	Gates           RegressionGates  `yaml:"gates" json:"gates"`
	Concurrency     int              `yaml:"concurrency" json:"concurrency"`
	DefaultTimeout  int              `yaml:"default_timeout" json:"default_timeout"`
	TasksDir        string           `yaml:"tasks_dir,omitempty" json:"tasks_dir,omitempty"`
	Tasks           []TaskDefinition `yaml:"tasks,omitempty" json:"tasks,omitempty"`
}
