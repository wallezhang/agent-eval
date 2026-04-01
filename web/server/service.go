// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"io"
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
	"gopkg.in/yaml.v3"
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

// sanitizeFilename validates that the filename does not escape the project directory.
// It allows relative paths like "tasks/sample.yaml" but rejects ".." traversal.
func sanitizeFilename(filename string) error {
	cleaned := filepath.Clean(filename)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		return fmt.Errorf("invalid filename: %q", filename)
	}
	return nil
}

// FileNode represents a file or directory in the project file tree.
type FileNode struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"`               // "file" or "dir"
	Path     string     `json:"path"`               // relative to project root
	Children *[]FileNode `json:"children,omitempty"` // non-nil for dirs (even if empty), nil for files
}

// ListFileTree returns the recursive directory/file tree for a project.
// Excludes the output dir (from eval.yaml), hidden dirs, and node_modules.
// Only includes .yaml/.yml files. Directories first, then files, alphabetical.
func (s *Service) ListFileTree(projectName string) ([]FileNode, error) {
	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return nil, err
	}

	excludeDirs := s.resolveExcludeDirs(projectDir)
	return s.buildFileTree(projectDir, projectDir, excludeDirs)
}

// resolveExcludeDirs returns the set of directory names to exclude from the file tree.
func (s *Service) resolveExcludeDirs(projectDir string) map[string]bool {
	exclude := map[string]bool{
		"node_modules": true,
	}
	// Read output.dir from eval config
	for _, name := range []string{"eval.yaml", "eval.yml"} {
		configPath := filepath.Join(projectDir, name)
		suite, err := config.Load(configPath)
		if err == nil && suite.Output.Dir != "" {
			dir := filepath.Clean(suite.Output.Dir)
			// Handle relative paths like "./results"
			dir = strings.TrimPrefix(dir, "./")
			// Only exclude top-level dir name
			parts := strings.SplitN(dir, string(filepath.Separator), 2)
			if len(parts) > 0 {
				exclude[parts[0]] = true
			}
			break
		}
	}
	// Default: exclude "results" if no config found
	if len(exclude) == 1 {
		exclude["results"] = true
	}
	return exclude
}

// buildFileTree recursively builds the file tree for a directory.
func (s *Service) buildFileTree(root, dir string, excludeDirs map[string]bool) ([]FileNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dirs, files []FileNode
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden entries
		if strings.HasPrefix(name, ".") {
			continue
		}
		rel, _ := filepath.Rel(root, filepath.Join(dir, name))

		if entry.IsDir() {
			if excludeDirs[name] {
				continue
			}
			children, err := s.buildFileTree(root, filepath.Join(dir, name), excludeDirs)
			if err != nil {
				continue
			}
			if children == nil {
				children = []FileNode{}
			}
			dirs = append(dirs, FileNode{
				Name:     name,
				Type:     "dir",
				Path:     rel,
				Children: &children,
			})
		} else {
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".yaml" || ext == ".yml" {
				files = append(files, FileNode{
					Name: name,
					Type: "file",
					Path: rel,
				})
			}
		}
	}

	// Directories first, then files (both already alphabetical from ReadDir)
	result := make([]FileNode, 0, len(dirs)+len(files))
	result = append(result, dirs...)
	result = append(result, files...)
	return result, nil
}

// CreateDir creates a directory under the project root.
func (s *Service) CreateDir(projectName, dirPath string) error {
	if err := sanitizeFilename(dirPath); err != nil {
		return err
	}

	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return err
	}

	fullPath := filepath.Join(projectDir, dirPath)
	return os.MkdirAll(fullPath, 0o755)
}

