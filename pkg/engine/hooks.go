// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// Hooks executes lifecycle hook commands at various stages of evaluation.
type Hooks struct {
	config model.HooksConfig
	logger *log.Logger
}

// NewHooks creates a new Hooks instance.
func NewHooks(config model.HooksConfig, logger *log.Logger) *Hooks {
	return &Hooks{config: config, logger: logger}
}

// hookContext is the data passed to hook commands via stdin as JSON.
type hookContext struct {
	Event     string         `json:"event"`
	RunID     string         `json:"run_id,omitempty"`
	SuiteName string         `json:"suite_name,omitempty"`
	TaskID    string         `json:"task_id,omitempty"`
	TaskName  string         `json:"task_name,omitempty"`
	TrialIdx  int            `json:"trial_index,omitempty"`
	Status    string         `json:"status,omitempty"`
	Score     float64        `json:"score,omitempty"`
	Pass      bool           `json:"pass,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"`
}

// BeforeRun executes the before_run hook.
func (h *Hooks) BeforeRun(ctx context.Context, suite *model.EvalSuite) error {
	if h.config.BeforeRun == "" {
		return nil
	}
	return h.run(ctx, h.config.BeforeRun, hookContext{
		Event:     "before_run",
		SuiteName: suite.Name,
	})
}

// AfterRun executes the after_run hook.
func (h *Hooks) AfterRun(ctx context.Context, run *model.EvalRun) error {
	if h.config.AfterRun == "" {
		return nil
	}
	return h.run(ctx, h.config.AfterRun, hookContext{
		Event:     "after_run",
		RunID:     run.ID,
		SuiteName: run.SuiteName,
		Extra: map[string]any{
			"pass_rate": run.Summary.OverallPassRate,
			"avg_score": run.Summary.AvgScore,
			"total":     run.Summary.TotalTrials,
			"passed":    run.Summary.PassedTrials,
		},
	})
}

// BeforeTask executes the before_task hook.
func (h *Hooks) BeforeTask(ctx context.Context, task model.Task) error {
	if h.config.BeforeTask == "" {
		return nil
	}
	return h.run(ctx, h.config.BeforeTask, hookContext{
		Event:    "before_task",
		TaskID:   task.ID,
		TaskName: task.Name,
	})
}

// AfterTask executes the after_task hook.
func (h *Hooks) AfterTask(ctx context.Context, task model.Task, result model.TaskResult) error {
	if h.config.AfterTask == "" {
		return nil
	}
	return h.run(ctx, h.config.AfterTask, hookContext{
		Event:    "after_task",
		TaskID:   task.ID,
		TaskName: task.Name,
		Score:    result.AvgScore,
		Extra: map[string]any{
			"pass_count":  result.PassCount,
			"fail_count":  result.FailCount,
			"error_count": result.ErrorCount,
		},
	})
}

// BeforeTrial executes the before_trial hook.
func (h *Hooks) BeforeTrial(ctx context.Context, task model.Task, trialIndex int) error {
	if h.config.BeforeTrial == "" {
		return nil
	}
	return h.run(ctx, h.config.BeforeTrial, hookContext{
		Event:    "before_trial",
		TaskID:   task.ID,
		TaskName: task.Name,
		TrialIdx: trialIndex,
	})
}

// AfterTrial executes the after_trial hook.
func (h *Hooks) AfterTrial(ctx context.Context, trial *model.Trial) error {
	if h.config.AfterTrial == "" {
		return nil
	}
	return h.run(ctx, h.config.AfterTrial, hookContext{
		Event:    "after_trial",
		TaskID:   trial.TaskID,
		TrialIdx: trial.Index,
		Status:   string(trial.Status),
		Score:    trial.Score,
		Pass:     trial.Pass,
	})
}

// HasAnyHook returns true if any hook is configured.
func (h *Hooks) HasAnyHook() bool {
	return h.config.BeforeRun != "" ||
		h.config.AfterRun != "" ||
		h.config.BeforeTask != "" ||
		h.config.AfterTask != "" ||
		h.config.BeforeTrial != "" ||
		h.config.AfterTrial != ""
}

func (h *Hooks) run(ctx context.Context, command string, hctx hookContext) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	payload, err := json.Marshal(hctx)
	if err != nil {
		return fmt.Errorf("marshaling hook context: %w", err)
	}
	cmd.Stdin = bytes.NewReader(payload)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	h.logger.Printf("  Hook [%s]: running %q", hctx.Event, command)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook %q failed: %w (stderr: %s)", hctx.Event, err, stderr.String())
	}

	return nil
}
