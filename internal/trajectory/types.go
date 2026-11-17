package trajectory

import "time"

// TaskInfo holds metadata about an evaluation task.
type TaskInfo struct {
	ID             string   `json:"id" yaml:"id"`
	Title          string   `json:"title" yaml:"title"`
	Description    string   `json:"description" yaml:"description"`
	Language       string   `json:"language" yaml:"language"` // "go", "python", "typescript"
	TargetFiles    []string `json:"target_files" yaml:"target_files"`
	TestCommand    string   `json:"test_command" yaml:"test_command"`
	TimeoutSeconds int      `json:"timeout_seconds" yaml:"timeout_seconds"`
}

// Trajectory contains the serial execution trace of an agent solving a task.
type Trajectory struct {
	SchemaVersion      string              `json:"schema_version"`
	TaskID             string              `json:"task_id"`
	AgentID            string              `json:"agent_id"`
	Model              string              `json:"model"`
	SystemPromptHash   string              `json:"system_prompt_hash,omitempty"`
	StartedAt          time.Time           `json:"started_at"`
	CompletedAt        time.Time           `json:"completed_at"`
	DurationMs         int64               `json:"duration_ms"`
	Status             string              `json:"status"` // "COMPLETED", "FAILED", "TIMEOUT", "ERROR"
	Verified           bool                `json:"verified"`
	VerificationOutput string              `json:"verification_output,omitempty"`
	Turns              []Turn              `json:"turns"`
	FinalPatch         string              `json:"final_patch,omitempty"`
	ASTMetrics         ASTMetrics          `json:"ast_metrics"`
	TokenUsage         TokenUsage          `json:"token_usage"`
	AntiPatterns       []AntiPattern       `json:"anti_patterns,omitempty"`
}

// Turn represents a single reasoning and action cycle.
type Turn struct {
	TurnIndex          int                 `json:"turn_index"`
	Timestamp          time.Time           `json:"timestamp"`
	DurationMs         int64               `json:"duration_ms"`
	Thought            string              `json:"thought"`
	ToolCalls          []ToolCall          `json:"tool_calls"`
	TerminalExecutions []TerminalExecution `json:"terminal_executions,omitempty"`
	FileMutations      []FileMutation      `json:"file_mutations,omitempty"`
}

// ToolCall represents an individual tool invocation by the agent.
type ToolCall struct {
	CallID           string                 `json:"call_id"`
	ToolName         string                 `json:"tool_name"`
	Arguments        map[string]interface{} `json:"arguments"`
	NormalizedTarget string                 `json:"normalized_target,omitempty"`
	Status           string                 `json:"status"` // "SUCCESS", "ERROR"
	DurationMs       int64                  `json:"duration_ms"`
	Output           string                 `json:"output"`
}