// ListConfigs returns all .yaml/.yml files in the project directory, recursively.
// Paths are returned relative to the project directory (e.g., "eval.yaml", "tasks/sample.yaml").
func (s *Service) ListConfigs(projectName string) ([]string, error) {
	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return nil, err
	}

	var configs []string
	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Skip hidden directories and common non-config directories
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "results" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext == ".yaml" || ext == ".yml" {
			rel, _ := filepath.Rel(projectDir, path)
			configs = append(configs, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking project directory: %w", err)
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
// ValidateConfig validates a config file. For main eval configs (eval.yaml/eval.yml
// in project root), it runs full config.Load() validation. For other YAML files
// (e.g., task files in subdirectories), it only checks YAML syntax.
func (s *Service) ValidateConfig(projectName, filename string) []string {
	if err := sanitizeFilename(filename); err != nil {
		return []string{err.Error()}
	}

	projectDir, err := s.projectPath(projectName)
	if err != nil {
		return []string{err.Error()}
	}

	fullPath := filepath.Join(projectDir, filename)

	// Main config files get full validation via config.Load()
	base := filepath.Base(filename)
	isMainConfig := (filename == base) && (base == "eval.yaml" || base == "eval.yml")
	if isMainConfig {
		_, err = config.Load(fullPath)
		if err != nil {
			return []string{err.Error()}
		}
		return nil
	}

	// Other YAML files: syntax check only
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return []string{fmt.Sprintf("reading file: %v", err)}
	}
	var yamlContent any
	if err := yaml.Unmarshal(data, &yamlContent); err != nil {
		return []string{fmt.Sprintf("YAML syntax error: %v", err)}
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

// ProjectInfo holds metadata about a project for the settings page.
type ProjectInfo struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	DBPath string `json:"db_path"`
}

// GetProjectInfo returns project metadata including the resolved DB path.
func (s *Service) GetProjectInfo(projectName string) (*ProjectInfo, error) {
	project, err := s.registry.Get(projectName)
	if err != nil {
		return nil, err
	}
	dbPath := s.resolveDBPath(project.Path)
	return &ProjectInfo{
		Name:   project.Name,
		Path:   project.Path,
		DBPath: dbPath,
	}, nil
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

		// Calculate total trials
		totalTrials := 0
		for _, task := range suite.Tasks {
			n := task.TrialsPerTask
			if n <= 0 {
				n = 1
			}
			totalTrials += n
		}

		rm.SendEvent(runID, SSEEvent{Type: "run_started", Data: map[string]any{
			"run_id":       runID,
			"suite":        suite.Name,
			"total_tasks":  len(suite.Tasks),
			"total_trials": totalTrials,
		}})

		// Create a logger that bridges engine log output to SSE events
		sseWriter := &sseLogWriter{runID: runID, rm: rm}
		// Tee to both the service logger's output and the SSE writer
		engineLogger := log.New(io.MultiWriter(s.logger.Writer(), sseWriter), s.logger.Prefix(), 0)

		eng := engine.New(suite, engineLogger)
		evalRun, err := eng.Execute(activeRun.Ctx)
		if err != nil {
			rm.SendEvent(runID, SSEEvent{Type: "run_error", Data: map[string]any{
				"run_id":  runID,
				"message": err.Error(),
			}})
			return
		}

		// Send progress summary
		rm.SendEvent(runID, SSEEvent{Type: "run_progress", Data: map[string]any{
			"completed":   evalRun.Summary.TotalTrials,
			"total":       evalRun.Summary.TotalTrials,
			"pass_count":  evalRun.Summary.PassedTrials,
			"fail_count":  evalRun.Summary.FailedTrials,
			"error_count": evalRun.Summary.ErrorTrials,
		}})

		evalRun.ID = runID
		if saveErr := store.SaveRun(context.Background(), evalRun); saveErr != nil {
			s.logger.Printf("failed to save run %s: %v", runID, saveErr)
		}

		rm.SendEvent(runID, SSEEvent{Type: "run_completed", Data: map[string]any{
			"run_id":  runID,
			"summary": evalRun.Summary,
		}})
	}()

	return runID, nil
}

// getStore returns a lazily-initialized SQLite store for the given project.
// The DB path is resolved from the project's main eval config (output.dir/agent-eval.db),
// matching the CLI behavior in cmd/run.go.
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

	dbPath := s.resolveDBPath(projectDir)
	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening store for project %q: %w", projectName, err)
	}

	s.stores[projectName] = store
	return store, nil
}

// resolveDBPath determines the SQLite database path for a project.
// It tries to load the main eval config to read output.dir, then uses
// <output_dir>/agent-eval.db (matching CLI behavior). Falls back to
// <project_dir>/results/agent-eval.db if no config is found.
func (s *Service) resolveDBPath(projectDir string) string {
	// Try common config filenames
	for _, name := range []string{"eval.yaml", "eval.yml"} {
		configPath := filepath.Join(projectDir, name)
		suite, err := config.Load(configPath)
		if err == nil && suite.Output.Dir != "" {
			outputDir := suite.Output.Dir
			if !filepath.IsAbs(outputDir) {
				outputDir = filepath.Join(projectDir, outputDir)
			}
			return filepath.Join(outputDir, "agent-eval.db")
		}
	}
	// Fallback: default output dir
	return filepath.Join(projectDir, "results", "agent-eval.db")
}
