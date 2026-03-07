// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// TableReporter outputs evaluation results as a formatted table to stdout.
type TableReporter struct{}

func (r *TableReporter) Generate(run *model.EvalRun, _ string) error {
	return r.WriteTo(os.Stdout, run)
}

// WriteTo writes the table report to the given writer.
func (r *TableReporter) WriteTo(w io.Writer, run *model.EvalRun) error {
	// Header.
	fmt.Fprintf(w, "\n=== Evaluation Report: %s ===\n", run.SuiteName)
	if run.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", run.Description)
	}
	fmt.Fprintf(w, "Agent: %s | Run ID: %s\n", run.AgentType, run.ID[:8])
	fmt.Fprintf(w, "Duration: %dms\n\n", run.DurationMS)

	// Task results table.
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TASK\tPASS\tFAIL\tERR\tAVG SCORE\tPASS@K\tPASS^K\tP50ms\tP90ms\tP99ms")
	fmt.Fprintln(tw, "----\t----\t----\t---\t---------\t------\t------\t-----\t-----\t-----")

	for _, tr := range run.TaskResults {
		name := tr.Task.Name
		if name == "" {
			name = tr.Task.ID
		}
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%.3f\t%.3f\t%.3f\t%d\t%d\t%d\n",
			name,
			tr.PassCount,
			tr.FailCount,
			tr.ErrorCount,
			tr.AvgScore,
			tr.PassAtK,
			tr.PassPowerK,
			tr.LatencyP50MS,
			tr.LatencyP90MS,
			tr.LatencyP99MS,
		)
	}
	tw.Flush()

	// Summary.
	s := run.Summary
	fmt.Fprintf(w, "\n--- Summary ---\n")
	fmt.Fprintf(w, "Tasks: %d | Trials: %d (passed: %d, failed: %d, error: %d)\n",
		s.TotalTasks, s.TotalTrials, s.PassedTrials, s.FailedTrials, s.ErrorTrials)
	fmt.Fprintf(w, "Overall Pass Rate: %.1f%% | Avg Score: %.3f\n",
		s.OverallPassRate*100, s.AvgScore)
	fmt.Fprintf(w, "Avg pass@k: %.3f | Avg pass^k: %.3f\n",
		s.AvgPassAtK, s.AvgPassPowerK)

	// Token usage summary.
	if s.Usage != nil {
		fmt.Fprintf(w, "\n--- Token Usage ---\n")
		fmt.Fprintf(w, "Input tokens: %d | Output tokens: %d | Total: %d\n",
			s.Usage.TotalInputTokens, s.Usage.TotalOutputTokens, s.Usage.TotalTokens)
		if s.Usage.EstimatedCostUSD > 0 {
			fmt.Fprintf(w, "Estimated cost: $%.4f\n", s.Usage.EstimatedCostUSD)
		}
	}

	// Per-task details (failures/errors).
	hasIssues := false
	for _, tr := range run.TaskResults {
		for _, trial := range tr.Trials {
			if trial.Status == model.TrialStatusFailed || trial.Status == model.TrialStatusError {
				if !hasIssues {
					fmt.Fprintf(w, "\n--- Failed/Error Trials ---\n")
					hasIssues = true
				}
				fmt.Fprintf(w, "[%s] trial #%d: %s\n", tr.Task.ID, trial.Index, trial.Status)
				if trial.Error != "" {
					fmt.Fprintf(w, "  Error: %s\n", trial.Error)
				}
				for _, g := range trial.Grades {
					if !g.Pass {
						reason := g.Reason
						if len(reason) > 100 {
							reason = reason[:97] + "..."
						}
						fmt.Fprintf(w, "  [%s] score=%.2f reason=%s\n", g.GraderType, g.Score, reason)
					}
				}
			}
		}
	}

	fmt.Fprintln(w, strings.Repeat("=", 50))
	return nil
}
