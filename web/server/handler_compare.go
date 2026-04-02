// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wallezhang/agent-eval/pkg/report"
)

func (s *Server) handleCompareRuns(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "name")

	runAID := strings.TrimSpace(r.URL.Query().Get("runA"))
	runBID := strings.TrimSpace(r.URL.Query().Get("runB"))

	if runAID == "" || runBID == "" {
		writeError(w, http.StatusBadRequest, "both runA and runB query parameters are required")
		return
	}

	runA, err := s.service.GetRun(r.Context(), projectName, runAID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "run A not found: "+runAID)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	runB, err := s.service.GetRun(r.Context(), projectName, runBID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "run B not found: "+runBID)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := report.CompareRuns(runA, runB)
	writeJSON(w, http.StatusOK, result)
}
