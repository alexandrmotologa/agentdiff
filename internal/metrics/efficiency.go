package metrics

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// AnalyzeAntiPatterns inspects a trajectory to discover operational regressions.
func AnalyzeAntiPatterns(traj *trajectory.Trajectory, targetFiles []string) []trajectory.AntiPattern {
	var antiPatterns []trajectory.AntiPattern

	// 1. Detect Redundant File Reads
	// Map of file -> last read turn index, cleared when file is edited/written
	lastReadTurn := make(map[string]int)
	for _, turn := range traj.Turns {
		// First check mutations in this turn to reset read tracking
		for _, mut := range turn.FileMutations {
			delete(lastReadTurn, mut.Path)
		}

		for _, call := range turn.ToolCalls {
			target := call.NormalizedTarget
			if target == "" {
				if p, ok := call.Arguments["path"].(string); ok {
					target = p
				}
			}

			if call.ToolName == "read_file" && target != "" {
				if prevTurn, exists := lastReadTurn[target]; exists {
					antiPatterns = append(antiPatterns, trajectory.AntiPattern{
						Kind:      "REDUNDANT_READ",
						Severity:  "WARNING",
						TurnIndex: turn.TurnIndex,
						ToolName:  "read_file",
						Details:   fmt.Sprintf("File %q read again at turn %d without intervening modification (first read at turn %d)", target, turn.TurnIndex, prevTurn),
					})
				} else {
					lastReadTurn[target] = turn.TurnIndex
				}
			} else if call.ToolName == "edit_file" || call.ToolName == "write_file" {
				if target != "" {
					delete(lastReadTurn, target)
				}
			}
		}
	}

	// 2. Detect Command Doom Loops (executing same failing command repeatedly)
	failingCmdCounts := make(map[string]int)
	for _, turn := range traj.Turns {
		// If code was edited, reset failing command memory
		if len(turn.FileMutations) > 0 {
			failingCmdCounts = make(map[string]int)
		}
		for _, exec := range turn.TerminalExecutions {
			trimmedCmd := strings.TrimSpace(exec.Command)
			if exec.ExitCode != 0 {
				failingCmdCounts[trimmedCmd]++
				if failingCmdCounts[trimmedCmd] >= 3 {
					antiPatterns = append(antiPatterns, trajectory.AntiPattern{
						Kind:      "DOOM_LOOP",
						Severity:  "CRITICAL",
						TurnIndex: turn.TurnIndex,
						ToolName:  "exec_bash",
						Details:   fmt.Sprintf("Command %q failed %d times consecutively without intermediate code edits", trimmedCmd, failingCmdCounts[trimmedCmd]),
					})
				}
			} else {
				delete(failingCmdCounts, trimmedCmd)
			}
		}
	}

	// 3. Detect Shotgun Surgery (touching files outside target files)
	if len(targetFiles) > 0 {
		targetMap := make(map[string]bool)
		for _, f := range targetFiles {
			targetMap[strings.ToLower(filepathBase(f))] = true
		}

		for _, turn := range traj.Turns {
			for _, mut := range turn.FileMutations {
				base := strings.ToLower(filepathBase(mut.Path))
				if !targetMap[base] && !strings.HasSuffix(base, "_test.go") && !strings.HasSuffix(base, "test.py") {
					antiPatterns = append(antiPatterns, trajectory.AntiPattern{
						Kind:      "SHOTGUN_SURGERY",
						Severity:  "WARNING",
						TurnIndex: turn.TurnIndex,
						Details:   fmt.Sprintf("Modified unexpected file %q outside specified target files", mut.Path),
					})
				}
			}
		}
	}

	return antiPatterns
}

func filepathBase(path string) string {
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// CalculateToolEfficiency calculates the ratio of essential tool calls to total tool calls.
func CalculateToolEfficiency(traj *trajectory.Trajectory) float64 {
	tokens := trajectory.ExtractToolTokens(traj)
	totalCalls := len(tokens)
	if totalCalls == 0 {
		return 1.0
	}

	// Identify redundant calls and failed repetitive commands
	redundantCount := 0
	readFiles := make(map[string]bool)

	for _, turn := range traj.Turns {
		for _, mut := range turn.FileMutations {
			delete(readFiles, mut.Path)
		}
		for _, call := range turn.ToolCalls {
			if call.ToolName == "read_file" {
				target := call.NormalizedTarget
				if target == "" {
					if p, ok := call.Arguments["path"].(string); ok {
						target = p
					}
				}
				if target != "" {
					if readFiles[target] {
						redundantCount++
					} else {
						readFiles[target] = true
					}
				}
			} else if call.ToolName == "edit_file" || call.ToolName == "write_file" {
				if p, ok := call.Arguments["path"].(string); ok {
					delete(readFiles, p)
				}
			}
		}
	}

	essential := totalCalls - redundantCount
	if essential < 1 {
		essential = 1
	}

	return float64(essential) / float64(totalCalls)
}
