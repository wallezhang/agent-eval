// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func init() {
	Register("command", func(config map[string]any) (Agent, error) {
		return newCommandAgent(config)
	})
}

// commandAgent executes an external command as the agent.
// The task prompt is passed via stdin by default. If any arg contains
// {{.Prompt}}, the prompt is substituted into the args instead and
// stdin is not used.
type commandAgent struct {
	command           string
	args              []string
	env               []string
	timeout           time.Duration
	hasPromptTemplate bool
}

func newCommandAgent(config map[string]any) (*commandAgent, error) {
	a := &commandAgent{
		timeout: 60 * time.Second,
	}

	if cmd, ok := config["command"].(string); ok {
		a.command = cmd
	} else {
		return nil, fmt.Errorf("command agent: command is required")
	}

	if args, ok := config["args"].([]any); ok {
		for _, arg := range args {
			a.args = append(a.args, fmt.Sprintf("%v", arg))
		}
	}

	for _, arg := range a.args {
		if strings.Contains(arg, "{{.Prompt}}") {
			a.hasPromptTemplate = true
			break
		}
	}

	if env, ok := config["env"].(map[string]any); ok {
		for k, v := range env {
			a.env = append(a.env, fmt.Sprintf("%s=%v", k, v))
		}
	}

	if t, ok := config["timeout"].(string); ok {
		if d, err := time.ParseDuration(t); err == nil {
			a.timeout = d
		}
	}

	return a, nil
}

func (a *commandAgent) Execute(ctx context.Context, input model.TaskInput) (*model.AgentOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	args := a.args
	if a.hasPromptTemplate {
		args = make([]string, len(a.args))
		for i, arg := range a.args {
			args[i] = strings.ReplaceAll(arg, "{{.Prompt}}", input.Prompt)
		}
	}

	cmd := exec.CommandContext(ctx, a.command, args...)

	if !a.hasPromptTemplate {
		cmd.Stdin = strings.NewReader(input.Prompt)
	}

	if len(a.env) > 0 {
		cmd.Env = append(cmd.Environ(), a.env...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
	}

	return &model.AgentOutput{
		Text: strings.TrimSpace(stdout.String()),
		Metadata: map[string]any{
			"stderr":    stderr.String(),
			"exit_code": 0,
		},
	}, nil
}

func (a *commandAgent) Close() error {
	return nil
}
