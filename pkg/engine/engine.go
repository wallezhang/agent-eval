// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/wallezhang/agent-eval/pkg/agent"
	"github.com/wallezhang/agent-eval/pkg/grader"
	"github.com/wallezhang/agent-eval/pkg/model"
)

// Engine orchestrates the evaluation process.
type Engine struct {
	suite  *model.EvalSuite
	agent  agent.Agent
	logger *log.Logger
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

// Execute runs the full evaluation and returns the results.
func (e *Engine) Execute(ctx context.Context) (*model.EvalRun, error) {
	startTime := time.Now()

	// Create the agent under test.
	ag, err := agent.Create(e.suite.Agent.Type, e.suite.Agent.Config)
	if err != nil {
		return nil, fmt.Errorf("creating agent: %w", err)
	}
	e.agent = ag
	defer ag.Close()

	// Build work items (task × trials).
	workItems := e.buildWorkItems()
	e.logger.Printf("Starting evaluation %q: %d tasks, %d total trials",
		e.suite.Name, len(e.suite.Tasks), len(workItems))

	// Resolve graders for each task.
	taskGraders, err := e.resolveGraders()
	if err != nil {
		return nil, fmt.Errorf("resolving graders: %w", err)
	}

	// Execute trials using the scheduler.
	sched := newScheduler(e.suite.Execution.Concurrency, e.suite.Execution.RateLimitRPS)
	trials, err := sched.Run(ctx, workItems, func(ctx context.Context, item workItem) (*model.Trial, error) {
		runner := &Runner{
			agent:   e.agent,
			graders: taskGraders[item.task.ID],
			timeout: e.suite.Execution.Timeout,
		}
		return runner.Run(ctx, item.task, item.trialIndex)
	})
	if err != nil {
		return nil, fmt.Errorf("executing trials: %w", err)
	}

	// Aggregate results by task.
	taskResults := e.aggregateResults(trials)

	// Compute summary.
	summary := e.computeSummary(taskResults)

	finishTime := time.Now()

	run := &model.EvalRun{
		ID:          uuid.New().String(),
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

	var results []model.TaskResult
	for _, task := range e.suite.Tasks {
		taskTrials := trialsByTask[task.ID]
		tr := model.TaskResult{
			Task: task,
		}

		var totalScore float64
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
		}

		n := len(taskTrials)
		if n > 0 {
			tr.AvgScore = totalScore / float64(n)
			k := min(n, task.TrialsPerTask)
			tr.PassAtK = model.ComputePassAtK(n, tr.PassCount, k)
			tr.PassPowerK = model.ComputePassPowerK(n, tr.PassCount, k)
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
	for _, r := range results {
		s.TotalTrials += len(r.Trials)
		s.PassedTrials += r.PassCount
		s.FailedTrials += r.FailCount
		s.ErrorTrials += r.ErrorCount
		totalScore += r.AvgScore
		totalPassAtK += r.PassAtK
		totalPassPowerK += r.PassPowerK
	}

	if s.TotalTrials > 0 {
		s.OverallPassRate = float64(s.PassedTrials) / float64(s.TotalTrials)
	}
	if s.TotalTasks > 0 {
		s.AvgScore = totalScore / float64(s.TotalTasks)
		s.AvgPassAtK = totalPassAtK / float64(s.TotalTasks)
		s.AvgPassPowerK = totalPassPowerK / float64(s.TotalTasks)
	}

	return s
}
