// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
	_ "modernc.org/sqlite"
)

// SQLiteStore persists evaluation runs to a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite store at the given path.
func NewSQLite(dbPath string) (*SQLiteStore, error) {
	// Ensure the directory exists.
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating directory for database: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS eval_runs (
			id TEXT PRIMARY KEY,
			suite_name TEXT NOT NULL,
			description TEXT,
			agent_type TEXT NOT NULL,
			agent_config TEXT,
			task_results TEXT NOT NULL,
			summary TEXT NOT NULL,
			started_at TEXT NOT NULL,
			finished_at TEXT NOT NULL,
			duration_ms INTEGER NOT NULL
		)
	`)
	return err
}

func (s *SQLiteStore) SaveRun(ctx context.Context, run *model.EvalRun) error {
	agentConfigJSON, err := json.Marshal(run.AgentConfig)
	if err != nil {
		return fmt.Errorf("marshaling agent config: %w", err)
	}

	taskResultsJSON, err := json.Marshal(run.TaskResults)
	if err != nil {
		return fmt.Errorf("marshaling task results: %w", err)
	}

	summaryJSON, err := json.Marshal(run.Summary)
	if err != nil {
		return fmt.Errorf("marshaling summary: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO eval_runs
		(id, suite_name, description, agent_type, agent_config, task_results, summary, started_at, finished_at, duration_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		run.ID,
		run.SuiteName,
		run.Description,
		run.AgentType,
		string(agentConfigJSON),
		string(taskResultsJSON),
		string(summaryJSON),
		run.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		run.FinishedAt.Format("2006-01-02T15:04:05Z07:00"),
		run.DurationMS,
	)
	return err
}

func (s *SQLiteStore) GetRun(ctx context.Context, id string) (*model.EvalRun, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, suite_name, description, agent_type, agent_config, task_results, summary, started_at, finished_at, duration_ms
		FROM eval_runs WHERE id = ?
	`, id)

	return scanRun(row)
}

func (s *SQLiteStore) ListRuns(ctx context.Context) ([]model.EvalRun, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, suite_name, description, agent_type, agent_config, task_results, summary, started_at, finished_at, duration_ms
		FROM eval_runs ORDER BY started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []model.EvalRun
	for rows.Next() {
		run, err := scanRunFromRows(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}
	return runs, rows.Err()
}

func (s *SQLiteStore) DeleteRun(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM eval_runs WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("run %q not found", id)
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRunFields(s scanner) (*model.EvalRun, error) {
	var (
		run             model.EvalRun
		agentConfigJSON string
		taskResultsJSON string
		summaryJSON     string
		startedAtStr    string
		finishedAtStr   string
	)

	err := s.Scan(
		&run.ID,
		&run.SuiteName,
		&run.Description,
		&run.AgentType,
		&agentConfigJSON,
		&taskResultsJSON,
		&summaryJSON,
		&startedAtStr,
		&finishedAtStr,
		&run.DurationMS,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(agentConfigJSON), &run.AgentConfig); err != nil {
		return nil, fmt.Errorf("unmarshaling agent config: %w", err)
	}
	if err := json.Unmarshal([]byte(taskResultsJSON), &run.TaskResults); err != nil {
		return nil, fmt.Errorf("unmarshaling task results: %w", err)
	}
	if err := json.Unmarshal([]byte(summaryJSON), &run.Summary); err != nil {
		return nil, fmt.Errorf("unmarshaling summary: %w", err)
	}

	run.StartedAt, _ = time.Parse(time.RFC3339, startedAtStr)
	run.FinishedAt, _ = time.Parse(time.RFC3339, finishedAtStr)

	return &run, nil
}

func scanRun(row *sql.Row) (*model.EvalRun, error) {
	return scanRunFields(row)
}

func scanRunFromRows(rows *sql.Rows) (*model.EvalRun, error) {
	return scanRunFields(rows)
}
