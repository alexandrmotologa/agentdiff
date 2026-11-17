package metrics

import (
	"fmt"
	"math"

	"github.com/alexandrmotologa/agentdiff/internal/config"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// CompareTrajectories evaluates candidate vs baseline trajectories against configured gates.
func CompareTrajectories(candidate, baseline *trajectory.Trajectory, targetFiles []string, gates config.RegressionGates) *trajectory.ComparisonResult {
	taskID := candidate.TaskID
	if taskID == "" {
		taskID = baseline.TaskID
	}

	candTools := trajectory.ExtractToolTokens(candidate)
	baseTools := trajectory.ExtractToolTokens(baseline)

	candToolCount := len(candTools)
	baseToolCount := len(baseTools)

	var bloatFactor float64 = 1.0
	if baseToolCount > 0 {
		bloatFactor = float64(candToolCount) / float64(baseToolCount)
	} else if candToolCount > 0 {
		bloatFactor = float64(candToolCount)
	}
	bloatFactor = math.Round(bloatFactor*100) / 100

	candEff := CalculateToolEfficiency(candidate)
	baseEff := CalculateToolEfficiency(baseline)
	effDelta := math.Round((candEff-baseEff)*1000) / 1000

	durDelta := candidate.DurationMs - baseline.DurationMs
	speedup := 1.0
	if candidate.DurationMs > 0 && baseline.DurationMs > 0 {
		speedup = float64(baseline.DurationMs) / float64(candidate.DurationMs)
		speedup = math.Round(speedup*100) / 100
	}

	surgicalDelta := math.Round((candidate.ASTMetrics.SurgicalScore-baseline.ASTMetrics.SurgicalScore)*1000) / 1000

	editDist, normDist, divTurn, steps := trajectory.AlignTrajectories(candidate, baseline)

	candAntiPatterns := AnalyzeAntiPatterns(candidate, targetFiles)
	baseAntiPatterns := AnalyzeAntiPatterns(baseline, targetFiles)

	// Combine detected anti-patterns with existing trajectory ones
	candAntiPatterns = append(candidate.AntiPatterns, candAntiPatterns...)
	baseAntiPatterns = append(baseline.AntiPatterns, baseAntiPatterns...)

	var regressions []string
	verdict := "PASS"

	// 1. Pass Rate Parity Gate
	if gates.RequirePassRateParity {
		if baseline.Verified && !candidate.Verified {
			regressions = append(regressions, "Candidate failed task verification while baseline succeeded")
			verdict = "REGRESSION"
		}
	}

	// 2. Tool Bloat Gate
	if gates.MaxToolBloatFactor > 0 && bloatFactor > gates.MaxToolBloatFactor {
		pct := int((bloatFactor - 1.0) * 100)
		regressions = append(regressions, fmt.Sprintf("Tool call bloat (+%d%%) exceeds maximum allowed factor (%.2f)", pct, gates.MaxToolBloatFactor))
		verdict = "REGRESSION"
	}

	// 3. Minimum Surgical Score Gate
	if gates.MinSurgicalScore > 0 && candidate.ASTMetrics.SurgicalScore < gates.MinSurgicalScore {
		regressions = append(regressions, fmt.Sprintf("Candidate AST surgical precision (%.2f) below minimum threshold (%.2f)", candidate.ASTMetrics.SurgicalScore, gates.MinSurgicalScore))
		verdict = "REGRESSION"
	}

	// 4. Surgical Score Drop Gate
	if gates.MaxSurgicalScoreDrop > 0 && surgicalDelta < -gates.MaxSurgicalScoreDrop {
		regressions = append(regressions, fmt.Sprintf("Candidate AST surgical score dropped by %.2f compared to baseline", -surgicalDelta))
		verdict = "REGRESSION"
	}

	// 5. Cosmetic Churn Gate
	if gates.MaxCosmeticChurnRatio > 0 && candidate.ASTMetrics.CosmeticChurnRatio > gates.MaxCosmeticChurnRatio {
		regressions = append(regressions, fmt.Sprintf("Candidate cosmetic churn ratio (%.2f) exceeds threshold (%.2f)", candidate.ASTMetrics.CosmeticChurnRatio, gates.MaxCosmeticChurnRatio))
		verdict = "REGRESSION"
	}

	// 6. Latency Regression Gate
	if gates.MaxLatencyIncreaseRatio > 0 && baseline.DurationMs > 0 {
		latencyRatio := float64(candidate.DurationMs) / float64(baseline.DurationMs)
		if latencyRatio > gates.MaxLatencyIncreaseRatio {
			pct := int((latencyRatio - 1.0) * 100)
			regressions = append(regressions, fmt.Sprintf("Execution duration increased by +%d%% over baseline", pct))
			verdict = "REGRESSION"
		}
	}

	// 7. Doom Loop Gate
	if gates.FailOnDoomLoop {
		for _, ap := range candAntiPatterns {
			if ap.Kind == "DOOM_LOOP" {
				regressions = append(regressions, fmt.Sprintf("Command doom loop detected at turn %d: %s", ap.TurnIndex, ap.Details))
				verdict = "REGRESSION"
				break
			}
		}
	}

	// If candidate failed and baseline also failed, mark INCONCLUSIVE unless regressions triggered
	if !candidate.Verified && !baseline.Verified && len(regressions) == 0 {
		verdict = "INCONCLUSIVE"
	}

	costProj := CalculateCostProjection(candidate, baseline)
	diag := DiagnoseDivergence(candidate, baseline, divTurn, steps)

	return &trajectory.ComparisonResult{
		TaskID:                       taskID,
		CandidateModel:               candidate.Model,
		BaselineModel:                baseline.Model,
		CandidateStatus:              candidate.Status,
		BaselineStatus:               baseline.Status,
		CandidateVerified:            candidate.Verified,
		BaselineVerified:             baseline.Verified,
		CandidateDurationMs:          candidate.DurationMs,
		BaselineDurationMs:           baseline.DurationMs,
		DurationDeltaMs:              durDelta,
		SpeedupFactor:                speedup,
		CandidateToolCount:           candToolCount,
		BaselineToolCount:            baseToolCount,
		ToolBloatFactor:              bloatFactor,
		ToolEfficiencyCandidate:      candEff,
		ToolEfficiencyBaseline:       baseEff,
		ToolEfficiencyDelta:          effDelta,
		CandidateSurgicalScore:       candidate.ASTMetrics.SurgicalScore,
		BaselineSurgicalScore:        baseline.ASTMetrics.SurgicalScore,
		SurgicalScoreDelta:           surgicalDelta,
		TrajectoryEditDistance:       editDist,
		NormalizedTrajectoryDistance: normDist,
		DivergenceTurn:               divTurn,
		AlignmentSteps:               steps,
		CandidateAntiPatterns:        candAntiPatterns,
		BaselineAntiPatterns:         baseAntiPatterns,
		CostProjection:               &costProj,
		Diagnosis:                    &diag,
		Regressions:                  regressions,
		Verdict:                      verdict,
	}
}
