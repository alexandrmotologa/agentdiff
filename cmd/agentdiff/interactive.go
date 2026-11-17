package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// runInteractiveReplay provides an interactive step-by-step terminal viewer.
func runInteractiveReplay(traj *trajectory.Trajectory) {
	if len(traj.Turns) == 0 {
		fmt.Println("No turns available in trajectory to replay.")
		return
	}

	reader := bufio.NewReader(os.Stdin)
	currentIdx := 0
	totalTurns := len(traj.Turns)

	for {
		clearScreen()
		turn := traj.Turns[currentIdx]

		fmt.Println("================================================================================")
		fmt.Printf(" agentdiff Interactive Replay: %s\n", traj.TaskID)
		fmt.Printf(" Model: %s | Status: %s | Verified: %t | Total Duration: %dms\n",
			traj.Model, traj.Status, traj.Verified, traj.DurationMs)
		fmt.Println("================================================================================")

		// Progress bar
		pct := int(float64(currentIdx+1) / float64(totalTurns) * 100)
		barLen := 30
		filled := int(float64(currentIdx+1) / float64(totalTurns) * float64(barLen))
		if filled > barLen {
			filled = barLen
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)
		fmt.Printf(" Turn %d of %d [%s] %d%%\n", currentIdx+1, totalTurns, bar, pct)
		fmt.Println("--------------------------------------------------------------------------------")

		// Thought
		if turn.Thought != "" {
			fmt.Printf(" [Reasoning]\n %s\n\n", turn.Thought)
		} else {
			fmt.Printf(" [Reasoning]\n (No explicit thought emitted)\n\n")
		}

		// Tool Invocations
		fmt.Printf(" [Tool Invocations] (%d)\n", len(turn.ToolCalls))
		if len(turn.ToolCalls) == 0 {
			fmt.Println("   (none)")
		} else {
			for i, call := range turn.ToolCalls {
				statusTag := "[OK]"
				if call.Status != "SUCCESS" {
					statusTag = "[FAIL]"
				}
				target := call.NormalizedTarget
				if target == "" {
					target = fmt.Sprintf("%v", call.Arguments)
				}
				fmt.Printf("   %d. %s %s -> %s (%dms)\n", i+1, statusTag, call.ToolName, target, call.DurationMs)
				if call.Output != "" {
					lines := strings.Split(strings.TrimSpace(call.Output), "\n")
					preview := lines[0]
					if len(lines) > 1 {
						preview += fmt.Sprintf(" ... (+%d more lines)", len(lines)-1)
					}
					fmt.Printf("      Output: %s\n", preview)
				}
			}
		}
		fmt.Println()

		// Terminal Executions
		if len(turn.TerminalExecutions) > 0 {
			fmt.Printf(" [Terminal Executions] (%d)\n", len(turn.TerminalExecutions))
			for i, term := range turn.TerminalExecutions {
				exitColor := "0"
				if term.ExitCode != 0 {
					exitColor = fmt.Sprintf("ERR (%d)", term.ExitCode)
				}
				fmt.Printf("   %d. $ %s [Exit: %s, %dms]\n", i+1, term.Command, exitColor, term.DurationMs)
				if term.Stdout != "" {
					outLines := strings.Split(strings.TrimSpace(term.Stdout), "\n")
					fmt.Printf("      Stdout: %s\n", outLines[0])
				}
				if term.Stderr != "" {
					errLines := strings.Split(strings.TrimSpace(term.Stderr), "\n")
					fmt.Printf("      Stderr: %s\n", errLines[0])
				}
			}
			fmt.Println()
		}

		// File Mutations
		if len(turn.FileMutations) > 0 {
			fmt.Printf(" [File Mutations] (%d)\n", len(turn.FileMutations))
			for _, mut := range turn.FileMutations {
				fmt.Printf("   * %s (%s, +%d, -%d)\n", mut.Path, mut.ChangeType, mut.LinesAdded, mut.LinesRemoved)
			}
			fmt.Println()
		}

		fmt.Println("================================================================================")
		fmt.Println(" Controls: [Enter/n] Next | [p] Prev | [1] First | [0] Last | [q] Quit")
		fmt.Print(" > ")

		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		cmd := strings.ToLower(strings.TrimSpace(input))
		switch cmd {
		case "q", "exit":
			return
		case "p", "b":
			if currentIdx > 0 {
				currentIdx--
			}
		case "1":
			currentIdx = 0
		case "0", "end":
			currentIdx = totalTurns - 1
		default:
			// Default Enter or 'n' goes next
			if currentIdx < totalTurns-1 {
				currentIdx++
			} else {
				fmt.Println("Reached final turn. Press 'q' to quit or 'p' to go back.")
			}
		}
	}
}

func clearScreen() {
	// ANSI clear screen
	fmt.Print("\033[H\033[2J")
}
