// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wallezhang/agent-eval/pkg/storage"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List evaluation runs",
	Long:  "Display all stored evaluation runs from the database.",
	RunE:  listRuns,
}

func init() {
	listCmd.Flags().String("db", "", "Path to SQLite database (default: ./results/agent-eval.db)")
	rootCmd.AddCommand(listCmd)
}

func listRuns(cmd *cobra.Command, _ []string) error {
	dbPath, _ := cmd.Flags().GetString("db")
	if dbPath == "" {
		dbPath = filepath.Join("results", "agent-eval.db")
	}

	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	runs, err := store.ListRuns(context.Background())
	if err != nil {
		return fmt.Errorf("listing runs: %w", err)
	}

	if len(runs) == 0 {
		fmt.Println("No evaluation runs found.")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSUITE\tAGENT\tTASKS\tPASS RATE\tDURATION\tDATE")
	fmt.Fprintln(tw, "--\t-----\t-----\t-----\t---------\t--------\t----")

	for _, r := range runs {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%.1f%%\t%dms\t%s\n",
			r.ID[:8],
			r.SuiteName,
			r.AgentType,
			r.Summary.TotalTasks,
			r.Summary.OverallPassRate*100,
			r.DurationMS,
			r.StartedAt.Format("2006-01-02 15:04"),
		)
	}
	tw.Flush()

	return nil
}
