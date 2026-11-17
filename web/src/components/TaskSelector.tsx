import React, { useState } from "react";
import { TaskListItem } from "../types";
import { Search, AlertCircle, CheckCircle2, ChevronRight, Play } from "lucide-react";

interface TaskSelectorProps {
  tasks: TaskListItem[];
  selectedTaskId: string;
  onSelectTask: (taskId: string) => void;
  onRunBenchmark: () => void;
  isRunningBenchmark: boolean;
}

export const TaskSelector: React.FC<TaskSelectorProps> = ({
  tasks,
  selectedTaskId,
  onSelectTask,
  onRunBenchmark,
  isRunningBenchmark,
}) => {
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<"ALL" | "REGRESSION" | "PASS">("ALL");

  const filteredTasks = tasks.filter((t) => {
    const matchesSearch = t.task_id.toLowerCase().includes(search.toLowerCase());
    if (!matchesSearch) return false;
    if (filter === "REGRESSION") return t.verdict === "REGRESSION";
    if (filter === "PASS") return t.verdict === "PASS";
    return true;
  });

  return (
    <div className="bg-surface border border-surfaceBorder rounded-xl p-4 flex flex-col h-full space-y-4">
      <div className="flex items-center justify-between pb-3 border-b border-surfaceBorder">
        <h3 className="text-sm font-bold uppercase tracking-wider text-slate-300">
          Benchmark Tasks ({tasks.length})
        </h3>
        <button
          onClick={onRunBenchmark}
          disabled={isRunningBenchmark}
          className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold tracking-wide transition-all ${
            isRunningBenchmark
              ? "bg-slate-800 text-slate-500 cursor-not-allowed"
              : "bg-sky-500 hover:bg-sky-400 text-white shadow-lg shadow-sky-500/20 active:scale-95"
          }`}
        >
          <Play className={`w-3.5 h-3.5 ${isRunningBenchmark ? "animate-spin" : ""}`} />
          {isRunningBenchmark ? "Running..." : "Run Suite"}
        </button>
      </div>

      {/* Search Input */}
      <div className="relative">
        <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          type="text"
          placeholder="Filter tasks..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full bg-background/80 border border-surfaceBorder rounded-lg pl-9 pr-3 py-1.5 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-sky-500 transition-colors"
        />
      </div>

      {/* Filter Tabs */}
      <div className="flex items-center gap-1 bg-background/60 p-1 rounded-lg border border-surfaceBorder/60">
        <button
          onClick={() => setFilter("ALL")}
          className={`flex-1 py-1 rounded text-[11px] font-semibold transition-all ${
            filter === "ALL"
              ? "bg-surface text-white shadow-sm"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          All
        </button>
        <button
          onClick={() => setFilter("REGRESSION")}
          className={`flex-1 py-1 rounded text-[11px] font-semibold transition-all ${
            filter === "REGRESSION"
              ? "bg-rose-950/60 text-rose-300 border border-rose-900/50"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          Regressed
        </button>
        <button
          onClick={() => setFilter("PASS")}
          className={`flex-1 py-1 rounded text-[11px] font-semibold transition-all ${
            filter === "PASS"
              ? "bg-emerald-950/60 text-emerald-300 border border-emerald-900/50"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          Passed
        </button>
      </div>

      {/* Tasks List */}
      <div className="flex-1 overflow-y-auto space-y-1.5 pr-1">
        {filteredTasks.length === 0 ? (
          <div className="text-center py-8 text-xs text-slate-500">
            No matching tasks found.
          </div>
        ) : (
          filteredTasks.map((t) => {
            const isSelected = t.task_id === selectedTaskId;
            const isRegressed = t.verdict === "REGRESSION";

            return (
              <button
                key={t.task_id}
                onClick={() => onSelectTask(t.task_id)}
                className={`w-full text-left p-3 rounded-lg border transition-all flex items-center justify-between group ${
                  isSelected
                    ? "bg-sky-500/10 border-sky-500/50 shadow-md shadow-sky-500/5"
                    : "bg-background/40 hover:bg-background/80 border-surfaceBorder/70 text-slate-300"
                }`}
              >
                <div className="space-y-1 truncate pr-2">
                  <div className="flex items-center gap-2">
                    {isRegressed ? (
                      <AlertCircle className="w-3.5 h-3.5 text-rose-400 flex-shrink-0" />
                    ) : (
                      <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 flex-shrink-0" />
                    )}
                    <span className="font-mono text-xs font-semibold truncate text-white">
                      {t.task_id}
                    </span>
                  </div>
                  <div className="text-[10px] text-slate-400 flex items-center gap-2 pl-5.5">
                    <span>Bloat: <strong className={t.tool_bloat_factor > 1.25 ? "text-rose-400" : "text-slate-300"}>{t.tool_bloat_factor.toFixed(2)}x</strong></span>
                    <span>•</span>
                    <span>AST Δ: <strong className={t.surgical_score_delta < -0.1 ? "text-rose-400" : "text-slate-300"}>{t.surgical_score_delta.toFixed(2)}</strong></span>
                  </div>
                </div>
                <ChevronRight
                  className={`w-4 h-4 transition-transform flex-shrink-0 ${
                    isSelected
                      ? "text-sky-400 translate-x-0.5"
                      : "text-slate-600 group-hover:text-slate-400"
                  }`}
                />
              </button>
            );
          })
        )}
      </div>
    </div>
  );
};
