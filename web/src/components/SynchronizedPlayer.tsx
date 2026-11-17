import React, { useState, useEffect } from "react";
import { Trajectory, Turn, ToolCall } from "../types";
import { 
  Play, 
  Pause, 
  RotateCcw, 
  ChevronLeft, 
  ChevronRight, 
  Terminal, 
  Wrench, 
  FileCode2, 
  CheckCircle, 
  XCircle,
  Eye,
  Bot
} from "lucide-react";

interface SynchronizedPlayerProps {
  candidateTraj?: Trajectory;
  baselineTraj?: Trajectory;
  divergenceTurn: number;
}

export const SynchronizedPlayer: React.FC<SynchronizedPlayerProps> = ({
  candidateTraj,
  baselineTraj,
  divergenceTurn,
}) => {
  const [currentTurnIdx, setCurrentTurnIdx] = useState(1);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState<1 | 2 | 5>(1);

  const candTurns = candidateTraj?.turns || [];
  const baseTurns = baselineTraj?.turns || [];
  const maxTurns = Math.max(candTurns.length, baseTurns.length, 1);

  useEffect(() => {
    let timer: ReturnType<typeof setInterval>;
    if (isPlaying) {
      timer = setInterval(() => {
        setCurrentTurnIdx((prev) => {
          if (prev >= maxTurns) {
            setIsPlaying(false);
            return prev;
          }
          return prev + 1;
        });
      }, 2000 / playbackSpeed);
    }
    return () => clearInterval(timer);
  }, [isPlaying, maxTurns, playbackSpeed]);

  const activeCandTurn = candTurns.find((t) => t.turn_index === currentTurnIdx);
  const activeBaseTurn = baseTurns.find((t) => t.turn_index === currentTurnIdx);

  const isDiverged = divergenceTurn > 0 && currentTurnIdx >= divergenceTurn;

  return (
    <div className="bg-surface border border-surfaceBorder rounded-xl p-5 shadow-2xl flex flex-col space-y-5">
      {/* Player Header & Controls */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-4 pb-4 border-b border-surfaceBorder/80">
        <div className="flex items-center gap-3">
          <span className="text-sm font-bold uppercase tracking-wider text-slate-200">
            Synchronized Trajectory Player
          </span>
          {isDiverged && (
            <span className="bg-amber-500/15 text-amber-400 border border-amber-500/30 text-[11px] font-semibold px-2.5 py-0.5 rounded-full">
              Diverged Strategy (Turn ≥ {divergenceTurn})
            </span>
          )}
        </div>

        {/* Playback Controls */}
        <div className="flex items-center gap-2">
          <button
            onClick={() => setCurrentTurnIdx(1)}
            title="Reset to Turn 1"
            className="p-1.5 rounded-lg bg-background hover:bg-slate-800 text-slate-300 border border-surfaceBorder transition-all"
          >
            <RotateCcw className="w-4 h-4" />
          </button>
          <button
            onClick={() => setCurrentTurnIdx((prev) => Math.max(1, prev - 1))}
            disabled={currentTurnIdx <= 1}
            title="Previous Turn"
            className="p-1.5 rounded-lg bg-background hover:bg-slate-800 disabled:opacity-40 text-slate-300 border border-surfaceBorder transition-all"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <button
            onClick={() => setIsPlaying(!isPlaying)}
            title={isPlaying ? "Pause" : "Play"}
            className="px-3 py-1.5 rounded-lg bg-sky-500 hover:bg-sky-400 text-white font-semibold text-xs flex items-center gap-1.5 shadow-md shadow-sky-500/20 transition-all"
          >
            {isPlaying ? <Pause className="w-3.5 h-3.5" /> : <Play className="w-3.5 h-3.5" />}
            {isPlaying ? "Pause" : "Play"}
          </button>
          <button
            onClick={() => setCurrentTurnIdx((prev) => Math.min(maxTurns, prev + 1))}
            disabled={currentTurnIdx >= maxTurns}
            title="Next Turn"
            className="p-1.5 rounded-lg bg-background hover:bg-slate-800 disabled:opacity-40 text-slate-300 border border-surfaceBorder transition-all"
          >
            <ChevronRight className="w-4 h-4" />
          </button>

          {/* Speed Toggle */}
          <div className="flex items-center bg-background rounded-lg border border-surfaceBorder p-0.5 ml-2 text-xs">
            {([1, 2, 5] as const).map((speed) => (
              <button
                key={speed}
                onClick={() => setPlaybackSpeed(speed)}
                className={`px-2 py-0.5 rounded font-mono font-medium transition-all ${
                  playbackSpeed === speed
                    ? "bg-sky-500/20 text-sky-300 border border-sky-500/40"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                {speed}x
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Scrubber Slider */}
      <div className="space-y-1.5">
        <div className="flex items-center justify-between text-xs text-slate-400">
          <span>Turn {currentTurnIdx} of {maxTurns}</span>
          <span className="font-mono text-[11px] text-slate-500">
            {currentTurnIdx === divergenceTurn ? "⚡ Point of divergence" : ""}
          </span>
        </div>
        <input
          type="range"
          min={1}
          max={maxTurns}
          value={currentTurnIdx}
          onChange={(e) => setCurrentTurnIdx(Number(e.target.value))}
          className="w-full accent-sky-400 cursor-pointer h-1.5 bg-slate-800 rounded-lg appearance-none"
        />
      </div>

      {/* Side-by-Side Dual Agent Panes */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Left Pane: Candidate Model */}
        <TurnCard
          title="Candidate Model"
          modelName={candidateTraj?.model || "Candidate"}
          turn={activeCandTurn}
          isTurnPresent={Boolean(activeCandTurn)}
          isCompleted={currentTurnIdx > candTurns.length}
          accentColor="sky"
        />

        {/* Right Pane: Baseline Model */}
        <TurnCard
          title="Baseline Model"
          modelName={baselineTraj?.model || "Baseline"}
          turn={activeBaseTurn}
          isTurnPresent={Boolean(activeBaseTurn)}
          isCompleted={currentTurnIdx > baseTurns.length}
          accentColor="indigo"
        />
      </div>
    </div>
  );
};

interface TurnCardProps {
  title: string;
  modelName: string;
  turn?: Turn;
  isTurnPresent: boolean;
  isCompleted: boolean;
  accentColor: "sky" | "indigo";
}

const TurnCard: React.FC<TurnCardProps> = ({
  title,
  modelName,
  turn,
  isTurnPresent,
  isCompleted,
  accentColor,
}) => {
  return (
    <div
      className={`border rounded-xl p-4 flex flex-col space-y-4 bg-background/50 ${
        accentColor === "sky" ? "border-sky-500/30" : "border-indigo-500/30"
      }`}
    >
      <div className="flex items-center justify-between pb-2.5 border-b border-surfaceBorder/60">
        <div className="flex items-center gap-2">
          <Bot className={`w-4 h-4 ${accentColor === "sky" ? "text-sky-400" : "text-indigo-400"}`} />
          <span className="font-bold text-xs uppercase tracking-wider text-slate-200">
            {title}
          </span>
          <span className="text-[10px] text-slate-500 font-mono">({modelName})</span>
        </div>
        {turn && (
          <span className="text-[11px] font-mono text-slate-400">
            {turn.duration_ms}ms
          </span>
        )}
      </div>

      {!isTurnPresent || !turn ? (
        <div className="flex-1 flex items-center justify-center py-12 text-slate-500 text-xs italic">
          {isCompleted ? "Agent completed execution prior to this turn." : "No action in this turn."}
        </div>
      ) : (
        <div className="space-y-4 text-xs">
          {/* Reasoning / Thought */}
          {turn.thought && (
            <div className="space-y-1">
              <div className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                <Eye className="w-3.5 h-3.5 text-slate-400" />
                Reasoning
              </div>
              <p className="bg-surface/80 border border-surfaceBorder rounded-lg p-3 text-slate-300 leading-relaxed font-sans text-xs">
                {turn.thought}
              </p>
            </div>
          )}

          {/* Tool Invocations */}
          {turn.tool_calls && turn.tool_calls.length > 0 && (
            <div className="space-y-2">
              <div className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                <Wrench className="w-3.5 h-3.5 text-amber-400" />
                Tool Invocations ({turn.tool_calls.length})
              </div>
              <div className="space-y-2">
                {turn.tool_calls.map((call, idx) => (
                  <ToolCallItem key={idx} call={call} />
                ))}
              </div>
            </div>
          )}

          {/* Terminal Outputs */}
          {turn.terminal_executions && turn.terminal_executions.length > 0 && (
            <div className="space-y-2">
              <div className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                <Terminal className="w-3.5 h-3.5 text-emerald-400" />
                Terminal Executions
              </div>
              {turn.terminal_executions.map((term, idx) => (
                <div key={idx} className="bg-black/60 border border-surfaceBorder rounded-lg p-2.5 font-mono text-[11px] space-y-1.5">
                  <div className="flex items-center justify-between text-slate-400 text-[10px]">
                    <span className="text-sky-300 font-semibold">$ {term.command}</span>
                    <span className={term.exit_code === 0 ? "text-emerald-400" : "text-rose-400 font-bold"}>
                      Exit: {term.exit_code} ({term.duration_ms}ms)
                    </span>
                  </div>
                  {term.stdout && (
                    <pre className="text-slate-300 whitespace-pre-wrap max-h-36 overflow-y-auto">
                      {term.stdout}
                    </pre>
                  )}
                  {term.stderr && (
                    <pre className="text-rose-400 whitespace-pre-wrap max-h-36 overflow-y-auto">
                      {term.stderr}
                    </pre>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* File Mutations */}
          {turn.file_mutations && turn.file_mutations.length > 0 && (
            <div className="space-y-1.5">
              <div className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                <FileCode2 className="w-3.5 h-3.5 text-teal-400" />
                File Mutations ({turn.file_mutations.length})
              </div>
              {turn.file_mutations.map((mut, idx) => (
                <div key={idx} className="bg-surface/50 border border-surfaceBorder rounded-lg p-2 flex items-center justify-between">
                  <span className="font-mono text-slate-200 text-xs">{mut.path}</span>
                  <div className="flex items-center gap-2 text-[10px] font-mono">
                    <span className="text-emerald-400">+{mut.lines_added}</span>
                    <span className="text-rose-400">-{mut.lines_removed}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

const ToolCallItem: React.FC<{ call: ToolCall }> = ({ call }) => {
  const [expanded, setExpanded] = useState(false);
  const isSuccess = call.status === "SUCCESS";

  return (
    <div className="bg-surface/60 border border-surfaceBorder rounded-lg overflow-hidden text-xs">
      <div
        onClick={() => setExpanded(!expanded)}
        className="p-2.5 flex items-center justify-between cursor-pointer hover:bg-slate-800/40 transition-colors"
      >
        <div className="flex items-center gap-2 truncate pr-2">
          {isSuccess ? (
            <CheckCircle className="w-3.5 h-3.5 text-emerald-400 flex-shrink-0" />
          ) : (
            <XCircle className="w-3.5 h-3.5 text-rose-400 flex-shrink-0" />
          )}
          <span className="font-mono font-bold text-sky-300">{call.tool_name}</span>
          {call.normalized_target && (
            <span className="text-slate-400 font-mono text-[11px] truncate">
              {call.normalized_target}
            </span>
          )}
        </div>
        <span className="font-mono text-[10px] text-slate-500">{call.duration_ms}ms</span>
      </div>

      {expanded && (
        <div className="p-3 bg-black/40 border-t border-surfaceBorder/60 space-y-2 font-mono text-[11px]">
          <div>
            <div className="text-[10px] text-slate-500 font-semibold mb-1">ARGUMENTS:</div>
            <pre className="text-slate-300 whitespace-pre-wrap max-h-36 overflow-y-auto">
              {JSON.stringify(call.arguments, null, 2)}
            </pre>
          </div>
          {call.output && (
            <div>
              <div className="text-[10px] text-slate-500 font-semibold mb-1">OUTPUT:</div>
              <pre className="text-slate-300 whitespace-pre-wrap max-h-48 overflow-y-auto">
                {call.output}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
