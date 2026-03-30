// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	home := t.TempDir()
	s, err := New(home, nil)
	if err != nil {
		t.Fatalf("creating test server: %v", err)
	}
	return s
}

func TestHandlerListProjects_Empty(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var projects []Project
	if err := json.NewDecoder(w.Body).Decode(&projects); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("expected empty list, got %d projects", len(projects))
	}
}

func TestHandlerAddProject(t *testing.T) {
	s := newTestServer(t)

	// Create a temp dir for the project path
	projectDir := t.TempDir()

	body := `{"name":"myproject","path":"` + strings.ReplaceAll(projectDir, `\`, `\\`) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var project Project
	if err := json.NewDecoder(w.Body).Decode(&project); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if project.Name != "myproject" {
		t.Fatalf("expected name 'myproject', got %q", project.Name)
	}

	// Verify it appears in the list
	req = httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	var projects []Project
	if err := json.NewDecoder(w.Body).Decode(&projects); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "myproject" {
		t.Fatalf("expected 'myproject', got %q", projects[0].Name)
	}
}

func TestHandlerAddProjectDuplicate(t *testing.T) {
	s := newTestServer(t)
	projectDir := t.TempDir()

	body := `{"name":"dup","path":"` + strings.ReplaceAll(projectDir, `\`, `\\`) + `"}`

	// First add — should succeed
	req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first add: expected 201, got %d", w.Code)
	}

	// Second add — should conflict
	req = httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate add: expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerDeleteProject(t *testing.T) {
	s := newTestServer(t)
	projectDir := t.TempDir()

	// Add a project first
	body := `{"name":"todelete","path":"` + strings.ReplaceAll(projectDir, `\`, `\\`) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("add: expected 201, got %d", w.Code)
	}

	// Delete it
	req = httptest.NewRequest(http.MethodDelete, "/api/projects/todelete", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's gone
	req = httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	var projects []Project
	if err := json.NewDecoder(w.Body).Decode(&projects); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("expected empty list after delete, got %d", len(projects))
	}
}

func TestHandlerDeleteProjectNotFound(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/nonexistent", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerAddProjectBadRequest(t *testing.T) {
	s := newTestServer(t)

	// Empty name
	req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"","path":"/tmp"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty name: expected 400, got %d", w.Code)
	}

	// Empty path
	req = httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"test","path":""}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty path: expected 400, got %d", w.Code)
	}

	// Non-existent path
	req = httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"test","path":"/nonexistent/path/`+os.TempDir()+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad path: expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
