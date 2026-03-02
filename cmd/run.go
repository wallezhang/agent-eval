// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wallezhang/agent-eval/pkg/config"
	"github.com/wallezhang/agent-eval/pkg/engine"
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
	rootCmd.AddCommand(runCmd)
}

func runEval(cmd *cobra.Command, _ []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	dbPath, _ := cmd.Flags().GetString("db")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Load configuration.
	suite, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
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

	return nil
}
