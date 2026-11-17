export interface ToolCall {
  call_id: string;
  tool_name: string;
  arguments: Record<string, any>;
  normalized_target?: string;
  status: string;
  duration_ms: number;
  output: string;
}

export interface TerminalExecution {
  command: string;
  exit_code: number;
  stdout: string;
  stderr: string;
  duration_ms: number;
}

export interface ASTNodeRef {
  kind: string;
  identifier: string;
  line_start: number;
  line_end: number;
}

export interface FileMutation {
  path: string;
  change_type: string;
  lines_added: number;
  lines_removed: number;
  diff: string;
  ast_nodes_affected?: ASTNodeRef[];
}

export interface Turn {
  turn_index: number;
  timestamp: string;
  duration_ms: number;
  thought: string;
  tool_calls: ToolCall[];
  terminal_executions?: TerminalExecution[];
  file_mutations?: FileMutation[];
}

export interface ASTMetrics {
  total_lines_diff: number;
  ast_nodes_modified: number;
  surgical_score: number;
  cosmetic_churn_ratio: number;
  whitespace_lines: number;
  comment_lines: number;
}

export interface TokenUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  estimated_cost_usd: number;
}

export interface AntiPattern {
  kind: string;
  severity: "INFO" | "WARNING" | "CRITICAL";
  turn_index: number;
  tool_name?: string;
  details: string;
}

export interface Trajectory {
  schema_version: string;
  task_id: string;
  agent_id: string;
  model: string;
  started_at: string;
  completed_at: string;
  duration_ms: number;
  status: string;
  verified: boolean;
  verification_output?: string;
  turns: Turn[];
  final_patch?: string;
  ast_metrics: ASTMetrics;
  token_usage: TokenUsage;
  anti_patterns?: AntiPattern[];
}

export interface AlignmentStep {
  index: number;
  op: "MATCH" | "MISMATCH" | "CANDIDATE_INSERT" | "BASELINE_INSERT";
  candidate_tool?: ToolCall;
  baseline_tool?: ToolCall;
  candidate_turn?: number;
  baseline_turn?: number;
}

export interface ComparisonResult {
  task_id: string;
  candidate_model: string;
  baseline_model: string;
  candidate_status: string;
  baseline_status: string;
  candidate_verified: boolean;
  baseline_verified: boolean;
  candidate_duration_ms: number;
  baseline_duration_ms: number;
  duration_delta_ms: number;
  speedup_factor: number;
  candidate_tool_count: number;
  baseline_tool_count: number;
  tool_bloat_factor: number;
  tool_efficiency_candidate: number;
  tool_efficiency_baseline: number;
  tool_efficiency_delta: number;
  candidate_surgical_score: number;
  baseline_surgical_score: number;
  surgical_score_delta: number;
  trajectory_edit_distance: number;
  normalized_trajectory_distance: number;
  divergence_turn: number;
  alignment_steps: AlignmentStep[];
  candidate_anti_patterns: AntiPattern[];
  baseline_anti_patterns: AntiPattern[];
  regressions: string[];
  verdict: "PASS" | "REGRESSION" | "INCONCLUSIVE";
}

export interface TaskListItem {
  task_id: string;
  verdict: "PASS" | "REGRESSION" | "INCONCLUSIVE";
  candidate_model: string;
  baseline_model: string;
  candidate_verified: boolean;
  baseline_verified: boolean;
  tool_bloat_factor: number;
  surgical_score_delta: number;
  speedup_factor: number;
  trajectory_edit_distance: number;
  regression_count: number;
}
