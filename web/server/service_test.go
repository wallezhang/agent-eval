// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const testEvalYAML = `name: test-suite
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

// setupTestProject creates a temp dir with a valid eval.yaml, creates a Service,
// and registers the project. Returns the service and a cleanup function.
func setupTestProject(t *testing.T) (*Service, string) {
	t.Helper()

	homeDir := t.TempDir()
	projectDir := t.TempDir()

	// Write a valid eval.yaml in the project directory.
	err := os.WriteFile(filepath.Join(projectDir, "eval.yaml"), []byte(testEvalYAML), 0o644)
	if err != nil {
		t.Fatalf("writing eval.yaml: %v", err)
	}

	svc, err := NewService(homeDir)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	// Register the project.
	err = svc.Registry().Add("test-project", projectDir)
	if err != nil {
		t.Fatalf("registering project: %v", err)
	}

	return svc, projectDir
}

func TestService_ListConfigs(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	configs, err := svc.ListConfigs("test-project")
	if err != nil {
		t.Fatalf("ListConfigs: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0] != "eval.yaml" {
		t.Errorf("expected eval.yaml, got %q", configs[0])
	}
}

func TestService_GetConfig(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	content, err := svc.GetConfig("test-project", "eval.yaml")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("expected non-empty content")
	}
}

func TestService_GetConfigNotFound(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	_, err := svc.GetConfig("test-project", "nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent config")
	}
}

func TestService_SaveConfig(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	content := []byte("name: new-suite\n")
	err := svc.SaveConfig("test-project", "new.yaml", content)
	if err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got, err := svc.GetConfig("test-project", "new.yaml")
	if err != nil {
		t.Fatalf("GetConfig after save: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestService_DeleteConfig(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	// Save a config first.
	err := svc.SaveConfig("test-project", "to-delete.yaml", []byte("name: delete-me\n"))
	if err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	err = svc.DeleteConfig("test-project", "to-delete.yaml")
	if err != nil {
		t.Fatalf("DeleteConfig: %v", err)
	}

	_, err = svc.GetConfig("test-project", "to-delete.yaml")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestService_ValidateConfig(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	errs := svc.ValidateConfig("test-project", "eval.yaml")
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got: %v", errs)
	}
}

func TestService_ValidateConfigInvalid(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	// eval.yaml with missing required fields → full validation fails
	err := svc.SaveConfig("test-project", "eval.yaml", []byte("name: invalid\n"))
	if err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	errs := svc.ValidateConfig("test-project", "eval.yaml")
	if len(errs) == 0 {
		t.Error("expected validation errors for invalid eval.yaml")
	}

	// Non-eval file with bad YAML syntax → syntax check fails
	err = svc.SaveConfig("test-project", "bad-syntax.yaml", []byte("  bad:\n\t- mixed indent"))
	if err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	errs = svc.ValidateConfig("test-project", "bad-syntax.yaml")
	if len(errs) == 0 {
		t.Error("expected YAML syntax errors for bad-syntax.yaml")
	}

	// Non-eval file with valid YAML → syntax check passes
	err = svc.SaveConfig("test-project", "tasks.yaml", []byte("- id: test\n  name: test\n"))
	if err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	errs = svc.ValidateConfig("test-project", "tasks.yaml")
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid task YAML, got: %v", errs)
	}
}

func TestService_ListRuns(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	ctx := context.Background()
	runs, err := svc.ListRuns(ctx, "test-project")
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}

	if len(runs) != 0 {
		t.Errorf("expected 0 runs for new project, got %d", len(runs))
	}
}

func TestService_GetRunNotFound(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	ctx := context.Background()
	_, err := svc.GetRun(ctx, "test-project", "nonexistent-run-id")
	if err == nil {
		t.Fatal("expected error for nonexistent run")
	}
}

func TestService_PathTraversal(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	attacks := []string{"../../etc/passwd", "../secret.yaml", "/etc/shadow", "sub/dir/file.yaml"}
	for _, attack := range attacks {
		_, err := svc.GetConfig("test-project", attack)
		if err == nil {
			t.Errorf("GetConfig(%q) should have been rejected", attack)
		}
		err = svc.SaveConfig("test-project", attack, []byte("x"))
		if err == nil {
			t.Errorf("SaveConfig(%q) should have been rejected", attack)
		}
		err = svc.DeleteConfig("test-project", attack)
		if err == nil {
			t.Errorf("DeleteConfig(%q) should have been rejected", attack)
		}
	}
}

func TestService_ProjectNotFound(t *testing.T) {
	svc, _ := setupTestProject(t)
	defer svc.Close()

	_, err := svc.ListConfigs("nonexistent-project")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}

	_, err = svc.GetConfig("nonexistent-project", "eval.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}

	ctx := context.Background()
	_, err = svc.ListRuns(ctx, "nonexistent-project")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
}
