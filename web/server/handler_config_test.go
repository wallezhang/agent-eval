// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validEvalYAML = `name: test-suite
agent:
  type: openai
  config:
    api_key: test-key
    model: gpt-4o
tasks:
  - id: task1
    name: Test Task
    input:
      prompt: "Hello"
    expected:
      text: "World"
    graders:
      - type: exact_match
`

// newTestServerWithProject creates a Server with a registered project containing a valid eval.yaml.
func newTestServerWithProject(t *testing.T) (*Server, string) {
	t.Helper()

	home := t.TempDir()
	projectDir := filepath.Join(t.TempDir(), "myproject")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("creating project dir: %v", err)
	}

	// Write a valid eval.yaml
	if err := os.WriteFile(filepath.Join(projectDir, "eval.yaml"), []byte(validEvalYAML), 0o644); err != nil {
		t.Fatalf("writing eval.yaml: %v", err)
	}

	s, err := New(home, nil)
	if err != nil {
		t.Fatalf("creating server: %v", err)
	}

	// Register the project via POST /api/projects
	body := `{"name":"proj","path":"` + strings.ReplaceAll(projectDir, `\`, `\\`) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("registering project: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	return s, projectDir
}

func TestHandlerListConfigs(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/configs", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var configs []string
	if err := json.NewDecoder(w.Body).Decode(&configs); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(configs) != 1 || configs[0] != "eval.yaml" {
		t.Fatalf("expected [\"eval.yaml\"], got %v", configs)
	}
}

func TestHandlerGetConfig(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/configs/eval.yaml", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/yaml" {
		t.Fatalf("expected Content-Type text/yaml, got %q", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, "test-suite") {
		t.Fatalf("expected body to contain 'test-suite', got %q", body)
	}
}

func TestHandlerGetConfigNotFound(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/proj/configs/nonexistent.yaml", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCreateConfig(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	// Create a new config
	body := `{"filename":"new.yaml","content":"name: new-suite\n"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/proj/configs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it exists via GET
	req = httptest.NewRequest(http.MethodGet, "/api/projects/proj/configs/new.yaml", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get after create: expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "new-suite") {
		t.Fatalf("expected 'new-suite' in body, got %q", w.Body.String())
	}
}

func TestHandlerUpdateConfig(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	// Update existing eval.yaml
	body := `{"content":"name: updated-suite\n"}`
	req := httptest.NewRequest(http.MethodPut, "/api/projects/proj/configs/eval.yaml", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify content changed
	req = httptest.NewRequest(http.MethodGet, "/api/projects/proj/configs/eval.yaml", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "updated-suite") {
		t.Fatalf("expected 'updated-suite' in body, got %q", w.Body.String())
	}
}

func TestHandlerDeleteConfig(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/projects/proj/configs/eval.yaml", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerValidateConfig(t *testing.T) {
	s, _ := newTestServerWithProject(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/proj/configs/eval.yaml/validate", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if result["valid"] != true {
		t.Fatalf("expected valid=true, got %v", result)
	}
}

func TestHandlerValidateConfigInvalid(t *testing.T) {
	s, projectDir := newTestServerWithProject(t)

	// Write an invalid config
	invalidYAML := `name: bad-suite
# missing agent and tasks
`
	if err := os.WriteFile(filepath.Join(projectDir, "bad.yaml"), []byte(invalidYAML), 0o644); err != nil {
		t.Fatalf("writing bad.yaml: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/projects/proj/configs/bad.yaml/validate", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result struct {
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if result.Valid {
		t.Fatal("expected valid=false for invalid config")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error for invalid config")
	}
}
