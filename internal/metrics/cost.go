package metrics

import (
	"math"
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// ModelPricing defines pricing rates per million tokens in USD.
type ModelPricing struct {
	PromptPricePerM     float64
	CompletionPricePerM float64
}

// PricingCatalog holds standard market pricing for common frontier models.
var PricingCatalog = map[string]ModelPricing{
	// OpenAI
	"gpt-4o":      {PromptPricePerM: 2.50, CompletionPricePerM: 10.00},
	"gpt-4o-mini": {PromptPricePerM: 0.15, CompletionPricePerM: 0.60},
	"o1":          {PromptPricePerM: 15.00, CompletionPricePerM: 60.00},
	"o3-mini":     {PromptPricePerM: 1.10, CompletionPricePerM: 4.40},

	// Anthropic
	"claude-3-7-sonnet": {PromptPricePerM: 3.00, CompletionPricePerM: 15.00},
	"claude-3-5-sonnet": {PromptPricePerM: 3.00, CompletionPricePerM: 15.00},
	"claude-3-5-haiku":  {PromptPricePerM: 0.80, CompletionPricePerM: 4.00},

	// DeepSeek
	"deepseek-chat":     {PromptPricePerM: 0.14, CompletionPricePerM: 0.28},
	"deepseek-reasoner": {PromptPricePerM: 0.55, CompletionPricePerM: 2.19},

	// Google
	"gemini-2.0-flash": {PromptPricePerM: 0.10, CompletionPricePerM: 0.40},
	"gemini-1.5-pro":   {PromptPricePerM: 1.25, CompletionPricePerM: 5.00},

	// Local & Mock models are free
	"mock":   {PromptPricePerM: 0.00, CompletionPricePerM: 0.00},
	"ollama": {PromptPricePerM: 0.00, CompletionPricePerM: 0.00},
}

// EstimateTrajectoryCost computes the monetary cost of a trajectory based on token usage.
func EstimateTrajectoryCost(model string, usage trajectory.TokenUsage) float64 {
	pricing := ResolveModelPricing(model)
	promptCost := (float64(usage.PromptTokens) / 1_000_000.0) * pricing.PromptPricePerM
	completionCost := (float64(usage.CompletionTokens) / 1_000_000.0) * pricing.CompletionPricePerM
	return math.Round((promptCost+completionCost)*10000) / 10000
}

// ResolveModelPricing matches a model name to its catalog pricing entry.
func ResolveModelPricing(model string) ModelPricing {
	normalized := strings.ToLower(model)
	for key, p := range PricingCatalog {
		if strings.Contains(normalized, key) {
			return p
		}
	}
	return ModelPricing{PromptPricePerM: 2.00, CompletionPricePerM: 8.00}
}

// CalculateCostProjection compares candidate vs baseline expenditures.
func CalculateCostProjection(candTraj, baseTraj *trajectory.Trajectory) trajectory.CostProjection {
	candCost := candTraj.TokenUsage.EstimatedCostUSD
	if candCost == 0 {
		candCost = EstimateTrajectoryCost(candTraj.Model, candTraj.TokenUsage)
	}

	baseCost := baseTraj.TokenUsage.EstimatedCostUSD
	if baseCost == 0 {
		baseCost = EstimateTrajectoryCost(baseTraj.Model, baseTraj.TokenUsage)
	}

	costDelta := candCost - baseCost
	var costRatio float64 = 1.0
	if baseCost > 0 {
		costRatio = candCost / baseCost
	}

	var costPerPrecision float64 = 0.0
	if candTraj.ASTMetrics.SurgicalScore > 0 {
		costPerPrecision = candCost / candTraj.ASTMetrics.SurgicalScore
	}

	return trajectory.CostProjection{
		CandidateCostUSD:      candCost,
		BaselineCostUSD:       baseCost,
		CostDeltaUSD:          math.Round(costDelta*10000) / 10000,
		CostRatio:             math.Round(costRatio*100) / 100,
		MonthlyDiffAt1kTasks:  math.Round(costDelta*1000*100) / 100,
		MonthlyDiffAt10kTasks: math.Round(costDelta*10000*100) / 100,
		MonthlyDiffAt50kTasks: math.Round(costDelta*50000*100) / 100,
		CostPerPrecisionPoint: math.Round(costPerPrecision*10000) / 10000,
	}
}
