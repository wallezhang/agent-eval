// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wallezhang/agent-eval/web"
)

// Server is the web UI backend server.
type Server struct {
	router     chi.Router
	logger     *log.Logger
	homePath   string
	service    *Service
	runManager *RunManager
}

// New creates a new Server with the given home directory.
func New(homePath string, logger *log.Logger) (*Server, error) {
	if logger == nil {
		logger = log.New(os.Stderr, "[server] ", log.LstdFlags)
	}

	svc, err := NewService(homePath)
	if err != nil {
		return nil, fmt.Errorf("creating service: %w", err)
	}

	s := &Server{
		router:     chi.NewRouter(),
		logger:     logger,
		homePath:   homePath,
		service:    svc,
		runManager: NewRunManager(),
	}

	s.buildRouter()
	return s, nil
}

func (s *Server) buildRouter() {
	r := s.router

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Health
	r.Get("/api/health", s.handleHealth)

	// Projects
	r.Get("/api/projects", s.handleListProjects)
	r.Post("/api/projects", s.handleAddProject)

	// Per-project routes
	r.Route("/api/projects/{name}", func(r chi.Router) {
		r.Delete("/", s.handleDeleteProject)
		r.Get("/configs", s.handleListConfigs)
		r.Post("/configs", s.handleCreateConfig)
		r.Get("/configs/{filename}", s.handleGetConfig)
		r.Put("/configs/{filename}", s.handleUpdateConfig)
		r.Delete("/configs/{filename}", s.handleDeleteConfig)
		r.Post("/configs/{filename}/validate", s.handleValidateConfig)

		r.Post("/runs", s.handleStartRun)
		r.Get("/runs", s.handleListRuns)
		r.Get("/runs/active", s.handleListActiveRuns)
		r.Get("/runs/{id}", s.handleGetRun)
		r.Delete("/runs/{id}", s.handleDeleteRun)
		r.Post("/runs/{id}/cancel", s.handleCancelRun)
		r.Get("/runs/{id}/sse", s.handleSSE)
	})

	// Metadata
	r.Get("/api/agents", s.handleListAgentTypes)
	r.Get("/api/graders", s.handleListGraderTypes)

	// Serve embedded frontend (SPA fallback)
	frontendFS, err := fs.Sub(web.FrontendFS, "frontend/dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(frontendFS))
		r.NotFound(func(w http.ResponseWriter, req *http.Request) {
			// Try to serve static file first
			path := req.URL.Path
			if len(path) > 0 && path[0] == '/' {
				path = path[1:]
			}
			_, err := fs.Stat(frontendFS, path)
			if err == nil {
				fileServer.ServeHTTP(w, req)
				return
			}
			// SPA fallback: serve index.html for client-side routing
			req.URL.Path = "/"
			fileServer.ServeHTTP(w, req)
		})
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
