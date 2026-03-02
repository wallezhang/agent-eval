// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func init() {
	Register("command", func(config map[string]any) (Grader, error) {
		return newCommandGrader(config)
	})
}

// commandGrader runs an external command to evaluate the agent output.
// The agent output is passed via stdin as JSON.
// The command should exit 0 for pass, non-zero for fail.
// Optionally, stdout can contain a JSON object with "score", "pass", and "reason" fields.
type commandGrader struct {
	command string
	args    []string
	timeout time.Duration
}

func newCommandGrader(config map[string]any) (*commandGrader, error) {
	g := &commandGrader{
		timeout: 60 * time.Second,
	}

	if cmd, ok := config["command"].(string); ok {
		g.command = cmd
	} else {
		return nil, fmt.Errorf("command grader: command is required")
	}

	if args, ok := config["args"].([]any); ok {
		for _, arg := range args {
			g.args = append(g.args, fmt.Sprintf("%v", arg))
		}
	}

	if t, ok := config["timeout"].(string); ok {
		if d, err := time.ParseDuration(t); err == nil {
			g.timeout = d
		}
	}

	return g, nil
}

func (g *commandGrader) Type() string { return "command" }

func (g *commandGrader) Grade(ctx context.Context, input GradeInput) (*model.GradeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	// Prepare input as JSON for the command.
	cmdInput := map[string]any{
		"task_id":      input.Task.ID,
		"agent_output": input.AgentOutput.Text,
		"expected":     input.Task.Expected,
	}
	inputJSON, err := json.Marshal(cmdInput)
	if err != nil {
		return nil, fmt.Errorf("marshaling command input: %w", err)
	}

	cmd := exec.CommandContext(ctx, g.command, g.args...)
	cmd.Stdin = bytes.NewReader(inputJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Try to parse stdout as a grade result.
	var result struct {
		Score  float64 `json:"score"`
		Pass   bool    `json:"pass"`
		Reason string  `json:"reason"`
	}

	if jsonErr := json.Unmarshal(stdout.Bytes(), &result); jsonErr == nil {
		return &model.GradeResult{
			GraderType: g.Type(),
			Score:      result.Score,
			Pass:       result.Pass,
			Reason:     result.Reason,
		}, nil
	}

	// Fall back to exit code.
	if err != nil {
		return &model.GradeResult{
			GraderType: g.Type(),
			Score:      0,
			Pass:       false,
			Reason:     fmt.Sprintf("command failed: %v\nstderr: %s", err, strings.TrimSpace(stderr.String())),
		}, nil
	}

	return &model.GradeResult{
		GraderType: g.Type(),
		Score:      1.0,
		Pass:       true,
		Reason:     "command exited 0",
	}, nil
}
