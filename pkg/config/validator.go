// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// Validate checks the EvalSuite for configuration errors.
func Validate(suite *model.EvalSuite) error {
	if suite.Name == "" {
		return fmt.Errorf("suite name is required")
	}
	if suite.Agent.Type == "" {
		return fmt.Errorf("agent type is required")
	}
	if len(suite.Tasks) == 0 {
		return fmt.Errorf("at least one task is required")
	}

	taskIDs := make(map[string]bool)
	for i, task := range suite.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task at index %d: id is required", i)
		}
		if taskIDs[task.ID] {
			return fmt.Errorf("task at index %d: duplicate id %q", i, task.ID)
		}
		taskIDs[task.ID] = true

		if task.Input.Prompt == "" && len(task.Input.Messages) == 0 {
			return fmt.Errorf("task %q: input prompt or messages is required", task.ID)
		}

		if len(task.Graders) == 0 {
			return fmt.Errorf("task %q: at least one grader is required (set defaults or per-task)", task.ID)
		}

		for j, g := range task.Graders {
			if g.Type == "" {
				return fmt.Errorf("task %q: grader at index %d: type is required", task.ID, j)
			}
			if g.Weight < 0 {
				return fmt.Errorf("task %q: grader at index %d: weight must be non-negative", task.ID, j)
			}
		}
	}

	if suite.Execution.Concurrency < 0 {
		return fmt.Errorf("execution concurrency must be non-negative")
	}
	if suite.Execution.RateLimitRPS < 0 {
		return fmt.Errorf("execution rate_limit_rps must be non-negative")
	}

	return nil
}
