// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"context"
	"fmt"
	"strings"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func init() {
	Register("exact_match", func(config map[string]any) (Grader, error) {
		return newExactMatchGrader(config), nil
	})
}

type exactMatchGrader struct {
	ignoreCase       bool
	ignoreWhitespace bool
}

func newExactMatchGrader(config map[string]any) *exactMatchGrader {
	g := &exactMatchGrader{}
	if v, ok := config["ignore_case"].(bool); ok {
		g.ignoreCase = v
	}
	if v, ok := config["ignore_whitespace"].(bool); ok {
		g.ignoreWhitespace = v
	}
	return g
}

func (g *exactMatchGrader) Type() string { return "exact_match" }

func (g *exactMatchGrader) Grade(_ context.Context, input GradeInput) (*model.GradeResult, error) {
	if input.Task.Expected == nil {
		return &model.GradeResult{
			GraderType: g.Type(),
			Score:      0,
			Pass:       false,
			Reason:     "no expected output defined",
		}, nil
	}

	actual := input.AgentOutput.Text
	expected := input.Task.Expected.Text

	if g.ignoreWhitespace {
		actual = strings.TrimSpace(actual)
		expected = strings.TrimSpace(expected)
	}

	var match bool
	if g.ignoreCase {
		match = strings.EqualFold(actual, expected)
	} else {
		match = actual == expected
	}

	score := 0.0
	if match {
		score = 1.0
	}

	return &model.GradeResult{
		GraderType: g.Type(),
		Score:      score,
		Pass:       match,
		Reason:     fmt.Sprintf("expected=%q, got=%q", expected, actual),
	}, nil
}
