// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"sync"
	"testing"
	"time"
)

func TestRunManager_StartAndGet(t *testing.T) {
	rm := NewRunManager()

	run := rm.Start("run-1", "project-a")

	if run.ID != "run-1" {
		t.Errorf("expected ID run-1, got %s", run.ID)
	}
	if run.Project != "project-a" {
		t.Errorf("expected Project project-a, got %s", run.Project)
	}
	if run.Ctx == nil {
		t.Error("expected non-nil context")
	}
	if run.Cancel == nil {
		t.Error("expected non-nil cancel func")
	}
	if run.EventChan == nil {
		t.Error("expected non-nil event channel")
	}
	if cap(run.EventChan) != 100 {
		t.Errorf("expected channel capacity 100, got %d", cap(run.EventChan))
	}
	if run.StartedAt.IsZero() {
		t.Error("expected non-zero StartedAt")
	}

	got, ok := rm.Get("run-1")
	if !ok {
		t.Fatal("expected ok=true for existing run")
	}
	if got != run {
		t.Error("expected same run pointer")
	}
}

func TestRunManager_GetNotFound(t *testing.T) {
	rm := NewRunManager()

	_, ok := rm.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for nonexistent run")
	}
}

func TestRunManager_ListActive(t *testing.T) {
	rm := NewRunManager()

	rm.Start("run-1", "project-a")
	rm.Start("run-2", "project-b")
	rm.Start("run-3", "project-a")

	// Filter by project
	listA := rm.ListActive("project-a")
	if len(listA) != 2 {
		t.Fatalf("expected 2 runs for project-a, got %d", len(listA))
	}

	listB := rm.ListActive("project-b")
	if len(listB) != 1 {
		t.Fatalf("expected 1 run for project-b, got %d", len(listB))
	}

	// Empty string returns all
	all := rm.ListActive("")
	if len(all) != 3 {
		t.Fatalf("expected 3 total runs, got %d", len(all))
	}
}

func TestRunManager_Cancel(t *testing.T) {
	rm := NewRunManager()

	run := rm.Start("run-1", "project-a")
	rm.Cancel("run-1")

	select {
	case <-run.Ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Error("expected context to be done after cancel")
	}
}

func TestRunManager_Finish(t *testing.T) {
	rm := NewRunManager()

	rm.Start("run-1", "project-a")
	rm.Finish("run-1")

	_, ok := rm.Get("run-1")
	if ok {
		t.Error("expected run to be removed after finish")
	}
}

func TestRunManager_SendEvent(t *testing.T) {
	rm := NewRunManager()

	run := rm.Start("run-1", "project-a")

	evt := SSEEvent{Type: "progress", Data: map[string]string{"status": "running"}}
	rm.SendEvent("run-1", evt)

	select {
	case received := <-run.EventChan:
		if received.Type != "progress" {
			t.Errorf("expected event type progress, got %s", received.Type)
		}
	case <-time.After(time.Second):
		t.Error("expected to receive event on channel")
	}
}

func TestRunManager_SendEventToNonexistent(t *testing.T) {
	rm := NewRunManager()

	// Should not panic
	rm.SendEvent("nonexistent", SSEEvent{Type: "test", Data: nil})
}

func TestRunManager_ConcurrentSendAndFinish(t *testing.T) {
	rm := NewRunManager()
	rm.Start("run-race", "proj")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			rm.SendEvent("run-race", SSEEvent{Type: "tick", Data: i})
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		rm.Finish("run-race")
	}()
	wg.Wait()
}

func TestRunManager_FinishIdempotent(t *testing.T) {
	rm := NewRunManager()
	rm.Start("run-double", "proj")
	rm.Finish("run-double")
	rm.Finish("run-double") // should not panic
}

func TestRunManager_FinishCancelsContext(t *testing.T) {
	rm := NewRunManager()
	run := rm.Start("run-ctx", "proj")
	rm.Finish("run-ctx")
	select {
	case <-run.Ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Error("context should be canceled after Finish")
	}
}
