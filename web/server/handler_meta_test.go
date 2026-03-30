// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestHandlerListAgents(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var types []string
	if err := json.NewDecoder(w.Body).Decode(&types); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if len(types) == 0 {
		t.Fatal("expected non-empty agent types list")
	}

	if !slices.Contains(types, "openai") {
		t.Errorf("expected agent types to contain 'openai', got %v", types)
	}
}

func TestHandlerListGraders(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/graders", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var types []string
	if err := json.NewDecoder(w.Body).Decode(&types); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if len(types) == 0 {
		t.Fatal("expected non-empty grader types list")
	}

	if !slices.Contains(types, "exact_match") {
		t.Errorf("expected grader types to contain 'exact_match', got %v", types)
	}
}
