package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexandrmotologa/agentdiff/internal/ast"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// Workspace represents an isolated environment for an agent task run.
type Workspace struct {
	ID         string
	TaskID     string
	SourceDir  string
	WorkDir    string
	IsWorktree bool
	BranchName string
	InitialFiles map[string]string // relative path -> initial content
}

// NewWorkspace initializes an isolated workspace by copying or creating a git worktree.
func NewWorkspace(taskID, runID, sourceDir string) (*Workspace, error) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("agentdiff_%s_%s_*", taskID, runID))
	if err != nil {
		return nil, fmt.Errorf("creating temp workspace dir: %w", err)
	}

	ws := &Workspace{
		ID:           runID,
		TaskID:       taskID,
		SourceDir:    sourceDir,
		WorkDir:      tmpDir,
		InitialFiles: make(map[string]string),
	}

	// Copy all files from sourceDir into WorkDir
	if err := copyDir(sourceDir, tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("copying source dir %q to %q: %w", sourceDir, tmpDir, err)
	}

	// Record initial files and contents for diff calculations
	if err := ws.snapshotInitialFiles(); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("snapshotting initial files: %w", err)
	}

	return ws, nil
}

// Cleanup removes the workspace directory.
func (w *Workspace) Cleanup() error {
	if w.WorkDir != "" {
		return os.RemoveAll(w.WorkDir)
	}
	return nil
}

