// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/wallezhang/agent-eval/pkg/agent"
	"github.com/wallezhang/agent-eval/pkg/cache"
	"github.com/wallezhang/agent-eval/pkg/grader"
	"github.com/wallezhang/agent-eval/pkg/model"
)

// Engine orchestrates the evaluation process.
type Engine struct {
	suite       *model.EvalSuite
	agent       agent.Agent
	logger      *log.Logger
	tags        []string
	excludeTags []string
	checkpoint  checkpointStore
	resumeRunID string
}

// checkpointStore is the subset of storage.CheckpointStore used by the engine.
type checkpointStore interface {
	SaveCheckpoint(ctx context.Context, runID string, trial *model.Trial) error
	LoadCheckpoint(ctx context.Context, runID string) ([]*model.Trial, error)
	DeleteCheckpoint(ctx context.Context, runID string) error
}

// New creates a new evaluation engine.
func New(suite *model.EvalSuite, logger *log.Logger) *Engine {
	if logger == nil {
		logger = log.Default()
	}
	return &Engine{
		suite:  suite,
		logger: logger,
	}
}

// SetTagFilters sets the tag inclusion and exclusion filters.
func (e *Engine) SetTagFilters(tags, excludeTags []string) {
	e.tags = tags
	e.excludeTags = excludeTags
}

// SetCheckpointStore sets the checkpoint store for resume support.
func (e *Engine) SetCheckpointStore(store checkpointStore, resumeRunID string) {
	e.checkpoint = store
	e.resumeRunID = resumeRunID
}

