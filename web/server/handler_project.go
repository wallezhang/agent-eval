// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type addProjectRequest struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects := s.service.Registry().List()
	writeJSON(w, http.StatusOK, projects)
}

func (s *Server) handleAddProject(w http.ResponseWriter, r *http.Request) {
	var req addProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Path = strings.TrimSpace(req.Path)

	if req.Name == "" || req.Path == "" {
		writeError(w, http.StatusBadRequest, "name and path are required")
		return
	}

	if err := s.service.Registry().Add(req.Name, req.Path); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	project := Project{Name: req.Name, Path: req.Path}
	writeJSON(w, http.StatusCreated, project)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if err := s.service.Registry().Remove(name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("WARNING: writeJSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
