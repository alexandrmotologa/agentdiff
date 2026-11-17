package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// PreloadedReportData is the data package embedded directly into standalone HTML reports.
type PreloadedReportData struct {
	Tasks        []TaskListItem                                  `json:"tasks"`
	Comparisons  map[string]*trajectory.ComparisonResult         `json:"comparisons"`
	Trajectories map[string]map[string]*trajectory.Trajectory    `json:"trajectories"` // task_id -> "candidate"|"baseline" -> Trajectory
}

// GenerateStandaloneHTML bundles the React 19 application and all benchmark data into a single file.
func GenerateStandaloneHTML(tracesDir string) (string, error) {
	// 1. Gather all tasks, comparisons, and trajectories from tracesDir
	data := PreloadedReportData{
		Tasks:        make([]TaskListItem, 0),
		Comparisons:  make(map[string]*trajectory.ComparisonResult),
		Trajectories: make(map[string]map[string]*trajectory.Trajectory),
	}

	entries, err := os.ReadDir(tracesDir)
	if err != nil {
		return "", fmt.Errorf("reading traces dir %q: %w", tracesDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		taskID := entry.Name()
		taskDir := filepath.Join(tracesDir, taskID)

		// Load comparison
		compFile := filepath.Join(taskDir, "comparison.json")
		comp, err := trajectory.LoadComparison(compFile)
		if err == nil && comp != nil {
			data.Comparisons[taskID] = comp
			data.Tasks = append(data.Tasks, TaskListItem{
				TaskID:                 comp.TaskID,
				Verdict:                comp.Verdict,
				CandidateModel:         comp.CandidateModel,
				BaselineModel:          comp.BaselineModel,
				CandidateVerified:      comp.CandidateVerified,
				BaselineVerified:       comp.BaselineVerified,
				ToolBloatFactor:        comp.ToolBloatFactor,
				SurgicalScoreDelta:     comp.SurgicalScoreDelta,
				SpeedupFactor:          comp.SpeedupFactor,
				TrajectoryEditDistance: comp.TrajectoryEditDistance,
				RegressionCount:        len(comp.Regressions),
			})
		}

		// Load trajectories
		candFile := filepath.Join(taskDir, "candidate.trajectory.json")
		candTraj, _ := trajectory.LoadTrajectory(candFile)

		baseFile := filepath.Join(taskDir, "baseline.trajectory.json")
		baseTraj, _ := trajectory.LoadTrajectory(baseFile)

		if candTraj != nil || baseTraj != nil {
			data.Trajectories[taskID] = make(map[string]*trajectory.Trajectory)
			if candTraj != nil {
				data.Trajectories[taskID]["candidate"] = candTraj
			}
			if baseTraj != nil {
				data.Trajectories[taskID]["baseline"] = baseTraj
			}
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshaling preloaded data: %w", err)
	}

	// 2. Read embedded assets
	distFS, err := fs.Sub(embeddedFS, "dist")
	if err != nil {
		return "", fmt.Errorf("accessing embedded dist filesystem: %w", err)
	}

	indexBytes, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		return "", fmt.Errorf("reading index.html from embedded dist: %w", err)
	}

	// Locate and read CSS and JS files in assets
	var cssContent strings.Builder
	var jsContent strings.Builder

	_ = fs.WalkDir(distFS, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".css") {
			bytes, err := fs.ReadFile(distFS, path)
			if err == nil {
				cssContent.Write(bytes)
				cssContent.WriteString("\n")
			}
		} else if strings.HasSuffix(path, ".js") {
			bytes, err := fs.ReadFile(distFS, path)
			if err == nil {
				jsContent.Write(bytes)
				jsContent.WriteString("\n")
			}
		}
		return nil
	})

	htmlStr := string(indexBytes)

	// Replace external CSS links with inline <style>
	cssInjection := fmt.Sprintf("<style>\n%s\n</style>", cssContent.String())
	htmlStr = strings.Replace(htmlStr, `<link rel="stylesheet"`, `<!-- inlined css --><!-- `, 1)

	// Inject Preloaded Data and JS into HTML
	dataScript := fmt.Sprintf(`<script id="agentdiff-preloaded-data">window.__AGENTDIFF_DATA__ = %s;</script>`, string(jsonData))
	jsInjection := fmt.Sprintf("%s\n<script type=\"module\">\n%s\n</script>", dataScript, jsContent.String())

	// Remove script module tag from index.html and append inline JS
	if idx := strings.Index(htmlStr, `<script type="module"`); idx != -1 {
		endIdx := strings.Index(htmlStr[idx:], `</script>`)
		if endIdx != -1 {
			htmlStr = htmlStr[:idx] + htmlStr[idx+endIdx+9:]
		}
	}

	htmlStr = strings.Replace(htmlStr, "</head>", cssInjection+"\n</head>", 1)
	htmlStr = strings.Replace(htmlStr, "</body>", jsInjection+"\n</body>", 1)

	return htmlStr, nil
}

// ExportReportToFile writes a standalone report to the given destination path.
func ExportReportToFile(tracesDir, outputPath string) error {
	htmlContent, err := GenerateStandaloneHTML(tracesDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating directory for report %q: %w", outputPath, err)
	}

	return os.WriteFile(outputPath, []byte(htmlContent), 0644)
}
