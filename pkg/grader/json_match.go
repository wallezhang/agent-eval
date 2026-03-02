// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wallezhang/agent-eval/pkg/model"
)

func init() {
	Register("json_match", func(config map[string]any) (Grader, error) {
		return newJSONMatchGrader(config), nil
	})
}

type jsonMatchGrader struct {
	ignoreCase bool
}

func newJSONMatchGrader(config map[string]any) *jsonMatchGrader {
	g := &jsonMatchGrader{}
	if v, ok := config["ignore_case"].(bool); ok {
		g.ignoreCase = v
	}
	return g
}

func (g *jsonMatchGrader) Type() string { return "json_match" }

func (g *jsonMatchGrader) Grade(_ context.Context, input GradeInput) (*model.GradeResult, error) {
	if input.Task.Expected == nil || len(input.Task.Expected.Fields) == 0 {
		return &model.GradeResult{
			GraderType: g.Type(),
			Score:      0,
			Pass:       false,
			Reason:     "no expected fields defined",
		}, nil
	}

	// Parse agent output as JSON.
	var actual map[string]any
	if err := json.Unmarshal([]byte(input.AgentOutput.Text), &actual); err != nil {
		return &model.GradeResult{
			GraderType: g.Type(),
			Score:      0,
			Pass:       false,
			Reason:     fmt.Sprintf("agent output is not valid JSON: %v", err),
		}, nil
	}

	matched := 0
	total := len(input.Task.Expected.Fields)
	var mismatches []string

	for field, expectedVal := range input.Task.Expected.Fields {
		actualVal, ok := actual[field]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("missing field %q", field))
			continue
		}

		actualStr := fmt.Sprintf("%v", actualVal)
		if g.ignoreCase {
			if strings.EqualFold(actualStr, expectedVal) {
				matched++
			} else {
				mismatches = append(mismatches, fmt.Sprintf("field %q: expected=%q, got=%q", field, expectedVal, actualStr))
			}
		} else {
			if actualStr == expectedVal {
				matched++
			} else {
				mismatches = append(mismatches, fmt.Sprintf("field %q: expected=%q, got=%q", field, expectedVal, actualStr))
			}
		}
	}

	score := float64(matched) / float64(total)
	pass := matched == total

	reason := fmt.Sprintf("matched %d/%d fields", matched, total)
	if len(mismatches) > 0 {
		reason += ": " + strings.Join(mismatches, "; ")
	}

	return &model.GradeResult{
		GraderType: g.Type(),
		Score:      score,
		Pass:       pass,
		Reason:     reason,
	}, nil
}
