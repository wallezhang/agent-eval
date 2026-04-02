// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wallezhang/agent-eval/pkg/model"
	"github.com/wallezhang/agent-eval/pkg/report"
)

func saveTestRun(t *testing.T, s *Server, ctx context.Context, projectName, runID, suite, agent string) error {
	t.Helper()
	store, err := s.service.getStore(projectName)
	if err != nil {
		return err
	}
	run := &model.EvalRun{
		ID:         runID,
		SuiteName:  suite,
		AgentType:  agent,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
		Summary: model.EvalSummary{
			OverallPassRate: 0.75,
			AvgScore:        0.70,
		},
	}
	return store.SaveRun(ctx, run)
}

func TestHandlerCompareRuns_MissingParams(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/compare", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCompareRuns_MissingRunB(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/compare?runA=abc", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCompareRuns_RunNotFound(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/compare?runA=nonexistent-a&runB=nonexistent-b", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCompareRuns_ValidResponse(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	ctx := context.Background()
	if err := saveTestRun(t, s, ctx, "proj", "run-a-id", "suite1", "openai"); err != nil {
		t.Fatalf("saving run A: %v", err)
	}
	if err := saveTestRun(t, s, ctx, "proj", "run-b-id", "suite1", "anthropic"); err != nil {
		t.Fatalf("saving run B: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/compare?runA=run-a-id&runB=run-b-id", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result report.CompareResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if result.RunA.ID != "run-a-id" {
		t.Errorf("RunA.ID = %q, want run-a-id", result.RunA.ID)
	}
	if result.RunB.ID != "run-b-id" {
		t.Errorf("RunB.ID = %q, want run-b-id", result.RunB.ID)
	}
}
