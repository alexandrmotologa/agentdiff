import React, { useState } from "react";
import { Trajectory, FileMutation, ASTNodeRef } from "../types";
import { Code2, FileCode, ChevronRight, Layers, Sparkles } from "lucide-react";

interface AstDiffViewerProps {
  candidateTraj?: Trajectory;
  baselineTraj?: Trajectory;
}

export const AstDiffViewer: React.FC<AstDiffViewerProps> = ({
  candidateTraj,
  baselineTraj,
}) => {
  const [selectedAgent, setSelectedAgent] = useState<"candidate" | "baseline">("candidate");
  const [activeFileIdx, setActiveFileIdx] = useState(0);
  const [selectedNode, setSelectedNode] = useState<ASTNodeRef | null>(null);

  const currentTraj = selectedAgent === "candidate" ? candidateTraj : baselineTraj;
  const metrics = currentTraj?.ast_metrics;

  // Gather all mutations across turns
  const mutations: FileMutation[] = [];
  if (currentTraj?.turns) {
    for (const turn of currentTraj.turns) {
      if (turn.file_mutations) {
        mutations.push(...turn.file_mutations);
      }
    }
  }

  const activeMutation = mutations[activeFileIdx];

  return (
    <div className="bg-surface border border-surfaceBorder rounded-xl p-5 shadow-2xl flex flex-col space-y-4">
      {/* Header and Agent Selector */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3 pb-3 border-b border-surfaceBorder/80">
        <div className="flex items-center gap-2.5">
          <Code2 className="w-4 h-4 text-teal-400" />
          <h3 className="text-sm font-bold uppercase tracking-wider text-slate-200">
            AST Semantic Diff & Syntax Hierarchy Inspector
          </h3>
        </div>

        {/* Candidate vs Baseline Tab */}
        <div className="flex items-center bg-background rounded-lg border border-surfaceBorder p-0.5 text-xs">
          <button
            onClick={() => { setSelectedAgent("candidate"); setActiveFileIdx(0); setSelectedNode(null); }}
            className={`px-3 py-1 rounded font-semibold transition-all ${
              selectedAgent === "candidate"
                ? "bg-sky-500 text-white shadow"
                : "text-slate-400 hover:text-slate-200"
            }`}
          >
            Candidate Edits
          </button>
          <button
            onClick={() => { setSelectedAgent("baseline"); setActiveFileIdx(0); setSelectedNode(null); }}
            className={`px-3 py-1 rounded font-semibold transition-all ${
              selectedAgent === "baseline"
                ? "bg-indigo-500 text-white shadow"
                : "text-slate-400 hover:text-slate-200"
            }`}
          >
            Baseline Edits
          </button>
        </div>
      </div>

      {/* AST Precision Stats Bar */}
      {metrics && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 bg-background/50 border border-surfaceBorder/60 p-3 rounded-lg text-xs font-mono">
          <div>
            <span className="text-slate-500 block text-[10px]">Surgical Score</span>
            <span className="text-emerald-400 font-bold text-sm">
              {metrics.surgical_score.toFixed(3)}
            </span>
          </div>
          <div>
            <span className="text-slate-500 block text-[10px]">AST Functional Nodes</span>
            <span className="text-white font-bold text-sm">
              {metrics.ast_nodes_modified} nodes
            </span>
          </div>
          <div>
            <span className="text-slate-500 block text-[10px]">Total Line Diff</span>
            <span className="text-white font-bold text-sm">
              {metrics.total_lines_diff} lines
            </span>
          </div>
          <div>
            <span className="text-slate-500 block text-[10px]">Cosmetic Churn Ratio</span>
            <span className={`font-bold text-sm ${metrics.cosmetic_churn_ratio > 0.3 ? "text-amber-400" : "text-slate-300"}`}>
              {(metrics.cosmetic_churn_ratio * 100).toFixed(1)}%
            </span>
          </div>
        </div>
      )}

      {/* Files List & Diff Viewer */}
      {mutations.length === 0 ? (
        <div className="text-center py-12 text-slate-500 text-xs italic bg-background/30 rounded-lg border border-surfaceBorder/50">
          No file mutations were recorded for this agent trajectory.
        </div>
      ) : (
        <div className="space-y-3">
          {/* File tabs */}
          <div className="flex items-center gap-1.5 overflow-x-auto pb-1">
            {mutations.map((m, idx) => (
              <button
                key={idx}
                onClick={() => { setActiveFileIdx(idx); setSelectedNode(null); }}
                className={`px-3 py-1.5 rounded-lg border text-xs font-mono flex items-center gap-2 transition-all ${
                  activeFileIdx === idx
                    ? "bg-slate-800 border-sky-500/50 text-white shadow-sm"
                    : "bg-background/40 hover:bg-background border-surfaceBorder text-slate-400"
                }`}
              >
                <FileCode className="w-3.5 h-3.5 text-sky-400" />
                <span>{m.path}</span>
                <span className="text-[10px] text-emerald-400">+{m.lines_added}</span>
                <span className="text-[10px] text-rose-400">-{m.lines_removed}</span>
              </button>
            ))}
          </div>

          {/* AST Hierarchy Breadcrumb */}
          {activeMutation && (
            <div className="bg-slate-900/60 border border-surfaceBorder rounded-lg p-2.5 flex items-center justify-between text-xs font-mono">
              <div className="flex items-center gap-2 truncate text-slate-300">
                <Layers className="w-3.5 h-3.5 text-teal-400 flex-shrink-0" />
                <span>{activeMutation.path}</span>
                {selectedNode && (
                  <>
                    <ChevronRight className="w-3.5 h-3.5 text-slate-500 flex-shrink-0" />
                    <span className="text-teal-300 font-semibold">{selectedNode.kind}</span>
                    <ChevronRight className="w-3.5 h-3.5 text-slate-500 flex-shrink-0" />
                    <span className="text-sky-300 font-bold">{selectedNode.identifier}</span>
                    <span className="text-slate-500 text-[10px]">
                      (Lines {selectedNode.line_start}-{selectedNode.line_end})
                    </span>
                  </>
                )}
              </div>

              {selectedNode && (
                <button
                  onClick={() => setSelectedNode(null)}
                  className="text-[10px] text-slate-400 hover:text-white underline ml-2"
                >
                  Clear Filter
                </button>
              )}
            </div>
          )}

          {/* Diff Box */}
          {activeMutation && (
            <div className="bg-black/70 border border-surfaceBorder rounded-xl overflow-hidden font-mono text-xs shadow-inner">
              {/* Diff Header & AST Node Chips */}
              <div className="bg-slate-900/80 px-4 py-2.5 border-b border-surfaceBorder flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <span className="text-slate-300 font-semibold">{activeMutation.path}</span>
                {activeMutation.ast_nodes_affected && activeMutation.ast_nodes_affected.length > 0 && (
                  <div className="flex items-center gap-1.5 flex-wrap">
                    <span className="text-[10px] text-slate-500 uppercase flex items-center gap-1">
                      <Sparkles className="w-3 h-3 text-teal-400" />
                      AST Nodes:
                    </span>
                    {activeMutation.ast_nodes_affected.map((node, i) => {
                      const isNodeSelected = selectedNode?.identifier === node.identifier && selectedNode?.kind === node.kind;
                      return (
                        <button
                          key={i}
                          onClick={() => setSelectedNode(isNodeSelected ? null : node)}
                          className={`px-2 py-0.5 rounded text-[10px] font-semibold border transition-all ${
                            isNodeSelected
                              ? "bg-teal-500 text-black border-teal-300 shadow-sm shadow-teal-500/30 scale-105"
                              : "bg-teal-950/60 text-teal-300 hover:bg-teal-900/60 border-teal-800/60"
                          }`}
                        >
                          {node.kind}:{node.identifier}
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>

              {/* Diff Lines */}
              <div className="p-3 max-h-96 overflow-y-auto space-y-0.5 leading-relaxed">
                {activeMutation.diff.split("\n").map((line, idx) => {
                  let lineClass = "text-slate-400";
                  let bgClass = "";
                  if (line.startsWith("+") && !line.startsWith("+++")) {
                    lineClass = "text-emerald-300";
                    bgClass = "bg-emerald-950/30";
                  } else if (line.startsWith("-") && !line.startsWith("---")) {
                    lineClass = "text-rose-300";
                    bgClass = "bg-rose-950/30";
                  } else if (line.startsWith("@@")) {
                    lineClass = "text-sky-400 font-semibold";
                    bgClass = "bg-sky-950/20";
                  }

                  return (
                    <div
                      key={idx}
                      className={`px-2 py-0.5 rounded ${bgClass} ${lineClass} flex items-start`}
                    >
                      <span className="select-none text-slate-600 w-8 flex-shrink-0 text-right pr-2">
                        {idx + 1}
                      </span>
                      <span className="whitespace-pre-wrap">{line}</span>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
