// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"fmt"
	"io"
	"math"
	"text/tabwriter"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// CompareResult holds structured comparison data between two runs.
// Used by both CLI (text rendering) and Web (JSON API).
type CompareResult struct {
	RunA    RunMeta          `json:"run_a"`
	RunB    RunMeta          `json:"run_b"`
	Summary CompareSummary   `json:"summary"`
	Tasks   []TaskComparison `json:"tasks"`
}

// RunMeta holds identifying information about a run.
type RunMeta struct {
	ID        string    `json:"id"`
	SuiteName string    `json:"suite_name"`
	AgentType string    `json:"agent_type"`
	StartedAt time.Time `json:"started_at"`
}

// CompareSummary holds summary-level metric comparisons.
type CompareSummary struct {
	PassRate      MetricDiff `json:"pass_rate"`
	AvgScore      MetricDiff `json:"avg_score"`
	AvgPassAtK    MetricDiff `json:"avg_pass_at_k"`
	AvgPassPowerK MetricDiff `json:"avg_pass_power_k"`
}

// MetricDiff holds a single metric comparison (A, B, and their diff).
type MetricDiff struct {
	A    float64 `json:"a"`
	B    float64 `json:"b"`
	Diff float64 `json:"diff"`
}

// TaskComparison holds per-task comparison data.
type TaskComparison struct {
	TaskID  string        `json:"task_id"`
	ScoreA  float64       `json:"score_a"`
	ScoreB  float64       `json:"score_b"`
	Diff    float64       `json:"diff"`
	Status  string        `json:"status"` // "improved" / "regressed" / "unchanged"
	TrialsA []TrialDetail `json:"trials_a"`
	TrialsB []TrialDetail `json:"trials_b"`
}

// TrialDetail holds trial-level data for comparison display.
type TrialDetail struct {
	Status string        `json:"status"`
	Score  float64       `json:"score"`
	Grades []GradeDetail `json:"grades"`
}

// GradeDetail holds grade-level data for comparison display.
type GradeDetail struct {
	GraderType string  `json:"grader_type"`
	Score      float64 `json:"score"`
	Pass       bool    `json:"pass"`
	Reason     string  `json:"reason"`
}

// CompareRuns computes a structured comparison between two evaluation runs.
func CompareRuns(runA, runB *model.EvalRun) *CompareResult {
	result := &CompareResult{
		RunA: RunMeta{
			ID:        runA.ID,
			SuiteName: runA.SuiteName,
			AgentType: runA.AgentType,
			StartedAt: runA.StartedAt,
		},
		RunB: RunMeta{
			ID:        runB.ID,
			SuiteName: runB.SuiteName,
			AgentType: runB.AgentType,
			StartedAt: runB.StartedAt,
		},
		Summary: CompareSummary{
			PassRate:      newMetricDiff(runA.Summary.OverallPassRate, runB.Summary.OverallPassRate),
			AvgScore:      newMetricDiff(runA.Summary.AvgScore, runB.Summary.AvgScore),
			AvgPassAtK:    newMetricDiff(runA.Summary.AvgPassAtK, runB.Summary.AvgPassAtK),
			AvgPassPowerK: newMetricDiff(runA.Summary.AvgPassPowerK, runB.Summary.AvgPassPowerK),
		},
	}

	taskMapA := buildTaskMap(runA)
	taskMapB := buildTaskMap(runB)
	allTasks := mergeTaskIDs(taskMapA, taskMapB)

	for _, taskID := range allTasks {
		a, okA := taskMapA[taskID]
		b, okB := taskMapB[taskID]

		var scoreA, scoreB float64
		if okA {
			scoreA = a.AvgScore
		}
		if okB {
			scoreB = b.AvgScore
		}

		diff := scoreB - scoreA
		status := "unchanged"
		if diff > 0.01 {
			status = "improved"
		} else if diff < -0.01 {
			status = "regressed"
		}

		tc := TaskComparison{
			TaskID: taskID,
			ScoreA: scoreA,
			ScoreB: scoreB,
			Diff:   diff,
			Status: status,
		}

		if okA {
			tc.TrialsA = convertTrials(a.Trials)
		}
		if okB {
			tc.TrialsB = convertTrials(b.Trials)
		}

		result.Tasks = append(result.Tasks, tc)
	}

	return result
}

