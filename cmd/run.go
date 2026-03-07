// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wallezhang/agent-eval/pkg/config"
	"github.com/wallezhang/agent-eval/pkg/engine"
	"github.com/wallezhang/agent-eval/pkg/model"
	"github.com/wallezhang/agent-eval/pkg/report"
	"github.com/wallezhang/agent-eval/pkg/storage"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an evaluation suite",
	Long:  "Execute all tasks in an evaluation suite and generate reports.",
	RunE:  runEval,
}

func init() {
	runCmd.Flags().StringP("config", "c", "eval.yaml", "Path to the evaluation config file")
	runCmd.Flags().String("db", "", "Path to SQLite database (default: <output_dir>/agent-eval.db)")
	runCmd.Flags().Float64("fail-under", 0, "Minimum pass rate threshold (0-1). Exit with code 1 if below. Default 0 (no gating)")
	runCmd.Flags().StringSlice("tags", nil, "Only run tasks with these tags (comma-separated, OR logic)")
	runCmd.Flags().StringSlice("exclude-tags", nil, "Exclude tasks with these tags (comma-separated, takes precedence over --tags)")
	runCmd.Flags().Bool("no-cache", false, "Disable response cache even if configured")
	runCmd.Flags().String("resume", "", "Resume a previously interrupted run by run ID")
	rootCmd.AddCommand(runCmd)
}

func runEval(cmd *cobra.Command, _ []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	dbPath, _ := cmd.Flags().GetString("db")
	verbose, _ := cmd.Flags().GetBool("verbose")
	failUnder, _ := cmd.Flags().GetFloat64("fail-under")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	excludeTags, _ := cmd.Flags().GetStringSlice("exclude-tags")
	noCache, _ := cmd.Flags().GetBool("no-cache")
	resumeRunID, _ := cmd.Flags().GetString("resume")

	// Load configuration.
	suite, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Disable cache if --no-cache flag is set.
	if noCache {
		suite.Cache.Enabled = false
	}

	// Set up logger.
	logger := log.New(os.Stderr, "[agent-eval] ", log.LstdFlags)
	if !verbose {
		logger = log.New(os.Stderr, "", 0)
	}

	// Set up context with cancellation on interrupt.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Initialize storage.
	if dbPath == "" {
		dbPath = filepath.Join(suite.Output.Dir, "agent-eval.db")
	}
	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}
	defer store.Close()

	// Run the evaluation.
	eng := engine.New(suite, logger)

	// Apply tag filters.
	cleanTags := cleanStringSlice(tags)
	cleanExcludeTags := cleanStringSlice(excludeTags)
	if len(cleanTags) > 0 || len(cleanExcludeTags) > 0 {
		eng.SetTagFilters(cleanTags, cleanExcludeTags)
	}

	// Set up checkpoint/resume support.
	eng.SetCheckpointStore(store, resumeRunID)

	run, err := eng.Execute(ctx)
	if err != nil {
		return fmt.Errorf("evaluation failed: %w", err)
	}

	// Save results.
	if err := store.SaveRun(ctx, run); err != nil {
		logger.Printf("Warning: failed to save run: %v", err)
	}

	// Generate reports.
	if err := report.GenerateAll(run, suite.Output.Format, suite.Output.Dir); err != nil {
		return fmt.Errorf("generating reports: %w", err)
	}

	// Write machine-readable summary.json.
	if err := writeSummaryJSON(run, suite.Output.Dir); err != nil {
		logger.Printf("Warning: failed to write summary.json: %v", err)
	}

	// Check fail-under threshold.
	if failUnder > 0 && run.Summary.OverallPassRate < failUnder {
		return fmt.Errorf("pass rate %.1f%% is below threshold %.1f%%",
			run.Summary.OverallPassRate*100, failUnder*100)
	}

	return nil
}

// writeSummaryJSON writes a machine-readable summary file.
func writeSummaryJSON(run *model.EvalRun, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(run.Summary, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(outputDir, "summary.json"), data, 0o644)
}

// cleanStringSlice trims whitespace from each element and removes empty strings.
func cleanStringSlice(ss []string) []string {
	var result []string
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
