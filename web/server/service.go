// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/wallezhang/agent-eval/pkg/config"
	"github.com/wallezhang/agent-eval/pkg/engine"
	"github.com/wallezhang/agent-eval/pkg/model"
	"github.com/wallezhang/agent-eval/pkg/storage"
)

// Service bridges HTTP handlers to pkg/* packages. It holds the ProjectRegistry
// and provides methods for loading configs, listing runs, etc.
type Service struct {
	registry *ProjectRegistry
	logger   *log.Logger
	mu       sync.Mutex
	stores   map[string]*storage.SQLiteStore // project name → store (lazy init)
}

// NewService creates a new Service with a project registry at homePath/projects.json.
func NewService(homePath string) (*Service, error) {
	registryPath := filepath.Join(homePath, "projects.json")
	registry, err := NewProjectRegistry(registryPath)
	if err != nil {
		return nil, fmt.Errorf("creating project registry: %w", err)
	}

	return &Service{
		registry: registry,
		logger:   log.New(os.Stderr, "[service] ", log.LstdFlags),
		stores:   make(map[string]*storage.SQLiteStore),
	}, nil
}

// Registry returns the project registry.
func (s *Service) Registry() *ProjectRegistry {
	return s.registry
}

// sanitizeFilename validates that the filename is a base name only (no path separators, no "..").
func sanitizeFilename(filename string) error {
	if filename != filepath.Base(filename) {
		return fmt.Errorf("invalid filename: %q", filename)
	}
	return nil
}

// ListConfigs returns all .yaml/.yml files in the project directory.
func (s *Service) ListConfigs(projectName string) ([]string, error) {
	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, fmt.Errorf("reading project directory: %w", err)
	}

	var configs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".yaml" || ext == ".yml" {
			configs = append(configs, entry.Name())
		}
	}
	return configs, nil
}

// GetConfig reads and returns the content of a config file in the project directory.
func (s *Service) GetConfig(projectName, filename string) ([]byte, error) {
	if err := sanitizeFilename(filename); err != nil {
		return nil, err
	}

	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(projectDir, filename))
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", filename, err)
	}
	return data, nil
}

// SaveConfig writes content to a config file in the project directory.
func (s *Service) SaveConfig(projectName, filename string, content []byte) error {
	if err := sanitizeFilename(filename); err != nil {
		return err
	}

	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(projectDir, filename), content, 0o644); err != nil {
		return fmt.Errorf("writing config file %q: %w", filename, err)
	}
	return nil
}

// DeleteConfig deletes a config file from the project directory.
func (s *Service) DeleteConfig(projectName, filename string) error {
	if err := sanitizeFilename(filename); err != nil {
		return err
	}

	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(projectDir, filename)); err != nil {
		return fmt.Errorf("deleting config file %q: %w", filename, err)
	}
	return nil
}

// ValidateConfig loads a config file via config.Load() and returns any validation errors.
// An empty slice means the config is valid.
func (s *Service) ValidateConfig(projectName, filename string) []string {
	if err := sanitizeFilename(filename); err != nil {
		return []string{err.Error()}
	}

	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return []string{err.Error()}
	}

	_, err = config.Load(filepath.Join(projectDir, filename))
	if err != nil {
		return []string{err.Error()}
	}
	return nil
}

// ListRuns returns all evaluation runs for a project from SQLite storage.
func (s *Service) ListRuns(ctx context.Context, projectName string) ([]model.EvalRun, error) {
	store, err := s.getStore(projectName)
	if err != nil {
		return nil, err
	}

	return store.ListRuns(ctx)
}

// GetRun returns a single evaluation run by ID.
func (s *Service) GetRun(ctx context.Context, projectName, runID string) (*model.EvalRun, error) {
	store, err := s.getStore(projectName)
	if err != nil {
		return nil, err
	}

	return store.GetRun(ctx, runID)
}

// DeleteRun deletes an evaluation run by ID.
func (s *Service) DeleteRun(ctx context.Context, projectName, runID string) error {
	store, err := s.getStore(projectName)
	if err != nil {
		return err
	}

	return store.DeleteRun(ctx, runID)
}

// Close closes all open SQLite stores.
func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, store := range s.stores {
		if err := store.Close(); err != nil {
			s.logger.Printf("closing store for project %q: %v", name, err)
		}
	}
	s.stores = make(map[string]*storage.SQLiteStore)
}

// projectPath resolves the filesystem path for a project from the registry.
func (s *Service) projectPath(projectName string) (string, error) {
	project, err := s.registry.Get(projectName)
	if err != nil {
		return "", err
	}
	return project.Path, nil
}

// StartRunRequest holds the parameters for starting a new evaluation run.
type StartRunRequest struct {
	ConfigFile string `json:"config_file"`
}

// StartRun loads the config, creates an engine, and launches an evaluation run
// in a background goroutine. It returns the run ID immediately.
func (s *Service) StartRun(ctx context.Context, projectName string, configFile string, rm *RunManager) (string, error) {
	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(projectDir, configFile)
	suite, err := config.Load(configPath)
	if err != nil {
		return "", fmt.Errorf("loading config %q: %w", configFile, err)
	}

	store, err := s.getStore(projectName)
	if err != nil {
		return "", fmt.Errorf("getting store: %w", err)
	}

	runID := uuid.New().String()
	activeRun := rm.Start(runID, projectName)

	go func() {
		defer rm.Finish(runID)

		rm.SendEvent(runID, SSEEvent{Type: "run_started", Data: map[string]string{"run_id": runID}})

		eng := engine.New(suite, s.logger)
		evalRun, err := eng.Execute(activeRun.Ctx)
		if err != nil {
			rm.SendEvent(runID, SSEEvent{Type: "run_error", Data: map[string]string{
				"run_id": runID,
				"error":  err.Error(),
			}})
			return
		}

		evalRun.ID = runID
		if saveErr := store.SaveRun(context.Background(), evalRun); saveErr != nil {
			s.logger.Printf("failed to save run %s: %v", runID, saveErr)
		}

		rm.SendEvent(runID, SSEEvent{Type: "run_completed", Data: map[string]string{"run_id": runID}})
	}()

	return runID, nil
}

// getStore returns a lazily-initialized SQLite store for the given project.
func (s *Service) getStore(projectName string) (*storage.SQLiteStore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if store, ok := s.stores[projectName]; ok {
		return store, nil
	}

	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(projectDir, ".agent-eval", "results.db")
	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening store for project %q: %w", projectName, err)
	}

	s.stores[projectName] = store
	return store, nil
}
