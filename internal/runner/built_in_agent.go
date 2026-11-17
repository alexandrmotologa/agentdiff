package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/agentdiff/internal/config"
	"github.com/alexandrmotologa/agentdiff/internal/trajectory"
)

// RunBuiltInAgent runs an autonomous ReAct loop against an OpenAI-compatible or Anthropic API.
func RunBuiltInAgent(ws *Workspace, task config.TaskDefinition, modelCfg config.ModelConfig) (*trajectory.Trajectory, error) {
	rec := trajectory.NewRecorder(task.ID, modelCfg.ID, modelCfg.Model, modelCfg.SystemPrompt)

	endpoint := modelCfg.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	apiKey := modelCfg.APIKey
	if apiKey == "" && modelCfg.Provider != "ollama" {
		return nil, fmt.Errorf("API key required for provider %q (set %s or configure in YAML)", modelCfg.Provider, modelCfg.APIKeyEnv)
	}

	maxTurns := modelCfg.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 20
	}

	systemPrompt := modelCfg.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf("You are an expert autonomous software engineer. Your task is to resolve an issue in the repository. Target files: %v. Test command: %q. Use the available tools to inspect code, apply edits, and verify tests pass. Finish when verification succeeds.", task.TargetFiles, task.TestCommand)
	}

	messages := []map[string]interface{}{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": fmt.Sprintf("Task ID: %s\nTitle: %s\nDescription: %s\nTarget files: %v\nTest command: %s\nPlease fix the bug and verify that tests pass.", task.ID, task.Title, task.Description, task.TargetFiles, task.TestCommand)},
	}

	client := &http.Client{Timeout: 120 * time.Second}
	var totalPromptTokens, totalCompletionTokens int

	for turnIdx := 1; turnIdx <= maxTurns; turnIdx++ {
		// Call model endpoint
		reqBody := map[string]interface{}{
			"model":    modelCfg.Model,
			"messages": messages,
			"tools":    getAvailableTools(),
		}
		if modelCfg.Temperature != nil {
			reqBody["temperature"] = *modelCfg.Temperature
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshaling LLM request: %w", err)
		}

		req, err := http.NewRequestWithContext(context.Background(), "POST", endpoint, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("creating HTTP request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("LLM API request failed: %w", err)
		}

		respBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(respBytes))
		}

		var apiResp struct {
			Choices []struct {
				Message struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					ToolCalls []struct {
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"message"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal(respBytes, &apiResp); err != nil {
			return nil, fmt.Errorf("unmarshaling LLM response: %w", err)
		}

		totalPromptTokens += apiResp.Usage.PromptTokens
		totalCompletionTokens += apiResp.Usage.CompletionTokens

		if len(apiResp.Choices) == 0 {
			break
		}

		choice := apiResp.Choices[0].Message
		thought := choice.Content

		rec.StartTurn(thought)

		// If no tool calls, model may be finished or explaining
		if len(choice.ToolCalls) == 0 {
			rec.EndTurn()
			break
		}

		// Execute each tool call
		finished := false
		for _, tc := range choice.ToolCalls {
			toolName := tc.Function.Name
			var args map[string]interface{}
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)

			tCall := trajectory.ToolCall{
				CallID:     tc.ID,
				ToolName:   toolName,
				Arguments:  args,
				Status:     "SUCCESS",
				DurationMs: 0,
			}

			startTool := time.Now()
			output, err := executeToolOnWorkspace(ws, task, toolName, args)
			tCall.DurationMs = time.Since(startTool).Milliseconds()
			tCall.Output = output

			if err != nil {
				tCall.Status = "ERROR"
			}
			rec.RecordToolCall(tCall)

			// Append tool response to messages
			messages = append(messages, map[string]interface{}{
				"role":         "tool",
				"tool_call_id": tc.ID,
				"content":      output,
			})

			if toolName == "finish" {
				finished = true
			}
		}

		rec.EndTurn()
		if finished {
			break
		}
	}

	// Verification
	stdout, _, exitCode, _, _ := ws.ExecCommand(task.TestCommand, time.Duration(task.TimeoutSeconds)*time.Second)
	verified := (exitCode == 0)
	status := "COMPLETED"
	if !verified {
		status = "FAILED"
	}

	patch, mutations, astMetrics, _ := ws.ComputeDiff()
	rec.SetFinalPatch(patch)
	for _, m := range mutations {
		rec.RecordFileMutation(m)
	}
	rec.SetASTMetrics(astMetrics)
	rec.SetTokenUsage(trajectory.TokenUsage{
		PromptTokens:     totalPromptTokens,
		CompletionTokens: totalCompletionTokens,
		TotalTokens:      totalPromptTokens + totalCompletionTokens,
		EstimatedCostUSD: float64(totalPromptTokens)*0.000003 + float64(totalCompletionTokens)*0.000015,
	})

	return rec.Complete(status, verified, stdout), nil
}

