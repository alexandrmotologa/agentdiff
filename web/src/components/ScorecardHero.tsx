import React from "react";
import { ComparisonResult } from "../types";
import { 
  ShieldCheck, 
  AlertTriangle, 
  Clock, 
  Cpu, 
  Code2, 
  GitFork, 
  Flame, 
  CheckCircle2, 
  XCircle,
  Coins,
  Compass,
  Lightbulb
} from "lucide-react";

interface ScorecardHeroProps {
  comparison: ComparisonResult;
}

export const ScorecardHero: React.FC<ScorecardHeroProps> = ({ comparison }) => {
  const isRegression = comparison.verdict === "REGRESSION";
  const isPass = comparison.verdict === "PASS";

  const bloatPct = Math.round((comparison.tool_bloat_factor - 1) * 100);
  const candSec = (comparison.candidate_duration_ms / 1000).toFixed(1);
  const baseSec = (comparison.baseline_duration_ms / 1000).toFixed(1);

  const cost = comparison.cost_projection;
  const diag = comparison.diagnosis;

  return (
    <div className="bg-surface border border-surfaceBorder rounded-xl p-6 shadow-2xl space-y-6">
      {/* Top Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-4 border-b border-surfaceBorder/80">
        <div>
          <div className="flex items-center gap-3">
            <h2 className="text-xl font-bold tracking-tight text-white flex items-center gap-2">
              <span className="text-sky-400 font-mono text-base">{comparison.task_id}</span>
            </h2>
            <span
              className={`px-3 py-1 rounded-full text-xs font-semibold uppercase tracking-wider flex items-center gap-1.5 ${
                isRegression
                  ? "bg-rose-500/20 text-rose-400 border border-rose-500/40"
                  : isPass
                  ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                  : "bg-amber-500/20 text-amber-400 border border-amber-500/40"
              }`}
            >
              {isRegression ? (
                <AlertTriangle className="w-3.5 h-3.5" />
              ) : isPass ? (
                <ShieldCheck className="w-3.5 h-3.5" />
              ) : (
                <Flame className="w-3.5 h-3.5" />
              )}
              {comparison.verdict}
            </span>
          </div>
          <p className="text-xs text-slate-400 mt-1 flex items-center gap-2">
            <span>Candidate: <strong className="text-slate-200">{comparison.candidate_model}</strong></span>
            <span className="text-slate-600">vs</span>
            <span>Baseline: <strong className="text-slate-200">{comparison.baseline_model}</strong></span>
          </p>
        </div>

        {/* Regression count alert banner if regressed */}
        {comparison.regressions && comparison.regressions.length > 0 && (
          <div className="bg-rose-950/40 border border-rose-900/60 rounded-lg px-4 py-2 text-rose-300 text-xs flex items-center gap-2.5">
            <AlertTriangle className="w-4 h-4 text-rose-400 flex-shrink-0" />
            <span>
              <strong>{comparison.regressions.length} Regression(s) Detected</strong>: {comparison.regressions[0]}
            </span>
          </div>
        )}
      </div>

      {/* Root-Cause Diagnosis & Steering Banner */}
      {diag && diag.divergence_turn > 0 && (
        <div className="bg-amber-950/25 border border-amber-500/30 rounded-lg p-4 space-y-2">
          <div className="flex items-center gap-2 text-amber-400 font-semibold text-xs tracking-wide">
            <Compass className="w-4 h-4 text-amber-400 flex-shrink-0" />
            <span>Divergence Root Cause (Turn {diag.divergence_turn} • {diag.category.replace(/_/g, " ")})</span>
          </div>
          <p className="text-slate-200 text-xs leading-relaxed pl-6">
            {diag.summary}
          </p>
          {diag.recommended_steering && (
            <div className="mt-2 pl-6 pt-2 border-t border-amber-500/20 text-xs flex items-start gap-2 text-amber-200/90 font-mono">
              <Lightbulb className="w-3.5 h-3.5 text-amber-300 mt-0.5 flex-shrink-0" />
              <span><strong>Prompt Steering:</strong> {diag.recommended_steering}</span>
            </div>
          )}
        </div>
      )}

      {/* Grid of Key Diagnostic Metrics */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
        {/* Metric 1: Verification */}
        <div className="bg-background/60 border border-surfaceBorder rounded-lg p-3.5 flex flex-col justify-between">
          <div className="text-[11px] font-medium text-slate-400 flex items-center justify-between">
            <span>Verification</span>
            {comparison.candidate_verified ? (
              <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
            ) : (
              <XCircle className="w-3.5 h-3.5 text-rose-400" />
            )}
          </div>
          <div className="mt-2">
            <div className="text-base font-bold text-white flex items-center gap-1.5">
              <span className={comparison.candidate_verified ? "text-emerald-400" : "text-rose-400"}>
                {comparison.candidate_verified ? "PASSED" : "FAILED"}
              </span>
            </div>
            <div className="text-[10px] text-slate-400 mt-0.5">
              Base: {comparison.baseline_verified ? "PASSED" : "FAILED"}
            </div>
          </div>
        </div>

        {/* Metric 2: Speedup & Latency */}
        <div className="bg-background/60 border border-surfaceBorder rounded-lg p-3.5 flex flex-col justify-between">
          <div className="text-[11px] font-medium text-slate-400 flex items-center justify-between">
            <span>Latency / Speedup</span>
            <Clock className="w-3.5 h-3.5 text-sky-400" />
          </div>
          <div className="mt-2">
            <div className="text-base font-bold text-white">
              {comparison.speedup_factor.toFixed(2)}x
            </div>
            <div className="text-[10px] text-slate-400 mt-0.5">
              {candSec}s <span className="text-slate-600">vs</span> {baseSec}s
            </div>
          </div>
        </div>

        {/* Metric 3: Tool Call Bloat */}
        <div className="bg-background/60 border border-surfaceBorder rounded-lg p-3.5 flex flex-col justify-between">
          <div className="text-[11px] font-medium text-slate-400 flex items-center justify-between">
            <span>Tool Bloat</span>
            <Cpu className="w-3.5 h-3.5 text-indigo-400" />
          </div>
          <div className="mt-2">
            <div className={`text-base font-bold ${bloatPct > 15 ? "text-rose-400" : "text-emerald-400"}`}>
              {bloatPct > 0 ? `+${bloatPct}%` : `${bloatPct}%`}
            </div>
            <div className="text-[10px] text-slate-400 mt-0.5">
              {comparison.candidate_tool_count} calls <span className="text-slate-600">vs</span> {comparison.baseline_tool_count}
            </div>
          </div>
        </div>

        {/* Metric 4: AST Surgical Precision */}
        <div className="bg-background/60 border border-surfaceBorder rounded-lg p-3.5 flex flex-col justify-between">
          <div className="text-[11px] font-medium text-slate-400 flex items-center justify-between">
            <span>AST Precision</span>
            <Code2 className="w-3.5 h-3.5 text-teal-400" />
          </div>
          <div className="mt-2">
            <div className="text-base font-bold text-white">
              {comparison.candidate_surgical_score.toFixed(3)}
            </div>
            <div className={`text-[10px] mt-0.5 ${comparison.surgical_score_delta < -0.1 ? "text-rose-400 font-medium" : "text-slate-400"}`}>
              Δ {comparison.surgical_score_delta >= 0 ? "+" : ""}{comparison.surgical_score_delta.toFixed(3)}
            </div>
          </div>
        </div>

        {/* Metric 5: Trajectory Edit Distance */}
        <div className="bg-background/60 border border-surfaceBorder rounded-lg p-3.5 flex flex-col justify-between">
          <div className="text-[11px] font-medium text-slate-400 flex items-center justify-between">
            <span>Trajectory Dist</span>
            <GitFork className="w-3.5 h-3.5 text-purple-400" />
          </div>
          <div className="mt-2">
            <div className="text-base font-bold text-white">
              {comparison.trajectory_edit_distance}
            </div>
            <div className="text-[10px] text-slate-400 mt-0.5">
              Norm: {comparison.normalized_trajectory_distance.toFixed(2)}
            </div>
          </div>
        </div>

        {/* Metric 6: Token Economics & Scale Impact */}
        <div className="bg-background/60 border border-surfaceBorder rounded-lg p-3.5 flex flex-col justify-between">
          <div className="text-[11px] font-medium text-slate-400 flex items-center justify-between">
            <span>Token Economics</span>
            <Coins className="w-3.5 h-3.5 text-amber-400" />
          </div>
          <div className="mt-2">
            <div className="text-base font-bold text-white">
              {cost ? `$${cost.candidate_cost_usd.toFixed(4)}` : "$0.00"}
            </div>
            <div className="text-[10px] text-slate-400 mt-0.5 font-mono">
              {cost ? `10k: ${cost.monthly_diff_at_10k_tasks >= 0 ? "+" : ""}$${cost.monthly_diff_at_10k_tasks.toFixed(1)}/mo` : "Local/Zero"}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
