package trajectory

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// Recorder captures an agent's execution timeline in a thread-safe manner.
type Recorder struct {
	mu         sync.Mutex
	trajectory *Trajectory
	startTime  time.Time
	currentTurn *Turn
}

// NewRecorder initializes a new trajectory recorder.
func NewRecorder(taskID, agentID, model, systemPrompt string) *Recorder {
	now := time.Now()
	var promptHash string
	if systemPrompt != "" {
		hash := sha256.Sum256([]byte(systemPrompt))
		promptHash = fmt.Sprintf("sha256:%x", hash)
	}

	return &Recorder{
		trajectory: &Trajectory{
			SchemaVersion:    "1.0.0",
			TaskID:           taskID,
			AgentID:          agentID,
			Model:            model,
			SystemPromptHash: promptHash,
			StartedAt:        now,
			Status:           "RUNNING",
			Turns:            make([]Turn, 0),
			AntiPatterns:     make([]AntiPattern, 0),
		},
		startTime: now,
	}
}

// StartTurn opens a new reasoning and action turn.
func (r *Recorder) StartTurn(thought string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	turnIdx := len(r.trajectory.Turns) + 1
	r.currentTurn = &Turn{
		TurnIndex:          turnIdx,
		Timestamp:          time.Now(),
		Thought:            thought,
		ToolCalls:          make([]ToolCall, 0),
		TerminalExecutions: make([]TerminalExecution, 0),
		FileMutations:      make([]FileMutation, 0),
	}
	return turnIdx
}

// EndTurn finishes the active turn, recording its duration.
func (r *Recorder) EndTurn() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentTurn == nil {
		return
	}
	r.currentTurn.DurationMs = time.Since(r.currentTurn.Timestamp).Milliseconds()
	r.trajectory.Turns = append(r.trajectory.Turns, *r.currentTurn)
	r.currentTurn = nil
}

// RecordToolCall appends a tool invocation to the active turn.
func (r *Recorder) RecordToolCall(call ToolCall) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentTurn == nil {
		r.currentTurn = &Turn{
			TurnIndex: len(r.trajectory.Turns) + 1,
			Timestamp: time.Now(),
		}
	}
	if call.CallID == "" {
		call.CallID = fmt.Sprintf("call_%d_%d", r.currentTurn.TurnIndex, len(r.currentTurn.ToolCalls)+1)
	}
	r.currentTurn.ToolCalls = append(r.currentTurn.ToolCalls, call)
}

// RecordTerminalExecution appends a command output to the active turn.
func (r *Recorder) RecordTerminalExecution(exec TerminalExecution) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentTurn == nil {
		r.currentTurn = &Turn{
			TurnIndex: len(r.trajectory.Turns) + 1,
			Timestamp: time.Now(),
		}
	}
	r.currentTurn.TerminalExecutions = append(r.currentTurn.TerminalExecutions, exec)
}

// RecordFileMutation appends a file mutation to the active turn.
func (r *Recorder) RecordFileMutation(mutation FileMutation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentTurn != nil {
		r.currentTurn.FileMutations = append(r.currentTurn.FileMutations, mutation)
		return
	}
	if len(r.trajectory.Turns) > 0 {
		lastIdx := len(r.trajectory.Turns) - 1
		r.trajectory.Turns[lastIdx].FileMutations = append(r.trajectory.Turns[lastIdx].FileMutations, mutation)
		return
	}
	r.currentTurn = &Turn{
		TurnIndex:     1,
		Timestamp:     time.Now(),
		FileMutations: []FileMutation{mutation},
	}
}

// AddAntiPattern flags an anti-pattern on the trajectory.
func (r *Recorder) AddAntiPattern(ap AntiPattern) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trajectory.AntiPatterns = append(r.trajectory.AntiPatterns, ap)
}

// SetTokenUsage sets the cumulative token count and cost.
func (r *Recorder) SetTokenUsage(usage TokenUsage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trajectory.TokenUsage = usage
}

// SetFinalPatch sets the overall git diff.
func (r *Recorder) SetFinalPatch(patch string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trajectory.FinalPatch = patch
}

// SetASTMetrics updates structural AST metrics.
func (r *Recorder) SetASTMetrics(metrics ASTMetrics) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trajectory.ASTMetrics = metrics
}

// Complete finalizes the trajectory record with status and verification output.
func (r *Recorder) Complete(status string, verified bool, verificationOutput string) *Trajectory {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentTurn != nil {
		r.currentTurn.DurationMs = time.Since(r.currentTurn.Timestamp).Milliseconds()
		r.trajectory.Turns = append(r.trajectory.Turns, *r.currentTurn)
		r.currentTurn = nil
	}

	r.trajectory.CompletedAt = time.Now()
	r.trajectory.DurationMs = time.Since(r.startTime).Milliseconds()
	r.trajectory.Status = status
	r.trajectory.Verified = verified
	r.trajectory.VerificationOutput = verificationOutput

	return r.trajectory
}

// Trajectory returns a snapshot copy of the current trajectory.
func (r *Recorder) Trajectory() *Trajectory {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.trajectory
}
