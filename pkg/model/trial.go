// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import "time"

// TrialStatus represents the status of a single trial.
type TrialStatus string

const (
	TrialStatusPending TrialStatus = "pending"
	TrialStatusRunning TrialStatus = "running"
	TrialStatusPassed  TrialStatus = "passed"
	TrialStatusFailed  TrialStatus = "failed"
	TrialStatusError   TrialStatus = "error"
)

// Trial represents a single execution of a task.
type Trial struct {
	ID              string        `json:"id"`
	TaskID          string        `json:"task_id"`
	Index           int           `json:"index"`
	Status          TrialStatus   `json:"status"`
	AgentOutput     *AgentOutput  `json:"agent_output,omitempty"`
	Transcript      *Transcript   `json:"transcript,omitempty"`
	Grades          []GradeResult `json:"grades,omitempty"`
	Score           float64       `json:"score"`
	Pass            bool          `json:"pass"`
	Error           string        `json:"error,omitempty"`
	StartedAt       time.Time     `json:"started_at"`
	FinishedAt      time.Time     `json:"finished_at"`
	DurationMS      int64         `json:"duration_ms"`
	AgentDurationMS int64         `json:"agent_duration_ms,omitempty"`
	StepCount       int           `json:"step_count,omitempty"`
}

// AgentOutput represents the output from an agent execution.
type AgentOutput struct {
	Text     string         `json:"text"`
	Messages []Message      `json:"messages,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Transcript records the full interaction between the harness and the agent.
type Transcript struct {
	Steps []TranscriptStep `json:"steps"`
}

// TranscriptStep represents a single step in the transcript.
type TranscriptStep struct {
	Type      string    `json:"type"`
	Role      string    `json:"role,omitempty"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// GradeResult holds the output from a single grader.
type GradeResult struct {
	GraderType string  `json:"grader_type"`
	Score      float64 `json:"score"`
	Pass       bool    `json:"pass"`
	Weight     float64 `json:"weight"`
	Reason     string  `json:"reason,omitempty"`
	Error      string  `json:"error,omitempty"`
}
