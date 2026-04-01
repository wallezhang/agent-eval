// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"regexp"
	"strconv"
	"strings"
)

// sseLogWriter is an io.Writer that parses engine log lines and sends
// corresponding SSE events to a RunManager. It bridges the engine's
// logger output to the SSE event stream without modifying pkg/engine.
type sseLogWriter struct {
	runID string
	rm    *RunManager
}

// Patterns matching engine log output (from pkg/engine/runner.go, scheduler.go, engine.go)
var (
	// [1/4] Task "sample-task" trial #1: failed (score=0.00, 965ms)
	trialCompletedRe = regexp.MustCompile(`\[(\d+)/(\d+)\] Task "([^"]+)" trial #(\d+): (\w+) \(score=([\d.]+), (\d+)ms\)`)
	// Task "sample-task" trial #1: executing agent...
	trialStartedRe = regexp.MustCompile(`Task "([^"]+)" trial #(\d+): executing agent`)
	// Task "sample-task" trial #1: grading with 1 grader(s)...
	trialGradingRe = regexp.MustCompile(`Task "([^"]+)" trial #(\d+): grading with (\d+) grader`)
)

func (w *sseLogWriter) Write(p []byte) (n int, err error) {
	line := strings.TrimSpace(string(p))
	if line == "" {
		return len(p), nil
	}

	// Try to parse structured events from log lines
	if m := trialCompletedRe.FindStringSubmatch(line); m != nil {
		completed, _ := strconv.Atoi(m[1])
		total, _ := strconv.Atoi(m[2])
		taskID := m[3]
		trialIndex, _ := strconv.Atoi(m[4])
		status := m[5]
		score, _ := strconv.ParseFloat(m[6], 64)
		durationMS, _ := strconv.Atoi(m[7])

		w.rm.SendEvent(w.runID, SSEEvent{Type: "trial_completed", Data: map[string]any{
			"task_id":     taskID,
			"trial_index": trialIndex,
			"status":      status,
			"score":       score,
			"duration_ms": durationMS,
		}})

		// Also send progress update
		w.rm.SendEvent(w.runID, SSEEvent{Type: "run_progress", Data: map[string]any{
			"completed": completed,
			"total":     total,
		}})
	} else if m := trialStartedRe.FindStringSubmatch(line); m != nil {
		taskID := m[1]
		trialIndex, _ := strconv.Atoi(m[2])
		w.rm.SendEvent(w.runID, SSEEvent{Type: "trial_started", Data: map[string]any{
			"task_id":     taskID,
			"trial_index": trialIndex,
		}})
	}

	// Always send the raw log line as a log event
	w.rm.SendEvent(w.runID, SSEEvent{Type: "log", Data: map[string]any{
		"message": line,
	}})

	return len(p), nil
}
