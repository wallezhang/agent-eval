// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"context"
	"fmt"
	"regexp"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func init() {
	Register("regex", func(config map[string]any) (Grader, error) {
		return newRegexGrader(config)
	})
}

type regexGrader struct {
	pattern *regexp.Regexp
}

func newRegexGrader(config map[string]any) (*regexGrader, error) {
	patternStr, ok := config["pattern"].(string)
	if !ok || patternStr == "" {
		return nil, fmt.Errorf("regex grader: pattern is required")
	}

	re, err := regexp.Compile(patternStr)
	if err != nil {
		return nil, fmt.Errorf("regex grader: invalid pattern %q: %w", patternStr, err)
	}

	return &regexGrader{pattern: re}, nil
}

func (g *regexGrader) Type() string { return "regex" }

func (g *regexGrader) Grade(_ context.Context, input GradeInput) (*model.GradeResult, error) {
	match := g.pattern.MatchString(input.AgentOutput.Text)

	score := 0.0
	if match {
		score = 1.0
	}

	return &model.GradeResult{
		GraderType: g.Type(),
		Score:      score,
		Pass:       match,
		Reason:     fmt.Sprintf("pattern=%q, match=%v", g.pattern.String(), match),
	}, nil
}
