// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wallezhang/agent-eval/pkg/model"
	"github.com/wallezhang/agent-eval/pkg/report"
	"github.com/wallezhang/agent-eval/pkg/storage"
)

var compareCmd = &cobra.Command{
	Use:   "compare <runA> <runB>",
	Short: "Compare two evaluation runs",
	Long:  "Display a side-by-side comparison of two evaluation runs.",
	Args:  cobra.ExactArgs(2),
	RunE:  compareRuns,
}

func init() {
	compareCmd.Flags().String("db", "", "Path to SQLite database (default: ./results/agent-eval.db)")
	rootCmd.AddCommand(compareCmd)
}

func compareRuns(cmd *cobra.Command, args []string) error {
	dbPath, _ := cmd.Flags().GetString("db")
	if dbPath == "" {
		dbPath = filepath.Join("results", "agent-eval.db")
	}

	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Support prefix matching for run IDs.
	runA, err := findRun(ctx, store, args[0])
	if err != nil {
		return fmt.Errorf("finding run A: %w", err)
	}

	runB, err := findRun(ctx, store, args[1])
	if err != nil {
		return fmt.Errorf("finding run B: %w", err)
	}

	result := report.CompareRuns(runA, runB)
	return report.FormatCompareText(result, cmd.OutOrStdout())
}

func findRun(ctx context.Context, store *storage.SQLiteStore, idPrefix string) (*model.EvalRun, error) {
	// Try exact match first.
	run, err := store.GetRun(ctx, idPrefix)
	if err == nil {
		return run, nil
	}

	// Try prefix match.
	runs, err := store.ListRuns(ctx)
	if err != nil {
		return nil, err
	}

	for i := range runs {
		if strings.HasPrefix(runs[i].ID, idPrefix) {
			return &runs[i], nil
		}
	}

	return nil, fmt.Errorf("no run found matching %q", idPrefix)
}
