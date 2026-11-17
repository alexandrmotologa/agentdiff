import React from "react";
import { AntiPattern } from "../types";
import { AlertTriangle, AlertCircle, Info, Flame, Repeat, FileQuestion } from "lucide-react";

interface AntiPatternAlertsProps {
  candidatePatterns?: AntiPattern[];
  baselinePatterns?: AntiPattern[];
}

export const AntiPatternAlerts: React.FC<AntiPatternAlertsProps> = ({
  candidatePatterns = [],
  baselinePatterns = [],
}) => {
  const getIcon = (kind: string) => {
    switch (kind) {
      case "DOOM_LOOP":
        return <Flame className="w-4 h-4 text-rose-400" />;
      case "REDUNDANT_READ":
        return <Repeat className="w-4 h-4 text-amber-400" />;
      case "SHOTGUN_SURGERY":
        return <FileQuestion className="w-4 h-4 text-orange-400" />;
      default:
        return <AlertTriangle className="w-4 h-4 text-yellow-400" />;
    }
  };

  const hasAny = candidatePatterns.length > 0 || baselinePatterns.length > 0;

  if (!hasAny) {
    return (
      <div className="bg-surface border border-surfaceBorder rounded-xl p-4 flex items-center justify-between text-xs text-slate-400">
        <span className="flex items-center gap-2">
          <Info className="w-4 h-4 text-emerald-400" />
          <span>Zero behavioral anti-patterns or doom loops detected in either trajectory.</span>
        </span>
        <span className="bg-emerald-950/60 border border-emerald-800 text-emerald-400 px-2 py-0.5 rounded text-[11px] font-semibold">
          Clean Execution
        </span>
      </div>
    );
  }

  return (
    <div className="bg-surface border border-surfaceBorder rounded-xl p-5 shadow-2xl space-y-4">
      <div className="flex items-center gap-2 pb-3 border-b border-surfaceBorder/80">
        <AlertTriangle className="w-4 h-4 text-rose-400" />
        <h3 className="text-sm font-bold uppercase tracking-wider text-slate-200">
          Detected Behavioral Anti-Patterns & Operational Waste
        </h3>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
        {/* Candidate Anti-Patterns */}
        <div className="space-y-2">
          <h4 className="font-semibold text-sky-300 text-xs flex items-center justify-between">
            <span>Candidate Model Anomalies</span>
            <span className="text-[11px] font-mono text-slate-500">({candidatePatterns.length})</span>
          </h4>
          {candidatePatterns.length === 0 ? (
            <div className="bg-background/40 border border-surfaceBorder rounded-lg p-3 text-slate-500 text-xs italic">
              No anomalies detected for candidate model.
            </div>
          ) : (
            <div className="space-y-2">
              {candidatePatterns.map((ap, idx) => (
                <div
                  key={idx}
                  className={`p-3 rounded-lg border flex items-start gap-2.5 ${
                    ap.severity === "CRITICAL"
                      ? "bg-rose-950/30 border-rose-900/60 text-rose-200"
                      : "bg-amber-950/30 border-amber-900/60 text-amber-200"
                  }`}
                >
                  <div className="mt-0.5 flex-shrink-0">{getIcon(ap.kind)}</div>
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-[11px] uppercase tracking-wide">
                        {ap.kind.replace("_", " ")}
                      </span>
                      <span className="text-[10px] font-mono px-1.5 py-0.2 rounded bg-black/40 border border-white/10">
                        Turn {ap.turn_index}
                      </span>
                    </div>
                    <p className="text-slate-300 text-[11px] leading-relaxed">
                      {ap.details}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Baseline Anti-Patterns */}
        <div className="space-y-2">
          <h4 className="font-semibold text-indigo-300 text-xs flex items-center justify-between">
            <span>Baseline Model Anomalies</span>
            <span className="text-[11px] font-mono text-slate-500">({baselinePatterns.length})</span>
          </h4>
          {baselinePatterns.length === 0 ? (
            <div className="bg-background/40 border border-surfaceBorder rounded-lg p-3 text-slate-500 text-xs italic">
              No anomalies detected for baseline model.
            </div>
          ) : (
            <div className="space-y-2">
              {baselinePatterns.map((ap, idx) => (
                <div
                  key={idx}
                  className={`p-3 rounded-lg border flex items-start gap-2.5 ${
                    ap.severity === "CRITICAL"
                      ? "bg-rose-950/30 border-rose-900/60 text-rose-200"
                      : "bg-amber-950/30 border-amber-900/60 text-amber-200"
                  }`}
                >
                  <div className="mt-0.5 flex-shrink-0">{getIcon(ap.kind)}</div>
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-[11px] uppercase tracking-wide">
                        {ap.kind.replace("_", " ")}
                      </span>
                      <span className="text-[10px] font-mono px-1.5 py-0.2 rounded bg-black/40 border border-white/10">
                        Turn {ap.turn_index}
                      </span>
                    </div>
                    <p className="text-slate-300 text-[11px] leading-relaxed">
                      {ap.details}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