// FormatCompareText renders a CompareResult as human-readable text for CLI output.
func FormatCompareText(result *CompareResult, w io.Writer) error {
	if _, err := fmt.Fprintf(w, "\n=== Run Comparison ===\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Run A: %s (%s)\n", truncateID(result.RunA.ID), result.RunA.SuiteName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Run B: %s (%s)\n\n", truncateID(result.RunB.ID), result.RunB.SuiteName); err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "METRIC\tRUN A\tRUN B\tDIFF"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(tw, "------\t-----\t-----\t----"); err != nil {
		return err
	}

	formatMetricLine(tw, "Pass Rate", result.Summary.PassRate, true)
	formatMetricLine(tw, "Avg Score", result.Summary.AvgScore, false)
	formatMetricLine(tw, "Avg pass@k", result.Summary.AvgPassAtK, false)
	formatMetricLine(tw, "Avg pass^k", result.Summary.AvgPassPowerK, false)
	if err := tw.Flush(); err != nil {
		return err
	}

	if len(result.Tasks) > 0 {
		if _, err := fmt.Fprintf(w, "\n--- Per-Task Comparison ---\n"); err != nil {
			return err
		}
		tw2 := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(tw2, "TASK\tSCORE A\tSCORE B\tDIFF\tSTATUS"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(tw2, "----\t-------\t-------\t----\t------"); err != nil {
			return err
		}

		for _, task := range result.Tasks {
			id := task.TaskID
			if len(id) > 25 {
				id = id[:22] + "..."
			}
			if _, err := fmt.Fprintf(tw2, "%s\t%.3f\t%.3f\t%+.3f\t%s\n", id, task.ScoreA, task.ScoreB, task.Diff, task.Status); err != nil {
				return err
			}
		}
		if err := tw2.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func newMetricDiff(a, b float64) MetricDiff {
	return MetricDiff{A: a, B: b, Diff: b - a}
}

func formatMetricLine(w io.Writer, metric string, md MetricDiff, isPct bool) {
	var aStr, bStr, diffStr string
	if isPct {
		aStr = fmt.Sprintf("%.1f%%", md.A*100)
		bStr = fmt.Sprintf("%.1f%%", md.B*100)
		diffStr = fmt.Sprintf("%+.1f%%", md.Diff*100)
	} else {
		aStr = fmt.Sprintf("%.3f", md.A)
		bStr = fmt.Sprintf("%.3f", md.B)
		diffStr = fmt.Sprintf("%+.3f", md.Diff)
	}
	indicator := ""
	if math.Abs(md.Diff) > 0.01 {
		if md.Diff > 0 {
			indicator = " [+]"
		} else {
			indicator = " [-]"
		}
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s%s\n", metric, aStr, bStr, diffStr, indicator)
}

func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func convertTrials(trials []model.Trial) []TrialDetail {
	details := make([]TrialDetail, len(trials))
	for i, t := range trials {
		grades := make([]GradeDetail, len(t.Grades))
		for j, g := range t.Grades {
			grades[j] = GradeDetail{
				GraderType: g.GraderType,
				Score:      g.Score,
				Pass:       g.Pass,
				Reason:     g.Reason,
			}
		}
		details[i] = TrialDetail{
			Status: string(t.Status),
			Score:  t.Score,
			Grades: grades,
		}
	}
	return details
}

func buildTaskMap(run *model.EvalRun) map[string]model.TaskResult {
	m := make(map[string]model.TaskResult)
	for _, tr := range run.TaskResults {
		m[tr.Task.ID] = tr
	}
	return m
}

func mergeTaskIDs(a, b map[string]model.TaskResult) []string {
	seen := make(map[string]bool)
	var ids []string
	for id := range a {
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	for id := range b {
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	return ids
}
