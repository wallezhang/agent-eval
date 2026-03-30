// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"sync"
	"time"
)

// SSEEvent represents a server-sent event for real-time run updates.
type SSEEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// ActiveRun tracks a single in-progress evaluation run.
type ActiveRun struct {
	ID        string
	Project   string
	Ctx       context.Context
	Cancel    context.CancelFunc
	EventChan chan SSEEvent
	StartedAt time.Time
	mu        sync.Mutex // protects closed and EventChan close
	closed    bool
}

// RunManager tracks concurrent evaluation runs.
type RunManager struct {
	mu     sync.RWMutex
	active map[string]*ActiveRun
}

// NewRunManager creates a new RunManager.
func NewRunManager() *RunManager {
	return &RunManager{
		active: make(map[string]*ActiveRun),
	}
}

// Start creates a new active run with its own context, cancel function,
// and buffered event channel, then registers it in the active map.
func (rm *RunManager) Start(runID, project string) *ActiveRun {
	ctx, cancel := context.WithCancel(context.Background())
	run := &ActiveRun{
		ID:        runID,
		Project:   project,
		Ctx:       ctx,
		Cancel:    cancel,
		EventChan: make(chan SSEEvent, 100),
		StartedAt: time.Now(),
	}

	rm.mu.Lock()
	rm.active[runID] = run
	rm.mu.Unlock()

	return run
}

// Get looks up an active run by ID.
func (rm *RunManager) Get(runID string) (*ActiveRun, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	run, ok := rm.active[runID]
	return run, ok
}

// ListActive returns active runs filtered by project.
// An empty project string returns all active runs.
func (rm *RunManager) ListActive(project string) []*ActiveRun {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var result []*ActiveRun
	for _, run := range rm.active {
		if project == "" || run.Project == project {
			result = append(result, run)
		}
	}
	return result
}

// Cancel cancels the context of the specified run.
func (rm *RunManager) Cancel(runID string) {
	rm.mu.RLock()
	run, ok := rm.active[runID]
	rm.mu.RUnlock()

	if ok {
		run.Cancel()
	}
}

// Finish removes a run from the active map, cancels its context,
// and closes its event channel. It is safe to call multiple times.
func (rm *RunManager) Finish(runID string) {
	rm.mu.Lock()
	run, ok := rm.active[runID]
	if ok {
		delete(rm.active, runID)
	}
	rm.mu.Unlock()

	if ok {
		run.Cancel()
		run.mu.Lock()
		if !run.closed {
			run.closed = true
			close(run.EventChan)
		}
		run.mu.Unlock()
	}
}

// SendEvent performs a non-blocking send of an event to the run's event channel.
// It is a no-op if the run is not found, already finished, or the channel is full.
func (rm *RunManager) SendEvent(runID string, event SSEEvent) {
	rm.mu.RLock()
	run, ok := rm.active[runID]
	rm.mu.RUnlock()

	if !ok {
		return
	}

	run.mu.Lock()
	defer run.mu.Unlock()

	if run.closed {
		return
	}

	select {
	case run.EventChan <- event:
	default:
	}
}
