import React, { useState, useEffect } from "react";
import { Trophy, TrendingUp, Cpu, Code2, Clock, Coins, CheckCircle, AlertTriangle } from "lucide-react";

export interface LeaderboardEntry {
  rank: number;
  model_name: string;
  total_tasks: number;
  tasks_solved: number;
  pass_rate_pct: number;
  mean_surgical_score: number;
  mean_efficiency_ratio: number;
  mean_duration_ms: number;
  total_cost_usd: number;
  mean_cost_usd: number;
  total_regressions_count: number;
}

export const LeaderboardMatrix: React.FC = () => {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchLeaderboard = async () => {
      try {
        if ((window as any).__AGENTDIFF_DATA__) {
          const preloaded = (window as any).__AGENTDIFF_DATA__;
          // Compute client-side if offline standalone report
          const comparisons = Object.values(preloaded.comparisons || {}) as any[];
          const modelMap: Record<string, any> = {};

          comparisons.forEach((c) => {
            const cand = c.candidate_model;
            if (!modelMap[cand]) {
              modelMap[cand] = { count: 0, solved: 0, surg: 0, eff: 0, dur: 0, cost: 0, reg: 0 };
            }
            modelMap[cand].count++;
            if (c.candidate_verified) modelMap[cand].solved++;
            modelMap[cand].surg += c.candidate_surgical_score || 0;
            modelMap[cand].eff += c.tool_efficiency_candidate || 0;
            modelMap[cand].dur += c.candidate_duration_ms || 0;
            modelMap[cand].cost += c.cost_projection?.candidate_cost_usd || 0;
            modelMap[cand].reg += (c.regressions || []).length;

            const base = c.baseline_model;
            if (base && !modelMap[base]) {
              modelMap[base] = { count: 0, solved: 0, surg: 0, eff: 0, dur: 0, cost: 0, reg: 0 };
            }
            if (base) {
              modelMap[base].count++;
              if (c.baseline_verified) modelMap[base].solved++;
              modelMap[base].surg += c.baseline_surgical_score || 0;
              modelMap[base].eff += c.tool_efficiency_baseline || 0;
              modelMap[base].dur += c.baseline_duration_ms || 0;
              modelMap[base].cost += c.cost_projection?.baseline_cost_usd || 0;
            }
          });

          const list: LeaderboardEntry[] = Object.keys(modelMap).map((m, idx) => {
            const acc = modelMap[m];
            return {
              rank: idx + 1,
              model_name: m,
              total_tasks: acc.count,
              tasks_solved: acc.solved,
              pass_rate_pct: Math.round((acc.solved / acc.count) * 1000) / 10,
              mean_surgical_score: Math.round((acc.surg / acc.count) * 1000) / 1000,
              mean_efficiency_ratio: Math.round((acc.eff / acc.count) * 1000) / 1000,
              mean_duration_ms: Math.round(acc.dur / acc.count),
              total_cost_usd: Math.round(acc.cost * 10000) / 10000,
              mean_cost_usd: Math.round((acc.cost / acc.count) * 10000) / 10000,
              total_regressions_count: acc.reg,
            };
          });

          list.sort((a, b) => b.pass_rate_pct - a.pass_rate_pct || b.mean_surgical_score - a.mean_surgical_score);
          list.forEach((e, i) => (e.rank = i + 1));
          setEntries(list);
          setLoading(false);
          return;
        }

        const res = await fetch("/api/leaderboard");
        if (res.ok) {
          const data: LeaderboardEntry[] = await res.json();
          setEntries(data || []);
        }
      } catch (err) {
        console.error("Failed to load leaderboard:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchLeaderboard();
  }, []);

  if (loading) {
    return (
      <div className="bg-surface border border-surfaceBorder rounded-xl p-8 text-center text-slate-400 text-xs">
        Compiling multi-model leaderboard...
      </div>
    );
  }

  return (
    <div className="bg-surface border border-surfaceBorder rounded-xl p-6 shadow-2xl space-y-6">
      <div className="flex items-center justify-between pb-4 border-b border-surfaceBorder/80">
        <div className="flex items-center gap-2.5">
          <Trophy className="w-5 h-5 text-amber-400" />
          <div>
            <h3 className="text-base font-bold text-white">
              Multi-Model Benchmark Leaderboard & Frontier Matrix
            </h3>
            <p className="text-xs text-slate-400 mt-0.5">
              Aggregated pass rates, AST surgical precision, and token economics across benchmarked agent models.
            </p>
          </div>
        </div>
        <span className="text-xs font-mono text-slate-400 bg-background/80 px-3 py-1 rounded-lg border border-surfaceBorder">
          {entries.length} Models Benchmarked
        </span>
      </div>

      {entries.length === 0 ? (
        <div className="text-center py-12 text-slate-500 text-xs italic">
          No benchmark results available to generate leaderboard. Run a benchmark suite first.
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs font-sans">
            <thead>
              <tr className="border-b border-surfaceBorder/80 text-slate-400 uppercase tracking-wider text-[11px] font-semibold bg-background/30">
                <th className="py-3 px-3">Rank</th>
                <th className="py-3 px-3">Model</th>
                <th className="py-3 px-3">Pass Rate</th>
                <th className="py-3 px-3">AST Precision</th>
                <th className="py-3 px-3">Tool Efficiency</th>
                <th className="py-3 px-3">Avg Duration</th>
                <th className="py-3 px-3">Avg Cost</th>
                <th className="py-3 px-3">Regressions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surfaceBorder/50 font-mono text-xs">
              {entries.map((e) => (
                <tr
                  key={e.model_name}
                  className={`hover:bg-slate-800/30 transition-colors ${
                    e.rank === 1 ? "bg-amber-500/5 font-semibold" : ""
                  }`}
                >
                  <td className="py-3.5 px-3">
                    <span
                      className={`w-6 h-6 rounded-full inline-flex items-center justify-center font-bold text-xs ${
                        e.rank === 1
                          ? "bg-amber-400 text-black shadow-md shadow-amber-400/20"
                          : e.rank === 2
                          ? "bg-slate-300 text-black"
                          : e.rank === 3
                          ? "bg-amber-700 text-white"
                          : "bg-slate-800 text-slate-400"
                      }`}
                    >
                      {e.rank}
                    </span>
                  </td>
                  <td className="py-3.5 px-3">
                    <span className="text-white font-bold">{e.model_name}</span>
                    <span className="text-slate-500 block text-[10px] font-sans">
                      {e.tasks_solved} of {e.total_tasks} tasks passed
                    </span>
                  </td>
                  <td className="py-3.5 px-3">
                    <div className="flex items-center gap-2">
                      <span className={e.pass_rate_pct >= 80 ? "text-emerald-400 font-bold" : "text-amber-400"}>
                        {e.pass_rate_pct.toFixed(1)}%
                      </span>
                      <div className="w-16 bg-slate-800 rounded-full h-1.5 overflow-hidden">
                        <div
                          className="bg-emerald-400 h-full rounded-full"
                          style={{ width: `${e.pass_rate_pct}%` }}
                        ></div>
                      </div>
                    </div>
                  </td>
                  <td className="py-3.5 px-3">
                    <span className="text-teal-300 font-semibold">
                      {e.mean_surgical_score.toFixed(3)}
                    </span>
                  </td>
                  <td className="py-3.5 px-3">
                    <span className="text-slate-300">
                      {(e.mean_efficiency_ratio * 100).toFixed(1)}%
                    </span>
                  </td>
                  <td className="py-3.5 px-3 text-slate-300">
                    {(e.mean_duration_ms / 1000).toFixed(1)}s
                  </td>
                  <td className="py-3.5 px-3 text-amber-300">
                    ${e.mean_cost_usd.toFixed(4)}
                  </td>
                  <td className="py-3.5 px-3">
                    {e.total_regressions_count > 0 ? (
                      <span className="bg-rose-950/60 border border-rose-900 text-rose-400 px-2 py-0.5 rounded text-[10px] font-bold">
                        {e.total_regressions_count} regressed
                      </span>
                    ) : (
                      <span className="text-emerald-400 text-[11px] font-bold">0</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
