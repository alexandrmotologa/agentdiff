package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alexandrmotologa/agentdiff/internal/config"
	"github.com/alexandrmotologa/agentdiff/internal/runner"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// Server handles REST and static assets for Trajectory Studio.
type Server struct {
	TracesDir  string
	ConfigPath string
	Broker     *SSEBroker
	mu         sync.RWMutex
	benchmarks map[string]*trajectory.ComparisonResult
}

// NewServer initializes the HTTP API server.
func NewServer(tracesDir, configPath string) *Server {
	s := &Server{
		TracesDir:  tracesDir,
		ConfigPath: configPath,
		Broker:     NewSSEBroker(),
		benchmarks: make(map[string]*trajectory.ComparisonResult),
	}
	s.loadHistoricalTraces()
	return s
}

// RegisterRoutes sets up HTTP routes on a mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/comparison", s.handleComparison)
	mux.HandleFunc("/api/trajectory", s.handleTrajectory)
	mux.HandleFunc("/api/bench/run", s.handleRunBenchmark)
	mux.HandleFunc("/api/bench/stream", s.Broker.HandleSSE)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"service": "agentdiff-studio",
	})
}

// TaskListItem represents an overview of an evaluated task.
type TaskListItem struct {
	TaskID                 string  `json:"task_id"`
	Verdict                string  `json:"verdict"`
	CandidateModel         string  `json:"candidate_model"`
	BaselineModel          string  `json:"baseline_model"`
	CandidateVerified      bool    `json:"candidate_verified"`
	BaselineVerified       bool    `json:"baseline_verified"`
	ToolBloatFactor        float64 `json:"tool_bloat_factor"`
	SurgicalScoreDelta     float64 `json:"surgical_score_delta"`
	SpeedupFactor          float64 `json:"speedup_factor"`
	TrajectoryEditDistance int     `json:"trajectory_edit_distance"`
	RegressionCount        int     `json:"regression_count"`
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []TaskListItem
	for _, comp := range s.benchmarks {
		list = append(list, TaskListItem{
			TaskID:                 comp.TaskID,
			Verdict:                comp.Verdict,
			CandidateModel:         comp.CandidateModel,
			BaselineModel:          comp.BaselineModel,
			CandidateVerified:      comp.CandidateVerified,
			BaselineVerified:       comp.BaselineVerified,
			ToolBloatFactor:        comp.ToolBloatFactor,
			SurgicalScoreDelta:     comp.SurgicalScoreDelta,
			SpeedupFactor:          comp.SpeedupFactor,
			TrajectoryEditDistance: comp.TrajectoryEditDistance,
			RegressionCount:        len(comp.Regressions),
		})
	}

	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleComparison(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "missing task_id parameter", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	comp, exists := s.benchmarks[taskID]
	s.mu.RUnlock()

	if !exists {
		// Attempt reading from disk
		path := filepath.Join(s.TracesDir, taskID, "comparison.json")
		loaded, err := trajectory.LoadComparison(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("comparison for task %q not found", taskID), http.StatusNotFound)
			return
		}
		comp = loaded
		s.mu.Lock()
		s.benchmarks[taskID] = loaded
		s.mu.Unlock()
	}

	writeJSON(w, http.StatusOK, comp)
}

func (s *Server) handleTrajectory(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	agent := strings.ToLower(r.URL.Query().Get("agent")) // "candidate" or "baseline"

	if taskID == "" || (agent != "candidate" && agent != "baseline") {
		http.Error(w, "invalid query parameters: require task_id and agent=(candidate|baseline)", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("%s.trajectory.json", agent)
	path := filepath.Join(s.TracesDir, taskID, filename)

	traj, err := trajectory.LoadTrajectory(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("trajectory not found at %s: %v", path, err), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, traj)
}

func (s *Server) handleRunBenchmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	configPath := s.ConfigPath
	if configPath == "" {
		configPath = "examples/agentdiff.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("loading config %q: %v", configPath, err), http.StatusBadRequest)
		return
	}

	// Trigger run asynchronously
	go func() {
		s.Broker.Broadcast("bench_started", map[string]interface{}{
			"suite_name":  cfg.Name,
			"total_tasks": len(cfg.Tasks),
		})

		progressCb := func(taskID string, status string, comp *trajectory.ComparisonResult) {
			if comp != nil {
				s.mu.Lock()
				s.benchmarks[taskID] = comp
				s.mu.Unlock()
			}
			s.Broker.Broadcast("task_progress", map[string]interface{}{
				"task_id":    taskID,
				"status":     status,
				"comparison": comp,
			})
		}

		results, err := runner.RunBenchmarkSuite(cfg, s.TracesDir, progressCb)
		if err != nil {
			s.Broker.Broadcast("bench_error", map[string]string{"error": err.Error()})
			return
		}

		s.Broker.Broadcast("bench_completed", map[string]interface{}{
			"total_tasks": len(results),
			"results":     results,
		})
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "Benchmark execution launched",
		"config":  configPath,
	})
}

func (s *Server) loadHistoricalTraces() {
	if s.TracesDir == "" {
		return
	}

	entries, err := os.ReadDir(s.TracesDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		compFile := filepath.Join(s.TracesDir, entry.Name(), "comparison.json")
		comp, err := trajectory.LoadComparison(compFile)
		if err == nil && comp != nil {
			s.benchmarks[comp.TaskID] = comp
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
