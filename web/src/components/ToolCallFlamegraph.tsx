import React from "react";
import { Trajectory, ToolCall } from "../types";
import { Clock, Cpu } from "lucide-react";

interface ToolCallFlamegraphProps {
  candidateTraj?: Trajectory;
  baselineTraj?: Trajectory;
}

export const ToolCallFlamegraph: React.FC<ToolCallFlamegraphProps> = ({
  candidateTraj,
  baselineTraj,
}) => {
  const candTools: { turn: number; tool: ToolCall }[] = [];
  candidateTraj?.turns?.forEach((t) => {
    t.tool_calls?.forEach((c) => {
      candTools.push({ turn: t.turn_index, tool: c });
    });
  });

  const baseTools: { turn: number; tool: ToolCall }[] = [];
  baselineTraj?.turns?.forEach((t) => {
    t.tool_calls?.forEach((c) => {
      baseTools.push({ turn: t.turn_index, tool: c });
    });
  });

  const getToolColor = (toolName: string) => {
    switch (toolName) {
      case "read_file":
        return "bg-sky-500/20 text-sky-300 border-sky-500/40";
      case "edit_file":
      case "write_file":
        return "bg-emerald-500/20 text-emerald-300 border-emerald-500/40";
      case "run_tests":
      case "exec_bash":
        return "bg-purple-500/20 text-purple-300 border-purple-500/40";
      case "list_dir":
        return "bg-amber-500/20 text-amber-300 border-amber-500/40";
      default:
        return "bg-slate-700/50 text-slate-300 border-slate-600";
    }
  };

  return (
    <div className="bg-surface border border-surfaceBorder rounded-xl p-5 shadow-2xl flex flex-col space-y-4">
      <div className="flex items-center justify-between pb-3 border-b border-surfaceBorder/80">
        <div className="flex items-center gap-2">
          <Cpu className="w-4 h-4 text-indigo-400" />
          <h3 className="text-sm font-bold uppercase tracking-wider text-slate-200">
            Tool Call Execution Timeline & Latency Distribution
          </h3>
        </div>
        <div className="flex items-center gap-3 text-[11px] text-slate-400 font-mono">
          <div className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-sm bg-sky-500/40 border border-sky-500 inline-block"></span>
            <span>read</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-sm bg-emerald-500/40 border border-emerald-500 inline-block"></span>
            <span>edit/write</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-sm bg-purple-500/40 border border-purple-500 inline-block"></span>
            <span>exec/test</span>
          </div>
        </div>
      </div>

      <div className="space-y-4 text-xs">
        {/* Candidate Timeline */}
        <div className="space-y-1.5">
          <div className="flex items-center justify-between text-[11px] text-slate-400">
            <span className="font-semibold text-sky-300">
              Candidate: {candidateTraj?.model || "Candidate"} ({candTools.length} calls)
            </span>
            <span className="font-mono">{candidateTraj?.duration_ms || 0}ms total</span>
          </div>
          <div className="bg-background/60 border border-surfaceBorder/70 rounded-lg p-2.5 flex flex-wrap gap-1.5 min-h-[52px] items-center">
            {candTools.length === 0 ? (
              <span className="text-slate-600 italic">No tool calls recorded</span>
            ) : (
              candTools.map((item, idx) => (
                <div
                  key={idx}
                  title={`Turn ${item.turn}: ${item.tool.tool_name} (${item.tool.duration_ms}ms)`}
                  className={`px-2.5 py-1 rounded border text-[11px] font-mono flex items-center gap-1.5 shadow-sm ${getToolColor(
                    item.tool.tool_name
                  )}`}
                >
                  <span className="opacity-60 text-[9px]">T{item.turn}</span>
                  <span className="font-semibold">{item.tool.tool_name}</span>
                  <span className="text-[10px] opacity-75">{item.tool.duration_ms}ms</span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Baseline Timeline */}
        <div className="space-y-1.5">
          <div className="flex items-center justify-between text-[11px] text-slate-400">
            <span className="font-semibold text-indigo-300">
              Baseline: {baselineTraj?.model || "Baseline"} ({baseTools.length} calls)
            </span>
            <span className="font-mono">{baselineTraj?.duration_ms || 0}ms total</span>
          </div>
          <div className="bg-background/60 border border-surfaceBorder/70 rounded-lg p-2.5 flex flex-wrap gap-1.5 min-h-[52px] items-center">
            {baseTools.length === 0 ? (
              <span className="text-slate-600 italic">No tool calls recorded</span>
            ) : (
              baseTools.map((item, idx) => (
                <div
                  key={idx}
                  title={`Turn ${item.turn}: ${item.tool.tool_name} (${item.tool.duration_ms}ms)`}
                  className={`px-2.5 py-1 rounded border text-[11px] font-mono flex items-center gap-1.5 shadow-sm ${getToolColor(
                    item.tool.tool_name
                  )}`}
                >
                  <span className="opacity-60 text-[9px]">T{item.turn}</span>
                  <span className="font-semibold">{item.tool.tool_name}</span>
                  <span className="text-[10px] opacity-75">{item.tool.duration_ms}ms</span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