// Execute runs the full evaluation and returns the results.
func (e *Engine) Execute(ctx context.Context) (*model.EvalRun, error) {
	startTime := time.Now()

	// Set up lifecycle hooks.
	hooks := NewHooks(e.suite.Hooks, e.logger)

	// Run before_run hook.
	if err := hooks.BeforeRun(ctx, e.suite); err != nil {
		e.logger.Printf("Warning: before_run hook failed: %v", err)
	}

	// Create the agent under test.
	ag, err := agent.Create(e.suite.Agent.Type, e.suite.Agent.Config)
	if err != nil {
		return nil, fmt.Errorf("creating agent: %w", err)
	}
	defer ag.Close()

	// Wrap agent with caching if configured.
	var finalAgent agent.Agent = ag
	if e.suite.Cache.Enabled {
		cacheDir := e.suite.Cache.Dir
		if cacheDir == "" {
			cacheDir = ".cache"
		}
		cacheTTL := 24 * time.Hour
		if e.suite.Cache.TTL != "" {
			if d, err := time.ParseDuration(e.suite.Cache.TTL); err == nil {
				cacheTTL = d
			}
		}
		finalAgent = cache.Wrap(ag, e.suite.Agent.Type, e.suite.Agent.Config, cacheDir, cacheTTL)
		e.logger.Printf("Response caching enabled (dir=%s, ttl=%v)", cacheDir, cacheTTL)
	}
	e.agent = finalAgent

	// Filter tasks by tags if specified.
	e.filterTasksByTags()

	if len(e.suite.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks match the specified tag filters")
	}

	// Build work items (task × trials).
	workItems := e.buildWorkItems()

	// Handle resume: load checkpoint and filter out completed work items.
	var resumedTrials []*model.Trial
	runID := uuid.New().String()
	if e.resumeRunID != "" && e.checkpoint != nil {
		runID = e.resumeRunID
		var err error
		resumedTrials, err = e.checkpoint.LoadCheckpoint(ctx, e.resumeRunID)
		if err != nil {
			e.logger.Printf("Warning: failed to load checkpoint: %v", err)
		} else if len(resumedTrials) > 0 {
			completed := make(map[string]bool)
			for _, t := range resumedTrials {
				key := fmt.Sprintf("%s:%d", t.TaskID, t.Index)
				completed[key] = true
			}
			var remaining []workItem
			for _, item := range workItems {
				key := fmt.Sprintf("%s:%d", item.task.ID, item.trialIndex)
				if !completed[key] {
					remaining = append(remaining, item)
				}
			}
			e.logger.Printf("Resuming run %s: %d/%d trials already completed, %d remaining",
				e.resumeRunID[:8], len(resumedTrials), len(workItems), len(remaining))
			workItems = remaining
		}
	}

	e.logger.Printf("Starting evaluation %q: %d tasks, %d total trials",
		e.suite.Name, len(e.suite.Tasks), len(workItems))

	// Resolve graders for each task.
	taskGraders, err := e.resolveGraders()
	if err != nil {
		return nil, fmt.Errorf("resolving graders: %w", err)
	}

	// Parse retry delay.
	retryDelay := time.Second // default 1s
	if e.suite.Execution.RetryDelay != "" {
		if d, err := time.ParseDuration(e.suite.Execution.RetryDelay); err == nil {
			retryDelay = d
		}
	}

	// Execute trials using the scheduler.
	sched := newScheduler(e.suite.Execution.Concurrency, e.suite.Execution.RateLimitRPS, e.logger)
	trials, err := sched.Run(ctx, workItems, func(ctx context.Context, item workItem) (*model.Trial, error) {
		// Run before_trial hook.
		if err := hooks.BeforeTrial(ctx, item.task, item.trialIndex); err != nil {
			e.logger.Printf("Warning: before_trial hook failed: %v", err)
		}

		runner := &Runner{
			agent:      e.agent,
			graders:    taskGraders[item.task.ID],
			timeout:    e.suite.Execution.Timeout,
			maxRetries: e.suite.Execution.MaxRetries,
			retryDelay: retryDelay,
			logger:     e.logger,
		}
		trial, err := runner.Run(ctx, item.task, item.trialIndex)

		// Run after_trial hook.
		if trial != nil {
			if hookErr := hooks.AfterTrial(ctx, trial); hookErr != nil {
				e.logger.Printf("Warning: after_trial hook failed: %v", hookErr)
			}
			// Save checkpoint.
			if e.checkpoint != nil {
				if cpErr := e.checkpoint.SaveCheckpoint(ctx, runID, trial); cpErr != nil {
					e.logger.Printf("Warning: failed to save checkpoint: %v", cpErr)
				}
			}
		}

		return trial, err
	})
	if err != nil {
		return nil, fmt.Errorf("executing trials: %w", err)
	}

	// Merge resumed trials with newly completed ones.
	if len(resumedTrials) > 0 {
		allTrials := make([]*model.Trial, 0, len(resumedTrials)+len(trials))
		allTrials = append(allTrials, resumedTrials...)
		allTrials = append(allTrials, trials...)
		trials = allTrials
	}

	// Aggregate results by task.
	taskResults := e.aggregateResults(trials)

	// Compute summary.
	summary := e.computeSummary(taskResults)

	finishTime := time.Now()

	run := &model.EvalRun{
		ID:          runID,
		SuiteName:   e.suite.Name,
		Description: e.suite.Description,
		AgentType:   e.suite.Agent.Type,
		AgentConfig: e.suite.Agent.Config,
		TaskResults: taskResults,
		Summary:     summary,
		StartedAt:   startTime,
		FinishedAt:  finishTime,
		DurationMS:  finishTime.Sub(startTime).Milliseconds(),
	}

	e.logger.Printf("Evaluation complete: %d/%d trials passed (%.1f%%)",
		summary.PassedTrials, summary.TotalTrials, summary.OverallPassRate*100)

	// Run after_run hook.
	if err := hooks.AfterRun(ctx, run); err != nil {
		e.logger.Printf("Warning: after_run hook failed: %v", err)
	}

	// Clean up checkpoint after successful completion.
	if e.checkpoint != nil {
		if err := e.checkpoint.DeleteCheckpoint(ctx, runID); err != nil {
			e.logger.Printf("Warning: failed to clean up checkpoint: %v", err)
		}
	}

	return run, nil
}

