// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemory()
	ctx := context.Background()

	run := &model.EvalRun{
		ID:         "test-run-id",
		SuiteName:  "test-suite",
		AgentType:  "http",
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Summary: model.EvalSummary{
			TotalTasks:  2,
			TotalTrials: 4,
		},
	}

	// Save.
	if err := store.SaveRun(ctx, run); err != nil {
		t.Fatalf("SaveRun: %v", err)
	}

	// Get.
	got, err := store.GetRun(ctx, "test-run-id")
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if got.SuiteName != "test-suite" {
		t.Errorf("got suite_name=%q, want %q", got.SuiteName, "test-suite")
	}

	// List.
	runs, err := store.ListRuns(ctx)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("got %d runs, want 1", len(runs))
	}

	// Delete.
	if err := store.DeleteRun(ctx, "test-run-id"); err != nil {
		t.Fatalf("DeleteRun: %v", err)
	}
	runs, _ = store.ListRuns(ctx)
	if len(runs) != 0 {
		t.Errorf("got %d runs after delete, want 0", len(runs))
	}

	// Delete non-existent.
	if err := store.DeleteRun(ctx, "nonexistent"); err == nil {
		t.Error("expected error deleting non-existent run")
	}
}

func TestSQLiteStore(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	run := &model.EvalRun{
		ID:          "sqlite-test-id",
		SuiteName:   "test-suite",
		AgentType:   "http",
		AgentConfig: map[string]any{"url": "http://localhost"},
		TaskResults: []model.TaskResult{
			{
				Task: model.Task{ID: "t1", Name: "Task 1"},
				Trials: []model.Trial{
					{ID: "tr1", TaskID: "t1", Status: model.TrialStatusPassed, Score: 1.0},
				},
				PassCount: 1,
				AvgScore:  1.0,
			},
		},
		Summary: model.EvalSummary{
			TotalTasks:      1,
			TotalTrials:     1,
			PassedTrials:    1,
			OverallPassRate: 1.0,
			AvgScore:        1.0,
		},
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		DurationMS: 100,
	}

	// Save.
	if err := store.SaveRun(ctx, run); err != nil {
		t.Fatalf("SaveRun: %v", err)
	}

	// Get.
	got, err := store.GetRun(ctx, "sqlite-test-id")
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if got.SuiteName != "test-suite" {
		t.Errorf("got suite_name=%q, want %q", got.SuiteName, "test-suite")
	}
	if got.Summary.TotalTasks != 1 {
		t.Errorf("got total_tasks=%d, want 1", got.Summary.TotalTasks)
	}
	if len(got.TaskResults) != 1 {
		t.Errorf("got %d task results, want 1", len(got.TaskResults))
	}

	// List.
	runs, err := store.ListRuns(ctx)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("got %d runs, want 1", len(runs))
	}

	// Delete.
	if err := store.DeleteRun(ctx, "sqlite-test-id"); err != nil {
		t.Fatalf("DeleteRun: %v", err)
	}
	runs, _ = store.ListRuns(ctx)
	if len(runs) != 0 {
		t.Errorf("got %d runs after delete, want 0", len(runs))
	}
}
