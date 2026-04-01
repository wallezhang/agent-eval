// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Fatalf("expected status %d, got %d: %s", expected, w.Code, w.Body.String())
	}
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestIntegration_FullFlow(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := t.TempDir()

	srv, err := New(homeDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	// 1. List projects — should be empty
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	var projects []Project
	if err := json.NewDecoder(w.Body).Decode(&projects); err != nil {
		t.Fatalf("decoding projects: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("expected empty project list, got %d", len(projects))
	}

	// 2. Add a project
	body := fmt.Sprintf(`{"name":"integration","path":%s}`, jsonEscape(projectDir))
	req = httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusCreated)

	var createdProject Project
	if err := json.NewDecoder(w.Body).Decode(&createdProject); err != nil {
		t.Fatalf("decoding created project: %v", err)
	}
	if createdProject.Name != "integration" {
		t.Fatalf("expected project name 'integration', got %q", createdProject.Name)
	}

	// 3. Create a config
	configContent := `name: integration-test
agent:
  type: openai
  config:
    api_key: ${OPENAI_API_KEY}
    model: gpt-4o
tasks:
  - id: hello
    name: Hello World
    input:
      prompt: "Say hello"
    expected:
      text: "Hello"
    graders:
      - type: exact_match
`
	createConfigBody := fmt.Sprintf(`{"filename":"test.yaml","content":%s}`, jsonEscape(configContent))
	req = httptest.NewRequest(http.MethodPost, "/api/projects/integration/configs", strings.NewReader(createConfigBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusCreated)

	// 4. List configs — should contain test.yaml
	req = httptest.NewRequest(http.MethodGet, "/api/projects/integration/configs", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	var configs []string
	if err := json.NewDecoder(w.Body).Decode(&configs); err != nil {
		t.Fatalf("decoding configs: %v", err)
	}
	if len(configs) != 1 || configs[0] != "test.yaml" {
		t.Fatalf("expected [test.yaml], got %v", configs)
	}

	// 5. Get config content
	req = httptest.NewRequest(http.MethodGet, "/api/projects/integration/configs/test.yaml", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	gotContent := w.Body.String()
	if gotContent != configContent {
		t.Fatalf("config content mismatch:\ngot:  %q\nwant: %q", gotContent, configContent)
	}

	// 6. Validate config
	validateBody := `{"filename":"test.yaml"}`
	req = httptest.NewRequest(http.MethodPost, "/api/projects/integration/configs/validate", strings.NewReader(validateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	var validation map[string]any
	if err := json.NewDecoder(w.Body).Decode(&validation); err != nil {
		t.Fatalf("decoding validation: %v", err)
	}
	if valid, ok := validation["valid"].(bool); !ok || !valid {
		t.Fatalf("expected valid=true, got %v", validation)
	}

	// 7. List runs — should be empty
	req = httptest.NewRequest(http.MethodGet, "/api/projects/integration/runs", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	var runs []any
	if err := json.NewDecoder(w.Body).Decode(&runs); err != nil {
		t.Fatalf("decoding runs: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected empty runs, got %d", len(runs))
	}

	// 8. List agents
	req = httptest.NewRequest(http.MethodGet, "/api/agents", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	var agents []string
	if err := json.NewDecoder(w.Body).Decode(&agents); err != nil {
		t.Fatalf("decoding agents: %v", err)
	}
	if len(agents) == 0 {
		t.Fatal("expected non-empty agent types")
	}

	// 9. List graders
	req = httptest.NewRequest(http.MethodGet, "/api/graders", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	var graders []string
	if err := json.NewDecoder(w.Body).Decode(&graders); err != nil {
		t.Fatalf("decoding graders: %v", err)
	}
	if len(graders) == 0 {
		t.Fatal("expected non-empty grader types")
	}

	// 10. Health check
	req = httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusOK)

	var health map[string]string
	if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
		t.Fatalf("decoding health: %v", err)
	}
	if health["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", health["status"])
	}

	// 11. Delete project
	req = httptest.NewRequest(http.MethodDelete, "/api/projects/integration", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, http.StatusNoContent)

	// 12. Verify project directory still exists on disk
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Fatal("project directory should still exist after registry removal")
	}

	// Verify config file still exists on disk
	configPath := filepath.Join(projectDir, "test.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file should still exist on disk after project removal")
	}
}
