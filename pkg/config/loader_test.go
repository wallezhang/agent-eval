// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temp directory with config and task files.
	dir := t.TempDir()

	taskDir := filepath.Join(dir, "tasks")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a task file.
	taskContent := `- id: test-task
  name: "Test Task"
  input:
    prompt: "Hello?"
  expected:
    text: "World"
`
	if err := os.WriteFile(filepath.Join(taskDir, "test.yaml"), []byte(taskContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write the main config.
	configContent := `name: "test-suite"
description: "Test suite"
agent:
  type: http
  config:
    url: http://localhost:8080
defaults:
  trials_per_task: 2
  graders:
    - type: exact_match
      config:
        ignore_case: true
execution:
  concurrency: 2
  timeout: 30s
task_files:
  - tasks/*.yaml
output:
  format: json
  dir: ./out
`
	configPath := filepath.Join(dir, "eval.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	suite, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if suite.Name != "test-suite" {
		t.Errorf("got name=%q, want %q", suite.Name, "test-suite")
	}

	if suite.Agent.Type != "http" {
		t.Errorf("got agent type=%q, want %q", suite.Agent.Type, "http")
	}

	if len(suite.Tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(suite.Tasks))
	}

	task := suite.Tasks[0]
	if task.ID != "test-task" {
		t.Errorf("got task id=%q, want %q", task.ID, "test-task")
	}
	if task.TrialsPerTask != 2 {
		t.Errorf("got trials_per_task=%d, want 2", task.TrialsPerTask)
	}
	if len(task.Graders) != 1 {
		t.Fatalf("got %d graders, want 1", len(task.Graders))
	}
	if task.Graders[0].Type != "exact_match" {
		t.Errorf("got grader type=%q, want %q", task.Graders[0].Type, "exact_match")
	}
	if task.Graders[0].Weight != 1.0 {
		t.Errorf("got grader weight=%f, want 1.0", task.Graders[0].Weight)
	}
}

func TestExpandEnvVars(t *testing.T) {
	t.Setenv("TEST_VAR", "hello")

	result := expandEnvVars("value=${TEST_VAR} and ${MISSING_VAR}")
	expected := "value=hello and ${MISSING_VAR}"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestValidate_MissingName(t *testing.T) {
	configContent := `name: ""
agent:
  type: http
  config:
    url: http://localhost:8080
tasks:
  - id: t1
    input:
      prompt: "hi"
    graders:
      - type: exact_match
`
	dir := t.TempDir()
	path := filepath.Join(dir, "eval.yaml")
	if err := os.WriteFile(path, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestValidate_DuplicateTaskID(t *testing.T) {
	configContent := `name: "test"
agent:
  type: http
  config:
    url: http://localhost:8080
tasks:
  - id: dup
    input:
      prompt: "hi"
    graders:
      - type: exact_match
  - id: dup
    input:
      prompt: "hi"
    graders:
      - type: exact_match
`
	dir := t.TempDir()
	path := filepath.Join(dir, "eval.yaml")
	if err := os.WriteFile(path, []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate task IDs, got nil")
	}
}