// TerminalExecution records the execution of a shell command.
type TerminalExecution struct {
	Command    string `json:"command"`
	ExitCode   int    `json:"exit_code"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMs int64  `json:"duration_ms"`
}

// FileMutation represents changes made to a file.
type FileMutation struct {
	Path              string        `json:"path"`
	ChangeType        string        `json:"change_type"` // "ADDED", "MODIFIED", "DELETED"
	LinesAdded        int           `json:"lines_added"`
	LinesRemoved      int           `json:"lines_removed"`
	Diff              string        `json:"diff"`
	ASTNodesAffected  []ASTNodeRef  `json:"ast_nodes_affected,omitempty"`
}

// ASTNodeRef locates an AST element modified in the file.
type ASTNodeRef struct {
	Kind       string `json:"kind"` // "FuncDecl", "IfStmt", "AssignStmt", etc.
	Identifier string `json:"identifier"`
	LineStart  int    `json:"line_start"`
	LineEnd    int    `json:"line_end"`
}

// ASTMetrics captures structural precision metrics of code mutations.
type ASTMetrics struct {
	TotalLinesDiff     int     `json:"total_lines_diff"`
	ASTNodesModified   int     `json:"ast_nodes_modified"`
	SurgicalScore      float64 `json:"surgical_score"`
	CosmeticChurnRatio float64 `json:"cosmetic_churn_ratio"`
	WhitespaceLines    int     `json:"whitespace_lines"`
	CommentLines       int     `json:"comment_lines"`
}

// TokenUsage tracks token counts and estimated costs.
type TokenUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// AntiPattern records an operational pathology detected during execution.
type AntiPattern struct {
	Kind      string `json:"kind"` // "REDUNDANT_READ", "DOOM_LOOP", "SHOTGUN_SURGERY", "TOOL_BLOAT", "CONTEXT_BLOAT"
	Severity  string `json:"severity"` // "INFO", "WARNING", "CRITICAL"
	TurnIndex int    `json:"turn_index"`
	ToolName  string `json:"tool_name,omitempty"`
	Details   string `json:"details"`
}

// AlignmentOp represents an alignment action between two trajectory steps.
type AlignmentOp string

const (
	OpMatch        AlignmentOp = "MATCH"
	OpMismatch     AlignmentOp = "MISMATCH"
	OpCandidateGap AlignmentOp = "CANDIDATE_INSERT"
	OpBaselineGap  AlignmentOp = "BASELINE_INSERT"
)

// AlignmentStep is a single matched or gapped entry in sequence alignment.
type AlignmentStep struct {
	Index          int         `json:"index"`
	Op             AlignmentOp `json:"op"`
	CandidateTool  *ToolCall   `json:"candidate_tool,omitempty"`
	BaselineTool   *ToolCall   `json:"baseline_tool,omitempty"`
	CandidateTurn  int         `json:"candidate_turn,omitempty"`
	BaselineTurn   int         `json:"baseline_turn,omitempty"`
}

// ComparisonResult stores the comparative analysis between candidate and baseline.
type ComparisonResult struct {
	TaskID                       string          `json:"task_id"`
	CandidateModel               string          `json:"candidate_model"`
	BaselineModel                string          `json:"baseline_model"`
	CandidateStatus              string          `json:"candidate_status"`
	BaselineStatus               string          `json:"baseline_status"`
	CandidateVerified            bool            `json:"candidate_verified"`
	BaselineVerified             bool            `json:"baseline_verified"`
	CandidateDurationMs          int64           `json:"candidate_duration_ms"`
	BaselineDurationMs           int64           `json:"baseline_duration_ms"`
	DurationDeltaMs              int64           `json:"duration_delta_ms"`
	SpeedupFactor                float64         `json:"speedup_factor"`
	CandidateToolCount           int             `json:"candidate_tool_count"`
	BaselineToolCount            int             `json:"baseline_tool_count"`
	ToolBloatFactor              float64         `json:"tool_bloat_factor"`
	ToolEfficiencyCandidate      float64         `json:"tool_efficiency_candidate"`
	ToolEfficiencyBaseline       float64         `json:"tool_efficiency_baseline"`
	ToolEfficiencyDelta          float64         `json:"tool_efficiency_delta"`
	CandidateSurgicalScore       float64         `json:"candidate_surgical_score"`
	BaselineSurgicalScore        float64         `json:"baseline_surgical_score"`
	SurgicalScoreDelta           float64         `json:"surgical_score_delta"`
	TrajectoryEditDistance       int             `json:"trajectory_edit_distance"`
	NormalizedTrajectoryDistance float64         `json:"normalized_trajectory_distance"`
	DivergenceTurn               int             `json:"divergence_turn"`
	AlignmentSteps               []AlignmentStep `json:"alignment_steps"`
	CandidateAntiPatterns        []AntiPattern   `json:"candidate_anti_patterns"`
	BaselineAntiPatterns         []AntiPattern   `json:"baseline_anti_patterns"`
	Regressions                  []string        `json:"regressions"`
	Verdict                      string          `json:"verdict"` // "PASS", "REGRESSION", "INCONCLUSIVE"
}
