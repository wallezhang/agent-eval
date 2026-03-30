// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
)

type createConfigRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type updateConfigRequest struct {
	Content string `json:"content"`
}

func (s *Server) handleListConfigs(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")

	configs, err := s.service.ListConfigs(projectName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return empty array instead of null
	if configs == nil {
		configs = []string{}
	}
	writeJSON(w, http.StatusOK, configs)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")
	filename := chi.URLParam(r, "filename")

	data, err := s.service.GetConfig(projectName, filename)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			writeError(w, http.StatusNotFound, "config file not found")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleCreateConfig(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")

	var req createConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Filename = strings.TrimSpace(req.Filename)
	if req.Filename == "" {
		writeError(w, http.StatusBadRequest, "filename is required")
		return
	}

	if err := s.service.SaveConfig(projectName, req.Filename, []byte(req.Content)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")
	filename := chi.URLParam(r, "filename")

	var req updateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.service.SaveConfig(projectName, filename, []byte(req.Content)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")
	filename := chi.URLParam(r, "filename")

	if err := s.service.DeleteConfig(projectName, filename); err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			writeError(w, http.StatusNotFound, "config file not found")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleValidateConfig(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")
	filename := chi.URLParam(r, "filename")

	errors := s.service.ValidateConfig(projectName, filename)

	valid := len(errors) == 0
	result := map[string]any{
		"valid":  valid,
		"errors": errors,
	}
	if errors == nil {
		result["errors"] = []string{}
	}

	writeJSON(w, http.StatusOK, result)
}
