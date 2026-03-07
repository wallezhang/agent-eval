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

func TestTagFiltering(t *testing.T) {
	suite := &model.EvalSuite{
		Tasks: []model.Task{
			{ID: "t1", Tags: []string{"safety", "regression"}},
			{ID: "t2", Tags: []string{"regression"}},
			{ID: "t3", Tags: []string{"safety"}},
			{ID: "t4", Tags: nil},
		},
	}

	tests := []struct {
		name        string
		tags        []string
		excludeTags []string
		wantIDs     []string
	}{
		{
			name:    "include safety",
			tags:    []string{"safety"},
			wantIDs: []string{"t1", "t3"},
		},
		{
			name:    "include regression",
			tags:    []string{"regression"},
			wantIDs: []string{"t1", "t2"},
		},
		{
			name:    "include both (OR logic)",
			tags:    []string{"safety", "regression"},
			wantIDs: []string{"t1", "t2", "t3"},
		},
		{
			name:        "exclude safety",
			excludeTags: []string{"safety"},
			wantIDs:     []string{"t2", "t4"},
		},
		{
			name:        "include regression, exclude safety",
			tags:        []string{"regression"},
			excludeTags: []string{"safety"},
			wantIDs:     []string{"t2"},
		},
		{
			name:    "no filters",
			wantIDs: []string{"t1", "t2", "t3", "t4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Deep copy tasks to avoid mutation between subtests.
			suiteCopy := *suite
			suiteCopy.Tasks = make([]model.Task, len(suite.Tasks))
			copy(suiteCopy.Tasks, suite.Tasks)

			eng := &Engine{
				suite:       &suiteCopy,
				logger:      log.Default(),
				tags:        tt.tags,
				excludeTags: tt.excludeTags,
			}
			eng.filterTasksByTags()

			var gotIDs []string
			for _, task := range eng.suite.Tasks {
				gotIDs = append(gotIDs, task.ID)
			}

			if len(gotIDs) != len(tt.wantIDs) {
				t.Fatalf("got %d tasks %v, want %d tasks %v", len(gotIDs), gotIDs, len(tt.wantIDs), tt.wantIDs)
			}
			for i, id := range gotIDs {
				if id != tt.wantIDs[i] {
					t.Errorf("task[%d] = %q, want %q", i, id, tt.wantIDs[i])
				}
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	tests := []struct {
		name   string
		sorted []int64
		p      int
		want   int64
	}{
		{"empty", nil, 50, 0},
		{"single", []int64{100}, 50, 100},
		{"p50 of two", []int64{100, 200}, 50, 100},
		{"p50 of three", []int64{100, 200, 300}, 50, 200},
		{"p90 of ten", []int64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, 90, 90},
		{"p99 of ten", []int64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, 99, 90},
		{"p0", []int64{10, 20, 30}, 0, 10},
		{"p100", []int64{10, 20, 30}, 100, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentile(tt.sorted, tt.p)
			if got != tt.want {
				t.Errorf("percentile(%v, %d) = %d, want %d", tt.sorted, tt.p, got, tt.want)
			}
		})
	}
}

func TestAggregateResultsUsage(t *testing.T) {
	suite := &model.EvalSuite{
		Agent: model.AgentConfig{
			CostPerInputToken:  0.001,
			CostPerOutputToken: 0.002,
		},
		Tasks: []model.Task{
			{ID: "t1", TrialsPerTask: 2},
		},
	}
	eng := &Engine{suite: suite, logger: log.Default()}

	trials := []*model.Trial{
		{
			TaskID:          "t1",
			Index:           0,
			Status:          model.TrialStatusPassed,
			Score:           1.0,
			Pass:            true,
			DurationMS:      100,
			AgentDurationMS: 80,
			StepCount:       3,
			AgentOutput: &model.AgentOutput{
				Text: "result",
				Metadata: map[string]any{
					"usage": map[string]any{
						"input_tokens":  float64(100),
						"output_tokens": float64(50),
					},
				},
			},
		},
		{
			TaskID:          "t1",
			Index:           1,
			Status:          model.TrialStatusPassed,
			Score:           1.0,
			Pass:            true,
			DurationMS:      200,
			AgentDurationMS: 160,
			StepCount:       5,
			AgentOutput: &model.AgentOutput{
				Text: "result",
				Metadata: map[string]any{
					"usage": map[string]any{
						"input_tokens":  float64(200),
						"output_tokens": float64(100),
					},
				},
			},
		},
	}

	results := eng.aggregateResults(trials)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	r := results[0]

	// Check usage.
	if r.Usage == nil {
		t.Fatal("expected usage to be non-nil")
	}
	if r.Usage.TotalInputTokens != 300 {
		t.Errorf("TotalInputTokens = %d, want 300", r.Usage.TotalInputTokens)
	}
	if r.Usage.TotalOutputTokens != 150 {
		t.Errorf("TotalOutputTokens = %d, want 150", r.Usage.TotalOutputTokens)
	}
	if r.Usage.TotalTokens != 450 {
		t.Errorf("TotalTokens = %d, want 450", r.Usage.TotalTokens)
	}
	expectedCost := 300*0.001 + 150*0.002
	if abs64(r.Usage.EstimatedCostUSD-expectedCost) > 0.001 {
		t.Errorf("EstimatedCostUSD = %.4f, want %.4f", r.Usage.EstimatedCostUSD, expectedCost)
	}

	// Check latency percentiles.
	if r.LatencyP50MS != 80 {
		t.Errorf("LatencyP50MS = %d, want 80", r.LatencyP50MS)
	}

	// Check step count.
	if abs64(r.AvgStepCount-4.0) > 0.01 {
		t.Errorf("AvgStepCount = %.2f, want 4.0", r.AvgStepCount)
	}
}

func TestExtractStepCount(t *testing.T) {
	tests := []struct {
		name   string
		output *model.AgentOutput
		trans  *model.Transcript
		want   int
	}{
		{
			name:   "from metadata step_count",
			output: &model.AgentOutput{Metadata: map[string]any{"step_count": float64(5)}},
			trans:  nil,
			want:   5,
		},
		{
			name:   "from metadata steps",
			output: &model.AgentOutput{Metadata: map[string]any{"steps": float64(3)}},
			trans:  nil,
			want:   3,
		},
		{
			name:   "from transcript",
			output: &model.AgentOutput{},
			trans: &model.Transcript{Steps: []model.TranscriptStep{
				{Type: "input"},
				{Type: "output"},
				{Type: "output"},
			}},
			want: 2,
		},
		{
			name:   "no data",
			output: &model.AgentOutput{},
			trans:  &model.Transcript{Steps: []model.TranscriptStep{{Type: "input"}}},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractStepCount(tt.output, tt.trans)
			if got != tt.want {
				t.Errorf("extractStepCount = %d, want %d", got, tt.want)
			}
		})
	}
}

func abs64(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
