// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"context"
	"testing"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func TestConstraintGrader_PatternMustMatch(t *testing.T) {
	g, err := newConstraintGrader(map[string]any{
		"checks": []any{
			map[string]any{
				"name":       "has_greeting",
				"pattern":    "(?i)hello",
				"must_match": true,
			},
		},
	})
	if err != nil {
		t.Fatalf("newConstraintGrader: %v", err)
	}

	tests := []struct {
		name     string
		text     string
		wantPass bool
	}{
		{"match", "Hello world", true},
		{"no_match", "Goodbye world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := g.Grade(context.Background(), GradeInput{
				AgentOutput: model.AgentOutput{Text: tt.text},
			})
			if err != nil {
				t.Fatalf("Grade: %v", err)
			}
			if result.Pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v (reason: %s)", result.Pass, tt.wantPass, result.Reason)
			}
		})
	}
}

func TestConstraintGrader_PatternMustNotMatch(t *testing.T) {
	g, err := newConstraintGrader(map[string]any{
		"checks": []any{
			map[string]any{
				"name":           "no_pii",
				"pattern":        `(?i)(ssn|credit card)`,
				"must_not_match": true,
			},
		},
	})
	if err != nil {
		t.Fatalf("newConstraintGrader: %v", err)
	}

	tests := []struct {
		name     string
		text     string
		wantPass bool
	}{
		{"clean", "Here is your account info", true},
		{"has_ssn", "Your SSN is 123-45-6789", false},
		{"has_credit_card", "Your Credit Card number is...", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := g.Grade(context.Background(), GradeInput{
				AgentOutput: model.AgentOutput{Text: tt.text},
			})
			if err != nil {
				t.Fatalf("Grade: %v", err)
			}
			if result.Pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v (reason: %s)", result.Pass, tt.wantPass, result.Reason)
			}
		})
	}
}

func TestConstraintGrader_WordLimit(t *testing.T) {
	g, err := newConstraintGrader(map[string]any{
		"checks": []any{
			map[string]any{
				"name":      "word_limit",
				"max_words": 5,
			},
		},
	})
	if err != nil {
		t.Fatalf("newConstraintGrader: %v", err)
	}

	tests := []struct {
		name     string
		text     string
		wantPass bool
	}{
		{"under_limit", "one two three", true},
		{"at_limit", "one two three four five", true},
		{"over_limit", "one two three four five six", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := g.Grade(context.Background(), GradeInput{
				AgentOutput: model.AgentOutput{Text: tt.text},
			})
			if err != nil {
				t.Fatalf("Grade: %v", err)
			}
			if result.Pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v (reason: %s)", result.Pass, tt.wantPass, result.Reason)
			}
		})
	}
}

func TestConstraintGrader_MultipleChecks(t *testing.T) {
	g, err := newConstraintGrader(map[string]any{
		"checks": []any{
			map[string]any{
				"name":           "no_pii",
				"pattern":        `(?i)ssn`,
				"must_not_match": true,
			},
			map[string]any{
				"name":       "has_disclaimer",
				"pattern":    `(?i)disclaimer`,
				"must_match": true,
			},
			map[string]any{
				"name":      "word_limit",
				"max_words": 100,
			},
		},
	})
	if err != nil {
		t.Fatalf("newConstraintGrader: %v", err)
	}

	tests := []struct {
		name      string
		text      string
		wantPass  bool
		wantScore float64
	}{
		{"all_pass", "Disclaimer: No sensitive data here.", true, 1.0},
		{"missing_disclaimer", "No sensitive data here.", false, 2.0 / 3.0},
		{"has_pii", "Disclaimer: Your SSN is here", false, 2.0 / 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := g.Grade(context.Background(), GradeInput{
				AgentOutput: model.AgentOutput{Text: tt.text},
			})
			if err != nil {
				t.Fatalf("Grade: %v", err)
			}
			if result.Pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v (reason: %s)", result.Pass, tt.wantPass, result.Reason)
			}
			if abs(result.Score-tt.wantScore) > 0.01 {
				t.Errorf("got score=%.3f, want %.3f", result.Score, tt.wantScore)
			}
		})
	}
}

func TestConstraintGrader_MinWords(t *testing.T) {
	g, err := newConstraintGrader(map[string]any{
		"checks": []any{
			map[string]any{
				"name":      "min_length",
				"min_words": 3,
			},
		},
	})
	if err != nil {
		t.Fatalf("newConstraintGrader: %v", err)
	}

	tests := []struct {
		name     string
		text     string
		wantPass bool
	}{
		{"enough_words", "one two three", true},
		{"too_few", "one two", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := g.Grade(context.Background(), GradeInput{
				AgentOutput: model.AgentOutput{Text: tt.text},
			})
			if err != nil {
				t.Fatalf("Grade: %v", err)
			}
			if result.Pass != tt.wantPass {
				t.Errorf("got pass=%v, want %v", result.Pass, tt.wantPass)
			}
		})
	}
}

func TestConstraintGrader_InvalidConfig(t *testing.T) {
	_, err := newConstraintGrader(map[string]any{})
	if err == nil {
		t.Error("expected error for missing checks")
	}

	_, err = newConstraintGrader(map[string]any{
		"checks": []any{
			map[string]any{
				"pattern": "[invalid",
			},
		},
	})
	if err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestConstraintGrader_RegisteredType(t *testing.T) {
	g, err := Create("constraint", map[string]any{
		"checks": []any{
			map[string]any{
				"name":       "test",
				"pattern":    "hello",
				"must_match": true,
			},
		},
	})
	if err != nil {
		t.Fatalf("Create constraint grader: %v", err)
	}
	if g.Type() != "constraint" {
		t.Errorf("got type=%q, want %q", g.Type(), "constraint")
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
