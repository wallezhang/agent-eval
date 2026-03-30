// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectRegistry_EmptyOnFirstLoad(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	projects := reg.List()
	if len(projects) != 0 {
		t.Errorf("expected empty list, got %d projects", len(projects))
	}
}

func TestProjectRegistry_AddAndList(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")
	projectDir := t.TempDir()

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	if err := reg.Add("myproject", projectDir); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	projects := reg.List()
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "myproject" {
		t.Errorf("expected name 'myproject', got %q", projects[0].Name)
	}
	if projects[0].Path != projectDir {
		t.Errorf("expected path %q, got %q", projectDir, projects[0].Path)
	}
}

func TestProjectRegistry_AddDuplicateName(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")
	projectDir := t.TempDir()

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	if err := reg.Add("myproject", projectDir); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err = reg.Add("myproject", projectDir)
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
}

func TestProjectRegistry_AddNonexistentPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	err = reg.Add("myproject", "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

func TestProjectRegistry_AddEmptyName(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")
	projectDir := t.TempDir()

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	err = reg.Add("", projectDir)
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestProjectRegistry_AddFilePath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")

	// Create a regular file to use as the project path
	tmpFile := filepath.Join(dir, "not-a-dir.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	err = reg.Add("myproject", tmpFile)
	if err == nil {
		t.Fatal("expected error for file path (not directory), got nil")
	}
}

func TestProjectRegistry_Remove(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")
	projectDir := t.TempDir()

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	if err := reg.Add("myproject", projectDir); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if err := reg.Remove("myproject"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	projects := reg.List()
	if len(projects) != 0 {
		t.Errorf("expected empty list after remove, got %d projects", len(projects))
	}

	// Verify the directory still exists (deregister only, no deletion)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Error("expected project directory to still exist after Remove")
	}
}

func TestProjectRegistry_RemoveNotFound(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	err = reg.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error for removing nonexistent project, got nil")
	}
}

func TestProjectRegistry_Get(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "projects.json")
	projectDir := t.TempDir()

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	if err := reg.Add("myproject", projectDir); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	p, err := reg.Get("myproject")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if p.Name != "myproject" {
		t.Errorf("expected name 'myproject', got %q", p.Name)
	}
	if p.Path != projectDir {
		t.Errorf("expected path %q, got %q", projectDir, p.Path)
	}

	// Error for missing
	_, err = reg.Get("missing")
	if err == nil {
		t.Fatal("expected error for missing project, got nil")
	}
}

func TestProjectRegistry_PersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "subdir", "projects.json")
	projectDir := t.TempDir()

	reg, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() error = %v", err)
	}

	if err := reg.Add("myproject", projectDir); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Create a new registry from the same file
	reg2, err := NewProjectRegistry(filePath)
	if err != nil {
		t.Fatalf("NewProjectRegistry() second load error = %v", err)
	}

	projects := reg2.List()
	if len(projects) != 1 {
		t.Fatalf("expected 1 project after reload, got %d", len(projects))
	}
	if projects[0].Name != "myproject" {
		t.Errorf("expected name 'myproject', got %q", projects[0].Name)
	}
	if projects[0].Path != projectDir {
		t.Errorf("expected path %q, got %q", projectDir, projects[0].Path)
	}
}
