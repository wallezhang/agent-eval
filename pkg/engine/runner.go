// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/wallezhang/agent-eval/pkg/agent"
	"github.com/wallezhang/agent-eval/pkg/grader"
	"github.com/wallezhang/agent-eval/pkg/model"
)

// Runner executes a single trial: calling the agent and grading the output.
type Runner struct {
	agent      agent.Agent
	graders    []grader.Grader
	timeout    string
	maxRetries int
	retryDelay time.Duration
	logger     *log.Logger
}

// Run executes one trial for the given task.
func (r *Runner) Run(ctx context.Context, task model.Task, trialIndex int) (*model.Trial, error) {
	// Apply timeout.
	if r.timeout != "" {
		if d, err := time.ParseDuration(r.timeout); err == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d)
			defer cancel()
		}
	}

	trial := &model.Trial{
		ID:        uuid.New().String(),
		TaskID:    task.ID,
		Index:     trialIndex,
		Status:    model.TrialStatusRunning,
		StartedAt: time.Now(),
	}

	// Build transcript.
	transcript := &model.Transcript{}
	transcript.Steps = append(transcript.Steps, model.TranscriptStep{
		Type:      "input",
		Role:      "user",
		Content:   task.Input.Prompt,
		Timestamp: time.Now(),
	})

	// Execute the agent with retry logic.
	r.logger.Printf("  Task %q trial #%d: executing agent...", task.ID, trialIndex+1)

	var output *model.AgentOutput
	var agentErr error
	retries := 0
	agentStart := time.Now()

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			delay := r.retryDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			r.logger.Printf("  Task %q trial #%d: retrying (attempt %d/%d) after %v...",
				task.ID, trialIndex+1, attempt, r.maxRetries, delay)
			select {
			case <-ctx.Done():
				agentErr = ctx.Err()
				break
			case <-time.After(delay):
			}
			retries++
		}

		output, agentErr = r.agent.Execute(ctx, task.Input)
		if agentErr == nil {
			break
		}
	}

	agentDuration := time.Since(agentStart)

	if agentErr != nil {
		r.logger.Printf("  Task %q trial #%d: agent error: %v", task.ID, trialIndex+1, agentErr)
		trial.Status = model.TrialStatusError
		trial.Error = agentErr.Error()
		trial.AgentDurationMS = agentDuration.Milliseconds()
		trial.FinishedAt = time.Now()
		trial.DurationMS = trial.FinishedAt.Sub(trial.StartedAt).Milliseconds()
		trial.Transcript = transcript
		if retries > 0 {
			if trial.AgentOutput == nil {
				trial.AgentOutput = &model.AgentOutput{}
			}
			if trial.AgentOutput.Metadata == nil {
				trial.AgentOutput.Metadata = make(map[string]any)
			}
			trial.AgentOutput.Metadata["retries"] = retries
		}
		return trial, nil
	}

	trial.AgentDurationMS = agentDuration.Milliseconds()
	trial.AgentOutput = output
	if retries > 0 {
		if trial.AgentOutput.Metadata == nil {
			trial.AgentOutput.Metadata = make(map[string]any)
		}
		trial.AgentOutput.Metadata["retries"] = retries
	}
	transcript.Steps = append(transcript.Steps, model.TranscriptStep{
		Type:      "output",
		Role:      "assistant",
		Content:   output.Text,
		Timestamp: time.Now(),
	})
	trial.Transcript = transcript

	// Grade the output.
	r.logger.Printf("  Task %q trial #%d: grading with %d grader(s)...", task.ID, trialIndex+1, len(r.graders))
	grades, err := r.grade(ctx, task, *output, transcript)
	if err != nil {
		trial.Status = model.TrialStatusError
		trial.Error = fmt.Sprintf("grading failed: %v", err)
		trial.FinishedAt = time.Now()
		trial.DurationMS = trial.FinishedAt.Sub(trial.StartedAt).Milliseconds()
		return trial, nil
	}

	trial.Grades = grades
	trial.Score, trial.Pass = computeWeightedScore(grades, task.Graders)

	if trial.Pass {
		trial.Status = model.TrialStatusPassed
	} else {
		trial.Status = model.TrialStatusFailed
	}

	// Extract step count from agent metadata or transcript.
	trial.StepCount = extractStepCount(output, transcript)

	trial.FinishedAt = time.Now()
	trial.DurationMS = trial.FinishedAt.Sub(trial.StartedAt).Milliseconds()

	return trial, nil
}

func (r *Runner) grade(ctx context.Context, task model.Task, output model.AgentOutput, transcript *model.Transcript) ([]model.GradeResult, error) {
	var results []model.GradeResult
	for i, g := range r.graders {
		input := grader.GradeInput{
			Task:        task,
			AgentOutput: output,
			Transcript:  transcript,
		}

		result, err := g.Grade(ctx, input)
		if err != nil {
			results = append(results, model.GradeResult{
				GraderType: g.Type(),
				Score:      0,
				Pass:       false,
				Weight:     task.Graders[i].Weight,
				Error:      err.Error(),
			})
			continue
		}

		result.Weight = task.Graders[i].Weight
		results = append(results, *result)
	}
	return results, nil
}

// computeWeightedScore calculates the overall score from multiple grader results.
// A trial passes only if all graders pass.
func computeWeightedScore(grades []model.GradeResult, refs []model.GraderRef) (float64, bool) {
	if len(grades) == 0 {
		return 0, false
	}

	var totalWeight, weightedScore float64
	allPass := true
	for _, g := range grades {
		w := g.Weight
		if w == 0 {
			w = 1.0
		}
		totalWeight += w
		weightedScore += g.Score * w
		if !g.Pass {
			allPass = false
		}
	}

	if totalWeight == 0 {
		return 0, false
	}

	return weightedScore / totalWeight, allPass
}

// extractStepCount extracts the step count from agent metadata or transcript.
// Priority: metadata["step_count"] > metadata["steps"] > transcript step count.
func extractStepCount(output *model.AgentOutput, transcript *model.Transcript) int {
	if output != nil && output.Metadata != nil {
		if v, ok := output.Metadata["step_count"]; ok {
			if n, ok := toInt(v); ok {
				return n
			}
		}
		if v, ok := output.Metadata["steps"]; ok {
			if n, ok := toInt(v); ok {
				return n
			}
		}
	}

	// Fall back to transcript step count (excluding the initial input).
	if transcript != nil && len(transcript.Steps) > 1 {
		return len(transcript.Steps) - 1
	}

	return 0
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
