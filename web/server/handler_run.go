// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wallezhang/agent-eval/pkg/model"
)

func (s *Server) handleStartRun(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")

	var req StartRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.ConfigFile = strings.TrimSpace(req.ConfigFile)
	if req.ConfigFile == "" {
		writeError(w, http.StatusBadRequest, "config_file is required")
		return
	}

	runID, err := s.service.StartRun(r.Context(), projectName, req.ConfigFile, s.runManager)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"run_id": runID,
		"status": "started",
	})
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")

	runs, err := s.service.ListRuns(r.Context(), projectName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if runs == nil {
		runs = []model.EvalRun{}
	}
	writeJSON(w, http.StatusOK, runs)
}

func (s *Server) handleListActiveRuns(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")

	activeRuns := s.runManager.ListActive(projectName)

	type activeRunResponse struct {
		ID        string `json:"id"`
		Project   string `json:"project"`
		StartedAt string `json:"started_at"`
	}

	result := make([]activeRunResponse, 0, len(activeRuns))
	for _, ar := range activeRuns {
		result = append(result, activeRunResponse{
			ID:        ar.ID,
			Project:   ar.Project,
			StartedAt: ar.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")
	runID := chi.URLParam(r, "id")

	run, err := s.service.GetRun(r.Context(), projectName, runID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if run == nil {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}

	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleDeleteRun(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")
	runID := chi.URLParam(r, "id")

	if err := s.service.DeleteRun(r.Context(), projectName, runID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCancelRun(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "id")

	_, ok := s.runManager.Get(runID)
	if !ok {
		writeError(w, http.StatusNotFound, "active run not found")
		return
	}

	s.runManager.Cancel(runID)
	writeJSON(w, http.StatusOK, map[string]string{
		"run_id": runID,
		"status": "cancelled",
	})
}
