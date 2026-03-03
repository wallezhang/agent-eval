// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package engine

import (
	"context"
	"log"
	"testing"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func TestComputeWeightedScore(t *testing.T) {
	tests := []struct {
		name      string
		grades    []model.GradeResult
		wantScore float64
		wantPass  bool
	}{
		{
			name: "single pass",
			grades: []model.GradeResult{
				{Score: 1.0, Pass: true, Weight: 1.0},
			},
			wantScore: 1.0,
			wantPass:  true,
		},
		{
			name: "single fail",
			grades: []model.GradeResult{
				{Score: 0.0, Pass: false, Weight: 1.0},
			},
			wantScore: 0.0,
			wantPass:  false,
		},
		{
			name: "weighted average",
			grades: []model.GradeResult{
				{Score: 1.0, Pass: true, Weight: 2.0},
				{Score: 0.0, Pass: false, Weight: 1.0},
			},
			wantScore: 2.0 / 3.0,
			wantPass:  false, // one grader failed
		},
		{
			name:      "empty grades",
			grades:    nil,
			wantScore: 0.0,
			wantPass:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, pass := computeWeightedScore(tt.grades, nil)
			if score != tt.wantScore {
				t.Errorf("got score=%f, want %f", score, tt.wantScore)
			}
			if pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v", pass, tt.wantPass)
			}
		})
	}
}

func TestSchedulerConcurrency(t *testing.T) {
	items := make([]workItem, 10)
	for i := range items {
		items[i] = workItem{
			task:       model.Task{ID: "task-1"},
			trialIndex: i,
		}
	}

	sched := newScheduler(3, 0, log.Default())
	results, err := sched.Run(context.Background(), items, func(ctx context.Context, item workItem) (*model.Trial, error) {
		return &model.Trial{
			TaskID: item.task.ID,
			Index:  item.trialIndex,
			Status: model.TrialStatusPassed,
			Score:  1.0,
			Pass:   true,
		}, nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(results) != 10 {
		t.Errorf("got %d results, want 10", len(results))
	}
	for i, r := range results {
		if r.Index != i {
			t.Errorf("result %d has index=%d", i, r.Index)
		}
	}
}
