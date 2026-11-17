package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alexandrmotologa/agentdiff/internal/config"
	"github.com/alexandrmotologa/agentdiff/internal/metrics"
	"github.com/alexandrmotologa/agentdiff/internal/runner"
	"github.com/alexandrmotologa/agentdiff/internal/server"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "bench":
		runBench(os.Args[2:])
	case "compare":
		runCompare(os.Args[2:])
	case "replay":
		runReplay(os.Args[2:])
	case "studio":
		runStudio(os.Args[2:])
	case "export":
		runExport(os.Args[2:])
	case "eval":
		runEval(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Println("agentdiff version 1.0.0 (Go runtime: " + runtime.Version() + ")")
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q. Run 'agentdiff help' for usage.\n", command)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`agentdiff: Comparative Benchmarking & Trajectory Regression Platform for AI Coding Agents

Usage:
  agentdiff <command> [flags]

Available Commands:
  bench      Execute comparative benchmark suite across Candidate and Baseline models
  compare    Compare two .trajectory.json traces directly and evaluate regressions
  replay     Play back an agent trajectory trace turn-by-turn in the terminal
  studio     Launch the embedded React 19 Trajectory Studio web workbench (HTTP :7676)
  export     Generate a self-contained offline HTML interactive report with all traces
  eval       Evaluate a single trajectory against regression quality gates
  version    Show version and build information

Use "agentdiff <command> --help" for more information about a command.`)
}

func runBench(args []string) {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	configPath := fs.String("config", "examples/agentdiff.yaml", "Path to benchmark configuration file")
	outputDir := fs.String("output-dir", "./eval_results", "Directory to write serialized trajectory traces and reports")
	failOnRegression := fs.Bool("fail-on-regression", false, "Exit with code 1 if any task fails or exhibits regression")
	githubSummary := fs.Bool("github-summary", false, "Append summary table to $GITHUB_STEP_SUMMARY environment file")
	jsonOutput := fs.String("json-output", "", "Optional path to export raw comparison JSON dataset")
	fs.Parse(args)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Starting agentdiff Benchmark Suite: %s\n", cfg.Name)
	fmt.Printf("Candidate: %s (%s) vs Baseline: %s (%s)\n",
		cfg.Candidate.Model, cfg.Candidate.Provider, cfg.Baseline.Model, cfg.Baseline.Provider)
	fmt.Printf("Tasks: %d | Concurrency: %d | Output: %s\n\n", len(cfg.Tasks), cfg.Concurrency, *outputDir)

	progressCb := func(taskID string, status string, comp *trajectory.ComparisonResult) {
		if comp != nil {
			fmt.Printf(" -> [%s] %s (Bloat: %.2fx, Surgical Δ: %+.3f, Verdict: %s)\n",
				taskID, status, comp.ToolBloatFactor, comp.SurgicalScoreDelta, comp.Verdict)
		} else {
			fmt.Printf(" -> [%s] %s\n", taskID, status)
		}
	}

	results, err := runner.RunBenchmarkSuite(cfg, *outputDir, progressCb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark suite completed with errors: %v\n", err)
	}

	fmt.Println()
	fmt.Println(metrics.FormatSuiteTerminal(cfg.Name, results))

	// Markdown report
	suiteMarkdown := metrics.FormatSuiteMarkdown(cfg.Name, results)
	_ = os.WriteFile(filepath.Join(*outputDir, "SUMMARY.md"), []byte(suiteMarkdown), 0644)

	// GitHub Actions Step Summary
	if *githubSummary {
		if summaryFile := os.Getenv("GITHUB_STEP_SUMMARY"); summaryFile != "" {
			f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err == nil {
				_, _ = f.WriteString("\n" + suiteMarkdown + "\n")
				_ = f.Close()
			}
		}
	}

	if *jsonOutput != "" {
		data, _ := json.MarshalIndent(results, "", "  ")
		_ = os.WriteFile(*jsonOutput, data, 0644)
	}

	hasRegression := false
	for _, r := range results {
		if r != nil && (r.Verdict == "REGRESSION" || !r.CandidateVerified) {
			hasRegression = true
			break
		}
	}

	if *failOnRegression && hasRegression {
		fmt.Println("CI Quality Gate Failed: Operational or trajectory regressions detected.")
		os.Exit(1)
	}
}

func runCompare(args []string) {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	candPath := fs.String("candidate", "", "Path to candidate .trajectory.json")
	basePath := fs.String("baseline", "", "Path to baseline .trajectory.json")
	targets := fs.String("target-files", "", "Comma-separated target files")
	mdReport := fs.String("markdown-report", "", "Optional path to write markdown summary")
	failOnRegression := fs.Bool("fail-on-regression", false, "Exit with code 1 if regression detected")
	fs.Parse(args)

	if *candPath == "" || *basePath == "" {
		fmt.Fprintln(os.Stderr, "Both --candidate and --baseline trajectory paths must be specified.")
		os.Exit(2)
	}

	candTraj, err := trajectory.LoadTrajectory(*candPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading candidate trajectory: %v\n", err)
		os.Exit(2)
	}

	baseTraj, err := trajectory.LoadTrajectory(*basePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading baseline trajectory: %v\n", err)
		os.Exit(2)
	}

	var targetList []string
	if *targets != "" {
		for _, t := range strings.Split(*targets, ",") {
			targetList = append(targetList, strings.TrimSpace(t))
		}
	}

	gates := config.DefaultGates()
	comp := metrics.CompareTrajectories(candTraj, baseTraj, targetList, gates)

	fmt.Print(metrics.FormatTerminalReport(comp))

	if *mdReport != "" {
		md := metrics.FormatMarkdownSummary(comp)
		_ = os.WriteFile(*mdReport, []byte(md), 0644)
	}

	if *failOnRegression && comp.Verdict == "REGRESSION" {
		os.Exit(1)
	}
}

func runReplay(args []string) {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	tracePath := fs.String("trace", "", "Path to .trajectory.json trace file")
	delayMs := fs.Int("delay", 500, "Millisecond delay between turn playback")
	interactive := fs.Bool("interactive", false, "Interactive turn-by-turn stepping with keyboard")
	fs.BoolVar(interactive, "i", false, "Shorthand for --interactive")
	fs.Parse(args)

	if *tracePath == "" {
		fmt.Fprintln(os.Stderr, "Path to --trace file is required.")
		os.Exit(2)
	}

	traj, err := trajectory.LoadTrajectory(*tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading trajectory: %v\n", err)
		os.Exit(2)
	}

	if *interactive {
		runInteractiveReplay(traj)
		return
	}

	fmt.Println("================================================================================")
	fmt.Printf(" REPLAYING TRAJECTORY: %s\n", traj.TaskID)
	fmt.Printf(" Model: %s | Duration: %dms | Status: %s | Verified: %t\n",
		traj.Model, traj.DurationMs, traj.Status, traj.Verified)
	fmt.Println("================================================================================")

	for _, turn := range traj.Turns {
		fmt.Printf("\n--- [TURN %d] (Duration: %dms) ---\n", turn.TurnIndex, turn.DurationMs)
		if turn.Thought != "" {
			fmt.Printf("Reasoning: %s\n", turn.Thought)
		}
		for _, call := range turn.ToolCalls {
			statusIcon := "[OK]"
			if call.Status != "SUCCESS" {
				statusIcon = "[FAIL]"
			}
			fmt.Printf("Tool: %s %s (Target: %s, %dms)\n",
				statusIcon, call.ToolName, call.NormalizedTarget, call.DurationMs)
		}
		for _, exec := range turn.TerminalExecutions {
			fmt.Printf("Bash: $ %s (Exit: %d)\n", exec.Command, exec.ExitCode)
		}
		for _, mut := range turn.FileMutations {
			fmt.Printf("Diff: %s (+%d, -%d)\n", mut.Path, mut.LinesAdded, mut.LinesRemoved)
		}
		time.Sleep(time.Duration(*delayMs) * time.Millisecond)
	}

	fmt.Println("\n================================================================================")
	fmt.Printf(" REPLAY FINISHED: Verified=%t (Surgical Score: %.3f)\n",
		traj.Verified, traj.ASTMetrics.SurgicalScore)
	fmt.Println("================================================================================")
}

func runStudio(args []string) {
	fs := flag.NewFlagSet("studio", flag.ExitOnError)
	port := fs.Int("port", 7676, "HTTP port for Trajectory Studio server")
	tracesDir := fs.String("traces-dir", "./eval_results", "Directory containing serialized benchmark traces")
	configPath := fs.String("config", "examples/agentdiff.yaml", "Optional benchmark configuration file")
	openBrowser := fs.Bool("open", false, "Automatically open Studio in the default browser")
	fs.Parse(args)

	absTracesDir, err := filepath.Abs(*tracesDir)
	if err != nil {
		absTracesDir = *tracesDir
	}

	srv := server.NewServer(absTracesDir, *configPath)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	mux.Handle("/", srv.FileServerHandler())

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)

	fmt.Println("================================================================================")
	fmt.Println(" agentdiff Trajectory Studio")
	fmt.Printf(" Local Web Workbench: %s\n", url)
	fmt.Printf(" Traces Directory   : %s\n", absTracesDir)
	fmt.Println("================================================================================")
	fmt.Println("Press Ctrl+C to terminate.")

	if *openBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openURL(url)
		}()
	}

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		os.Exit(1)
	}
}

func runExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	tracesDir := fs.String("traces-dir", "./eval_results", "Directory containing serialized benchmark traces")
	outputPath := fs.String("output", "./eval_results/report.html", "Path to output standalone interactive HTML report")
	openBrowser := fs.Bool("open", false, "Automatically open generated report in the default browser")
	fs.Parse(args)

	absTracesDir, err := filepath.Abs(*tracesDir)
	if err != nil {
		absTracesDir = *tracesDir
	}

	absOut, err := filepath.Abs(*outputPath)
	if err != nil {
		absOut = *outputPath
	}

	fmt.Printf("Generating self-contained HTML report from %q...\n", absTracesDir)
	if err := server.ExportReportToFile(absTracesDir, absOut); err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Standalone interactive report exported to: %s\n", absOut)

	if *openBrowser {
		openURL(absOut)
	}
}

func runEval(args []string) {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	tracePath := fs.String("trace", "", "Path to .trajectory.json to evaluate")
	minSurgical := fs.Float64("min-surgical", 0.40, "Minimum required AST surgical score")
	fs.Parse(args)

	if *tracePath == "" {
		fmt.Fprintln(os.Stderr, "Path to --trace is required.")
		os.Exit(2)
	}

	traj, err := trajectory.LoadTrajectory(*tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading trajectory: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Evaluating Trajectory: %s (%s)\n", traj.TaskID, traj.Model)
	fmt.Printf("Verified: %t | Duration: %dms | Turns: %d\n", traj.Verified, traj.DurationMs, len(traj.Turns))
	fmt.Printf("AST Surgical Score: %.3f (Min Required: %.3f)\n", traj.ASTMetrics.SurgicalScore, *minSurgical)

	if !traj.Verified {
		fmt.Println("EVAL RESULT: FAIL (Task not verified)")
		os.Exit(1)
	}
	if traj.ASTMetrics.SurgicalScore < *minSurgical {
		fmt.Printf("EVAL RESULT: FAIL (Surgical precision %.3f < %.3f)\n", traj.ASTMetrics.SurgicalScore, *minSurgical)
		os.Exit(1)
	}

	fmt.Println("EVAL RESULT: PASS")
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
