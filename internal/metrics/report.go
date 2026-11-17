package metrics

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// FormatTerminalReport renders an individual comparison scorecard for terminal output.
func FormatTerminalReport(comp *trajectory.ComparisonResult) string {
	var sb strings.Builder

	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf(" agentdiff Comparison: %s\n", comp.TaskID))
	sb.WriteString(fmt.Sprintf(" Candidate: %s  vs  Baseline: %s\n", comp.CandidateModel, comp.BaselineModel))
	sb.WriteString("================================================================================\n")

	// Status & Verification
	sb.WriteString(fmt.Sprintf(" Verification : Candidate: %s (Verified: %t) | Baseline: %s (Verified: %t)\n",
		comp.CandidateStatus, comp.CandidateVerified, comp.BaselineStatus, comp.BaselineVerified))

	// Duration
	durSign := "+"
	if comp.DurationDeltaMs < 0 {
		durSign = ""
	}
	sb.WriteString(fmt.Sprintf(" Duration     : Cand: %.1fs | Base: %.1fs | Delta: %s%.1fs (Speedup: %.2fx)\n",
		float64(comp.CandidateDurationMs)/1000.0, float64(comp.BaselineDurationMs)/1000.0,
		durSign, float64(comp.DurationDeltaMs)/1000.0, comp.SpeedupFactor))

	// Tool Calls
	bloatPct := int((comp.ToolBloatFactor - 1.0) * 100)
	bloatSign := "+"
	if bloatPct < 0 {
		bloatSign = ""
	}
	sb.WriteString(fmt.Sprintf(" Tool Calls   : Cand: %d | Base: %d | Bloat: %s%d%% (Factor: %.2f)\n",
		comp.CandidateToolCount, comp.BaselineToolCount, bloatSign, bloatPct, comp.ToolBloatFactor))

	// Efficiency
	sb.WriteString(fmt.Sprintf(" Efficiency   : Cand: %.1f%% | Base: %.1f%% | Delta: %+.1f%%\n",
		comp.ToolEfficiencyCandidate*100.0, comp.ToolEfficiencyBaseline*100.0, comp.ToolEfficiencyDelta*100.0))

	// AST Surgical Score
	sb.WriteString(fmt.Sprintf(" Surgical AST : Cand: %.3f | Base: %.3f | Delta: %+.3f\n",
		comp.CandidateSurgicalScore, comp.BaselineSurgicalScore, comp.SurgicalScoreDelta))

	// Trajectory Distance & Divergence
	sb.WriteString(fmt.Sprintf(" Alignment    : Edit Dist: %d (Norm: %.2f) | Divergence Turn: %d\n",
		comp.TrajectoryEditDistance, comp.NormalizedTrajectoryDistance, comp.DivergenceTurn))

	// Cost & Scale
	if comp.CostProjection != nil {
		sb.WriteString(fmt.Sprintf(" Cost & Scale : Cand: $%.4f | Base: $%.4f (Delta: %s) | 10k/mo impact: %s\n",
			comp.CostProjection.CandidateCostUSD, comp.CostProjection.BaselineCostUSD,
			formatUSDDelta(comp.CostProjection.CostDeltaUSD), formatUSDShort(comp.CostProjection.MonthlyDiffAt10kTasks)))
	}

	// Diagnosis
	if comp.Diagnosis != nil && comp.Diagnosis.DivergenceTurn > 0 {
		sb.WriteString(fmt.Sprintf(" Diagnosis    : [%s] %s\n", comp.Diagnosis.Category, comp.Diagnosis.Summary))
		if comp.Diagnosis.RecommendedSteering != "" {
			sb.WriteString(fmt.Sprintf(" Steering Tip : %s\n", comp.Diagnosis.RecommendedSteering))
		}
	}

	// Anti-Patterns
	if len(comp.CandidateAntiPatterns) > 0 {
		sb.WriteString(" Anti-Patterns: Candidate:\n")
		for _, ap := range comp.CandidateAntiPatterns {
			sb.WriteString(fmt.Sprintf("   - [%s] Turn %d (%s): %s\n", ap.Severity, ap.TurnIndex, ap.Kind, ap.Details))
		}
	}

	// Regressions
	sb.WriteString("--------------------------------------------------------------------------------\n")
	if len(comp.Regressions) > 0 {
		sb.WriteString(fmt.Sprintf(" VERDICT: REGRESSION (%d violations)\n", len(comp.Regressions)))
		for _, reg := range comp.Regressions {
			sb.WriteString(fmt.Sprintf("   * %s\n", reg))
		}
	} else {
		sb.WriteString(fmt.Sprintf(" VERDICT: %s\n", comp.Verdict))
	}
	sb.WriteString("================================================================================\n")

	return sb.String()
}

