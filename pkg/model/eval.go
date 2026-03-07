// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import "time"

// EvalSuite represents the complete evaluation configuration.
type EvalSuite struct {
	Name        string          `json:"name" yaml:"name"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	Agent       AgentConfig     `json:"agent" yaml:"agent"`
	Defaults    DefaultsConfig  `json:"defaults,omitempty" yaml:"defaults,omitempty"`
	Execution   ExecutionConfig `json:"execution,omitempty" yaml:"execution,omitempty"`
	TaskFiles   []string        `json:"task_files,omitempty" yaml:"task_files,omitempty"`
	Tasks       []Task          `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	Output      OutputConfig    `json:"output,omitempty" yaml:"output,omitempty"`
	Cache       CacheConfig     `json:"cache,omitempty" yaml:"cache,omitempty"`
	Hooks       HooksConfig     `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}

// AgentConfig configures the agent under test.
type AgentConfig struct {
	Type              string         `json:"type" yaml:"type"`
	Config            map[string]any `json:"config,omitempty" yaml:"config,omitempty"`
	CostPerInputToken float64        `json:"cost_per_input_token,omitempty" yaml:"cost_per_input_token,omitempty"`
	CostPerOutputToken float64       `json:"cost_per_output_token,omitempty" yaml:"cost_per_output_token,omitempty"`
}

// DefaultsConfig provides default settings applied to all tasks.
type DefaultsConfig struct {
	TrialsPerTask int         `json:"trials_per_task,omitempty" yaml:"trials_per_task,omitempty"`
	Graders       []GraderRef `json:"graders,omitempty" yaml:"graders,omitempty"`
	PassThreshold float64     `json:"pass_threshold,omitempty" yaml:"pass_threshold,omitempty"`
}

// ExecutionConfig controls how trials are executed.
type ExecutionConfig struct {
	Concurrency  int    `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	RateLimitRPS int    `json:"rate_limit_rps,omitempty" yaml:"rate_limit_rps,omitempty"`
	Timeout      string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	MaxRetries   int    `json:"max_retries,omitempty" yaml:"max_retries,omitempty"`
	RetryDelay   string `json:"retry_delay,omitempty" yaml:"retry_delay,omitempty"`
}

// OutputConfig controls report generation.
type OutputConfig struct {
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
	Dir    string `json:"dir,omitempty" yaml:"dir,omitempty"`
}

// CacheConfig controls response caching.
type CacheConfig struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Dir     string `json:"dir,omitempty" yaml:"dir,omitempty"`
	TTL     string `json:"ttl,omitempty" yaml:"ttl,omitempty"`
}

// HooksConfig defines lifecycle hook commands.
type HooksConfig struct {
	BeforeRun   string `json:"before_run,omitempty" yaml:"before_run,omitempty"`
	AfterRun    string `json:"after_run,omitempty" yaml:"after_run,omitempty"`
	BeforeTask  string `json:"before_task,omitempty" yaml:"before_task,omitempty"`
	AfterTask   string `json:"after_task,omitempty" yaml:"after_task,omitempty"`
	BeforeTrial string `json:"before_trial,omitempty" yaml:"before_trial,omitempty"`
	AfterTrial  string `json:"after_trial,omitempty" yaml:"after_trial,omitempty"`
}

// EvalRun represents a completed evaluation run.
type EvalRun struct {
	ID          string         `json:"id"`
	SuiteName   string         `json:"suite_name"`
	Description string         `json:"description,omitempty"`
	AgentType   string         `json:"agent_type"`
	AgentConfig map[string]any `json:"agent_config,omitempty"`
	TaskResults []TaskResult   `json:"task_results"`
	Summary     EvalSummary    `json:"summary"`
	StartedAt   time.Time      `json:"started_at"`
	FinishedAt  time.Time      `json:"finished_at"`
	DurationMS  int64          `json:"duration_ms"`
}

// TaskResult aggregates trials for a single task.
type TaskResult struct {
	Task         Task    `json:"task"`
	Trials       []Trial `json:"trials"`
	PassCount    int     `json:"pass_count"`
	FailCount    int     `json:"fail_count"`
	ErrorCount   int     `json:"error_count"`
	AvgScore     float64 `json:"avg_score"`
	PassAtK      float64 `json:"pass_at_k"`
	PassPowerK   float64 `json:"pass_power_k"`
	LatencyP50MS int64   `json:"latency_p50_ms,omitempty"`
	LatencyP90MS int64   `json:"latency_p90_ms,omitempty"`
	LatencyP99MS int64   `json:"latency_p99_ms,omitempty"`
	AvgStepCount float64 `json:"avg_step_count,omitempty"`
	Usage        *UsageSummary `json:"usage,omitempty"`
}

// UsageSummary aggregates token usage and estimated cost.
type UsageSummary struct {
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	EstimatedCostUSD  float64 `json:"estimated_cost_usd,omitempty"`
}

// EvalSummary provides aggregate metrics for the entire run.
type EvalSummary struct {
	TotalTasks      int     `json:"total_tasks"`
	TotalTrials     int     `json:"total_trials"`
	PassedTrials    int     `json:"passed_trials"`
	FailedTrials    int     `json:"failed_trials"`
	ErrorTrials     int     `json:"error_trials"`
	OverallPassRate float64 `json:"overall_pass_rate"`
	AvgScore        float64 `json:"avg_score"`
	AvgPassAtK      float64 `json:"avg_pass_at_k"`
	AvgPassPowerK   float64 `json:"avg_pass_power_k"`
	Usage           *UsageSummary `json:"usage,omitempty"`
}
