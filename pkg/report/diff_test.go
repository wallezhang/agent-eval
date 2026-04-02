// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func makeTestRun(id, suite, agent string, passRate, avgScore, passAtK, passPowerK float64, tasks []model.TaskResult) *model.EvalRun {
	return &model.EvalRun{
		ID:        id,
		SuiteName: suite,
		AgentType: agent,
		StartedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
		Summary: model.EvalSummary{
			OverallPassRate: passRate,
			AvgScore:        avgScore,
			AvgPassAtK:      passAtK,
			AvgPassPowerK:   passPowerK,
		},
		TaskResults: tasks,
	}
}

func TestCompareRuns_Summary(t *testing.T) {
	runA := makeTestRun("aaaa1111-0000-0000-0000-000000000000", "suite1", "openai", 0.75, 0.68, 0.80, 0.60, nil)
	runB := makeTestRun("bbbb2222-0000-0000-0000-000000000000", "suite1", "anthropic", 0.85, 0.79, 0.90, 0.70, nil)

	result := CompareRuns(runA, runB)

	if result.RunA.ID != runA.ID {
		t.Errorf("RunA.ID = %q, want %q", result.RunA.ID, runA.ID)
	}
	if result.RunB.ID != runB.ID {
		t.Errorf("RunB.ID = %q, want %q", result.RunB.ID, runB.ID)
	}
	if result.Summary.PassRate.A != 0.75 {
		t.Errorf("PassRate.A = %v, want 0.75", result.Summary.PassRate.A)
	}
	if result.Summary.PassRate.B != 0.85 {
		t.Errorf("PassRate.B = %v, want 0.85", result.Summary.PassRate.B)
	}
	if diff := result.Summary.PassRate.Diff; diff < 0.099 || diff > 0.101 {
		t.Errorf("PassRate.Diff = %v, want ~0.10", diff)
	}
}

func TestCompareRuns_Tasks(t *testing.T) {
	tasksA := []model.TaskResult{
		{Task: model.Task{ID: "task-1"}, AvgScore: 0.5, Trials: []model.Trial{
			{Status: model.TrialStatusPassed, Score: 0.5, Grades: []model.GradeResult{{GraderType: "exact_match", Score: 0.5, Pass: false, Reason: "partial"}}},
		}},
		{Task: model.Task{ID: "task-2"}, AvgScore: 1.0, Trials: []model.Trial{
			{Status: model.TrialStatusPassed, Score: 1.0, Grades: []model.GradeResult{{GraderType: "exact_match", Score: 1.0, Pass: true, Reason: "match"}}},
		}},
	}
	tasksB := []model.TaskResult{
		{Task: model.Task{ID: "task-1"}, AvgScore: 0.8, Trials: []model.Trial{
			{Status: model.TrialStatusPassed, Score: 0.8, Grades: []model.GradeResult{{GraderType: "exact_match", Score: 0.8, Pass: true, Reason: "close"}}},
		}},
		{Task: model.Task{ID: "task-2"}, AvgScore: 1.0, Trials: []model.Trial{
			{Status: model.TrialStatusPassed, Score: 1.0, Grades: []model.GradeResult{{GraderType: "exact_match", Score: 1.0, Pass: true, Reason: "match"}}},
		}},
	}
	runA := makeTestRun("aaaa", "s", "a", 0.75, 0.75, 0.8, 0.6, tasksA)
	runB := makeTestRun("bbbb", "s", "a", 0.90, 0.90, 0.9, 0.7, tasksB)

	result := CompareRuns(runA, runB)

	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Tasks))
	}

	// Find tasks by ID (order is non-deterministic due to map iteration)
	taskByID := make(map[string]TaskComparison)
	for _, tc := range result.Tasks {
		taskByID[tc.TaskID] = tc
	}

	// task-1: improved (0.5 → 0.8)
	t1, ok := taskByID["task-1"]
	if !ok {
		t.Fatal("task-1 not found in results")
	}
	if t1.Status != "improved" {
		t.Errorf("task-1 status = %q, want improved", t1.Status)
	}
	if len(t1.TrialsA) != 1 || len(t1.TrialsB) != 1 {
		t.Errorf("task-1 trials: A=%d B=%d, want 1 each", len(t1.TrialsA), len(t1.TrialsB))
	}
	if t1.TrialsA[0].Grades[0].GraderType != "exact_match" {
		t.Errorf("trial grade type = %q, want exact_match", t1.TrialsA[0].Grades[0].GraderType)
	}

	t2, ok := taskByID["task-2"]
	if !ok {
		t.Fatal("task-2 not found in results")
	}
	if t2.Status != "unchanged" {
		t.Errorf("task-2 status = %q, want unchanged", t2.Status)
	}
}

func TestCompareRuns_TaskOnlyInOneRun(t *testing.T) {
	tasksA := []model.TaskResult{
		{Task: model.Task{ID: "task-only-a"}, AvgScore: 0.5, Trials: []model.Trial{}},
	}
	tasksB := []model.TaskResult{
		{Task: model.Task{ID: "task-only-b"}, AvgScore: 0.8, Trials: []model.Trial{}},
	}
	runA := makeTestRun("a", "s", "a", 0.5, 0.5, 0.5, 0.5, tasksA)
	runB := makeTestRun("b", "s", "a", 0.8, 0.8, 0.8, 0.8, tasksB)

	result := CompareRuns(runA, runB)

	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Tasks))
	}
}

func TestFormatCompareText(t *testing.T) {
	runA := makeTestRun("aaaa1111-0000-0000-0000-000000000000", "suite1", "openai", 0.75, 0.68, 0.80, 0.60, nil)
	runB := makeTestRun("bbbb2222-0000-0000-0000-000000000000", "suite1", "anthropic", 0.85, 0.79, 0.90, 0.70, nil)

	result := CompareRuns(runA, runB)

	var buf bytes.Buffer
	if err := FormatCompareText(result, &buf); err != nil {
		t.Fatalf("FormatCompareText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Run Comparison") {
		t.Errorf("output missing 'Run Comparison' header")
	}
	if !strings.Contains(output, "aaaa1111") {
		t.Errorf("output missing run A ID prefix")
	}
	if !strings.Contains(output, "Pass Rate") {
		t.Errorf("output missing 'Pass Rate' metric")
	}
}
