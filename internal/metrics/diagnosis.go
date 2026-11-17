package metrics

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// DiagnoseDivergence inspects the exact point where candidate and baseline parted ways.
func DiagnoseDivergence(candidate, baseline *trajectory.Trajectory, divergenceTurn int, alignmentSteps []trajectory.AlignmentStep) trajectory.DivergenceDiagnosis {
	if divergenceTurn <= 0 || len(alignmentSteps) == 0 {
		return trajectory.DivergenceDiagnosis{
			DivergenceTurn:      0,
			Category:            "IDENTICAL_STRATEGY",
			Summary:             "Both agent versions followed aligned tool and decision sequences.",
			RecommendedSteering: "No prompt adjustments needed. Trajectories are structurally aligned.",
		}
	}

	// Locate the first non-matching alignment step
	var divergedStep *trajectory.AlignmentStep
	for _, step := range alignmentSteps {
		if step.Op != trajectory.OpMatch {
			s := step
			divergedStep = &s
			break
		}
	}

	candTurn := getTurn(candidate, divergenceTurn)
	baseTurn := getTurn(baseline, divergenceTurn)
	_ = baseTurn

	candAct := "No action"
	baseAct := "No action"

	if divergedStep != nil {
		if divergedStep.CandidateTool != nil {
			candAct = trajectory.ToolSignature(*divergedStep.CandidateTool)
		}
		if divergedStep.BaselineTool != nil {
			baseAct = trajectory.ToolSignature(*divergedStep.BaselineTool)
		}
	}

	// Check if candidate failed a terminal command at or just before divergence
	if candTurn != nil {
		for _, exec := range candTurn.TerminalExecutions {
			if exec.ExitCode != 0 {
				return trajectory.DivergenceDiagnosis{
					DivergenceTurn:      divergenceTurn,
					Category:            "COMMAND_FAILURE",
					Summary:             fmt.Sprintf("Candidate terminal command %q failed with exit code %d, derailing execution.", exec.Command, exec.ExitCode),
					CandidateAction:     fmt.Sprintf("Ran failing command: %s", exec.Command),
					BaselineAction:      baseAct,
					RecommendedSteering: "Add system prompt instruction to verify command syntax before execution or check environment requirements.",
				}
			}
		}
	}

	// Check if candidate engaged in redundant reads
	if divergedStep != nil && divergedStep.Op == trajectory.OpCandidateGap && divergedStep.CandidateTool != nil {
		if divergedStep.CandidateTool.ToolName == "read_file" {
			return trajectory.DivergenceDiagnosis{
				DivergenceTurn:      divergenceTurn,
				Category:            "REDUNDANT_READ",
				Summary:             fmt.Sprintf("Candidate performed unneeded file inspection on %q while baseline progressed directly to modifications.", candAct),
				CandidateAction:     candAct,
				BaselineAction:      baseAct,
				RecommendedSteering: "Instruct the candidate agent in the system prompt to avoid re-reading files when their context is already present in memory.",
			}
		} else if divergedStep.CandidateTool.ToolName == "list_dir" {
			return trajectory.DivergenceDiagnosis{
				DivergenceTurn:      divergenceTurn,
				Category:            "TOOL_CHOICE_MISMATCH",
				Summary:             "Candidate wasted turns exploring directory listings instead of targeting specified source files.",
				CandidateAction:     candAct,
				BaselineAction:      baseAct,
				RecommendedSteering: "Instruct the candidate to read target files directly when target files are specified in the task description.",
			}
		}
	}

	// Check if baseline performed an edit while candidate stalled
	if divergedStep != nil && divergedStep.Op == trajectory.OpBaselineGap && divergedStep.BaselineTool != nil {
		if strings.Contains(divergedStep.BaselineTool.ToolName, "edit") || strings.Contains(divergedStep.BaselineTool.ToolName, "write") {
			return trajectory.DivergenceDiagnosis{
				DivergenceTurn:      divergenceTurn,
				Category:            "STRATEGY_DEVIATION",
				Summary:             fmt.Sprintf("Baseline applied code modifications at Turn %d, whereas Candidate delayed edits by %d turns.", divergenceTurn, len(candidate.Turns)-len(baseline.Turns)),
				CandidateAction:     candAct,
				BaselineAction:      baseAct,
				RecommendedSteering: "Encourage direct patch application after initial inspection without redundant verification loops.",
			}
		}
	}

	return trajectory.DivergenceDiagnosis{
		DivergenceTurn:      divergenceTurn,
		Category:            "STRATEGY_DEVIATION",
		Summary:             fmt.Sprintf("Decision path parted at Turn %d: Candidate invoked %q whereas Baseline invoked %q.", divergenceTurn, candAct, baseAct),
		CandidateAction:     candAct,
		BaselineAction:      baseAct,
		RecommendedSteering: "Review tool invocation preferences in system prompt for candidate model.",
	}
}

func getTurn(traj *trajectory.Trajectory, idx int) *trajectory.Turn {
	if traj == nil {
		return nil
	}
	for i := range traj.Turns {
		if traj.Turns[i].TurnIndex == idx {
			return &traj.Turns[i]
		}
	}
	return nil
}