func executeToolOnWorkspace(ws *Workspace, task config.TaskDefinition, name string, args map[string]interface{}) (string, error) {
	switch name {
	case "read_file":
		path, _ := args["path"].(string)
		return ws.ReadFile(path)
	case "write_file":
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		err := ws.WriteFile(path, content)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
	case "edit_file":
		path, _ := args["path"].(string)
		target, _ := args["target"].(string)
		replacement, _ := args["replacement"].(string)
		err := ws.EditFile(path, target, replacement)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Successfully edited %s", path), nil
	case "list_dir":
		path, _ := args["path"].(string)
		entries, err := ws.ListDir(path)
		if err != nil {
			return "", err
		}
		return strings.Join(entries, "\n"), nil
	case "exec_bash":
		cmd, _ := args["command"].(string)
		stdout, stderr, exitCode, durMs, err := ws.ExecCommand(cmd, 60*time.Second)
		res := fmt.Sprintf("Exit Code: %d (Duration: %dms)\nStdout:\n%s\nStderr:\n%s", exitCode, durMs, stdout, stderr)
		return res, err
	case "run_tests":
		cmd := task.TestCommand
		if c, ok := args["command"].(string); ok && c != "" {
			cmd = c
		}
		stdout, stderr, exitCode, durMs, _ := ws.ExecCommand(cmd, time.Duration(task.TimeoutSeconds)*time.Second)
		res := fmt.Sprintf("Exit Code: %d (Duration: %dms)\nOutput:\n%s\n%s", exitCode, durMs, stdout, stderr)
		return res, nil
	case "finish":
		summary, _ := args["summary"].(string)
		return fmt.Sprintf("Task concluded: %s", summary), nil
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func getAvailableTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_file",
				"description": "Read full contents of a file in the workspace.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{"type": "string", "description": "Relative path to file"},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "edit_file",
				"description": "Perform an exact replacement in a file.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":        map[string]interface{}{"type": "string"},
						"target":      map[string]interface{}{"type": "string", "description": "Exact text to replace"},
						"replacement": map[string]interface{}{"type": "string", "description": "New replacement text"},
					},
					"required": []string{"path", "target", "replacement"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "write_file",
				"description": "Write entire contents to a file.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":    map[string]interface{}{"type": "string"},
						"content": map[string]interface{}{"type": "string"},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_dir",
				"description": "List files and subdirectories.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{"type": "string", "description": "Directory relative path, default '.'"},
					},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "exec_bash",
				"description": "Execute a shell command inside workspace.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{"type": "string"},
					},
					"required": []string{"command"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "run_tests",
				"description": "Execute the task verification test suite.",
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "finish",
				"description": "Signal that the problem has been solved.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"summary": map[string]interface{}{"type": "string", "description": "Summary of resolution"},
					},
					"required": []string{"summary"},
				},
			},
		},
	}
}
