package metrics

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// LeaderboardEntry aggregates benchmark performance for a model.
type LeaderboardEntry struct {
	Rank                  int     `json:"rank"`
	ModelName             string  `json:"model_name"`
	TotalTasks            int     `json:"total_tasks"`
	TasksSolved           int     `json:"tasks_solved"`
	PassRatePct           float64 `json:"pass_rate_pct"`
	MeanSurgicalScore     float64 `json:"mean_surgical_score"`
	MeanEfficiencyRatio   float64 `json:"mean_efficiency_ratio"`
	MeanDurationMs        int64   `json:"mean_duration_ms"`
	TotalCostUSD          float64 `json:"total_cost_usd"`
	MeanCostUSD           float64 `json:"mean_cost_usd"`
	TotalRegressionsCount int     `json:"total_regressions_count"`
}

// GenerateLeaderboard builds a model leaderboard from comparison results.
func GenerateLeaderboard(results []*trajectory.ComparisonResult) []LeaderboardEntry {
	type modelAccumulator struct {
		tasksCount       int
		tasksSolved      int
		totalSurgical    float64
		totalEfficiency  float64
		totalDurationMs  int64
		totalCostUSD     float64
		totalRegressions int
	}

	modelStats := make(map[string]*modelAccumulator)

	for _, r := range results {
		if r == nil {
			continue
		}

		// Candidate stats
		candModel := r.CandidateModel
		if candModel != "" {
			acc, exists := modelStats[candModel]
			if !exists {
				acc = &modelAccumulator{}
				modelStats[candModel] = acc
			}
			acc.tasksCount++
			if r.CandidateVerified {
				acc.tasksSolved++
			}
			acc.totalSurgical += r.CandidateSurgicalScore
			acc.totalEfficiency += r.ToolEfficiencyCandidate
			acc.totalDurationMs += r.CandidateDurationMs
			acc.totalRegressions += len(r.Regressions)
			if r.CostProjection != nil {
				acc.totalCostUSD += r.CostProjection.CandidateCostUSD
			}
		}

		// Baseline stats
		baseModel := r.BaselineModel
		if baseModel != "" && baseModel != candModel {
			acc, exists := modelStats[baseModel]
			if !exists {
				acc = &modelAccumulator{}
				modelStats[baseModel] = acc
			}
			acc.tasksCount++
			if r.BaselineVerified {
				acc.tasksSolved++
			}
			acc.totalSurgical += r.BaselineSurgicalScore
			acc.totalEfficiency += r.ToolEfficiencyBaseline
			acc.totalDurationMs += r.BaselineDurationMs
			if r.CostProjection != nil {
				acc.totalCostUSD += r.CostProjection.BaselineCostUSD
			}
		}
	}

	var leaderboard []LeaderboardEntry
	for model, acc := range modelStats {
		if acc.tasksCount == 0 {
			continue
		}
		passRate := (float64(acc.tasksSolved) / float64(acc.tasksCount)) * 100.0
		meanSurgical := acc.totalSurgical / float64(acc.tasksCount)
		meanEff := acc.totalEfficiency / float64(acc.tasksCount)
		meanDur := acc.totalDurationMs / int64(acc.tasksCount)
		meanCost := acc.totalCostUSD / float64(acc.tasksCount)

		leaderboard = append(leaderboard, LeaderboardEntry{
			ModelName:             model,
			TotalTasks:            acc.tasksCount,
			TasksSolved:           acc.tasksSolved,
			PassRatePct:           math.Round(passRate*10) / 10,
			MeanSurgicalScore:     math.Round(meanSurgical*1000) / 1000,
			MeanEfficiencyRatio:   math.Round(meanEff*1000) / 1000,
			MeanDurationMs:        meanDur,
			TotalCostUSD:          math.Round(acc.totalCostUSD*10000) / 10000,
			MeanCostUSD:           math.Round(meanCost*10000) / 10000,
			TotalRegressionsCount: acc.totalRegressions,
		})
	}

	// Sort leaderboard: PassRate desc, then MeanSurgicalScore desc, then TotalRegressions asc
	sort.Slice(leaderboard, func(i, j int) bool {
		if leaderboard[i].PassRatePct != leaderboard[j].PassRatePct {
			return leaderboard[i].PassRatePct > leaderboard[j].PassRatePct
		}
		if leaderboard[i].MeanSurgicalScore != leaderboard[j].MeanSurgicalScore {
			return leaderboard[i].MeanSurgicalScore > leaderboard[j].MeanSurgicalScore
		}
		return leaderboard[i].TotalRegressionsCount < leaderboard[j].TotalRegressionsCount
	})

	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	return leaderboard
}

// FormatLeaderboardTerminal renders the leaderboard as an ASCII table.
func FormatLeaderboardTerminal(leaderboard []LeaderboardEntry) string {
	var sb strings.Builder

	sb.WriteString("===================================================================================================\n")
	sb.WriteString(" agentdiff Multi-Model Evaluation Leaderboard\n")
	sb.WriteString("===================================================================================================\n")
	sb.WriteString(fmt.Sprintf("%-4s | %-24s | %-12s | %-14s | %-12s | %-10s | %-10s\n",
		"RANK", "MODEL", "PASS RATE", "AST SURGICAL", "EFFICIENCY", "AVG DUR", "AVG COST"))
	sb.WriteString("---------------------------------------------------------------------------------------------------\n")

	for _, e := range leaderboard {
		sb.WriteString(fmt.Sprintf("%-4d | %-24s | %5.1f%% (%d/%d) | %-14.3f | %-12.1f%% | %6.1fs   | $%.4f\n",
			e.Rank, truncate(e.ModelName, 24), e.PassRatePct, e.TasksSolved, e.TotalTasks,
			e.MeanSurgicalScore, e.MeanEfficiencyRatio*100.0, float64(e.MeanDurationMs)/1000.0, e.MeanCostUSD))
	}

	sb.WriteString("===================================================================================================\n")
	return sb.String()
}
