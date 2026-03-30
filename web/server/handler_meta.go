// Copyright 2026 wallezhang. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"

	"github.com/wallezhang/agent-eval/pkg/agent"
	"github.com/wallezhang/agent-eval/pkg/grader"
)

func (s *Server) handleListAgentTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, agent.Types())
}

func (s *Server) handleListGraderTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, grader.Types())
}
