// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/wallezhang/agent-eval/pkg/model"
)

// MemoryStore is an in-memory implementation of Store, useful for testing.
type MemoryStore struct {
	mu   sync.RWMutex
	runs map[string]*model.EvalRun
}

// NewMemory creates a new in-memory store.
func NewMemory() *MemoryStore {
	return &MemoryStore{
		runs: make(map[string]*model.EvalRun),
	}
}

func (s *MemoryStore) SaveRun(_ context.Context, run *model.EvalRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.runs[run.ID] = run
	return nil
}

func (s *MemoryStore) GetRun(_ context.Context, id string) (*model.EvalRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	run, ok := s.runs[id]
	if !ok {
		return nil, fmt.Errorf("run %q not found", id)
	}
	return run, nil
}

func (s *MemoryStore) ListRuns(_ context.Context) ([]model.EvalRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runs := make([]model.EvalRun, 0, len(s.runs))
	for _, run := range s.runs {
		runs = append(runs, *run)
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].StartedAt.After(runs[j].StartedAt)
	})

	return runs, nil
}

func (s *MemoryStore) DeleteRun(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.runs[id]; !ok {
		return fmt.Errorf("run %q not found", id)
	}
	delete(s.runs, id)
	return nil
}

func (s *MemoryStore) Close() error {
	return nil
}
