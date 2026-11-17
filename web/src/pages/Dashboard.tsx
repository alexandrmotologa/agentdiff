import React, { useState, useEffect } from "react";
import { TaskListItem, ComparisonResult, Trajectory } from "../types";
import { ScorecardHero } from "../components/ScorecardHero";
import { TaskSelector } from "../components/TaskSelector";
import { SynchronizedPlayer } from "../components/SynchronizedPlayer";
import { AstDiffViewer } from "../components/AstDiffViewer";
import { ToolCallFlamegraph } from "../components/ToolCallFlamegraph";
import { AntiPatternAlerts } from "../components/AntiPatternAlerts";
import { LeaderboardMatrix } from "../components/LeaderboardMatrix";
import { 
  GitCompare, 
  Layers, 
  Code2, 
  Clock, 
  AlertTriangle,
  RefreshCw,
  Trophy
} from "lucide-react";

export const Dashboard: React.FC = () => {
  const [tasks, setTasks] = useState<TaskListItem[]>([]);
  const [selectedTaskId, setSelectedTaskId] = useState<string>("");
  const [comparison, setComparison] = useState<ComparisonResult | null>(null);
  const [candidateTraj, setCandidateTraj] = useState<Trajectory | undefined>(undefined);
  const [baselineTraj, setBaselineTraj] = useState<Trajectory | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"player" | "ast" | "timeline" | "antipatterns" | "leaderboard">("player");
  const [isRunningBenchmark, setIsRunningBenchmark] = useState(false);
  const [liveLog, setLiveLog] = useState<string>("");

  // 1. Fetch task list on load
  const fetchTasks = async () => {
    try {
      if ((window as any).__AGENTDIFF_DATA__) {
        const preloaded = (window as any).__AGENTDIFF_DATA__;
        const taskList = preloaded.tasks || [];
        setTasks(taskList);
        if (taskList.length > 0 && !selectedTaskId) {
          setSelectedTaskId(taskList[0].task_id);
        }
        setLoading(false);
        return;
      }

      const res = await fetch("/api/tasks");
      if (res.ok) {
        const data: TaskListItem[] = await res.json();
        setTasks(data || []);
        if (data && data.length > 0 && !selectedTaskId) {
          setSelectedTaskId(data[0].task_id);
        }
      }
    } catch (e) {
      console.error("Failed to fetch tasks:", e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  // 2. Fetch comparison & trajectories when task changes
  useEffect(() => {
    if (!selectedTaskId) return;

    const loadTaskData = async () => {
      try {
        if ((window as any).__AGENTDIFF_DATA__) {
          const preloaded = (window as any).__AGENTDIFF_DATA__;
          if (preloaded.comparisons && preloaded.comparisons[selectedTaskId]) {
            setComparison(preloaded.comparisons[selectedTaskId]);
          }
          if (preloaded.trajectories && preloaded.trajectories[selectedTaskId]) {
            setCandidateTraj(preloaded.trajectories[selectedTaskId].candidate);
            setBaselineTraj(preloaded.trajectories[selectedTaskId].baseline);
          }
          return;
        }

        const compRes = await fetch(`/api/comparison?task_id=${selectedTaskId}`);
        if (compRes.ok) {
          const compData: ComparisonResult = await compRes.json();
          setComparison(compData);
        }

        const candRes = await fetch(`/api/trajectory?task_id=${selectedTaskId}&agent=candidate`);
        if (candRes.ok) {
          const candData: Trajectory = await candRes.json();
          setCandidateTraj(candData);
        }

        const baseRes = await fetch(`/api/trajectory?task_id=${selectedTaskId}&agent=baseline`);
        if (baseRes.ok) {
          const baseData: Trajectory = await baseRes.json();
          setBaselineTraj(baseData);
        }
      } catch (e) {
        console.error("Error loading task data:", e);
      }
    };

    loadTaskData();
  }, [selectedTaskId]);

  // 3. Connect to SSE Stream
  useEffect(() => {
    if ((window as any).__AGENTDIFF_DATA__) {
      // Offline standalone report mode: skip SSE
      return;
    }
    const sse = new EventSource("/api/bench/stream");

    sse.addEventListener("connected", () => {
      console.log("SSE connected to agentdiff studio");
    });

    sse.addEventListener("task_progress", (e) => {
      try {
        const payload = JSON.parse(e.data);
        setLiveLog(`Task ${payload.task_id}: ${payload.status}`);
        if (payload.status === "COMPLETED") {
          fetchTasks();
          if (payload.task_id === selectedTaskId && payload.comparison) {
            setComparison(payload.comparison);
          }
        }
      } catch (err) {
        console.error("Error parsing SSE event:", err);
      }
    });

    sse.addEventListener("bench_completed", () => {
      setIsRunningBenchmark(false);
      setLiveLog("Benchmark run completed successfully.");
      fetchTasks();
    });

    return () => {
      sse.close();
    };
  }, [selectedTaskId]);

  const handleRunBenchmark = async () => {
    setIsRunningBenchmark(true);
    setLiveLog("Launching benchmark suite...");
    try {
      await fetch("/api/bench/run", { method: "POST" });
    } catch (e) {
      console.error("Failed to run benchmark:", e);
      setIsRunningBenchmark(false);
    }
  };

  return (
    <div className="min-h-screen bg-background text-slate-100 flex flex-col">
      {/* Top Navigation Bar */}
      <header className="border-b border-surfaceBorder/80 bg-surface/70 backdrop-blur sticky top-0 z-50 px-6 py-3.5 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-sky-500/20 border border-sky-400/40 flex items-center justify-center text-sky-400 shadow-md shadow-sky-500/10">
            <GitCompare className="w-4 h-4" />
          </div>
          <div>
            <h1 className="text-base font-extrabold tracking-tight text-white flex items-center gap-2">
              <span>agentdiff</span>
              <span className="text-[11px] font-mono text-sky-400 font-semibold px-2 py-0.5 rounded bg-sky-950/50 border border-sky-800">
                v1.0.0
              </span>
            </h1>
            <p className="text-[11px] text-slate-400">
              Autonomous Agent Trajectory Regression & AST Variance Studio
            </p>
          </div>
        </div>

        <div className="flex items-center gap-4 text-xs">
          {liveLog && (
            <div className="hidden sm:flex items-center gap-2 px-3 py-1 bg-slate-900 border border-surfaceBorder rounded-full font-mono text-[11px] text-sky-300">
              <span className="w-2 h-2 rounded-full bg-sky-400 animate-pulse"></span>
              <span>{liveLog}</span>
            </div>
          )}

          <button
            onClick={fetchTasks}
            title="Refresh Benchmarks"
            className="p-2 rounded-lg bg-surface hover:bg-slate-800 border border-surfaceBorder text-slate-300 transition-colors"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </button>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 p-6 max-w-[1700px] w-full mx-auto grid grid-cols-1 xl:grid-cols-12 gap-6">
        {/* Left Sidebar: Task Selector */}
        <aside className="xl:col-span-3 min-h-[500px]">
          <TaskSelector
            tasks={tasks}
            selectedTaskId={selectedTaskId}
            onSelectTask={setSelectedTaskId}
            onRunBenchmark={handleRunBenchmark}
            isRunningBenchmark={isRunningBenchmark}
          />
        </aside>

        {/* Center / Right Content: Scorecard & Visualizers */}
        <section className="xl:col-span-9 space-y-6">
          {loading ? (
            <div className="bg-surface border border-surfaceBorder rounded-xl p-12 text-center text-slate-400 text-sm">
              Loading benchmark data...
            </div>
          ) : !comparison ? (
            <div className="bg-surface border border-surfaceBorder rounded-xl p-12 text-center text-slate-400 text-sm">
              Select a benchmark task from the left sidebar to inspect trajectory differences.
            </div>
          ) : (
            <>
              {/* Executive Scorecard */}
              <ScorecardHero comparison={comparison} />

              {/* View Switcher Tabs */}
              <div className="flex items-center gap-1 bg-surface p-1 rounded-xl border border-surfaceBorder w-fit">
                <button
                  onClick={() => setActiveTab("player")}
                  className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                    activeTab === "player"
                      ? "bg-sky-500 text-white shadow-md shadow-sky-500/20"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <Layers className="w-3.5 h-3.5" />
                  Synchronized Player
                </button>
                <button
                  onClick={() => setActiveTab("ast")}
                  className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                    activeTab === "ast"
                      ? "bg-teal-500 text-white shadow-md shadow-teal-500/20"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <Code2 className="w-3.5 h-3.5" />
                  AST Semantic Diff
                </button>
                <button
                  onClick={() => setActiveTab("timeline")}
                  className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                    activeTab === "timeline"
                      ? "bg-indigo-500 text-white shadow-md shadow-indigo-500/20"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <Clock className="w-3.5 h-3.5" />
                  Tool Flamegraph
                </button>
                <button
                  onClick={() => setActiveTab("antipatterns")}
                  className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                    activeTab === "antipatterns"
                      ? "bg-rose-500 text-white shadow-md shadow-rose-500/20"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <AlertTriangle className="w-3.5 h-3.5" />
                  Anti-Patterns
                  {(comparison.candidate_anti_patterns.length > 0 || comparison.baseline_anti_patterns.length > 0) && (
                    <span className="w-2 h-2 rounded-full bg-rose-400"></span>
                  )}
                </button>
                <button
                  onClick={() => setActiveTab("leaderboard")}
                  className={`flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                    activeTab === "leaderboard"
                      ? "bg-amber-500 text-black shadow-md shadow-amber-500/20 font-bold"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <Trophy className="w-3.5 h-3.5" />
                  Leaderboard
                </button>
              </div>

              {/* View Components */}
              {activeTab === "player" && (
                <SynchronizedPlayer
                  candidateTraj={candidateTraj}
                  baselineTraj={baselineTraj}
                  divergenceTurn={comparison.divergence_turn}
                />
              )}

              {activeTab === "ast" && (
                <AstDiffViewer
                  candidateTraj={candidateTraj}
                  baselineTraj={baselineTraj}
                />
              )}

              {activeTab === "timeline" && (
                <ToolCallFlamegraph
                  candidateTraj={candidateTraj}
                  baselineTraj={baselineTraj}
                />
              )}

              {activeTab === "antipatterns" && (
                <AntiPatternAlerts
                  candidatePatterns={comparison.candidate_anti_patterns}
                  baselinePatterns={comparison.baseline_anti_patterns}
                />
              )}

              {activeTab === "leaderboard" && (
                <LeaderboardMatrix />
              )}
            </>
          )}
        </section>
      </main>
    </div>
  );
};
