package trajectory

import (
	"fmt"
	"math"
	"strings"
)

// ToolToken represents a tool call signature used for sequence alignment.
type ToolToken struct {
	Signature string
	TurnIndex int
	ToolCall  ToolCall
}

// ExtractToolTokens extracts all tool calls across turns in chronological order.
func ExtractToolTokens(traj *Trajectory) []ToolToken {
	var tokens []ToolToken
	for _, turn := range traj.Turns {
		for _, call := range turn.ToolCalls {
			sig := ToolSignature(call)
			tokens = append(tokens, ToolToken{
				Signature: sig,
				TurnIndex: turn.TurnIndex,
				ToolCall:  call,
			})
		}
	}
	return tokens
}

// ToolSignature formats a tool call into a canonical token string.
func ToolSignature(c ToolCall) string {
	target := c.NormalizedTarget
	if target == "" {
		// Attempt to extract target from common argument keys
		if p, ok := c.Arguments["path"].(string); ok && p != "" {
			target = p
		} else if f, ok := c.Arguments["file"].(string); ok && f != "" {
			target = f
		} else if cmd, ok := c.Arguments["command"].(string); ok && cmd != "" {
			// Normalize command to the binary name or first 2 words
			parts := strings.Fields(cmd)
			if len(parts) > 0 {
				target = parts[0]
				if len(parts) > 1 && !strings.HasPrefix(parts[1], "-") {
					target = parts[0] + " " + parts[1]
				}
			}
		}
	}
	if target != "" {
		return fmt.Sprintf("%s:%s", c.ToolName, target)
	}
	return c.ToolName
}

// AlignTrajectories performs Needleman-Wunsch global alignment and Levenshtein distance.
func AlignTrajectories(candidate, baseline *Trajectory) (int, float64, int, []AlignmentStep) {
	candTokens := ExtractToolTokens(candidate)
	baseTokens := ExtractToolTokens(baseline)

	n := len(candTokens)
	m := len(baseTokens)

	// Compute Levenshtein distance
	editDistance := ComputeLevenshtein(candTokens, baseTokens)
	maxLen := n
	if m > maxLen {
		maxLen = m
	}
	var normDistance float64
	if maxLen > 0 {
		normDistance = float64(editDistance) / float64(maxLen)
	}

	// If either trajectory has no tools, handle trivial alignment
	if n == 0 && m == 0 {
		return 0, 0, 0, nil
	}
	if n == 0 {
		var steps []AlignmentStep
		for i, bt := range baseTokens {
			steps = append(steps, AlignmentStep{
				Index:        i + 1,
				Op:           OpBaselineGap,
				BaselineTool: &bt.ToolCall,
				BaselineTurn: bt.TurnIndex,
			})
		}
		return editDistance, normDistance, 1, steps
	}
	if m == 0 {
		var steps []AlignmentStep
		for i, ct := range candTokens {
			steps = append(steps, AlignmentStep{
				Index:         i + 1,
				Op:            OpCandidateGap,
				CandidateTool: &ct.ToolCall,
				CandidateTurn: ct.TurnIndex,
			})
		}
		return editDistance, normDistance, 1, steps
	}

	// Needleman-Wunsch Scoring
	const (
		matchScore    = 2
		mismatchScore = -1
		gapPenalty    = -1
	)

	// dp matrix
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := 0; i <= n; i++ {
		dp[i][0] = i * gapPenalty
	}
	for j := 0; j <= m; j++ {
		dp[0][j] = j * gapPenalty
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			score := mismatchScore
			if candTokens[i-1].Signature == baseTokens[j-1].Signature {
				score = matchScore
			}
			diag := dp[i-1][j-1] + score
			up := dp[i-1][j] + gapPenalty
			left := dp[i][j-1] + gapPenalty

			maxVal := diag
			if up > maxVal {
				maxVal = up
			}
			if left > maxVal {
				maxVal = left
			}
			dp[i][j] = maxVal
		}
	}

	// Traceback
	var reversedSteps []AlignmentStep
	i, j := n, m
	for i > 0 || j > 0 {
		if i > 0 && j > 0 {
			score := mismatchScore
			op := OpMismatch
			if candTokens[i-1].Signature == baseTokens[j-1].Signature {
				score = matchScore
				op = OpMatch
			}
			if dp[i][j] == dp[i-1][j-1]+score {
				cTool := candTokens[i-1].ToolCall
				bTool := baseTokens[j-1].ToolCall
				reversedSteps = append(reversedSteps, AlignmentStep{
					Op:            op,
					CandidateTool: &cTool,
					BaselineTool:  &bTool,
					CandidateTurn: candTokens[i-1].TurnIndex,
					BaselineTurn:  baseTokens[j-1].TurnIndex,
				})
				i--
				j--
				continue
			}
		}
		if i > 0 && dp[i][j] == dp[i-1][j]+gapPenalty {
			cTool := candTokens[i-1].ToolCall
			reversedSteps = append(reversedSteps, AlignmentStep{
				Op:            OpCandidateGap,
				CandidateTool: &cTool,
				CandidateTurn: candTokens[i-1].TurnIndex,
			})
			i--
		} else {
			bTool := baseTokens[j-1].ToolCall
			reversedSteps = append(reversedSteps, AlignmentStep{
				Op:           OpBaselineGap,
				BaselineTool: &bTool,
				BaselineTurn: baseTokens[j-1].TurnIndex,
			})
			j--
		}
	}

	// Reverse steps to forward order and index them
	steps := make([]AlignmentStep, len(reversedSteps))
	divergenceTurn := 0
	for k := 0; k < len(reversedSteps); k++ {
		revIdx := len(reversedSteps) - 1 - k
		step := reversedSteps[revIdx]
		step.Index = k + 1
		steps[k] = step

		// First non-match is the divergence turn
		if divergenceTurn == 0 && step.Op != OpMatch {
			if step.CandidateTurn > 0 {
				divergenceTurn = step.CandidateTurn
			} else if step.BaselineTurn > 0 {
				divergenceTurn = step.BaselineTurn
			} else {
				divergenceTurn = 1
			}
		}
	}

	return editDistance, normDistance, divergenceTurn, steps
}

// ComputeLevenshtein calculates the edit distance between two token sequences.
func ComputeLevenshtein(a, b []ToolToken) int {
	la := len(a)
	lb := len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	v0 := make([]int, lb+1)
	v1 := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		v0[j] = j
	}

	for i := 0; i < la; i++ {
		v1[0] = i + 1
		for j := 0; j < lb; j++ {
			cost := 1
			if a[i].Signature == b[j].Signature {
				cost = 0
			}
			minVal := int(math.Min(float64(v1[j]+1), float64(v0[j+1]+1)))
			v1[j+1] = int(math.Min(float64(minVal), float64(v0[j]+cost)))
		}
		copy(v0, v1)
	}

	return v0[lb]
}
