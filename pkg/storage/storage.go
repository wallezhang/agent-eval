// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// Store provides persistence for evaluation runs.
type Store interface {
	// SaveRun persists a completed evaluation run.
	SaveRun(ctx context.Context, run *model.EvalRun) error
	// GetRun retrieves a run by ID.
	GetRun(ctx context.Context, id string) (*model.EvalRun, error)
	// ListRuns returns all stored runs, ordered by start time descending.
	ListRuns(ctx context.Context) ([]model.EvalRun, error)
	// DeleteRun removes a run by ID.
	DeleteRun(ctx context.Context, id string) error
	// Close releases any resources.
	Close() error
}