// FormatMarkdownSummary generates a GitHub-flavored Markdown report.
func FormatMarkdownSummary(comp *trajectory.ComparisonResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### agentdiff Benchmark: `%s`\n\n", comp.TaskID))
	sb.WriteString(fmt.Sprintf("**Candidate:** `%s` | **Baseline:** `%s`\n\n", comp.CandidateModel, comp.BaselineModel))

	sb.WriteString("| Metric | Baseline | Candidate | Delta | Status |\n")
	sb.WriteString("|---|---|---|---|---|\n")

	// Verification
	candVer := "PASS"
	if !comp.CandidateVerified {
		candVer = "FAIL"
	}
	baseVer := "PASS"
	if !comp.BaselineVerified {
		baseVer = "FAIL"
	}
	verStatus := "PASS"
	if baseVer == "PASS" && candVer == "FAIL" {
		verStatus = "REGRESSION"
	}
	sb.WriteString(fmt.Sprintf("| Verification | %s | %s | - | %s |\n", baseVer, candVer, verStatus))

	// Duration
	durStatus := "OK"
	if comp.SpeedupFactor < 0.8 {
		durStatus = "SLOWER"
	} else if comp.SpeedupFactor > 1.2 {
		durStatus = "FASTER"
	}
	sb.WriteString(fmt.Sprintf("| Duration | %.1fs | %.1fs | %+.1fs (%.2fx) | %s |\n",
		float64(comp.BaselineDurationMs)/1000.0, float64(comp.CandidateDurationMs)/1000.0,
		float64(comp.DurationDeltaMs)/1000.0, comp.SpeedupFactor, durStatus))

	// Tool Calls & Bloat
	bloatStatus := "OK"
	if comp.ToolBloatFactor > 1.25 {
		bloatStatus = "BLOAT"
	}
	sb.WriteString(fmt.Sprintf("| Tool Invocations | %d | %d | %+.1f%% | %s |\n",
		comp.BaselineToolCount, comp.CandidateToolCount, (comp.ToolBloatFactor-1.0)*100, bloatStatus))

	// Surgical AST
	surgStatus := "OK"
	if comp.SurgicalScoreDelta < -0.15 {
		surgStatus = "REGRESSION"
	}
	sb.WriteString(fmt.Sprintf("| AST Surgical Precision | %.3f | %.3f | %+.3f | %s |\n",
		comp.BaselineSurgicalScore, comp.CandidateSurgicalScore, comp.SurgicalScoreDelta, surgStatus))

	// Trajectory Alignment
	sb.WriteString(fmt.Sprintf("| Trajectory Edit Distance | - | %d | Norm: %.2f | Div Turn: %d |\n",
		comp.TrajectoryEditDistance, comp.NormalizedTrajectoryDistance, comp.DivergenceTurn))

	if comp.CostProjection != nil {
		sb.WriteString(fmt.Sprintf("| Token Cost (USD) | $%.4f | $%.4f | %s (%.2fx) | 10k: %s |\n\n",
			comp.CostProjection.BaselineCostUSD, comp.CostProjection.CandidateCostUSD,
			formatUSDDelta(comp.CostProjection.CostDeltaUSD), comp.CostProjection.CostRatio,
			formatUSDShort(comp.CostProjection.MonthlyDiffAt10kTasks)))
	} else {
		sb.WriteString("\n")
	}

	if comp.Diagnosis != nil && comp.Diagnosis.DivergenceTurn > 0 {
		sb.WriteString(fmt.Sprintf("> [!WARNING] **Divergence Root Cause (Turn %d - %s)**: %s\n",
			comp.Diagnosis.DivergenceTurn, comp.Diagnosis.Category, comp.Diagnosis.Summary))
		if comp.Diagnosis.RecommendedSteering != "" {
			sb.WriteString(fmt.Sprintf("> **Recommended Prompt Steering**: %s\n\n", comp.Diagnosis.RecommendedSteering))
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString(fmt.Sprintf("**Verdict:** `%s`\n", comp.Verdict))
	if len(comp.Regressions) > 0 {
		sb.WriteString("\n**Detected Regressions:**\n")
		for _, r := range comp.Regressions {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	return sb.String()
}

// FormatSuiteMarkdown formats a multi-task benchmark suite summary table.
func FormatSuiteMarkdown(suiteName string, results []*trajectory.ComparisonResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## agentdiff Suite Evaluation: %s\n\n", suiteName))

	totalTasks := len(results)
	regressions := 0
	passed := 0

	for _, r := range results {
		if r.Verdict == "REGRESSION" {
			regressions++
		} else if r.Verdict == "PASS" {
			passed++
		}
	}

	sb.WriteString(fmt.Sprintf("- **Total Tasks:** %d\n", totalTasks))
	sb.WriteString(fmt.Sprintf("- **Passed:** %d\n", passed))
	sb.WriteString(fmt.Sprintf("- **Regressions:** %d\n\n", regressions))

	sb.WriteString("| Task ID | Verdict | Candidate Status | Tool Bloat | AST Precision $\\Delta$ | Speedup |\n")
	sb.WriteString("|---|---|---|---|---|---|\n")

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("| `%s` | **%s** | %s (Ver: %t) | %.2fx | %+.3f | %.2fx |\n",
			r.TaskID, r.Verdict, r.CandidateStatus, r.CandidateVerified, r.ToolBloatFactor, r.SurgicalScoreDelta, r.SpeedupFactor))
	}

	sb.WriteString("\n---\n")
	for _, r := range results {
		if len(r.Regressions) > 0 {
			sb.WriteString(fmt.Sprintf("### Task `%s` Regressions\n", r.TaskID))
			for _, reg := range r.Regressions {
				sb.WriteString(fmt.Sprintf("- %s\n", reg))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// FormatSuiteTerminal formats a multi-task benchmark suite summary table for the terminal.
func FormatSuiteTerminal(suiteName string, results []*trajectory.ComparisonResult) string {
	var sb strings.Builder

	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf(" agentdiff Benchmark Suite: %s\n", suiteName))
	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf("%-24s | %-10s | %-10s | %-10s | %-12s\n", "TASK ID", "VERDICT", "BLOAT", "SURGICAL", "SPEEDUP"))
	sb.WriteString("--------------------------------------------------------------------------------\n")

	regCount := 0
	for _, r := range results {
		if r.Verdict == "REGRESSION" {
			regCount++
		}
		sb.WriteString(fmt.Sprintf("%-24s | %-10s | %-10.2fx | %-+10.3f | %-12.2fx\n",
			truncate(r.TaskID, 24), r.Verdict, r.ToolBloatFactor, r.SurgicalScoreDelta, r.SpeedupFactor))
	}

	sb.WriteString("================================================================================\n")
	sb.WriteString(fmt.Sprintf(" Total Tasks: %d | Regressions: %d\n", len(results), regCount))
	sb.WriteString("================================================================================\n")

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatUSDDelta(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+$%.4f", v)
	}
	return fmt.Sprintf("-$%.4f", -v)
}

func formatUSDShort(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+$%.2f", v)
	}
	return fmt.Sprintf("-$%.2f", -v)
}
