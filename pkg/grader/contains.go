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
	Register("contains", func(config map[string]any) (Grader, error) {
		return newContainsGrader(config), nil
	})
}

type containsGrader struct {
	ignoreCase bool
	keywords   []string // Optional additional keywords to check
}

func newContainsGrader(config map[string]any) *containsGrader {
	g := &containsGrader{}
	if v, ok := config["ignore_case"].(bool); ok {
		g.ignoreCase = v
	}
	if kws, ok := config["keywords"].([]any); ok {
		for _, kw := range kws {
			g.keywords = append(g.keywords, fmt.Sprintf("%v", kw))
		}
	}
	return g
}

func (g *containsGrader) Type() string { return "contains" }

func (g *containsGrader) Grade(_ context.Context, input GradeInput) (*model.GradeResult, error) {
	actual := input.AgentOutput.Text

	// Check expected text contains.
	var targets []string
	if input.Task.Expected != nil && input.Task.Expected.Text != "" {
		targets = append(targets, input.Task.Expected.Text)
	}
	targets = append(targets, g.keywords...)

	if len(targets) == 0 {
		return &model.GradeResult{
			GraderType: g.Type(),
			Score:      0,
			Pass:       false,
			Reason:     "no expected text or keywords defined",
		}, nil
	}

	matched := 0
	var missing []string
	for _, target := range targets {
		if g.containsCheck(actual, target) {
			matched++
		} else {
			missing = append(missing, target)
		}
	}

	score := float64(matched) / float64(len(targets))
	pass := matched == len(targets)

	reason := fmt.Sprintf("matched %d/%d targets", matched, len(targets))
	if len(missing) > 0 {
		reason += fmt.Sprintf(", missing: %v", missing)
	}

	return &model.GradeResult{
		GraderType: g.Type(),
		Score:      score,
		Pass:       pass,
		Reason:     reason,
	}, nil
}

func (g *containsGrader) containsCheck(haystack, needle string) bool {
	if g.ignoreCase {
		return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
	}
	return strings.Contains(haystack, needle)
}
