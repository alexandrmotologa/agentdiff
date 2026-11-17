package runner

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/alexandrmotologa/agentdiff/internal/config"
	"github.com/alexandrmotologa/agentdiff/internal/metrics"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// ProgressCallback informs callers of real-time benchmark execution status.
type ProgressCallback func(taskID string, status string, comp *trajectory.ComparisonResult)

// RunBenchmarkSuite executes the complete benchmark across candidate and baseline models.
func RunBenchmarkSuite(cfg *config.BenchmarkConfig, outputDir string, progressCb ProgressCallback) ([]*trajectory.ComparisonResult, error) {
	if len(cfg.Tasks) == 0 {
		return nil, fmt.Errorf("no benchmark tasks provided")
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 2
	}

	results := make([]*trajectory.ComparisonResult, len(cfg.Tasks))
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	var errMu sync.Mutex
	var firstErr error

	for idx, task := range cfg.Tasks {
		wg.Add(1)
		go func(taskIdx int, t config.TaskDefinition) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if progressCb != nil {
				progressCb(t.ID, "STARTING", nil)
			}

			comp, err := RunSingleTaskPair(cfg, t, outputDir)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("task %q failed: %w", t.ID, err)
				}
				errMu.Unlock()
				if progressCb != nil {
					progressCb(t.ID, "ERROR", nil)
				}
				return
			}

			results[taskIdx] = comp
			if progressCb != nil {
				progressCb(t.ID, "COMPLETED", comp)
			}
		}(idx, task)
	}

	wg.Wait()

	if firstErr != nil {
		return results, firstErr
	}

	return results, nil
}

// RunSingleTaskPair executes candidate and baseline models in isolated workspaces on one task.
func RunSingleTaskPair(cfg *config.BenchmarkConfig, task config.TaskDefinition, outputDir string) (*trajectory.ComparisonResult, error) {
	sourceDir := task.WorkingDir
	if sourceDir == "" {
		return nil, fmt.Errorf("task %q has empty working_dir", task.ID)
	}

	// 1. Create Workspace for Candidate
	wsCand, err := NewWorkspace(task.ID, "candidate", sourceDir)
	if err != nil {
		return nil, fmt.Errorf("creating candidate workspace: %w", err)
	}
	defer wsCand.Cleanup()

	// 2. Create Workspace for Baseline
	wsBase, err := NewWorkspace(task.ID, "baseline", sourceDir)
	if err != nil {
		return nil, fmt.Errorf("creating baseline workspace: %w", err)
	}
	defer wsBase.Cleanup()

	// 3. Execute Candidate and Baseline concurrently
	var candTraj, baseTraj *trajectory.Trajectory
	var candErr, baseErr error
	var runWg sync.WaitGroup
	runWg.Add(2)

	go func() {
		defer runWg.Done()
		candTraj, candErr = ExecuteAgent(wsCand, task, cfg.Candidate)
	}()

	go func() {
		defer runWg.Done()
		baseTraj, baseErr = ExecuteAgent(wsBase, task, cfg.Baseline)
	}()

	runWg.Wait()

	if candErr != nil {
		return nil, fmt.Errorf("candidate agent execution failed: %w", candErr)
	}
	if baseErr != nil {
		return nil, fmt.Errorf("baseline agent execution failed: %w", baseErr)
	}

	// 4. Compare Trajectories
	comp := metrics.CompareTrajectories(candTraj, baseTraj, task.TargetFiles, cfg.Gates)

	// 5. Persist traces if outputDir specified
	if outputDir != "" {
		taskDir := filepath.Join(outputDir, task.ID)
		_ = trajectory.SaveTrajectory(filepath.Join(taskDir, "candidate.trajectory.json"), candTraj)
		_ = trajectory.SaveTrajectory(filepath.Join(taskDir, "baseline.trajectory.json"), baseTraj)
		_ = trajectory.SaveComparison(filepath.Join(taskDir, "comparison.json"), comp)
		_ = trajectory.AppendJSONL(filepath.Join(outputDir, "agent_eval.jsonl"), comp)
	}

	return comp, nil
}
