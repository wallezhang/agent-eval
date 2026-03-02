// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"

// SetVersion sets the CLI version string.
func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:   "agent-eval",
	Short: "A universal agent evaluation framework",
	Long: `agent-eval is a CLI tool for evaluating AI agents.

It supports multiple agent types (HTTP, OpenAI, Anthropic, CLI commands),
various grading strategies (exact match, contains, regex, JSON, LLM-based),
and produces reports in table, JSON, and HTML formats.

Configuration is driven by YAML files. Results are stored in SQLite.`,
	Version: version,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
}
