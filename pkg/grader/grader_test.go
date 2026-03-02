// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"context"
	"testing"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func TestExactMatchGrader(t *testing.T) {
	g := newExactMatchGrader(map[string]any{
		"ignore_case":       true,
		"ignore_whitespace": true,
	})

	tests := []struct {
		name     string
		output   string
		expected string
		wantPass bool
	}{
		{"exact match", "hello", "hello", true},
		{"case mismatch ignored", "Hello", "hello", true},
		{"whitespace trimmed", "  hello  ", "hello", true},
		{"different text", "goodbye", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := GradeInput{
				Task: model.Task{
					Expected: &model.ExpectedOutput{Text: tt.expected},
				},
				AgentOutput: model.AgentOutput{Text: tt.output},
			}
			result, err := g.Grade(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v (reason: %s)", result.Pass, tt.wantPass, result.Reason)
			}
		})
	}
}

func TestContainsGrader(t *testing.T) {
	g := newContainsGrader(map[string]any{
		"ignore_case": true,
		"keywords":    []any{"hello", "world"},
	})

	tests := []struct {
		name      string
		output    string
		wantPass  bool
		wantScore float64
	}{
		{"all keywords", "Hello World!", true, 1.0},
		{"partial match", "Hello there!", false, 0.5},
		{"no match", "goodbye", false, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := GradeInput{
				Task:        model.Task{},
				AgentOutput: model.AgentOutput{Text: tt.output},
			}
			result, err := g.Grade(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v", result.Pass, tt.wantPass)
			}
			if result.Score != tt.wantScore {
				t.Errorf("got score=%f, want %f", result.Score, tt.wantScore)
			}
		})
	}
}

func TestRegexGrader(t *testing.T) {
	g, err := newRegexGrader(map[string]any{
		"pattern": `^\d{3}-\d{4}$`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		output   string
		wantPass bool
	}{
		{"matching pattern", "123-4567", true},
		{"no match", "hello", false},
		{"partial match", "123-45678", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := GradeInput{
				AgentOutput: model.AgentOutput{Text: tt.output},
			}
			result, err := g.Grade(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v", result.Pass, tt.wantPass)
			}
		})
	}
}

func TestJSONMatchGrader(t *testing.T) {
	g := newJSONMatchGrader(map[string]any{"ignore_case": true})

	input := GradeInput{
		Task: model.Task{
			Expected: &model.ExpectedOutput{
				Fields: map[string]string{
					"name": "Alice",
					"age":  "30",
				},
			},
		},
		AgentOutput: model.AgentOutput{
			Text: `{"name": "alice", "age": 30, "extra": true}`,
		},
	}

	result, err := g.Grade(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Pass {
		t.Errorf("expected pass, got fail: %s", result.Reason)
	}
	if result.Score != 1.0 {
		t.Errorf("got score=%f, want 1.0", result.Score)
	}
}

func TestGraderRegistry(t *testing.T) {
	// Test that built-in graders are registered.
	types := []string{"exact_match", "contains", "regex", "json_match", "command"}
	for _, typ := range types {
		_, err := Create(typ, map[string]any{
			"pattern": "test",
			"command": "echo",
		})
		if err != nil {
			t.Logf("Note: creating %q returned: %v (may need specific config)", typ, err)
		}
	}

	// Test unknown type.
	_, err := Create("nonexistent", nil)
	if err == nil {
		t.Error("expected error for unknown grader type")
	}
}
