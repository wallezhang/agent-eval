// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"fmt"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// Agent represents a system under test that can execute tasks.
type Agent interface {
	// Execute sends the task input to the agent and returns its output.
	Execute(ctx context.Context, input model.TaskInput) (*model.AgentOutput, error)
	// Close releases any resources held by the agent.
	Close() error
}

// Factory creates an Agent from a configuration map.
type Factory func(config map[string]any) (Agent, error)

var registry = make(map[string]Factory)

// Register adds an agent factory to the registry.
func Register(typeName string, factory Factory) {
	registry[typeName] = factory
}

// Create instantiates an agent of the given type with the given config.
func Create(typeName string, config map[string]any) (Agent, error) {
	factory, ok := registry[typeName]
	if !ok {
		return nil, fmt.Errorf("unknown agent type: %q (registered: %v)", typeName, registeredTypes())
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
