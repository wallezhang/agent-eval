// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func init() {
	Register("constraint", func(config map[string]any) (Grader, error) {
		return newConstraintGrader(config)
	})
}

type constraintGrader struct {
	checks []constraintCheck
}

type constraintCheck struct {
	name         string
	pattern      *regexp.Regexp
	mustMatch    bool
	mustNotMatch bool
	maxWords     int
	minWords     int
}

func newConstraintGrader(config map[string]any) (*constraintGrader, error) {
	checksRaw, ok := config["checks"].([]any)
	if !ok || len(checksRaw) == 0 {
		return nil, fmt.Errorf("constraint grader: checks list is required")
	}

	var checks []constraintCheck
	for i, raw := range checksRaw {
		checkMap, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("constraint grader: check at index %d must be a map", i)
		}

		check := constraintCheck{}

		if name, ok := checkMap["name"].(string); ok {
			check.name = name
		} else {
			check.name = fmt.Sprintf("check_%d", i)
		}

		if pattern, ok := checkMap["pattern"].(string); ok && pattern != "" {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("constraint grader: check %q: invalid pattern %q: %w", check.name, pattern, err)
			}
			check.pattern = re

			if v, ok := checkMap["must_match"].(bool); ok {
				check.mustMatch = v
			}
			if v, ok := checkMap["must_not_match"].(bool); ok {
				check.mustNotMatch = v
			}

			// Default: if neither specified and pattern is set, treat as must_match.
			if !check.mustMatch && !check.mustNotMatch {
				check.mustMatch = true
			}
		}

		if v, ok := toIntValue(checkMap["max_words"]); ok {
			check.maxWords = v
		}
		if v, ok := toIntValue(checkMap["min_words"]); ok {
			check.minWords = v
		}

		checks = append(checks, check)
	}

	return &constraintGrader{checks: checks}, nil
}

func (g *constraintGrader) Type() string { return "constraint" }

func (g *constraintGrader) Grade(_ context.Context, input GradeInput) (*model.GradeResult, error) {
	text := input.AgentOutput.Text
	wordCount := len(strings.Fields(text))

	var violations []string
	passed := 0

	for _, check := range g.checks {
		if check.pattern != nil {
			matches := check.pattern.MatchString(text)
			if check.mustMatch && !matches {
				violations = append(violations, fmt.Sprintf("%s: pattern %q must match but didn't", check.name, check.pattern.String()))
				continue
			}
			if check.mustNotMatch && matches {
				violations = append(violations, fmt.Sprintf("%s: pattern %q must not match but did", check.name, check.pattern.String()))
				continue
			}
		}

		if check.maxWords > 0 && wordCount > check.maxWords {
			violations = append(violations, fmt.Sprintf("%s: word count %d exceeds max %d", check.name, wordCount, check.maxWords))
			continue
		}

		if check.minWords > 0 && wordCount < check.minWords {
			violations = append(violations, fmt.Sprintf("%s: word count %d below min %d", check.name, wordCount, check.minWords))
			continue
		}

		passed++
	}

	total := len(g.checks)
	score := float64(passed) / float64(total)
	pass := len(violations) == 0

	reason := fmt.Sprintf("%d/%d checks passed", passed, total)
	if len(violations) > 0 {
		reason += "; violations: " + strings.Join(violations, "; ")
	}

	return &model.GradeResult{
		GraderType: g.Type(),
		Score:      score,
		Pass:       pass,
		Reason:     reason,
	}, nil
}

func toIntValue(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
