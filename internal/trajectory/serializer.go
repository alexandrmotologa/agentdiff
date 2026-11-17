package trajectory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveTrajectory writes a Trajectory to a formatted JSON file.
func SaveTrajectory(path string, traj *Trajectory) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for trajectory: %w", err)
	}

	data, err := json.MarshalIndent(traj, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling trajectory: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// LoadTrajectory reads a Trajectory from a JSON file.
func LoadTrajectory(path string) (*Trajectory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading trajectory file %q: %w", path, err)
	}

	var traj Trajectory
	if err := json.Unmarshal(data, &traj); err != nil {
		return nil, fmt.Errorf("unmarshaling trajectory from %q: %w", path, err)
	}

	return &traj, nil
}

// SaveComparison writes a ComparisonResult to a formatted JSON file.
func SaveComparison(path string, comp *ComparisonResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for comparison: %w", err)
	}

	data, err := json.MarshalIndent(comp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling comparison: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// LoadComparison reads a ComparisonResult from a JSON file.
func LoadComparison(path string) (*ComparisonResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading comparison file %q: %w", path, err)
	}

	var comp ComparisonResult
	if err := json.Unmarshal(data, &comp); err != nil {
		return nil, fmt.Errorf("unmarshaling comparison from %q: %w", path, err)
	}

	return &comp, nil
}

// AppendJSONL appends an item to a JSONL file.
func AppendJSONL(path string, record interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for jsonl: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening jsonl file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshaling jsonl record: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("writing jsonl record: %w", err)
	}

	return nil
}

// ReadJSONL parses lines of a JSONL file into a slice of type T.
func ReadJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening jsonl file %q: %w", path, err)
	}
	defer f.Close()

	var records []T
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var record T
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, fmt.Errorf("decoding jsonl line %d in %s: %w", lineNum, path, err)
		}
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning jsonl file: %w", err)
	}

	return records, nil
}