type workItem struct {
	task       model.Task
	trialIndex int
}

func (e *Engine) buildWorkItems() []workItem {
	var items []workItem
	for _, task := range e.suite.Tasks {
		for i := 0; i < task.TrialsPerTask; i++ {
			items = append(items, workItem{
				task:       task,
				trialIndex: i,
			})
		}
	}
	return items
}

func (e *Engine) resolveGraders() (map[string][]grader.Grader, error) {
	result := make(map[string][]grader.Grader)
	for _, task := range e.suite.Tasks {
		var graders []grader.Grader
		for _, ref := range task.Graders {
			g, err := grader.Create(ref.Type, ref.Config)
			if err != nil {
				return nil, fmt.Errorf("task %q: creating grader %q: %w", task.ID, ref.Type, err)
			}
			graders = append(graders, g)
		}
		result[task.ID] = graders
	}
	return result, nil
}

func (e *Engine) aggregateResults(trials []*model.Trial) []model.TaskResult {
	// Group trials by task ID.
	trialsByTask := make(map[string][]*model.Trial)
	for _, t := range trials {
		trialsByTask[t.TaskID] = append(trialsByTask[t.TaskID], t)
	}

	costPerInput := e.suite.Agent.CostPerInputToken
	costPerOutput := e.suite.Agent.CostPerOutputToken

	var results []model.TaskResult
	for _, task := range e.suite.Tasks {
		taskTrials := trialsByTask[task.ID]
		tr := model.TaskResult{
			Task: task,
		}

		var totalScore float64
		var durations []int64
		var totalSteps int
		var stepsCount int
		var usage model.UsageSummary
		hasUsage := false

		for _, trial := range taskTrials {
			tr.Trials = append(tr.Trials, *trial)
			totalScore += trial.Score
			switch {
			case trial.Status == model.TrialStatusPassed:
				tr.PassCount++
			case trial.Status == model.TrialStatusError:
				tr.ErrorCount++
			default:
				tr.FailCount++
			}

			// Collect latency data (use agent duration if available, else total).
			if trial.AgentDurationMS > 0 {
				durations = append(durations, trial.AgentDurationMS)
			} else if trial.DurationMS > 0 {
				durations = append(durations, trial.DurationMS)
			}

			// Collect step count.
			if trial.StepCount > 0 {
				totalSteps += trial.StepCount
				stepsCount++
			}

			// Collect token usage from agent metadata.
			if trial.AgentOutput != nil && trial.AgentOutput.Metadata != nil {
				if u, ok := trial.AgentOutput.Metadata["usage"]; ok {
					if usageMap, ok := u.(map[string]any); ok {
						hasUsage = true
						if v, ok := toIntFromMeta(usageMap["input_tokens"]); ok {
							usage.TotalInputTokens += int64(v)
						}
						if v, ok := toIntFromMeta(usageMap["prompt_tokens"]); ok {
							usage.TotalInputTokens += int64(v)
						}
						if v, ok := toIntFromMeta(usageMap["output_tokens"]); ok {
							usage.TotalOutputTokens += int64(v)
						}
						if v, ok := toIntFromMeta(usageMap["completion_tokens"]); ok {
							usage.TotalOutputTokens += int64(v)
						}
					}
				}
			}
		}

		n := len(taskTrials)
		if n > 0 {
			tr.AvgScore = totalScore / float64(n)
			k := min(n, task.TrialsPerTask)
			tr.PassAtK = model.ComputePassAtK(n, tr.PassCount, k)
			tr.PassPowerK = model.ComputePassPowerK(n, tr.PassCount, k)
		}

		// Compute latency percentiles.
		if len(durations) > 0 {
			sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
			tr.LatencyP50MS = percentile(durations, 50)
			tr.LatencyP90MS = percentile(durations, 90)
			tr.LatencyP99MS = percentile(durations, 99)
		}

		// Compute average step count.
		if stepsCount > 0 {
			tr.AvgStepCount = float64(totalSteps) / float64(stepsCount)
		}

		// Set usage summary.
		if hasUsage {
			usage.TotalTokens = usage.TotalInputTokens + usage.TotalOutputTokens
			if costPerInput > 0 || costPerOutput > 0 {
				usage.EstimatedCostUSD = float64(usage.TotalInputTokens)*costPerInput +
					float64(usage.TotalOutputTokens)*costPerOutput
			}
			tr.Usage = &usage
		}

		results = append(results, tr)
	}

	return results
}

