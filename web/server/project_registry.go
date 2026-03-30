// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Project represents a registered project with its name and filesystem path.
type Project struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// ProjectRegistry manages a list of projects persisted to a JSON file.
type ProjectRegistry struct {
	mu       sync.RWMutex
	filePath string
	projects []Project
}

// NewProjectRegistry loads or creates a project registry from the given JSON file path.
// If the file does not exist, an empty registry is created. Parent directories are
// created if needed.
func NewProjectRegistry(filePath string) (*ProjectRegistry, error) {
	r := &ProjectRegistry{
		filePath: filePath,
		projects: []Project{},
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return r, nil
		}
		return nil, fmt.Errorf("reading registry file: %w", err)
	}

	if err := json.Unmarshal(data, &r.projects); err != nil {
		return nil, fmt.Errorf("parsing registry file: %w", err)
	}

	return r, nil
}

// List returns a copy of all registered projects.
func (r *ProjectRegistry) List() []Project {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Project, len(r.projects))
	copy(out, r.projects)
	return out
}

// Get finds a project by name. Returns an error if not found.
func (r *ProjectRegistry) Get(name string) (Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.projects {
		if p.Name == name {
			return p, nil
		}
	}
	return Project{}, fmt.Errorf("project %q not found", name)
}

// Add registers a new project. The path must exist and be a directory.
// Duplicate names are rejected. The registry is persisted to disk after adding.
func (r *ProjectRegistry) Add(name, path string) error {
	if name == "" {
		return fmt.Errorf("project name must not be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path %q does not exist: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}

	for _, p := range r.projects {
		if p.Name == name {
			return fmt.Errorf("project %q already exists", name)
		}
	}

	r.projects = append(r.projects, Project{Name: name, Path: path})
	return r.save()
}

// Remove deregisters a project by name. The project's directory is not deleted.
// Returns an error if the project is not found.
func (r *ProjectRegistry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.projects {
		if p.Name == name {
			r.projects = append(r.projects[:i], r.projects[i+1:]...)
			return r.save()
		}
	}
	return fmt.Errorf("project %q not found", name)
}

// save persists the current project list to disk. Must be called with mu held.
func (r *ProjectRegistry) save() error {
	if err := os.MkdirAll(filepath.Dir(r.filePath), 0o755); err != nil {
		return fmt.Errorf("creating registry directory: %w", err)
	}

	data, err := json.MarshalIndent(r.projects, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling registry: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0o644); err != nil {
		return fmt.Errorf("writing registry file: %w", err)
	}

	return nil
}
