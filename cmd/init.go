// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize a new evaluation project",
	Long:  "Create a scaffold for a new agent evaluation project with example config and task files.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  initProject,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

const defaultEvalYAML = `name: "my-eval"
description: "Agent evaluation suite"

agent:
  type: openai
  config:
    model: gpt-4
    api_key: ${OPENAI_API_KEY}
    base_url: https://api.openai.com/v1
    temperature: 0.0

defaults:
  trials_per_task: 3
  graders:
    - type: exact_match
      config:
        ignore_case: true

execution:
  concurrency: 4
  rate_limit_rps: 10
  timeout: 60s

task_files:
  - tasks/*.yaml

output:
  format: all
  dir: ./results
`

const defaultTaskYAML = `- id: sample-task
  name: "Sample Task"
  tags: [sample]
  input:
    prompt: "What is 2 + 2? Answer with just the number."
  expected:
    text: "4"
  graders:
    - type: exact_match
      config:
        ignore_case: true
        ignore_whitespace: true
`

func initProject(_ *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	// Create directories.
	dirs := []string{
		filepath.Join(dir, "tasks"),
		filepath.Join(dir, "results"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	// Write default files.
	files := map[string]string{
		filepath.Join(dir, "eval.yaml"):            defaultEvalYAML,
		filepath.Join(dir, "tasks", "sample.yaml"): defaultTaskYAML,
	}

	for path, content := range files {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Skipping %s (already exists)\n", path)
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		fmt.Printf("Created %s\n", path)
	}

	fmt.Println("\nProject initialized. Edit eval.yaml and tasks/ to configure your evaluation.")
	fmt.Println("Run with: agent-eval run -c eval.yaml")
	return nil
}
