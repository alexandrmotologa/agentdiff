package ast

import (
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// DiffResult holds line diff indices and parsed nodes.
type DiffResult struct {
	OriginalLines []string
	ModifiedLines []string
	AddedIndices  []int // 0-based in ModifiedLines
	RemovedIndices []int // 0-based in OriginalLines
}

// ComputeLineDiff performs LCS line-based diff between original and modified text.
func ComputeLineDiff(original, modified string) DiffResult {
	origLines := strings.Split(original, "\n")
	modLines := strings.Split(modified, "\n")

	n := len(origLines)
	m := len(modLines)

	// Standard LCS table
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if origLines[i-1] == modLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack
	added := make([]int, 0)
	removed := make([]int, 0)
	i, j := n, m
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && origLines[i-1] == modLines[j-1] {
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			added = append([]int{j - 1}, added...)
			j--
		} else if i > 0 && (j == 0 || dp[i-1][j] > dp[i][j-1]) {
			removed = append([]int{i - 1}, removed...)
			i--
		}
	}

	return DiffResult{
		OriginalLines:  origLines,
		ModifiedLines:  modLines,
		AddedIndices:   added,
		RemovedIndices: removed,
	}
}

// CompareSourceFiles computes full AST and surgical diff between before and after files.
func CompareSourceFiles(filePath, originalContent, modifiedContent string) (*ASTDiffSummary, error) {
	origParsed, err := ParseSource(filePath, originalContent)
	if err != nil {
		origParsed, _ = ParseGeneric(originalContent)
	}

	modParsed, err := ParseSource(filePath, modifiedContent)
	if err != nil {
		modParsed, _ = ParseGeneric(modifiedContent)
	}

	diff := ComputeLineDiff(originalContent, modifiedContent)
	linesAdded := len(diff.AddedIndices)
	linesRemoved := len(diff.RemovedIndices)
	totalLinesDiff := linesAdded + linesRemoved

	whitespaceLines := 0
	commentLines := 0

	for _, idx := range diff.AddedIndices {
		line := strings.TrimSpace(diff.ModifiedLines[idx])
		if line == "" {
			whitespaceLines++
		} else if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") {
			commentLines++
		}
	}
	for _, idx := range diff.RemovedIndices {
		line := strings.TrimSpace(diff.OriginalLines[idx])
		if line == "" {
			whitespaceLines++
		} else if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") {
			commentLines++
		}
	}

	// Identify affected functional AST nodes in modified content
	affectedMap := make(map[string]trajectory.ASTNodeRef)
	for _, addedIdx := range diff.AddedIndices {
		lineNum := addedIdx + 1
		for _, node := range modParsed.Nodes {
			if node.Kind == KindComment {
				continue
			}
			if lineNum >= node.LineStart && lineNum <= node.LineEnd {
				key := string(node.Kind) + ":" + node.Identifier
				affectedMap[key] = trajectory.ASTNodeRef{
					Kind:       string(node.Kind),
					Identifier: node.Identifier,
					LineStart:  node.LineStart,
					LineEnd:    node.LineEnd,
				}
			}
		}
	}

	// Identify affected functional AST nodes in original content
	for _, remIdx := range diff.RemovedIndices {
		lineNum := remIdx + 1
		for _, node := range origParsed.Nodes {
			if node.Kind == KindComment {
				continue
			}
			if lineNum >= node.LineStart && lineNum <= node.LineEnd {
				key := string(node.Kind) + ":" + node.Identifier
				affectedMap[key] = trajectory.ASTNodeRef{
					Kind:       string(node.Kind),
					Identifier: node.Identifier,
					LineStart:  node.LineStart,
					LineEnd:    node.LineEnd,
				}
			}
		}
	}

	astNodesModified := len(affectedMap)
	affectedNodes := make([]trajectory.ASTNodeRef, 0, astNodesModified)
	for _, ref := range affectedMap {
		affectedNodes = append(affectedNodes, ref)
	}

	surgicalScore, cosmeticChurn := CalculateSurgicalScore(totalLinesDiff, astNodesModified, whitespaceLines, commentLines)

	return &ASTDiffSummary{
		TotalLinesDiff:     totalLinesDiff,
		LinesAdded:         linesAdded,
		LinesRemoved:       linesRemoved,
		ASTNodesModified:   astNodesModified,
		WhitespaceLines:    whitespaceLines,
		CommentLines:       commentLines,
		SurgicalScore:      surgicalScore,
		CosmeticChurnRatio: cosmeticChurn,
		AffectedNodes:      affectedNodes,
	}, nil
}