func (e *Engine) computeSummary(results []model.TaskResult) model.EvalSummary {
	s := model.EvalSummary{
		TotalTasks: len(results),
	}

	var totalPassAtK, totalPassPowerK, totalScore float64
	var totalUsage model.UsageSummary
	hasUsage := false

	for _, r := range results {
		s.TotalTrials += len(r.Trials)
		s.PassedTrials += r.PassCount
		s.FailedTrials += r.FailCount
		s.ErrorTrials += r.ErrorCount
		totalScore += r.AvgScore
		totalPassAtK += r.PassAtK
		totalPassPowerK += r.PassPowerK

		if r.Usage != nil {
			hasUsage = true
			totalUsage.TotalInputTokens += r.Usage.TotalInputTokens
			totalUsage.TotalOutputTokens += r.Usage.TotalOutputTokens
			totalUsage.TotalTokens += r.Usage.TotalTokens
			totalUsage.EstimatedCostUSD += r.Usage.EstimatedCostUSD
		}
	}

	if s.TotalTrials > 0 {
		s.OverallPassRate = float64(s.PassedTrials) / float64(s.TotalTrials)
	}
	if s.TotalTasks > 0 {
		s.AvgScore = totalScore / float64(s.TotalTasks)
		s.AvgPassAtK = totalPassAtK / float64(s.TotalTasks)
		s.AvgPassPowerK = totalPassPowerK / float64(s.TotalTasks)
	}

	if hasUsage {
		s.Usage = &totalUsage
	}

	return s
}

// filterTasksByTags filters tasks based on tag inclusion/exclusion rules.
// Tags use OR logic: a task matches if it has any of the specified tags.
// Exclude tags take precedence over include tags.
func (e *Engine) filterTasksByTags() {
	if len(e.tags) == 0 && len(e.excludeTags) == 0 {
		return
	}

	var filtered []model.Task
	for _, task := range e.suite.Tasks {
		if e.taskMatchesTags(task) {
			filtered = append(filtered, task)
		}
	}

	e.logger.Printf("Tag filter: %d/%d tasks selected", len(filtered), len(e.suite.Tasks))
	e.suite.Tasks = filtered
}

func (e *Engine) taskMatchesTags(task model.Task) bool {
	// Check exclusion first (takes precedence).
	if len(e.excludeTags) > 0 {
		for _, tag := range task.Tags {
			for _, excl := range e.excludeTags {
				if tag == excl {
					return false
				}
			}
		}
	}

	// If include tags specified, task must have at least one matching tag.
	if len(e.tags) > 0 {
		for _, tag := range task.Tags {
			for _, incl := range e.tags {
				if tag == incl {
					return true
				}
			}
		}
		return false
	}

	return true
}

// percentile computes the p-th percentile from a sorted slice of int64 values.
func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := float64(p) / 100.0 * float64(len(sorted)-1)
	lower := int(idx)
	if lower >= len(sorted)-1 {
		return sorted[len(sorted)-1]
	}
	return sorted[lower]
}

// toIntFromMeta converts a metadata value to int, handling common JSON numeric types.
func toIntFromMeta(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i), true
		}
	}
	return 0, false
}
