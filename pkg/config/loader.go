// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wallezhang/agent-eval/pkg/model"
	"gopkg.in/yaml.v3"
)

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// Load reads a YAML config file and returns an EvalSuite.
func Load(path string) (*model.EvalSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Expand environment variables in the form ${VAR_NAME}.
	expanded := expandEnvVars(string(data))

	var suite model.EvalSuite
	if err := yaml.Unmarshal([]byte(expanded), &suite); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

	// Load external task files.
	baseDir := filepath.Dir(path)
	if err := loadTaskFiles(&suite, baseDir); err != nil {
		return nil, fmt.Errorf("loading task files: %w", err)
	}

	// Apply defaults.
	applyDefaults(&suite)

	// Validate.
	if err := Validate(&suite); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &suite, nil
}

// expandEnvVars replaces ${VAR_NAME} with the corresponding environment variable value.
func expandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		return match // Keep the original if not set.
	})
}

// loadTaskFiles reads external task definition files matching glob patterns.
func loadTaskFiles(suite *model.EvalSuite, baseDir string) error {
	for _, pattern := range suite.TaskFiles {
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(baseDir, pattern)
		}

		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("glob pattern %q: %w", pattern, err)
		}

		for _, match := range matches {
			tasks, err := loadTaskFile(match)
			if err != nil {
				return fmt.Errorf("loading task file %q: %w", match, err)
			}
			suite.Tasks = append(suite.Tasks, tasks...)
		}
	}
	return nil
}

// loadTaskFile reads a single task definition YAML file.
// The file can contain a single task or a list of tasks.
func loadTaskFile(path string) ([]model.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	expanded := expandEnvVars(string(data))

	// Try to parse as a list of tasks first.
	var tasks []model.Task
	if err := yaml.Unmarshal([]byte(expanded), &tasks); err == nil && len(tasks) > 0 {
		return tasks, nil
	}

	// Try as a single task.
	var task model.Task
	if err := yaml.Unmarshal([]byte(expanded), &task); err != nil {
		return nil, fmt.Errorf("parsing task YAML: %w", err)
	}
	return []model.Task{task}, nil
}

// applyDefaults applies suite-level defaults to tasks that don't override them.
func applyDefaults(suite *model.EvalSuite) {
	if suite.Defaults.TrialsPerTask == 0 {
		suite.Defaults.TrialsPerTask = 1
	}
	if suite.Defaults.PassThreshold == 0 {
		suite.Defaults.PassThreshold = 0.5
	}
	if suite.Execution.Concurrency == 0 {
		suite.Execution.Concurrency = 1
	}
	if suite.Execution.Timeout == "" {
		suite.Execution.Timeout = "60s"
	}
	if suite.Output.Format == "" {
		suite.Output.Format = "table"
	}
	if suite.Output.Dir == "" {
		suite.Output.Dir = "./results"
	}

	for i := range suite.Tasks {
		if suite.Tasks[i].TrialsPerTask == 0 {
			suite.Tasks[i].TrialsPerTask = suite.Defaults.TrialsPerTask
		}
		if len(suite.Tasks[i].Graders) == 0 {
			suite.Tasks[i].Graders = suite.Defaults.Graders
		}
		// Ensure grader weights default to 1.0.
		for j := range suite.Tasks[i].Graders {
			if suite.Tasks[i].Graders[j].Weight == 0 {
				suite.Tasks[i].Graders[j].Weight = 1.0
			}
		}
	}
}
