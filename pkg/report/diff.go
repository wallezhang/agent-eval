// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"fmt"
	"io"
	"math"
	"os"
	"text/tabwriter"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// DiffReporter compares two evaluation runs.
type DiffReporter struct{}

// Diff holds the comparison between two runs.
type Diff struct {
	RunA *model.EvalRun
	RunB *model.EvalRun
}

// CompareTo writes a comparison table between two runs.
func (r *DiffReporter) CompareTo(w io.Writer, runA, runB *model.EvalRun) error {
	fmt.Fprintf(w, "\n=== Run Comparison ===\n")
	fmt.Fprintf(w, "Run A: %s (%s)\n", runA.ID[:8], runA.SuiteName)
	fmt.Fprintf(w, "Run B: %s (%s)\n\n", runB.ID[:8], runB.SuiteName)

	// Summary comparison.
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "METRIC\tRUN A\tRUN B\tDIFF")
	fmt.Fprintln(tw, "------\t-----\t-----\t----")

	diffLine(tw, "Pass Rate", runA.Summary.OverallPassRate, runB.Summary.OverallPassRate, true)
	diffLine(tw, "Avg Score", runA.Summary.AvgScore, runB.Summary.AvgScore, false)
	diffLine(tw, "Avg pass@k", runA.Summary.AvgPassAtK, runB.Summary.AvgPassAtK, false)
	diffLine(tw, "Avg pass^k", runA.Summary.AvgPassPowerK, runB.Summary.AvgPassPowerK, false)
	tw.Flush()

	// Per-task comparison.
	taskMapA := buildTaskMap(runA)
	taskMapB := buildTaskMap(runB)

	allTasks := mergeTaskIDs(taskMapA, taskMapB)

	fmt.Fprintf(w, "\n--- Per-Task Comparison ---\n")
	tw2 := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw2, "TASK\tSCORE A\tSCORE B\tDIFF\tSTATUS")
	fmt.Fprintln(tw2, "----\t-------\t-------\t----\t------")

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
		status := "="
		if diff > 0.01 {
			status = "improved"
		} else if diff < -0.01 {
			status = "regressed"
		}

		id := taskID
		if len(id) > 25 {
			id = id[:22] + "..."
		}

		fmt.Fprintf(tw2, "%s\t%.3f\t%.3f\t%+.3f\t%s\n", id, scoreA, scoreB, diff, status)
	}
	tw2.Flush()

	return nil
}

// Generate writes the comparison to stdout (for use as a Reporter).
func (r *DiffReporter) Generate(run *model.EvalRun, _ string) error {
	// DiffReporter requires two runs; this is a no-op when used as a single Reporter.
	return nil
}

func diffLine(w io.Writer, metric string, a, b float64, isPct bool) {
	diff := b - a
	var aStr, bStr, diffStr string
	if isPct {
		aStr = fmt.Sprintf("%.1f%%", a*100)
		bStr = fmt.Sprintf("%.1f%%", b*100)
		diffStr = fmt.Sprintf("%+.1f%%", diff*100)
	} else {
		aStr = fmt.Sprintf("%.3f", a)
		bStr = fmt.Sprintf("%.3f", b)
		diffStr = fmt.Sprintf("%+.3f", diff)
	}
	indicator := ""
	if math.Abs(diff) > 0.01 {
		if diff > 0 {
			indicator = " [+]"
		} else {
			indicator = " [-]"
		}
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s%s\n", metric, aStr, bStr, diffStr, indicator)
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

// CompareRuns is a convenience function for CLI usage.
func CompareRuns(runA, runB *model.EvalRun) error {
	r := &DiffReporter{}
	return r.CompareTo(os.Stdout, runA, runB)
}
