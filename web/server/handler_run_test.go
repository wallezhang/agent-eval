// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerListRuns_Empty(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/runs", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var runs []any
	if err := json.NewDecoder(w.Body).Decode(&runs); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected empty array, got %v", runs)
	}
}

func TestHandlerListActiveRuns_Empty(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/runs/active", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var runs []any
	if err := json.NewDecoder(w.Body).Decode(&runs); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected empty array, got %v", runs)
	}
}

func TestHandlerStartRun_InvalidConfig(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	body := `{"config_file":"nonexistent.yaml"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/proj/runs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerDeleteRun_NotFound(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/proj/runs/nonexistent-id", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// DeleteRun on a nonexistent ID: SQLite DELETE is idempotent, so it may return 204.
	// But since we don't have any runs, store might return an error or just succeed.
	// Accept either 204 (idempotent delete) or 404.
	if w.Code != http.StatusNoContent && w.Code != http.StatusNotFound {
		t.Fatalf("expected 204 or 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCancelRun_NotFound(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/proj/runs/nonexistent-id/cancel", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerGetRun_NotFound(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/runs/nonexistent-id", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