// ReadFile reads the contents of a file relative to the workspace root.
func (w *Workspace) ReadFile(relPath string) (string, error) {
	cleanPath := filepath.Clean(relPath)
	fullPath := filepath.Join(w.WorkDir, cleanPath)
	if !strings.HasPrefix(fullPath, filepath.Clean(w.WorkDir)) {
		return "", fmt.Errorf("path traversal forbidden: %s", relPath)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes full content to a file in the workspace.
func (w *Workspace) WriteFile(relPath, content string) error {
	cleanPath := filepath.Clean(relPath)
	fullPath := filepath.Join(w.WorkDir, cleanPath)
	if !strings.HasPrefix(fullPath, filepath.Clean(w.WorkDir)) {
		return fmt.Errorf("path traversal forbidden: %s", relPath)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// EditFile performs an exact string replacement in a file.
func (w *Workspace) EditFile(relPath, target, replacement string) error {
	content, err := w.ReadFile(relPath)
	if err != nil {
		return err
	}

	if !strings.Contains(content, target) {
		return fmt.Errorf("target string not found in %s", relPath)
	}

	count := strings.Count(content, target)
	if count > 1 {
		return fmt.Errorf("target string is ambiguous (matches %d times) in %s", count, relPath)
	}

	newContent := strings.Replace(content, target, replacement, 1)
	return w.WriteFile(relPath, newContent)
}

// ListDir returns directory entries relative to the workspace.
func (w *Workspace) ListDir(relPath string) ([]string, error) {
	targetDir := w.WorkDir
	if relPath != "" && relPath != "." {
		targetDir = filepath.Join(w.WorkDir, filepath.Clean(relPath))
	}
	if !strings.HasPrefix(targetDir, filepath.Clean(w.WorkDir)) {
		return nil, fmt.Errorf("path traversal forbidden: %s", relPath)
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		results = append(results, name)
	}
	return results, nil
}

// ExecCommand executes a command inside the workspace directory.
func (w *Workspace) ExecCommand(cmdStr string, timeout time.Duration) (stdout, stderr string, exitCode int, durationMs int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	var cmd *exec.Cmd

	// On Windows, execute via powershell / cmd; on Unix via sh -c
	if isWindows() {
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}
	cmd.Dir = w.WorkDir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	execErr := cmd.Run()
	durationMs = time.Since(start).Milliseconds()

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if execErr != nil {
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
		err = execErr
	} else {
		exitCode = 0
	}

	return stdout, stderr, exitCode, durationMs, err
}

// ComputeDiff calculates AST metrics and file diffs compared to initial files.
func (w *Workspace) ComputeDiff() (string, []trajectory.FileMutation, trajectory.ASTMetrics, error) {
	var finalPatchBuilder strings.Builder
	var mutations []trajectory.FileMutation

	totalLinesDiff := 0
	astNodesModified := 0
	whitespaceLines := 0
	commentLines := 0

	// Scan all current files in the workspace
	currentFiles := make(map[string]string)
	err := filepath.Walk(w.WorkDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(w.WorkDir, path)
		if err != nil {
			return nil
		}
		// Ignore hidden files and git files
		if strings.HasPrefix(rel, ".git") || strings.HasPrefix(filepath.Base(rel), ".") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err == nil {
			currentFiles[filepath.ToSlash(rel)] = string(data)
		}
		return nil
	})
	if err != nil {
		return "", nil, trajectory.ASTMetrics{}, err
	}

	// Compare modified and added files
	for relPath, currentContent := range currentFiles {
		initialContent, existed := w.InitialFiles[relPath]
		if !existed {
			// Newly added file
			summary, _ := ast.CompareSourceFiles(relPath, "", currentContent)
			lines := strings.Split(currentContent, "\n")
			mut := trajectory.FileMutation{
				Path:             relPath,
				ChangeType:       "ADDED",
				LinesAdded:       len(lines),
				LinesRemoved:     0,
				Diff:             fmt.Sprintf("+++ %s (new file, %d lines)", relPath, len(lines)),
				ASTNodesAffected: summary.AffectedNodes,
			}
			mutations = append(mutations, mut)
			finalPatchBuilder.WriteString(mut.Diff + "\n")
			totalLinesDiff += summary.TotalLinesDiff
			astNodesModified += summary.ASTNodesModified
			whitespaceLines += summary.WhitespaceLines
			commentLines += summary.CommentLines
		} else if initialContent != currentContent {
			// Modified file
			summary, _ := ast.CompareSourceFiles(relPath, initialContent, currentContent)
			diffResult := ast.ComputeLineDiff(initialContent, currentContent)
			diffText := formatDiff(relPath, diffResult)

			mut := trajectory.FileMutation{
				Path:             relPath,
				ChangeType:       "MODIFIED",
				LinesAdded:       summary.LinesAdded,
				LinesRemoved:     summary.LinesRemoved,
				Diff:             diffText,
				ASTNodesAffected: summary.AffectedNodes,
			}
			mutations = append(mutations, mut)
			finalPatchBuilder.WriteString(diffText + "\n")
			totalLinesDiff += summary.TotalLinesDiff
			astNodesModified += summary.ASTNodesModified
			whitespaceLines += summary.WhitespaceLines
			commentLines += summary.CommentLines
		}
	}

	// Check deleted files
	for relPath, initialContent := range w.InitialFiles {
		if _, exists := currentFiles[relPath]; !exists {
			lines := strings.Split(initialContent, "\n")
			mut := trajectory.FileMutation{
				Path:         relPath,
				ChangeType:   "DELETED",
				LinesAdded:   0,
				LinesRemoved: len(lines),
				Diff:         fmt.Sprintf("--- %s (deleted file, %d lines)", relPath, len(lines)),
			}
			mutations = append(mutations, mut)
			finalPatchBuilder.WriteString(mut.Diff + "\n")
			totalLinesDiff += len(lines)
		}
	}

	surgScore, cosmeticRatio := ast.CalculateSurgicalScore(totalLinesDiff, astNodesModified, whitespaceLines, commentLines)
	metrics := trajectory.ASTMetrics{
		TotalLinesDiff:     totalLinesDiff,
		ASTNodesModified:   astNodesModified,
		SurgicalScore:      surgScore,
		CosmeticChurnRatio: cosmeticRatio,
		WhitespaceLines:    whitespaceLines,
		CommentLines:       commentLines,
	}

	return finalPatchBuilder.String(), mutations, metrics, nil
}

func (w *Workspace) snapshotInitialFiles() error {
	return filepath.Walk(w.WorkDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(w.WorkDir, path)
		if err != nil {
			return nil
		}
		if strings.HasPrefix(rel, ".git") || strings.HasPrefix(filepath.Base(rel), ".") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err == nil {
			w.InitialFiles[filepath.ToSlash(rel)] = string(data)
		}
		return nil
	})
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func formatDiff(path string, diff ast.DiffResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", path, path))
	addedMap := make(map[int]bool)
	for _, idx := range diff.AddedIndices {
		addedMap[idx] = true
	}
	removedMap := make(map[int]bool)
	for _, idx := range diff.RemovedIndices {
		removedMap[idx] = true
	}

	for i, l := range diff.OriginalLines {
		if removedMap[i] {
			sb.WriteString("-" + l + "\n")
		}
	}
	for j, l := range diff.ModifiedLines {
		if addedMap[j] {
			sb.WriteString("+" + l + "\n")
		}
	}
	return sb.String()
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
