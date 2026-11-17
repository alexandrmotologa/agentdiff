package ast

import "math"

// CalculateSurgicalScore computes the precision score and cosmetic churn ratio.
// Formula: S_surgical = N_ast / (L_diff + 1)
// Cosmetic Churn = (whitespaceLines + commentLines) / max(1, totalLinesDiff)
func CalculateSurgicalScore(totalLinesDiff, astNodesModified, whitespaceLines, commentLines int) (float64, float64) {
	if totalLinesDiff == 0 {
		return 1.0, 0.0
	}

	surgical := float64(astNodesModified) / float64(totalLinesDiff+1)
	if surgical > 1.0 {
		surgical = 1.0
	}
	// Round to 3 decimal places
	surgical = math.Round(surgical*1000) / 1000

	cosmetic := float64(whitespaceLines+commentLines) / float64(totalLinesDiff)
	if cosmetic > 1.0 {
		cosmetic = 1.0
	}
	cosmetic = math.Round(cosmetic*1000) / 1000

	return surgical, cosmetic
}
