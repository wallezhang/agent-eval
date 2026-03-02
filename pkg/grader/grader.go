// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package grader

import (
	"context"
	"fmt"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// Grader evaluates an agent's output against expected results.
type Grader interface {
	// Grade evaluates the agent output and returns a grade result.
	Grade(ctx context.Context, input GradeInput) (*model.GradeResult, error)
	// Type returns the grader type name.
	Type() string
}

// GradeInput provides all context needed for grading.
type GradeInput struct {
	Task        model.Task
	AgentOutput model.AgentOutput
	Transcript  *model.Transcript
}

// Factory creates a Grader from a configuration map.
type Factory func(config map[string]any) (Grader, error)

var registry = make(map[string]Factory)

// Register adds a grader factory to the registry.
func Register(typeName string, factory Factory) {
	registry[typeName] = factory
}

// Create instantiates a grader of the given type with the given config.
func Create(typeName string, config map[string]any) (Grader, error) {
	factory, ok := registry[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown grader type: %q (registered: %v)", typeName, registeredTypes())
	}
	return factory(config)
}

func registeredTypes() []string {
	types := make([]string, 0, len(registry))
	for t := range registry {
		types = append(types, t)
	}
	return types
}
