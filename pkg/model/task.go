// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

// Task represents a single evaluation task.
type Task struct {
	ID       string            `json:"id" yaml:"id"`
	Name     string            `json:"name" yaml:"name"`
	Tags     []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Input    TaskInput         `json:"input" yaml:"input"`
	Expected *ExpectedOutput   `json:"expected,omitempty" yaml:"expected,omitempty"`
	Graders  []GraderRef       `json:"graders,omitempty" yaml:"graders,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// TrialsPerTask overrides the suite-level default if set.
	TrialsPerTask int `json:"trials_per_task,omitempty" yaml:"trials_per_task,omitempty"`
	// StepLimit is the maximum expected number of steps for efficiency scoring.
	StepLimit int `json:"step_limit,omitempty" yaml:"step_limit,omitempty"`
}

// TaskInput defines the input to be sent to the agent.
type TaskInput struct {
	Prompt   string            `json:"prompt" yaml:"prompt"`
	System   string            `json:"system,omitempty" yaml:"system,omitempty"`
	Messages []Message         `json:"messages,omitempty" yaml:"messages,omitempty"`
	Params   map[string]string `json:"params,omitempty" yaml:"params,omitempty"`
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role" yaml:"role"`
	Content string `json:"content" yaml:"content"`
}

// ExpectedOutput defines the expected result for grading.
type ExpectedOutput struct {
	Text   string            `json:"text,omitempty" yaml:"text,omitempty"`
	JSON   map[string]any    `json:"json,omitempty" yaml:"json,omitempty"`
	Fields map[string]string `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// GraderRef references a grader configuration for a task.
type GraderRef struct {
	Type   string         `json:"type" yaml:"type"`
	Weight float64        `json:"weight,omitempty" yaml:"weight,omitempty"`
	Config map[string]any `json:"config,omitempty" yaml:"config,omitempty"`
}
